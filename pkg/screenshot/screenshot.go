package screenshot

import (
	"context"
	"fmt"
	"os"

	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type Connect struct {
	wsURL string
}

// NewConnect 创建一个连接对象
func NewConnect(wsURL string) (*Connect, error) {
	if len(wsURL) == 0 {
		return nil, errors.New("devtools WsURL can not be empty")
	}
	// TODO: 检测连接是否可用
	return &Connect{wsURL: wsURL}, nil
}

// Screenshot 全屏截图，返回图片数据
func (c *Connect) Screenshot(opts Options) ([]byte, error) {
	if err := opts.check(); err != nil {
		return nil, errors.Wrap(err, "check options")
	}
	timeout, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()
	// 创建一个远程浏览器
	remoteCtx, cancel := chromedp.NewRemoteAllocator(timeout, c.wsURL)
	defer cancel()
	// 创建一个新标签页
	ctx, cancel := chromedp.NewContext(remoteCtx)
	defer cancel()

	var buf []byte
	return buf, chromedp.Run(ctx, fullScreenshot(opts, &buf))
}

// ScreenshotToPath 全屏截图，将图片保存至指定路径
func (c *Connect) ScreenshotToPath(opts Options) error {
	if len(opts.Path) == 0 {
		return errors.New("path can not be empty ")
	}
	data, err := c.Screenshot(opts)
	if err != nil {
		return err
	}
	if err := os.WriteFile(opts.Path, data, 0o644); err != nil {
		return errors.Wrapf(err, "write to %s", opts.Path)
	}
	return nil
}

func fullScreenshot(opts Options, res *[]byte) chromedp.Tasks {
	logger := logWithFields(logx.LogField{Key: "options", Value: fmt.Sprintf("%+v", opts)})
	return chromedp.Tasks{
		//network.Enable(),
		//runtime.Enable(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			logger.Infof("screenshot start")
			return nil
		}),
		// 设置窗口大小及缩放，缩放会影响分辨率
		chromedp.EmulateViewport(opts.ViewportWidth, opts.ViewportHeight, chromedp.EmulateScale(1+opts.Clarity)),
		chromedp.Navigate(opts.URL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 等待网页加载完成
			logger.Info("wait for events")
			return runBatch(ctx,
				waitForEventNetworkIdle(ctx, logger),
				waitForEventLoadingFinished(ctx, logger),
			)
		}),
		chromedp.FullScreenshot(res, opts.Quality),
		chromedp.ActionFunc(func(ctx context.Context) error {
			logger.Info("screenshot end")
			return nil
		}),
	}
}

func logWithFields(fields ...logx.LogField) logx.Logger {
	ctx := logx.ContextWithFields(context.Background(), fields...)
	return logx.WithContext(ctx)
}
