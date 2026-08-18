# car-manage-web 项目说明

## 项目用途

`car-manage-web` 是一个基于 Go 的停车场管理 Web 应用，用于管理停车场、车位和车辆，并完成车辆入场登记、出场结算、费用规则配置与停车数据展示。

项目主要功能包括：

- 停车场和车位的增删改查，以及空闲、占用、维护中等车位状态管理；
- 车辆信息登记与维护；
- 全局和停车场专属费用规则配置；
- 车辆入场、出场和停车费用计算；
- 停车记录、当前在场车辆和停车场运营数据展示；
- 使用 SQLite 保存业务数据，服务端渲染页面并通过内嵌资源提供前端内容。

## 环境要求

- Go 1.26.5 或兼容的较新 Go 版本；
- Docker（仅在使用评测镜像构建或容器运行时需要）。

项目依赖已记录在 `go.mod` 和 `go.sum` 中，正常构建前无需额外安装项目依赖。

## 标准构建、运行和测试命令

在项目根目录执行以下命令。

### 构建

```bash
go build -o carmanageweb ./cmd/carmanageweb
```

### 运行

直接运行源码并写入演示数据：

```bash
go run ./cmd/carmanageweb -addr :8080 -db parking.db -seed
```

或运行已构建的程序：

```bash
./carmanageweb -addr :8080 -db parking.db -seed
```

启动后访问：`http://localhost:8080`

不需要演示数据时，去掉 `-seed` 参数即可。`-db` 指定 SQLite 数据库文件路径，数据库不存在时会自动创建。

### 测试

```bash
go test ./...
```

### 格式化和静态检查

```bash
gofmt -w ./cmd ./internal
go vet ./...
```

## Docker 评测镜像

使用 `benzhi.Dockerfile` 和 `build_benzhi_docker.sh` 构建评测镜像：

```bash
./build_benzhi_docker.sh
```

也可以指定镜像名和目标平台：

```bash
./build_benzhi_docker.sh my-project linux/amd64
```

构建完成后进入容器：

```bash
docker run -it my-project:latest
```
