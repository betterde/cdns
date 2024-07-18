.DEFAULT_GOAL := help

BINARY_NAME ?= cdns
INSTALL_PATH ?= /usr/local/bin

VERSION ?= latest
MODULE=github.com/betterde/cdns
BUILD_AT=`date -u`
COMMIT_HASH=`git rev-parse HEAD`
BUILD_FLAG=-ldflags "-s -w -X '$(MODULE)/cmd.version=$(VERSION)' -X '$(MODULE)/cmd.build=$(BUILD_AT)' -X '$(MODULE)/cmd.commit=$(COMMIT_HASH)'"

export CGO_ENABLED = 0

all: test build

fmt: ## 格式化代码
	go fmt ./...

test: ## 运行竞态检测
	CGO_ENABLED=1 go test -race ./...

build: ## 打包生成二进制可执行文件
	go build $(BUILD_FLAG) -o $(BINARY_NAME) main.go

docker-build: ## 构建 Docker image
	docker build \
	--build-arg VERSION="$(VERSION)" \
	-t betterde/cdns:latest .

.PHONY: install
install: ## 将打包生成的二进制可执行文件安装到指定目录
	cp "$(BINARY_NAME)" "$(INSTALL_PATH)"
	chmod +x "$(BINARY_NAME)"

.PHONY: clean
clean: ## 删除测试生成的缓存文件
	go clean -testcache

.PHONY: help
help: ## 获取帮助信息
	@echo "Usage: \n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'