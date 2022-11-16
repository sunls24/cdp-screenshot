package screenshot

import (
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type Options struct {
	URL string
	// Clarity 截图清晰度<0-1>，主要影响图片分辨率，默认为0.8，清晰度越高截图文件越大
	Clarity float64
	// Quality 截图质量 <1-100>
	Quality int
	// Viewport 截图窗口大小
	ViewportWidth, ViewportHeight int64

	Timeout time.Duration
	Path    string
}

func (o Options) check() error {
	if len(o.URL) == 0 {
		return errors.New("URL can not be empty")
	}
	if o.Clarity < 0 || o.Clarity > 1 {
		logx.Slowf("clarity option is %0.2f, it can only be 0 to 1, it will be set as the default value %0.2f", o.Clarity, defaultClarity)
		o.Clarity = defaultClarity
	}
	return nil
}

func DefaultOptions() Options {
	return Options{
		Clarity:        defaultClarity,
		Quality:        defaultQuality,
		Timeout:        defaultTimeout,
		ViewportWidth:  defaultViewportWidth,
		ViewportHeight: defaultViewportHeight,
	}
}

func (o Options) WithURL(url string) Options {
	o.URL = url
	return o
}

func (o Options) WithClarity(clarity float64) Options {
	o.Clarity = clarity
	return o
}

func (o Options) WithViewport(width, height int64) Options {
	o.ViewportWidth = width
	o.ViewportHeight = height
	return o
}

func (o Options) WithQuality(quality int) Options {
	o.Quality = quality
	return o
}

func (o Options) WithTimeout(timeout time.Duration) Options {
	o.Timeout = timeout
	return o
}

func (o Options) WithPath(path string) Options {
	o.Path = path
	return o
}
