package main

import (
	"carmanageweb/internal/models"
	"carmanageweb/internal/store"
)

// seedDemo 在数据库为空时写入一组演示数据，便于直接体验。
func seedDemo(st *store.Store) error {
	// 仅在无停车场时填充
	lots, err := st.ListLots()
	if err != nil {
		return err
	}
	if len(lots) > 0 {
		return nil
	}

	// 两个停车场
	lot1ID, err := st.CreateLot("中央广场停车场", "人民路 88 号", 30)
	if err != nil {
		return err
	}
	lot2ID, err := st.CreateLot("滨江商业中心", "滨江大道 1 号", 50)
	if err != nil {
		return err
	}

	// 给停车场 1 创建 12 个车位（A 区）
	for i := 1; i <= 12; i++ {
		code := codeOf("A", i)
		if _, err := st.CreateSpot(lot1ID, code, models.SpotAvailable); err != nil {
			return err
		}
	}
	// 停车场 2 创建 10 个车位
	for i := 1; i <= 10; i++ {
		code := codeOf("B", i)
		if _, err := st.CreateSpot(lot2ID, code, models.SpotAvailable); err != nil {
			return err
		}
	}

	// 全局费用规则：免费 15 分钟，首段 1 小时 10 元，之后每 30 分钟 5 元
	if _, err := st.CreateFeeRule(models.FeeRule{
		Name: "默认计费", FreeMinutes: 15,
		FirstBlockMinutes: 60, FirstBlockPrice: 10,
		UnitMinutes: 30, UnitPrice: 5, DailyCap: 60, Active: true,
	}); err != nil {
		return err
	}
	// 停车场 2 专属规则（夜间封顶更低）
	if _, err := st.CreateFeeRule(models.FeeRule{
		Name: "滨江夜间封顶", FreeMinutes: 0,
		FirstBlockMinutes: 60, FirstBlockPrice: 8,
		UnitMinutes: 60, UnitPrice: 4, DailyCap: 40,
		LotID: &lot2ID, Active: true,
	}); err != nil {
		return err
	}
	return nil
}

// codeOf 生成车位编号，如 A01、A12。
func codeOf(prefix string, n int) string {
	if n < 10 {
		return prefix + "0" + itoa(n)
	}
	return prefix + itoa(n)
}

// itoa 简单整数转字符串（避免引入 strconv 在 main 包重复）。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
