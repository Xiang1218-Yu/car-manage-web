package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"carmanageweb/internal/models"
)

// --- 停车记录 ---

// ListRecords 列出停车记录，可按状态过滤。active 在前，按入场时间倒序。
func (s *Store) ListRecords(status string) ([]models.RecordDetail, error) {
	q := `
SELECT r.id, r.spot_id, r.vehicle_id, r.rule_id, r.check_in_time, r.check_out_time, r.fee, r.status, r.created_at,
       s.code, COALESCE(l.name,''), v.plate, COALESCE(fr.name,'')
FROM parking_records r
JOIN parking_spots s   ON s.id = r.spot_id
LEFT JOIN parking_lots l  ON l.id = s.lot_id
LEFT JOIN vehicles v      ON v.id = r.vehicle_id
LEFT JOIN fee_rules fr    ON fr.id = r.rule_id`
	var (
		rows *sql.Rows
		err  error
	)
	if status == "active" || status == "completed" {
		q += " WHERE r.status=? ORDER BY (r.status='active') DESC, r.check_in_time DESC"
		rows, err = s.db.Query(q, status)
	} else {
		q += " ORDER BY (r.status='active') DESC, r.check_in_time DESC"
		rows, err = s.db.Query(q)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.RecordDetail
	for rows.Next() {
		var d models.RecordDetail
		var ruleID sql.NullInt64
		var checkout sql.NullString
		if err := rows.Scan(&d.ID, &d.SpotID, &d.VehicleID, &ruleID, &d.CheckInTime, &checkout,
			&d.Fee, &d.Status, &d.CreatedAt, &d.SpotCode, &d.LotName, &d.Plate, &d.RuleName); err != nil {
			return nil, err
		}
		if ruleID.Valid {
			ri := ruleID.Int64
			d.RuleID = &ri
		}
		if checkout.Valid {
			t, _ := time.Parse(time.RFC3339, checkout.String)
			d.CheckOutTime = &t
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// GetRecord 按 ID 查询停车记录。
func (s *Store) GetRecord(id int64) (*models.ParkingRecord, error) {
	var r models.ParkingRecord
	var ruleID sql.NullInt64
	var checkout sql.NullString
	err := s.db.QueryRow(`SELECT id, spot_id, vehicle_id, rule_id, check_in_time, check_out_time, fee, status, created_at
		FROM parking_records WHERE id=?`, id).
		Scan(&r.ID, &r.SpotID, &r.VehicleID, &ruleID, &r.CheckInTime, &checkout, &r.Fee, &r.Status, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if ruleID.Valid {
		ri := ruleID.Int64
		r.RuleID = &ri
	}
	if checkout.Valid {
		t, _ := time.Parse(time.RFC3339, checkout.String)
		r.CheckOutTime = &t
	}
	return &r, nil
}

// CheckIn 车辆入场：在同一事务中校验车位可用性、绑定费用规则、写入记录并占用车位。
//
//	vehicle 已通过 UpsertVehicle 落库后传入。
//	返回新建记录的 ID。
func (s *Store) CheckIn(spotID, vehicleID int64) (int64, *models.FeeRule, error) {
	var (
		recID  int64
		rule   *models.FeeRule
		retErr error
	)
	err := s.inTx(func(tx *sql.Tx) error {
		// 锁定并读取车位
		var lotID int64
		var status string
		err := tx.QueryRow(`SELECT lot_id, status FROM parking_spots WHERE id=?`, spotID).Scan(&lotID, &status)
		if err == sql.ErrNoRows {
			retErr = ErrNotFound
			return err
		}
		if err != nil {
			return err
		}
		if status == string(models.SpotOccupied) {
			retErr = errors.New("车位已被占用")
			return retErr
		}
		if status == string(models.SpotMaintenance) {
			retErr = errors.New("车位维护中，不可停车")
			return retErr
		}
		// 解析生效费用规则
		rule, err = s.resolveFeeRuleTx(tx, lotID)
		if err != nil {
			retErr = fmt.Errorf("无可用费用规则: %w", err)
			return retErr
		}
		nowT := time.Now().UTC()
		res, err := tx.Exec(`INSERT INTO parking_records
			(spot_id, vehicle_id, rule_id, check_in_time, status, created_at)
			VALUES (?,?,?,?,?,?)`,
			spotID, vehicleID, rule.ID, nowT.Format(time.RFC3339), string(models.RecordActive), nowT.Format(time.RFC3339))
		if err != nil {
			return err
		}
		recID, err = res.LastInsertId()
		if err != nil {
			return err
		}
		return setSpotStatus(tx, spotID, models.SpotOccupied)
	})
	if err != nil && !errors.Is(err, retErr) {
		return 0, nil, err
	}
	if retErr != nil {
		return 0, rule, retErr
	}
	return recID, rule, nil
}

// CheckOut 车辆出场：在同一事务中计算费用、更新记录与车位状态。
//
//	checkOut 若为零值则取当前时间。
//	返回计算出的费用与计费明细。
func (s *Store) CheckOut(recordID int64, checkOut time.Time, calc FeeCalculator) (float64, *models.FeeBreakdown, error) {
	var (
		fee  float64
		bd   *models.FeeBreakdown
		rErr error
	)
	err := s.inTx(func(tx *sql.Tx) error {
		var (
			ruleID   sql.NullInt64
			checkInS string
			outS     sql.NullString
			status   string
			spotID   int64
			rule     *models.FeeRule
		)
		err := tx.QueryRow(`SELECT rule_id, check_in_time, check_out_time, status, spot_id
			FROM parking_records WHERE id=?`, recordID).
			Scan(&ruleID, &checkInS, &outS, &status, &spotID)
		if err == sql.ErrNoRows {
			rErr = ErrNotFound
			return err
		}
		if err != nil {
			return err
		}
		if status != string(models.RecordActive) {
			rErr = errors.New("该记录已出场，不可重复结算")
			return rErr
		}
		if checkOut.IsZero() {
			checkOut = time.Now().UTC()
		}
		checkIn, _ := time.Parse(time.RFC3339, checkInS)
		if checkOut.Before(checkIn) {
			rErr = errors.New("出场时间早于入场时间")
			return rErr
		}
		// 结算时重新解析停车场当前规则，使规则调整立即生效。
		rule, err = s.resolveFeeRuleBySpotTx(tx, spotID)
		if err != nil {
			rErr = fmt.Errorf("无可用费用规则: %w", err)
			return rErr
		}
		bd = calc(checkIn, checkOut, rule)
		fee = bd.TotalFee
		_, err = tx.Exec(`UPDATE parking_records SET check_out_time=?, fee=?, status=? WHERE id=?`,
			checkOut.Format(time.RFC3339), fee, string(models.RecordCompleted), recordID)
		if err != nil {
			return err
		}
		return setSpotStatus(tx, spotID, models.SpotAvailable)
	})
	if err != nil && !errors.Is(err, rErr) {
		return 0, nil, err
	}
	if rErr != nil {
		return 0, bd, rErr
	}
	return fee, bd, nil
}

// UpdateSpotStatus 单独更新车位状态（用于维护标记）。
func (s *Store) UpdateSpotStatus(id int64, status models.SpotStatus) error {
	if !validStatus(status) {
		return ErrInvalidInput
	}
	res, err := s.db.Exec(`UPDATE parking_spots SET status=? WHERE id=?`, string(status), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteRecord 删除停车记录。若记录处于在场状态，则先释放车位再删除。
func (s *Store) DeleteRecord(id int64) error {
	rec, err := s.GetRecord(id)
	if err != nil {
		return err
	}
	return s.inTx(func(tx *sql.Tx) error {
		if rec.Status == models.RecordActive {
			if err := setSpotStatus(tx, rec.SpotID, models.SpotAvailable); err != nil {
				return err
			}
		}
		_, err := tx.Exec(`DELETE FROM parking_records WHERE id=?`, id)
		return err
	})
}
