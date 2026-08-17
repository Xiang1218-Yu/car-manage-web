package store

import (
	"errors"
	"testing"
	"time"

	"carmanageweb/internal/fee"
	"carmanageweb/internal/models"
)

func TestCheckInRejectsVehicleWithAnotherActiveRecord(t *testing.T) {
	s := newTestStore(t)
	seedRule(t, s)
	lotID, err := s.CreateLot("双入口测试场", "", 4)
	if err != nil {
		t.Fatalf("创建停车场失败: %v", err)
	}
	firstSpotID, _ := s.CreateSpot(lotID, "A01", models.SpotAvailable)
	secondSpotID, _ := s.CreateSpot(lotID, "B01", models.SpotAvailable)
	vehicleID, err := s.UpsertVehicle(models.Vehicle{Plate: "京D00501"})
	if err != nil {
		t.Fatalf("创建车辆失败: %v", err)
	}

	firstRecordID, _, err := s.CheckIn(firstSpotID, vehicleID)
	if err != nil {
		t.Fatalf("第一次入场失败: %v", err)
	}
	if _, _, err := s.CheckIn(secondSpotID, vehicleID); !errors.Is(err, ErrVehicleAlreadyParked) {
		t.Fatalf("同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got %v", err)
	}

	secondSpot, err := s.GetSpot(secondSpotID)
	if err != nil {
		t.Fatalf("读取第二个车位失败: %v", err)
	}
	if secondSpot.Status != models.SpotAvailable {
		t.Fatalf("被拒绝的第二次入场不应占用车位，got %s", secondSpot.Status)
	}

	if _, _, err := s.CheckOut(firstRecordID, time.Now().UTC().Add(time.Minute), fee.New().Calc); err != nil {
		t.Fatalf("第一次记录出场失败: %v", err)
	}
	if _, _, err := s.CheckIn(secondSpotID, vehicleID); err != nil {
		t.Fatalf("原在场记录结束后应允许再次入场: %v", err)
	}
}
