package store

import (
	"database/sql"
	"testing"
	"time"

	"carmanageweb/internal/fee"
	"carmanageweb/internal/models"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("打开内存库失败: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// TestCheckInCheckOutFlow 验证入场→出场完整流程与费用计算。
func TestCheckInCheckOutFlow(t *testing.T) {
	s := newTestStore(t)
	calc := fee.New()

	// 准备：停车场 + 车位 + 全局规则
	lotID, _ := s.CreateLot("测试场", "地址", 10)
	spotID, _ := s.CreateSpot(lotID, "A01", models.SpotAvailable)
	_, err := s.CreateFeeRule(models.FeeRule{
		Name: "测试规则", FreeMinutes: 15,
		FirstBlockMinutes: 60, FirstBlockPrice: 10,
		UnitMinutes: 30, UnitPrice: 5, DailyCap: 0, Active: true,
	})
	if err != nil {
		t.Fatalf("创建规则失败: %v", err)
	}

	// 车辆入场
	vehicleID, _ := s.UpsertVehicle(models.Vehicle{Plate: "京A00001"})
	recID, rule, err := s.CheckIn(spotID, vehicleID)
	if err != nil {
		t.Fatalf("入场失败: %v", err)
	}
	if rule == nil || rule.Name != "测试规则" {
		t.Fatalf("入场应返回生效规则，got %+v", rule)
	}
	// 车位应变为占用
	sp, _ := s.GetSpot(spotID)
	if sp.Status != models.SpotOccupied {
		t.Fatalf("入场后车位应占用，got %s", sp.Status)
	}

	// 再次入场同一车位应失败
	_, _, err = s.CheckIn(spotID, vehicleID)
	if err == nil {
		t.Fatal("占用车位再次入场应失败")
	}

	// 出场：构造一个 75 分钟前的入场时间无法做到（已写入），改为直接用记录的入场时间算
	// 这里用当前时间出场，费用取决于真实经过时长，可能为免费或首段。
	// 为确定性，手动把 check_in 改为 75 分钟前。
	rec, _ := s.GetRecord(recID)
	past := time.Now().UTC().Add(-75 * time.Minute)
	_, err = s.db.Exec(`UPDATE parking_records SET check_in_time=? WHERE id=?`, past.Format(time.RFC3339), recID)
	if err != nil {
		t.Fatalf("更新入场时间失败: %v", err)
	}
	_ = rec

	feeAmt, bd, err := s.CheckOut(recID, time.Time{}, calc.Calc)
	if err != nil {
		t.Fatalf("出场失败: %v", err)
	}
	// 75 分钟：扣 15 免费 = 60，正好首段，应收 10
	if feeAmt != 10 {
		t.Fatalf("75分钟应收10，got %.2f (明细 %+v)", feeAmt, bd)
	}
	// 车位应恢复空闲
	sp, _ = s.GetSpot(spotID)
	if sp.Status != models.SpotAvailable {
		t.Fatalf("出场后车位应空闲，got %s", sp.Status)
	}
	// 记录状态应为已完成
	rec2, _ := s.GetRecord(recID)
	if rec2.Status != models.RecordCompleted {
		t.Fatalf("出场后记录应完成，got %s", rec2.Status)
	}
	if rec2.CheckOutTime == nil {
		t.Fatal("出场时间应已记录")
	}

	// 重复出场应失败
	_, _, err = s.CheckOut(recID, time.Time{}, calc.Calc)
	if err == nil {
		t.Fatal("重复出场应失败")
	}
}

// TestResolveFeeRule 验证规则解析优先级：停车场专属 > 全局。
func TestResolveFeeRule(t *testing.T) {
	s := newTestStore(t)
	lotID, _ := s.CreateLot("测试场", "", 5)

	// 仅全局规则
	s.CreateFeeRule(models.FeeRule{Name: "全局", FirstBlockMinutes: 60, FirstBlockPrice: 5, UnitMinutes: 60, UnitPrice: 3, Active: true})
	r, err := s.ResolveFeeRule(lotID)
	if err != nil || r.Name != "全局" {
		t.Fatalf("应解析到全局规则，got %v %v", r, err)
	}

	// 新增停车场专属规则
	s.CreateFeeRule(models.FeeRule{Name: "专属", LotID: &lotID, FirstBlockMinutes: 60, FirstBlockPrice: 8, UnitMinutes: 60, UnitPrice: 4, Active: true})
	r, err = s.ResolveFeeRule(lotID)
	if err != nil || r.Name != "专属" {
		t.Fatalf("应优先解析到专属规则，got %v %v", r, err)
	}

	// 无任何启用规则时应返回 ErrNotFound
	s2 := newTestStore(t)
	_, err = s2.ResolveFeeRule(999)
	if err != ErrNotFound {
		t.Fatalf("无规则应返回 ErrNotFound，got %v", err)
	}
}

// seedRule 写入一个启用的全局费用规则，供需要 CheckIn 的测试使用。
func seedRule(t *testing.T, s *Store) {
	t.Helper()
	_, err := s.CreateFeeRule(models.FeeRule{
		Name: "测试规则", FreeMinutes: 15,
		FirstBlockMinutes: 60, FirstBlockPrice: 10,
		UnitMinutes: 30, UnitPrice: 5, Active: true,
	})
	if err != nil {
		t.Fatalf("创建规则失败: %v", err)
	}
}

// TestDeleteSpotOccupied 验证占用中的车位不可删除。
func TestDeleteSpotOccupied(t *testing.T) {
	s := newTestStore(t)
	seedRule(t, s)
	lotID, _ := s.CreateLot("测试场", "", 5)
	spotID, _ := s.CreateSpot(lotID, "A01", models.SpotAvailable)
	vid, _ := s.UpsertVehicle(models.Vehicle{Plate: "京B00002"})
	if _, _, err := s.CheckIn(spotID, vid); err != nil {
		t.Fatalf("入场失败: %v", err)
	}
	err := s.DeleteSpot(spotID)
	if err == nil {
		t.Fatal("占用中车位应无法删除")
	}
}

// TestDeleteVehicleWithActiveRecord 验证有在场记录的车辆不可删除。
func TestDeleteVehicleWithActiveRecord(t *testing.T) {
	s := newTestStore(t)
	seedRule(t, s)
	lotID, _ := s.CreateLot("测试场", "", 5)
	spotID, _ := s.CreateSpot(lotID, "A01", models.SpotAvailable)
	vid, _ := s.UpsertVehicle(models.Vehicle{Plate: "京C00003"})
	if _, _, err := s.CheckIn(spotID, vid); err != nil {
		t.Fatalf("入场失败: %v", err)
	}
	if err := s.DeleteVehicle(vid); err == nil {
		t.Fatal("有在场记录的车辆应无法删除")
	}
}

// TestStats 验证统计数据。
func TestStats(t *testing.T) {
	s := newTestStore(t)
	seedRule(t, s)
	lotID, _ := s.CreateLot("测试场", "", 5)
	s.CreateSpot(lotID, "A01", models.SpotAvailable)
	s.CreateSpot(lotID, "A02", models.SpotAvailable)
	vid, _ := s.UpsertVehicle(models.Vehicle{Plate: "京D00004"})
	spotID, _ := s.CreateSpot(lotID, "A03", models.SpotAvailable)
	if _, _, err := s.CheckIn(spotID, vid); err != nil {
		t.Fatalf("入场失败: %v", err)
	}

	st, err := s.Stats()
	if err != nil {
		t.Fatalf("Stats 失败: %v", err)
	}
	if st.TotalLots != 1 || st.TotalSpots != 3 || st.OccupiedSpots != 1 || st.AvailableSpots != 2 || st.ActiveRecords != 1 {
		t.Fatalf("统计数据不符: %+v", st)
	}
}

// 避免未使用导入告警
var _ = sql.ErrNoRows
