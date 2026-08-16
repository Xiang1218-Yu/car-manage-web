// Package fee 实现停车费用计算引擎。
//
// 计费规则（参见 models.FeeRule 文档）：
//  1. 停车时长 <= FreeMinutes 时免费
//  2. 否则首段收取 FirstBlockPrice（覆盖前 FirstBlockMinutes 分钟）
//  3. 超出首段的部分按 UnitMinutes/UnitPrice 计费，不足一个单位按一个单位算
//  4. 每日费用不超过 DailyCap（0 表示不封顶），跨日分别累计
package fee

import (
	"math"
	"time"

	"carmanageweb/internal/models"
)

// Calculator 独立的费用计算器，不依赖数据库，便于单测。
type Calculator struct{}

// New 创建计算器。
func New() *Calculator { return &Calculator{} }

// Calc 根据入场时间、出场时间与规则计算费用明细。
// 入场晚于出场时返回 0 费用并标记免费。
func (c *Calculator) Calc(checkIn, checkOut time.Time, rule *models.FeeRule) *models.FeeBreakdown {
	bd := &models.FeeBreakdown{RuleName: rule.Name}
	if rule == nil {
		bd.IsFree = true
		return bd
	}
	if checkOut.Before(checkIn) || checkOut.Equal(checkIn) {
		bd.IsFree = true
		return bd
	}
	// 统一使用入场时刻所在时区进行日切计算。
	loc := checkIn.Location()
	durMin := int(checkOut.Sub(checkIn).Minutes())
	bd.DurationMinutes = durMin

	// 1) 免费时长
	if rule.FreeMinutes > 0 && durMin <= rule.FreeMinutes {
		bd.IsFree = true
		return bd
	}
	bd.IsFree = false

	if rule.DailyCap > 0 {
		bd.TotalFee = c.calcCapped(checkIn, checkOut, loc, rule, bd)
	} else {
		bd.TotalFee = c.calcUncapped(durMin, rule, bd)
	}
	bd.TotalFee = round2(bd.TotalFee)
	if bd.TotalFee == 0 {
		bd.IsFree = true
	}
	return bd
}

// calcUncapped 无封顶时的计算：免费时长后按首段 + 后续单位计费。
func (c *Calculator) calcUncapped(durMin int, rule *models.FeeRule, bd *models.FeeBreakdown) float64 {
	chargeable := durMin
	if rule.FreeMinutes > 0 {
		chargeable = durMin - rule.FreeMinutes
	}
	if chargeable <= 0 {
		return 0
	}
	bd.FirstBlockFee = rule.FirstBlockPrice
	// 首段覆盖的计费分钟数：扣掉免费时长后的剩余部分。
	firstCharge := rule.FirstBlockMinutes
	if chargeable <= firstCharge {
		return rule.FirstBlockPrice
	}
	rest := chargeable - firstCharge
	units := (rest + rule.UnitMinutes - 1) / rule.UnitMinutes // 向上取整
	bd.ExtraUnits = units
	bd.ExtraBlockFee = round2(float64(units) * rule.UnitPrice)
	return bd.FirstBlockFee + bd.ExtraBlockFee
}

// calcCapped 带每日封顶的计算：按自然日切分，每日分别计算并封顶，再汇总。
// 首段费用只在第一个计费日扣除。
func (c *Calculator) calcCapped(checkIn, checkOut time.Time, loc *time.Location, rule *models.FeeRule, bd *models.FeeBreakdown) float64 {
	bd.DailyCapApplied = true
	// 把时间归一到本地时区做日切
	ci := checkIn.In(loc)
	co := checkOut.In(loc)
	dayStart := time.Date(ci.Year(), ci.Month(), ci.Day(), 0, 0, 0, 0, loc)

	var total float64
	firstDay := true
	freeApplied := false
	// 遍历从入场日 0 点到出场日的每一天
	for d := dayStart; !d.After(co); d = d.AddDate(0, 0, 1) {
		dayEnd := d.AddDate(0, 0, 1)
		segStart := ci
		segEnd := co
		if segStart.Before(d) {
			segStart = d
		}
		if segEnd.After(dayEnd) {
			segEnd = dayEnd
		}
		if !segEnd.After(segStart) {
			continue
		}
		segMin := int(segEnd.Sub(segStart).Minutes())

		var dayFee float64
		if firstDay {
			// 该段扣减免费时长
			chargeable := segMin
			if rule.FreeMinutes > 0 && !freeApplied {
				freeApplied = true
				chargeable = segMin - rule.FreeMinutes
			}
			if chargeable > 0 {
				dayFee = c.dayFeeWithFirstBlock(chargeable, rule, bd)
			}
			firstDay = false
		} else {
			// 非首日不再有首段，全部按单位价计费
			if segMin > 0 {
				units := (segMin + rule.UnitMinutes - 1) / rule.UnitMinutes
				dayFee = float64(units) * rule.UnitPrice
			}
		}
		if dayFee > rule.DailyCap {
			dayFee = rule.DailyCap
		}
		total += dayFee
	}
	// 免费情况
	if rule.FreeMinutes > 0 && (checkOut.Sub(checkIn).Minutes()) <= float64(rule.FreeMinutes) {
		bd.IsFree = true
		return 0
	}
	return total
}

// dayFeeWithFirstBlock 首日计费：首段价 + 后续单位，返回该日费用（未封顶）。
func (c *Calculator) dayFeeWithFirstBlock(chargeable int, rule *models.FeeRule, bd *models.FeeBreakdown) float64 {
	bd.FirstBlockFee = rule.FirstBlockPrice
	if chargeable <= rule.FirstBlockMinutes {
		return rule.FirstBlockPrice
	}
	rest := chargeable - rule.FirstBlockMinutes
	units := (rest + rule.UnitMinutes - 1) / rule.UnitMinutes
	bd.ExtraUnits += units
	return rule.FirstBlockPrice + float64(units)*rule.UnitPrice
}

// round2 保留两位小数，避免浮点误差。
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
