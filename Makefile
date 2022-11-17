dockerfile: ## 生成默认的Dockerfile
	@echo "生成Dockerfile"
	@goctl docker --go internal/screenshot.go --version 1.19.3 --port 8888 --base alpine:3.17

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

V?=v1.0.0
.PHONY: build
build: ## 构建Docker镜像
	@echo "building cdp-screenshot..."
	@docker build -t cdp-screenshot:${V} .

CV?=107.0.5304.107
.PHONY: build.chromedp
build.chromedp: ## 构建chromedp镜像
	@echo "building chromedp..."
	@docker build --build-arg SHELL_TAG=${CV} -f Dockerfile.chromedp -t chromedp/headless-shell:${CV}_CN .

.PHONY: build.all
build.all:build build.chromedp ## 构建全部的Docker镜像

help:
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {\
	printf "\033[36m%-30s\033[0m %s\n", $$1, $$NF \
	}' $(MAKEFILE_LIST)
.DEFAULT_GOAL=help
.PHONY=help
