# 缺陷复现说明

## 缺陷现象
入口岗亭刚反馈：同一辆车从两个闸口重复登记时，后台会留下两笔未结算的停车单，还会占住两个车位。请修复这条入场流程——重复登记要明确拒绝且不占用新车位，原订单完成出场结算后仍可正常再次入场。

## 触发方式
# 1. 进入车辆入场验证项目；预期：后续命令在待修复环境执行。
cd /workplace/car-manage-web__005

# 2. 连续登记同一辆仍在场的车辆；预期：缺陷基线错误接受第二次入场。
go test -v ./internal/store -run TestCheckInRejectsVehicleWithAnotherActiveRecord -count=20

# 3. 执行完整回归；预期：修复后入场、结算和车位状态流程保持通过。
go test ./...

## 触发后的实际错误输出

```text
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
=== RUN   TestCheckInRejectsVehicleWithAnotherActiveRecord
    duplicate_checkin_test.go:31: 同一车辆已有在场记录时，第二次入场应返回 ErrVehicleAlreadyParked，got <nil>
--- FAIL: TestCheckInRejectsVehicleWithAnotherActiveRecord (0.00s)
FAIL
FAIL	carmanageweb/internal/store	0.012s
FAIL
```
