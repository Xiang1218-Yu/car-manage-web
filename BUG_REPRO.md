# 缺陷复现说明

## 缺陷现象
运营把某停车场的费用规则调价后，一些已经入场的车出场时也按新价收费，客服无法复现入场时展示的规则。请修复这类在场订单的结算，让入场后的规则编辑不会改写它应付的价格。

## 触发方式
# 1. 进入停车结算的测试项目；预期：后续命令在待修复环境中执行。
cd /workplace/car-manage-web__003

# 2. 反复验证规则更新后的在场订单；预期：缺陷基线每次将 10 元入场规则错误结算为更新后的 80 元。
go test ./internal/store -run TestCheckOutKeepsFeeRuleAtCheckIn -count=20

# 3. 检查未受影响的费用和存储逻辑；预期：修复后完整测试套件保持通过。
go test ./...

## 触发后的实际错误输出

```text
=== RUN   TestCheckOutKeepsFeeRuleAtCheckIn
    flow_test.go:238: 已入场记录应按入场时 10 元规则结算，got 80.00（明细 &{RuleName:调整后价格 DurationMinutes:20 IsFree:false FirstBlockFee:80 ExtraUnits:0 ExtraBlockFee:0 TotalFee:80 DailyCapApplied:false}）
--- FAIL: TestCheckOutKeepsFeeRuleAtCheckIn (0.00s)
FAIL
FAIL	carmanageweb/internal/store	0.602s
FAIL
```
