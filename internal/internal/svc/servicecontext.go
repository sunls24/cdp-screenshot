package svc

import (
	"cdp-screenshot/internal/internal/config"
	"cdp-screenshot/pkg/screenshot"
)

type ServiceContext struct {
	Config  config.Config
	Connect *screenshot.Connect
}

func NewServiceContext(c config.Config, connect *screenshot.Connect) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		Connect: connect,
	}
}
