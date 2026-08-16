# 缺陷复现说明

## 缺陷现象
费用预估接口偶尔会生成特别离谱的金额。请求里的 check_out 传错格式时接口没有报参数错误，反而返回成功结果，而且金额还会随着当前时间变化。文件先不要改，帮我查清楚是哪个处理分支把坏参数吞掉的，以及缺省出场时间和非法出场时间为什么会混为一谈。

## 触发方式
# 1. 进入费用预估接口的测试项目；预期：后续测试在题目环境运行。
cd /workplace/car-manage-web__002

# 2. 连续传入格式错误的 check_out；预期：缺陷基线每次错误地返回 200，定位结论能解释这一分支。
go test ./internal/server -run TestFeeEstimateRejectsMalformedCheckoutTime -count=20

# 3. 运行完整回归；预期：诊断不修改任何代码，既有功能保持可测。
go test ./...

## 触发后的实际错误输出

```text
=== RUN   TestFeeEstimateRejectsMalformedCheckoutTime
    fee_estimate_integrity_test.go:47: 无效 check_out 不应被当作当前时间继续计算，实际状态码 200，响应 {"rule_name":"预估测试规则","duration_minutes":394,"is_free":false,"first_block_fee":10,"extra_units":12,"extra_block_fee":60,"total_fee":70,"daily_cap_applied":false}
--- FAIL: TestFeeEstimateRejectsMalformedCheckoutTime (0.00s)
FAIL
FAIL	carmanageweb/internal/server	0.778s
FAIL
```
