package store

import (
	"database/sql"
	"fmt"
	"time"

	"carmanageweb/internal/models"
)

// inTx 在事务中执行 fn，发生错误自动回滚。
func (s *Store) inTx(fn func(*sql.Tx) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%w (回滚失败: %v)", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

// FeeCalculator 由费用引擎实现，供 CheckOut 在事务内调用。
// 入参为入场时间、出场时间、生效规则；返回费用明细。
type FeeCalculator func(checkIn, checkOut time.Time, rule *models.FeeRule) *models.FeeBreakdown

// resolveFeeRuleTx 在事务内解析停车场生效规则。
func (s *Store) resolveFeeRuleTx(tx *sql.Tx, lotID int64) (*models.FeeRule, error) {
	r, err := s.getFeeRuleByLotTx(tx, lotID)
	if err == nil {
		return r, nil
	}
	if err != ErrNotFound {
		return nil, err
	}
	return s.getGlobalFeeRuleTx(tx)
}

// resolveFeeRuleBySpotTx 通过车位反查停车场再解析规则。
func (s *Store) resolveFeeRuleBySpotTx(tx *sql.Tx, spotID int64) (*models.FeeRule, error) {
	var lotID sql.NullInt64
	err := tx.QueryRow(`SELECT lot_id FROM parking_spots WHERE id=?`, spotID).Scan(&lotID)
	if err != nil {
		return nil, err
	}
	if !lotID.Valid {
		return s.getGlobalFeeRuleTx(tx)
	}
	return s.resolveFeeRuleTx(tx, lotID.Int64)
}

func (s *Store) getFeeRuleByLotTx(tx *sql.Tx, lotID int64) (*models.FeeRule, error) {
	var r models.FeeRule
	var active int
	err := tx.QueryRow(`SELECT id, lot_id, name, free_minutes, first_block_minutes, first_block_price,
		unit_minutes, unit_price, daily_cap, active, created_at FROM fee_rules WHERE lot_id=? AND active=1 LIMIT 1`, lotID).
		Scan(&r.ID, &r.LotID, &r.Name, &r.FreeMinutes, &r.FirstBlockMinutes, &r.FirstBlockPrice,
			&r.UnitMinutes, &r.UnitPrice, &r.DailyCap, &active, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	r.Active = active == 1
	return &r, nil
}

func (s *Store) getGlobalFeeRuleTx(tx *sql.Tx) (*models.FeeRule, error) {
	var r models.FeeRule
	var lotID sql.NullInt64
	var active int
	err := tx.QueryRow(`SELECT id, lot_id, name, free_minutes, first_block_minutes, first_block_price,
		unit_minutes, unit_price, daily_cap, active, created_at FROM fee_rules WHERE lot_id IS NULL AND active=1 LIMIT 1`).
		Scan(&r.ID, &lotID, &r.Name, &r.FreeMinutes, &r.FirstBlockMinutes, &r.FirstBlockPrice,
			&r.UnitMinutes, &r.UnitPrice, &r.DailyCap, &active, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	r.Active = active == 1
	return &r, nil
}

func (s *Store) getFeeRuleTx(tx *sql.Tx, id int64) (*models.FeeRule, error) {
	var r models.FeeRule
	var lotID sql.NullInt64
	var active int
	err := tx.QueryRow(`SELECT id, lot_id, name, free_minutes, first_block_minutes, first_block_price,
		unit_minutes, unit_price, daily_cap, active, created_at FROM fee_rules WHERE id=?`, id).
		Scan(&r.ID, &lotID, &r.Name, &r.FreeMinutes, &r.FirstBlockMinutes, &r.FirstBlockPrice,
			&r.UnitMinutes, &r.UnitPrice, &r.DailyCap, &active, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if lotID.Valid {
		li := lotID.Int64
		r.LotID = &li
	}
	r.Active = active == 1
	return &r, nil
}

// Statistics 用于仪表盘统计。
type Statistics struct {
	TotalLots      int     `json:"total_lots"`
	TotalSpots     int     `json:"total_spots"`
	AvailableSpots int     `json:"available_spots"`
	OccupiedSpots  int     `json:"occupied_spots"`
	ActiveRecords  int     `json:"active_records"`
	TodayRevenue   float64 `json:"today_revenue"`
	TotalRecords   int     `json:"total_records"`
}

// Stats 返回仪表盘统计数据。
func (s *Store) Stats() (Statistics, error) {
	var st Statistics
	queries := []struct {
		query string
		dst   interface{}
	}{
		{`SELECT COUNT(*) FROM parking_lots`, &st.TotalLots},
		{`SELECT COUNT(*) FROM parking_spots`, &st.TotalSpots},
		{`SELECT COUNT(*) FROM parking_spots WHERE status='available'`, &st.AvailableSpots},
		{`SELECT COUNT(*) FROM parking_spots WHERE status='occupied'`, &st.OccupiedSpots},
		{`SELECT COUNT(*) FROM parking_records WHERE status='active'`, &st.ActiveRecords},
		{`SELECT COALESCE(SUM(fee),0) FROM parking_records WHERE status='completed' AND date(check_out_time)=date('now')`, &st.TodayRevenue},
		{`SELECT COUNT(*) FROM parking_records`, &st.TotalRecords},
	}
	for _, q := range queries {
		if err := s.db.QueryRow(q.query).Scan(q.dst); err != nil {
			return st, err
		}
	}
	return st, nil
}
