package store

import (
	"testing"
	"time"

	"carmanageweb/internal/fee"
	"carmanageweb/internal/models"
)

func TestCheckOutPreservesMaintenanceStatus(t *testing.T) {
	s := newTestStore(t)
	seedRule(t, s)
	lotID, err := s.CreateLot("巡检测试场", "", 2)
	if err != nil {
		t.Fatalf("创建停车场失败: %v", err)
	}
	spotID, err := s.CreateSpot(lotID, "M01", models.SpotAvailable)
	if err != nil {
		t.Fatalf("创建车位失败: %v", err)
	}
	vehicleID, err := s.UpsertVehicle(models.Vehicle{Plate: "京M00401"})
	if err != nil {
		t.Fatalf("创建车辆失败: %v", err)
	}
	recordID, _, err := s.CheckIn(spotID, vehicleID)
	if err != nil {
		t.Fatalf("车辆入场失败: %v", err)
	}

	if err := s.UpdateSpotStatus(spotID, models.SpotMaintenance); err != nil {
		t.Fatalf("巡检标记维护失败: %v", err)
	}
	if _, _, err := s.CheckOut(recordID, time.Now().UTC().Add(time.Minute), fee.New().Calc); err != nil {
		t.Fatalf("车辆出场失败: %v", err)
	}

	spot, err := s.GetSpot(spotID)
	if err != nil {
		t.Fatalf("读取车位失败: %v", err)
	}
	if spot.Status != models.SpotMaintenance {
		t.Fatalf("已被巡检标记维护的车位在出场后仍应保持 maintenance，got %s", spot.Status)
	}
}
