# 缺陷复现说明

## 缺陷现象
夜班巡检把仍有车辆停放的车位临时设成维修，车辆完成结算后，大屏却把它重新标成空闲，下一班调度会继续往里派车。文件先不要改，帮我查清这个状态为何会被覆盖，以及最小修正应在哪个转换环节守住维修标记。

## 触发方式
# 1. 进入车位状态验证项目；预期：后续命令在题目环境执行。
cd /workplace/car-manage-web__004

# 2. 连续触发维护车位的结算；预期：缺陷基线会把 maintenance 错误改成 available。
go test -v ./internal/store -run TestCheckOutPreservesMaintenanceStatus -count=20

# 3. 执行完整测试套件；预期：可确认该状态错误在套件中稳定暴露。
go test ./...

## 触发后的实际错误输出

```text
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
=== RUN   TestCheckOutPreservesMaintenanceStatus
    maintenance_checkout_test.go:43: 已被巡检标记维护的车位在出场后仍应保持 maintenance，got available
--- FAIL: TestCheckOutPreservesMaintenanceStatus (0.00s)
FAIL
FAIL	carmanageweb/internal/store	0.536s
FAIL
```
