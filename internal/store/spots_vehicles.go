package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"carmanageweb/internal/models"
)

// --- 车位 ---

// ListSpotsByLot 列出某停车场下所有车位（带占用车辆与所在记录信息）。
func (s *Store) ListSpotsByLot(lotID int64) ([]models.SpotDetail, error) {
	q := `
SELECT s.id, s.lot_id, s.code, s.status, s.created_at,
       COALESCE(l.name,''), v.plate, COALESCE(pr.id,0)
FROM parking_spots s
LEFT JOIN parking_lots l   ON l.id = s.lot_id
LEFT JOIN parking_records pr ON pr.spot_id = s.id AND pr.status='active'
LEFT JOIN vehicles v       ON v.id = pr.vehicle_id
WHERE s.lot_id = ?
ORDER BY s.code`
	rows, err := s.db.Query(q, lotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.SpotDetail
	for rows.Next() {
		var d models.SpotDetail
		var plate sql.NullString
		var recID int64
		if err := rows.Scan(&d.ID, &d.LotID, &d.Code, &d.Status, &d.CreatedAt,
			&d.LotName, &plate, &recID); err != nil {
			return nil, err
		}
		d.CurrentPlate = plate.String
		d.CurrentRecordID = recID
		out = append(out, d)
	}
	return out, rows.Err()
}

// GetSpot 按 ID 查询车位。
func (s *Store) GetSpot(id int64) (*models.ParkingSpot, error) {
	var sp models.ParkingSpot
	err := s.db.QueryRow(`SELECT id, lot_id, code, status, created_at FROM parking_spots WHERE id=?`, id).
		Scan(&sp.ID, &sp.LotID, &sp.Code, &sp.Status, &sp.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sp, nil
}

// CreateSpot 新建车位。
func (s *Store) CreateSpot(lotID int64, code string, status models.SpotStatus) (int64, error) {
	if code = strings.TrimSpace(code); code == "" {
		return 0, ErrInvalidInput
	}
	if !validStatus(status) {
		return 0, ErrInvalidInput
	}
	res, err := s.db.Exec(`INSERT INTO parking_spots (lot_id, code, status, created_at) VALUES (?,?,?,?)`,
		lotID, code, string(status), now())
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, fmt.Errorf("%w: 该停车场下车位编号 %s 已存在", ErrConflict, code)
		}
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateSpot 更新车位编号与状态。
func (s *Store) UpdateSpot(id int64, code string, status models.SpotStatus) error {
	code = strings.TrimSpace(code)
	if code == "" || !validStatus(status) {
		return ErrInvalidInput
	}
	res, err := s.db.Exec(`UPDATE parking_spots SET code=?, status=? WHERE id=?`,
		code, string(status), id)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return fmt.Errorf("%w: 该停车场下车位编号 %s 已存在", ErrConflict, code)
		}
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteSpot 删除车位。若车位处于占用中则拒绝。
func (s *Store) DeleteSpot(id int64) error {
	var status string
	err := s.db.QueryRow(`SELECT status FROM parking_spots WHERE id=?`, id).Scan(&status)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status == string(models.SpotOccupied) {
		return errors.New("车位占用中，无法删除")
	}
	_, err = s.db.Exec(`DELETE FROM parking_spots WHERE id=?`, id)
	return err
}

// setSpotStatus 修改车位状态（事务内调用）。
func setSpotStatus(tx *sql.Tx, id int64, status models.SpotStatus) error {
	_, err := tx.Exec(`UPDATE parking_spots SET status=? WHERE id=?`, string(status), id)
	return err
}

// releaseSpotAfterCheckout 释放已结算车位，但不覆盖巡检期间设置的维护状态。
func releaseSpotAfterCheckout(tx *sql.Tx, id int64) error {
	_, err := tx.Exec(`UPDATE parking_spots
		SET status=?
		WHERE id=? AND status<>?`,
		string(models.SpotAvailable), id, string(models.SpotMaintenance))
	return err
}

func validStatus(s models.SpotStatus) bool {
	switch s {
	case models.SpotAvailable, models.SpotOccupied, models.SpotMaintenance:
		return true
	}
	return false
}

// --- 车辆 ---

// ListVehicles 列出所有车辆。
func (s *Store) ListVehicles() ([]models.Vehicle, error) {
	rows, err := s.db.Query(`SELECT id, plate, vehicle_type, color, owner_name, owner_phone, created_at FROM vehicles ORDER BY plate`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Vehicle
	for rows.Next() {
		var v models.Vehicle
		if err := rows.Scan(&v.ID, &v.Plate, &v.VehicleType, &v.Color, &v.OwnerName, &v.OwnerPhone, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// GetVehicleByPlate 按车牌查询车辆。
func (s *Store) GetVehicleByPlate(plate string) (*models.Vehicle, error) {
	var v models.Vehicle
	err := s.db.QueryRow(`SELECT id, plate, vehicle_type, color, owner_name, owner_phone, created_at FROM vehicles WHERE plate=?`, plate).
		Scan(&v.ID, &v.Plate, &v.VehicleType, &v.Color, &v.OwnerName, &v.OwnerPhone, &v.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// GetVehicleByID 按 ID 查询车辆。
func (s *Store) GetVehicleByID(id int64) (*models.Vehicle, error) {
	var v models.Vehicle
	err := s.db.QueryRow(`SELECT id, plate, vehicle_type, color, owner_name, owner_phone, created_at FROM vehicles WHERE id=?`, id).
		Scan(&v.ID, &v.Plate, &v.VehicleType, &v.Color, &v.OwnerName, &v.OwnerPhone, &v.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// DeleteVehicle 删除车辆。若有在场的停车记录则拒绝。
func (s *Store) DeleteVehicle(id int64) error {
	var active int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM parking_records WHERE vehicle_id=? AND status='active'`, id).Scan(&active)
	if err != nil {
		return err
	}
	if active > 0 {
		return errors.New("车辆有在场记录，无法删除")
	}
	res, err := s.db.Exec(`DELETE FROM vehicles WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// UpsertVehicle 按车牌创建或更新车辆，返回车辆 ID。
func (s *Store) UpsertVehicle(v models.Vehicle) (int64, error) {
	v.Plate = strings.TrimSpace(strings.ToUpper(v.Plate))
	if v.Plate == "" {
		return 0, ErrInvalidInput
	}
	if v.VehicleType == "" {
		v.VehicleType = "car"
	}
	// 先尝试查已存在
	var id int64
	err := s.db.QueryRow(`SELECT id FROM vehicles WHERE plate=?`, v.Plate).Scan(&id)
	if err == nil {
		_, err = s.db.Exec(`UPDATE vehicles SET vehicle_type=?, color=?, owner_name=?, owner_phone=? WHERE id=?`,
			v.VehicleType, v.Color, v.OwnerName, v.OwnerPhone, id)
		if err != nil {
			return 0, err
		}
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	res, err := s.db.Exec(`INSERT INTO vehicles (plate, vehicle_type, color, owner_name, owner_phone, created_at) VALUES (?,?,?,?,?,?)`,
		v.Plate, v.VehicleType, v.Color, v.OwnerName, v.OwnerPhone, now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
