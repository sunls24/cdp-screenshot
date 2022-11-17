package main

import (
	"flag"
	"log"

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
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.MustSetup(c.Log)

	connect, err := screenshot.NewConnect(c.WsURL)
	if err != nil {
		log.Fatalf("unable to connect chromedp, please check WsURL: %s", c.WsURL)
	}

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c, connect)
	handler.RegisterHandlers(server, ctx)

	logx.Infof("Starting server at %s:%d...", c.Host, c.Port)
	server.Start()
}
