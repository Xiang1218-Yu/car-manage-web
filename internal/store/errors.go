package store

import "errors"

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
