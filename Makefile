.PHONY: build run dev clean test

# 构建项目
build:
	go build -o bin/palu-wiki cmd/server/main.go

# 运行项目
run: build
	./bin/palu-wiki

# 开发模式运行
dev:
	go run cmd/server/main.go

# 清理构建文件
clean:
	rm -rf bin/

# 运行测试
test:
	go test -v ./...

# 安装依赖
deps:
	go mod tidy
	go mod download

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run

# 创建.env文件
env:
	cp .env.example .env

# 数据库迁移
migrate:
	go run cmd/server/main.go