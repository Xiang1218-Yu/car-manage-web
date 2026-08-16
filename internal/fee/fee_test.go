package fee

import (
	"math"
	"testing"
	"time"

	"carmanageweb/internal/models"
)

func mustTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

// 标准规则：免费15分钟，首段1小时10元，后续每30分钟5元，无封顶。
func stdRule() *models.FeeRule {
	return &models.FeeRule{
		Name:              "标准",
		FreeMinutes:       15,
		FirstBlockMinutes: 60,
		FirstBlockPrice:   10,
		UnitMinutes:       30,
		UnitPrice:         5,
		DailyCap:          0,
	}
}

func TestFreeWithinFreeMinutes(t *testing.T) {
	c := New()
	bd := c.Calc(mustTime("2026-08-16T10:00:00Z"), mustTime("2026-08-16T10:10:00Z"), stdRule())
	if !bd.IsFree || bd.TotalFee != 0 {
		t.Fatalf("10分钟应免费，got fee=%.2f free=%v", bd.TotalFee, bd.IsFree)
	}
}

func TestExactlyFreeBoundary(t *testing.T) {
	c := New()
	// 正好等于免费时长（15分钟）应免费
	bd := c.Calc(mustTime("2026-08-16T10:00:00Z"), mustTime("2026-08-16T10:15:00Z"), stdRule())
	if !bd.IsFree {
		t.Fatalf("15分钟(边界)应免费，got fee=%.2f", bd.TotalFee)
	}
}

func TestJustOverFreePaysFirstBlock(t *testing.T) {
	c := New()
	// 16分钟：超免费时长，落在首段内，应收首段价 10
	bd := c.Calc(mustTime("2026-08-16T10:00:00Z"), mustTime("2026-08-16T10:16:00Z"), stdRule())
	if bd.TotalFee != 10 {
		t.Fatalf("16分钟应收首段价10，got %.2f", bd.TotalFee)
	}
}

func TestFirstBlockExact(t *testing.T) {
	c := New()
	// 扣15免费后剩60，正好首段，应收10。总时长75分钟。
	bd := c.Calc(mustTime("2026-08-16T10:00:00Z"), mustTime("2026-08-16T11:15:00Z"), stdRule())
	if bd.TotalFee != 10 {
		t.Fatalf("75分钟应收10，got %.2f", bd.TotalFee)
	}
}

func TestOneExtraUnit(t *testing.T) {
	c := New()
	// 总76分钟，扣15=61，首段60 + 1分钟超出，向上取整=1个单位。应收 10 + 5 = 15
	bd := c.Calc(mustTime("2026-08-16T10:00:00Z"), mustTime("2026-08-16T11:16:00Z"), stdRule())
	if bd.TotalFee != 15 {
		t.Fatalf("76分钟应收15，got %.2f", bd.TotalFee)
	}
}

func TestTwoExtraUnits(t *testing.T) {
	c := New()
	// 总120分钟，扣15=105，首段60 + 45超出 -> 2个单位。应收 10 + 10 = 20
	bd := c.Calc(mustTime("2026-08-16T10:00:00Z"), mustTime("2026-08-16T12:00:00Z"), stdRule())
	if bd.TotalFee != 20 {
		t.Fatalf("120分钟应收20，got %.2f", bd.TotalFee)
	}
}

func TestDailyCap(t *testing.T) {
	c := New()
	r := stdRule()
	r.DailyCap = 30
	// 总300分钟（5小时），扣15=285，首段60 + 225超出=8单位(240分钟) -> 10+40=50，封顶30
	bd := c.Calc(mustTime("2026-08-16T10:00:00Z"), mustTime("2026-08-16T15:00:00Z"), r)
	if !bd.DailyCapApplied {
		t.Fatal("应触发封顶")
	}
	if bd.TotalFee != 30 {
		t.Fatalf("封顶30，got %.2f", bd.TotalFee)
	}
}

func TestCrossDayCapped(t *testing.T) {
	c := New()
	r := stdRule()
	r.DailyCap = 30
	// 入场 2026-08-16 23:00，出场 2026-08-17 01:00，共120分钟，按自然日切分：
	//   首日 23:00-24:00 = 60分，扣15免费=45分，首段内 -> 10元 (<30)
	//   次日 00:00-01:00 = 60分，非首日无首段，2单位 -> 10元 (<30)
	//   合计 20
	bd := c.Calc(mustTime("2026-08-16T23:00:00Z"), mustTime("2026-08-17T01:00:00Z"), r)
	if bd.TotalFee != 20 {
		t.Fatalf("跨日封顶应20，got %.2f", bd.TotalFee)
	}
}

func TestCrossDayEachCapped(t *testing.T) {
	c := New()
	r := stdRule()
	r.DailyCap = 30
	// 入场 2026-08-16 10:00，出场 2026-08-17 22:00（共36小时）
	//   首日 10:00-24:00 = 14小时(840分)，扣15=825，首段60+765超出=26单位(向上取整780分)
	//       未封顶 = 10 + 26*5 = 140 -> 封顶30
	//   次日 00:00-22:00 = 22小时(1320分)，非首日，44单位 -> 220 -> 封顶30
	//   合计 60
	bd := c.Calc(mustTime("2026-08-16T10:00:00Z"), mustTime("2026-08-17T22:00:00Z"), r)
	if bd.TotalFee != 60 {
		t.Fatalf("两日均封顶应60，got %.2f", bd.TotalFee)
	}
}

func TestCrossDayFirstCappedSecondNot(t *testing.T) {
	c := New()
	r := &models.FeeRule{
		Name: "经济", FreeMinutes: 30,
		FirstBlockMinutes: 120, FirstBlockPrice: 5,
		UnitMinutes: 60, UnitPrice: 2, DailyCap: 20,
	}
	// 入场 08-16 04:56:30，出场 08-17 05:56:30（共1500分钟=25小时），按自然日切分：
	//   首日 04:56:30→24:00 = 1143分，扣30免费=1113，首段120内5+剩余993→17单位×2=34，首日39→封顶20
	//   次日 00:00→05:56:30 = 356分，非首日，ceil(356/60)=6单位×2=12，未达封顶20
	//   合计 32
	bd := c.Calc(mustTime("2026-08-16T04:56:30Z"), mustTime("2026-08-17T05:56:30Z"), r)
	if bd.TotalFee != 32 {
		t.Fatalf("首日封顶次日不封顶应32，got %.2f", bd.TotalFee)
	}
}

func TestCheckoutBeforeCheckin(t *testing.T) {
	c := New()
	bd := c.Calc(mustTime("2026-08-16T10:00:00Z"), mustTime("2026-08-16T09:00:00Z"), stdRule())
	if !bd.IsFree || bd.TotalFee != 0 {
		t.Fatalf("出场早于入场应0费用，got %.2f", bd.TotalFee)
	}
}

func TestZeroDuration(t *testing.T) {
	c := New()
	bd := c.Calc(mustTime("2026-08-16T10:00:00Z"), mustTime("2026-08-16T10:00:00Z"), stdRule())
	if !bd.IsFree {
		t.Fatalf("0时长应免费，got %.2f", bd.TotalFee)
	}
}

func TestNoFreeMinutesRule(t *testing.T) {
	c := New()
	r := &models.FeeRule{
		Name: "无免费", FreeMinutes: 0,
		FirstBlockMinutes: 60, FirstBlockPrice: 5,
		UnitMinutes: 60, UnitPrice: 3,
		DailyCap: 0,
	}
	// 10分钟：无免费，落在首段，应收5
	bd := c.Calc(mustTime("2026-08-16T10:00:00Z"), mustTime("2026-08-16T10:10:00Z"), r)
	if bd.TotalFee != 5 {
		t.Fatalf("无免费规则10分钟应收5，got %.2f", bd.TotalFee)
	}
}

func TestRound2(t *testing.T) {
	got := round2(10.005)
	want := 10.01
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("round2(10.005)=%.4f want %.2f", got, want)
	}
}
