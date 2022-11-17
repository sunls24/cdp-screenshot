NowTime = $(shell date "+%Y-%m-%d-%H:%M:%S")

# 判断系统和架构，目前只支持macOS/linux
ifeq ($(shell uname),Darwin)
  OS=darwin
else
  OS=linux
endif
ifeq ($(shell uname -m),x86_64)
  ARCH=amd64
else
  ARCH=arm64
endif

dockerfile: ## 生成默认的Dockerfile
	@echo "生成Dockerfile"
	@goctl docker --go internal/screenshot.go

.PHONY: genapi
genapi: ## 生成API和Swagger文件
	@echo "[1] 生成goctl文件"
	@goctl api go -api internal/api/screenshot.api -dir internal/
	@echo "[2] 生成Swagger文件"
	@goctl api plugin -plugin goctl-swagger="swagger -filename swagger.json -host 127.0.0.1:8888 -basepath /" -api internal/api/screenshot.api -dir . >/dev/null 2>&1
	@echo "✅ 完成"

.PHONY: run
run:_clean_run ## 启动服务
	@echo "[1]启动Swagger服务"
	@swagger serve -F=swagger swagger.json --port 8889 --host 0.0.0.0 --no-open > swagger.log 2>&1 &
	@echo "✅ http://localhost:8889/docs"
	@echo "[2]启动Screenshot服务"
	@go run internal/screenshot.go -f internal/etc/screenshot.yaml

_clean_run:
	@echo "cleaning up..."
	@if [ -f swagger.log ]; then rm -f *.log; fi
	@lsof -i:8889 | grep swagger | awk '{print $$2}' | xargs kill -9

clean:_clean_run ## 清理文件

.PHONY: build
build: ## 构建Docker镜像
	@echo "building..."
	@docker build -t cdp-screenshot:v0.1 .

help:
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {\
	printf "\033[36m%-30s\033[0m %s\n", $$1, $$NF \
	}' $(MAKEFILE_LIST)
.DEFAULT_GOAL=help
.PHONY=help
