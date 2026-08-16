// Package models 定义停车场管理系统的核心数据结构。
package models

import "time"

// ParkingLot 停车场
type ParkingLot struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Address    string `json:"address"`
	TotalSpots int    `json:"total_spots"`
	CreatedAt  Time   `json:"created_at"`
}

// SpotStatus 车位状态
type SpotStatus string

const (
	SpotAvailable   SpotStatus = "available"   // 空闲
	SpotOccupied    SpotStatus = "occupied"    // 占用
	SpotMaintenance SpotStatus = "maintenance" // 维护中
)

// ParkingSpot 车位
type ParkingSpot struct {
	ID        int64     `json:"id"`
	LotID     int64     `json:"lot_id"`
	Code      string    `json:"code"`
	Status    SpotStatus `json:"status"`
	CreatedAt Time      `json:"created_at"`
}

// Vehicle 车辆
type Vehicle struct {
	ID          int64  `json:"id"`
	Plate       string `json:"plate"`
	VehicleType string `json:"vehicle_type"` // car / suv / truck / motorcycle
	Color       string `json:"color"`
	OwnerName   string `json:"owner_name"`
	OwnerPhone  string `json:"owner_phone"`
	CreatedAt   Time   `json:"created_at"`
}

// FeeRule 费用规则
//
// 计费逻辑：
//  1. 停车时长 <= FreeMinutes 时免费
//  2. 否则计算首段费用 FirstBlockPrice（覆盖前 FirstBlockMinutes 分钟）
//  3. 超出首段的部分按 UnitMinutes/UnitPrice 计费，不足一个单位按一个单位算
//  4. 每日费用不超过 DailyCap（0 表示不封顶），跨日分别累计
type FeeRule struct {
	ID                int64   `json:"id"`
	LotID             *int64  `json:"lot_id"`              // nil 表示全局规则
	Name              string  `json:"name"`
	FreeMinutes       int     `json:"free_minutes"`
	FirstBlockMinutes int     `json:"first_block_minutes"`
	FirstBlockPrice   float64 `json:"first_block_price"`
	UnitMinutes       int     `json:"unit_minutes"`
	UnitPrice         float64 `json:"unit_price"`
	DailyCap          float64 `json:"daily_cap"`
	Active            bool    `json:"active"`
	CreatedAt         Time    `json:"created_at"`
}

// RecordStatus 停车记录状态
type RecordStatus string

const (
	RecordActive    RecordStatus = "active"    // 在场
	RecordCompleted RecordStatus = "completed" // 已离场
)

// ParkingRecord 停车记录
type ParkingRecord struct {
	ID           int64        `json:"id"`
	SpotID       int64        `json:"spot_id"`
	VehicleID    int64        `json:"vehicle_id"`
	RuleID       *int64       `json:"rule_id"`
	CheckInTime  Time         `json:"check_in_time"`
	CheckOutTime *time.Time   `json:"check_out_time"`
	Fee          float64      `json:"fee"`
	Status       RecordStatus `json:"status"`
	CreatedAt    Time         `json:"created_at"`
}

// --- 用于 HTTP 层的关联视图结构 ---

// SpotDetail 带停车场与当前在场车辆信息的车位视图
type SpotDetail struct {
	ParkingSpot
	LotName        string  `json:"lot_name"`
	CurrentPlate   string  `json:"current_plate"`   // 当前占用该车位的车辆牌号（若有）
	CurrentRecordID int64 `json:"current_record_id"` // 当前在场记录 ID（若有）
}

// RecordDetail 带关联信息的停车记录视图
type RecordDetail struct {
	ParkingRecord
	SpotCode  string `json:"spot_code"`
	LotName   string `json:"lot_name"`
	Plate     string `json:"plate"`
	RuleName  string `json:"rule_name"`
}

// FeeBreakdown 费用计算明细，用于在前端展示计算过程
type FeeBreakdown struct {
	RuleName        string  `json:"rule_name"`
	DurationMinutes int     `json:"duration_minutes"`
	IsFree          bool    `json:"is_free"`
	FirstBlockFee   float64 `json:"first_block_fee"`
	ExtraUnits      int     `json:"extra_units"`
	ExtraBlockFee   float64 `json:"extra_block_fee"`
	TotalFee        float64 `json:"total_fee"`
	DailyCapApplied  bool    `json:"daily_cap_applied"`
}
