APP_NAME := initra
ENV ?= local
NAME ?= change_me

.PHONY: help deps run dev test lint fmt vet up down migrate-new migrate-apply migrate-status migrate-diff jet new-module build docker-build

help:
	@echo "Available targets:"
	@echo "  make deps                     - 下载并整理 Go 依赖"
	@echo "  make run                      - 以当前环境直接运行服务"
	@echo "  make dev                      - 使用 APP_ENV=local 启动开发服务"
	@echo "  make test                     - 运行全部测试"
	@echo "  make lint                     - 运行 golangci-lint"
	@echo "  make fmt                      - 格式化全部 Go 代码"
	@echo "  make vet                      - 运行 go vet"
	@echo "  make up                       - 启动 PostgreSQL 与 Redis"
	@echo "  make down                     - 停止本地依赖"
	@echo "  make migrate-new NAME=xxx     - 创建 migration 模板"
	@echo "  make migrate-apply ENV=local  - 执行 migration"
	@echo "  make migrate-status ENV=local - 查看 migration 状态"
	@echo "  make migrate-diff NAME=xxx    - 生成 Atlas diff migration"
	@echo "  make jet                      - 重新生成 Jet 代码"
	@echo "  make new-module NAME=user     - 生成业务模块骨架"
	@echo "  make build                    - 构建服务二进制"
	@echo "  make docker-build             - 构建 Docker 镜像"

deps:
	go mod tidy

run:
	go run ./cmd/server

dev:
	APP_ENV=local go run ./cmd/server

test:
	go test ./...

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

up:
	docker compose up -d postgres redis

down:
	docker compose down

migrate-new:
	sh ./scripts/new_migration.sh "$(NAME)"

migrate-apply:
	sh ./scripts/atlas.sh migrate apply --env "$(ENV)"

migrate-status:
	sh ./scripts/atlas.sh migrate status --env "$(ENV)"

migrate-diff:
	sh ./scripts/atlas.sh migrate diff "$(NAME)" --env "$(ENV)"

jet:
	sh ./scripts/jet.sh

new-module:
	sh ./scripts/new_module.sh "$(NAME)"

build:
	go build -o ./bin/$(APP_NAME) ./cmd/server

docker-build:
	docker build -t $(APP_NAME):latest .
