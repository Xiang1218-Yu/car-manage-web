// Package store 封装停车场系统的 SQLite 数据访问。
package store

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"carmanageweb/internal/models"
)

//go:embed schema.sql
var schemaSQL string

// Store 持有数据库连接，提供所有数据访问方法。
type Store struct {
	db *sql.DB
}

// Open 打开（或创建）SQLite 数据库并执行建表语句。
// path 为 ":memory:" 时使用内存库，否则为文件路径。
func Open(path string) (*Store, error) {
	// busy_timeout 避免并发写时立即报锁错误。
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库: %w", err)
	}
	// SQLite 单连接足以支撑本系统规模，且避免多写连接冲突。
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("执行建表语句: %w", err)
	}
	return &Store{db: db}, nil
}

// Close 关闭数据库。
func (s *Store) Close() error { return s.db.Close() }

// now 返回 UTC 时间字符串，统一存储格式。
func now() string { return time.Now().UTC().Format(time.RFC3339) }

// --- 停车场 ---

// ListLots 列出所有停车场。
func (s *Store) ListLots() ([]models.ParkingLot, error) {
	rows, err := s.db.Query(`SELECT id, name, address, total_spots, created_at FROM parking_lots ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.ParkingLot
	for rows.Next() {
		var l models.ParkingLot
		if err := rows.Scan(&l.ID, &l.Name, &l.Address, &l.TotalSpots, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// GetLot 按 ID 查询停车场。
func (s *Store) GetLot(id int64) (*models.ParkingLot, error) {
	var l models.ParkingLot
	err := s.db.QueryRow(`SELECT id, name, address, total_spots, created_at FROM parking_lots WHERE id=?`, id).
		Scan(&l.ID, &l.Name, &l.Address, &l.TotalSpots, &l.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// CreateLot 新建停车场。
func (s *Store) CreateLot(name, address string, totalSpots int) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO parking_lots (name, address, total_spots, created_at) VALUES (?,?,?,?)`,
		name, address, totalSpots, now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateLot 更新停车场信息。
func (s *Store) UpdateLot(id int64, name, address string, totalSpots int) error {
	res, err := s.db.Exec(`UPDATE parking_lots SET name=?, address=?, total_spots=? WHERE id=?`,
		name, address, totalSpots, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteLot 删除停车场（关联车位级联删除）。
func (s *Store) DeleteLot(id int64) error {
	res, err := s.db.Exec(`DELETE FROM parking_lots WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
