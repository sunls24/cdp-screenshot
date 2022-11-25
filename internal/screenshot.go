package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cdp-screenshot/internal/internal/config"
	"cdp-screenshot/internal/internal/handler"
	"cdp-screenshot/internal/internal/svc"
	"cdp-screenshot/pkg/screenshot"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/screenshot.yaml", "the config file")

func init() {
	_ = logx.SetUp(logx.LogConf{
		Encoding: "plain",
		Level:    "debug",
	})
	logx.DisableStat()
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.MustSetup(c.Log)

	wsURL := c.WsURL
	if v := os.Getenv("WsURL"); len(v) != 0 {
		wsURL = v
	}
	connect, err := screenshot.NewConnect(wsURL)
	if err != nil {
		log.Fatalf("unable to connect chromedp, please check WsURL: %s", c.WsURL)
	}

	//监听指定信号 ctrl+c kill
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-kill
		// 收到中断信号，将浏览器关闭防止内存泄漏
		if connect.CancelAll() {
			logx.Debug("wait cancel all...")
			<-time.After(time.Second)
		}
		os.Exit(1)
	}()

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()

	ctx := svc.NewServiceContext(c, connect)
	handler.RegisterHandlers(server, ctx)

	logx.Infof("Starting server at %s:%d...", c.Host, c.Port)
	server.Start()
}
