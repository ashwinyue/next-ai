.PHONY: help build run test clean docker-build docker-build-all docker-up docker-down dev-up dev-down dev-logs fmt lint

# Next-AI Makefile

# Go 相关变量
BINARY_NAME=next-ai
MAIN_PATH=./cmd/next-ai
DOCKER_IMAGE=next-ai-app
DOCKER_TAG=latest

# 平台检测
ifeq ($(shell uname -m),x86_64)
    PLATFORM=linux/amd64
else ifeq ($(shell uname -m),aarch64)
    PLATFORM=linux/arm64
else ifeq ($(shell uname -m),arm64)
    PLATFORM=linux/arm64
else
    PLATFORM=linux/amd64
endif

help: ## 显示帮助信息
	@echo "Next-AI Makefile"
	@echo ""
	@echo "基础命令:"
	@echo "  make build           构建应用"
	@echo "  make run             运行应用"
	@echo "  make test            运行测试"
	@echo "  make clean           清理构建文件"
	@echo ""
	@echo "Docker 命令:"
	@echo "  make docker-build    构建 Docker 镜像"
	@echo "  make docker-build-all 构建所有 Docker 镜像 (含前端)"
	@echo "  make docker-up       启动所有服务"
	@echo "  make docker-down     停止所有服务"
	@echo "  make docker-logs     查看服务日志"
	@echo ""
	@echo "开发命令:"
	@echo "  make dev-up          启动开发环境 (仅基础设施)"
	@echo "  make dev-down        停止开发环境"
	@echo "  make dev-logs        查看开发环境日志"
	@echo ""
	@echo "代码质量:"
	@echo "  make fmt             格式化代码"
	@echo "  make lint            代码检查"
	@echo "  make deps            更新依赖"

build: ## 构建应用
	@echo "构建 $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

run: ## 运行应用
	@echo "运行 $(BINARY_NAME)..."
	go run $(MAIN_PATH)

test: ## 运行测试
	go test -v ./...

clean: ## 清理构建文件
	go clean
	rm -rf bin/

fmt: ## 格式化代码
	go fmt ./...
	goimports -w .

lint: ## 代码检查
	golangci-lint run ./...

deps: ## 更新依赖
	go mod tidy
	go mod download

# Docker 命令
docker-build: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像..."
	docker build -f docker/Dockerfile -t $(DOCKER_IMAGE):$(DOCKER_TAG) --platform $(PLATFORM) .

docker-build-frontend: ## 构建前端 Docker 镜像
	@echo "构建前端 Docker 镜像..."
	docker build -t next-ai-ui:$(DOCKER_TAG) ./frontend

docker-build-all: docker-build docker-build-frontend ## 构建所有 Docker 镜像

docker-up: ## 启动所有服务
	docker-compose up -d

docker-down: ## 停止所有服务
	docker-compose down

docker-logs: ## 查看服务日志
	docker-compose logs -f

docker-restart: ## 重启所有服务
	docker-compose restart

# 开发命令
dev-up: ## 启动开发环境 (仅基础设施)
	docker-compose up -d postgres redis elasticsearch

dev-down: ## 停止开发环境
	docker-compose down

dev-logs: ## 查看开发环境日志
	docker-compose logs -f postgres redis elasticsearch

dev-ps: ## 查看容器状态
	docker-compose ps

# 构建前端
frontend-build: ## 构建前端
	cd frontend && npm install && npm run build

frontend-dev: ## 运行前端开发服务器
	cd frontend && npm run dev
