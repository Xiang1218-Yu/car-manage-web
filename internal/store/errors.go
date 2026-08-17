package store

import (
	"errors"
	"strings"
)

// 数据访问层错误。
var (
	// ErrNotFound 表示查询的记录不存在。
	ErrNotFound = errors.New("记录不存在")
	// ErrConflict 表示违反唯一约束等冲突。
	ErrConflict = errors.New("数据冲突")
	// ErrInvalidInput 表示输入数据不合法。
	ErrInvalidInput = errors.New("输入不合法")
	// ErrVehicleAlreadyParked 表示同一车辆仍有未结算记录。
	ErrVehicleAlreadyParked = errors.New("车辆已有在场记录")
)

// isUniqueViolation 判断是否为 SQLite 唯一约束冲突。
// modernc.org/sqlite 在违反 UNIQUE 约束时返回的错误信息包含 "UNIQUE"。
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE")
}

