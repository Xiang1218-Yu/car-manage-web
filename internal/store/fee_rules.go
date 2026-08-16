package store

import (
	"database/sql"
	"strings"

	"carmanageweb/internal/models"
)

// --- 费用规则 ---

// ListFeeRules 列出所有费用规则。
func (s *Store) ListFeeRules() ([]models.FeeRule, error) {
	rows, err := s.db.Query(`SELECT id, lot_id, name, free_minutes, first_block_minutes, first_block_price,
		unit_minutes, unit_price, daily_cap, active, created_at FROM fee_rules ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.FeeRule
	for rows.Next() {
		var r models.FeeRule
		var lotID sql.NullInt64
		var active int
		if err := rows.Scan(&r.ID, &lotID, &r.Name, &r.FreeMinutes, &r.FirstBlockMinutes, &r.FirstBlockPrice,
			&r.UnitMinutes, &r.UnitPrice, &r.DailyCap, &active, &r.CreatedAt); err != nil {
			return nil, err
		}
		if lotID.Valid {
			li := lotID.Int64
			r.LotID = &li
		}
		r.Active = active == 1
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetFeeRule 按 ID 查询费用规则。
func (s *Store) GetFeeRule(id int64) (*models.FeeRule, error) {
	var r models.FeeRule
	var lotID sql.NullInt64
	var active int
	err := s.db.QueryRow(`SELECT id, lot_id, name, free_minutes, first_block_minutes, first_block_price,
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

// CreateFeeRule 新建费用规则。
func (s *Store) CreateFeeRule(r models.FeeRule) (int64, error) {
	if strings.TrimSpace(r.Name) == "" {
		return 0, ErrInvalidInput
	}
	if r.FirstBlockMinutes <= 0 || r.UnitMinutes <= 0 {
		return 0, ErrInvalidInput
	}
	active := 0
	if r.Active {
		active = 1
	}
	res, err := s.db.Exec(`INSERT INTO fee_rules
		(lot_id, name, free_minutes, first_block_minutes, first_block_price, unit_minutes, unit_price, daily_cap, active, created_at)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		r.LotID, r.Name, r.FreeMinutes, r.FirstBlockMinutes, r.FirstBlockPrice,
		r.UnitMinutes, r.UnitPrice, r.DailyCap, active, now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateFeeRule 更新费用规则。已存在的停车记录不受影响（记录在入场时已绑定规则快照逻辑）。
func (s *Store) UpdateFeeRule(id int64, r models.FeeRule) error {
	if strings.TrimSpace(r.Name) == "" || r.FirstBlockMinutes <= 0 || r.UnitMinutes <= 0 {
		return ErrInvalidInput
	}
	active := 0
	if r.Active {
		active = 1
	}
	res, err := s.db.Exec(`UPDATE fee_rules SET lot_id=?, name=?, free_minutes=?, first_block_minutes=?,
		first_block_price=?, unit_minutes=?, unit_price=?, daily_cap=?, active=? WHERE id=?`,
		r.LotID, r.Name, r.FreeMinutes, r.FirstBlockMinutes, r.FirstBlockPrice,
		r.UnitMinutes, r.UnitPrice, r.DailyCap, active, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteFeeRule 删除费用规则。
func (s *Store) DeleteFeeRule(id int64) error {
	res, err := s.db.Exec(`DELETE FROM fee_rules WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ResolveFeeRule 为给定停车场解析生效规则：优先该停车场专属规则，其次全局规则（lot_id 为空）。
func (s *Store) ResolveFeeRule(lotID int64) (*models.FeeRule, error) {
	// 先查停车场专属的启用规则
	r, err := s.GetFeeRuleByLot(lotID)
	if err == nil {
		return r, nil
	}
	if err != nil && err != ErrNotFound {
		return nil, err
	}
	// 回退到全局规则
	var fr models.FeeRule
	var lotID2 sql.NullInt64
	var active int
	err = s.db.QueryRow(`SELECT id, lot_id, name, free_minutes, first_block_minutes, first_block_price,
		unit_minutes, unit_price, daily_cap, active, created_at FROM fee_rules WHERE lot_id IS NULL AND active=1 LIMIT 1`).
		Scan(&fr.ID, &lotID2, &fr.Name, &fr.FreeMinutes, &fr.FirstBlockMinutes, &fr.FirstBlockPrice,
			&fr.UnitMinutes, &fr.UnitPrice, &fr.DailyCap, &active, &fr.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	fr.Active = active == 1
	return &fr, nil
}

// GetFeeRuleByLot 查询某停车场专属的启用规则。
func (s *Store) GetFeeRuleByLot(lotID int64) (*models.FeeRule, error) {
	var fr models.FeeRule
	var active int
	err := s.db.QueryRow(`SELECT id, lot_id, name, free_minutes, first_block_minutes, first_block_price,
		unit_minutes, unit_price, daily_cap, active, created_at FROM fee_rules WHERE lot_id=? AND active=1 LIMIT 1`, lotID).
		Scan(&fr.ID, &fr.LotID, &fr.Name, &fr.FreeMinutes, &fr.FirstBlockMinutes, &fr.FirstBlockPrice,
			&fr.UnitMinutes, &fr.UnitPrice, &fr.DailyCap, &active, &fr.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	fr.Active = active == 1
	return &fr, nil
}
