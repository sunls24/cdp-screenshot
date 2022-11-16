package screenshot

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	testWsURL = "ws://127.0.0.1:9222/"
	testURL   = "https://baidu.com"
	testPath  = "screenshot.png"
)

func init() {
	_ = logx.SetUp(logx.LogConf{
		Encoding: "plain",
		Level:    "debug",
	})
}

func TestScreenshot_ScreenshotToPath(t *testing.T) {
	c, err := NewConnect(testWsURL)
	if err != nil {
		t.Error(errors.Wrap(err, "NewConnect"))
	}
	err = c.ScreenshotToPath(DefaultOptions().
		WithURL(testURL).
		WithPath(testPath))
	if err != nil {
		t.Error(err)
	}
}
