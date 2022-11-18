# cdp-screenshot

#### 使用 chromedp 接口进行网页截图

## Run

[http://localhost:8889/docs](http://localhost:8889/docs) 查看接口文档

*配置文件中需指定 chromedp 地址*

```shell
make run
```

## Build
打包 Docker 镜像
```shell
# 打包 cdp-screenshot
make build
# 打包 chromedp
make build.chromedp
# 打包全部镜像
make build.all
```

## 接口调用
只有`url`是必填参数 
```shell
curl --request POST --url 'http://127.0.0.1:8888/screenshot?=' --header 'Content-Type: application/json' --data '{
        "url": "https://baidu.com"
}' -v -o baidu.png
```

## chromedp 镜像打包
> 由于官方镜像字体缺失，无法展示中文以及 emoji 表情，所以需要对其重新打包，安装对应字体。

```dockerfile
# 修正简体中文显示为异体（日文）字形
COPY font.conf /etc/fonts/conf.d/64-language-selector-prefer.conf
RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list \
    && apt-get update && apt-get install -y \
    fonts-noto-cjk \
    fonts-noto-color-emoji \
    && rm -rf /var/lib/apt/lists/* \
```

## 使用 docker compose

因为两个容器间需要通信，使用 docker-compose 进行连接

```shell
# 先打包镜像
make build.all
# 启动
docker-compose up -d
# 停止
docker-compose down
```