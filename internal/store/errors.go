package store

import (
	"errors"
	"fmt"

	"carmanageweb/internal/models"
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

// VehicleAlreadyParkedError 带出仍在场记录的位置，供调用方提示用户先结算原记录。
type VehicleAlreadyParkedError struct {
	Active models.ActiveParking
}

func (e *VehicleAlreadyParkedError) Error() string {
	place := e.Active.SpotCode
	if e.Active.LotName != "" {
		place = e.Active.LotName + " " + place
	}
	if place == "" {
		place = fmt.Sprintf("记录 #%d", e.Active.RecordID)
	}
	return "车辆已有在场记录（" + place + "），请先完成原记录的出场结算"
}

func (e *VehicleAlreadyParkedError) Unwrap() error {
	return ErrVehicleAlreadyParked
}
