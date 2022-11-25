package screenshot

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type Connect struct {
	wsURL string
	// 存储运行中上下文的cancel函数，在程序中断或退出时关闭浏览器，防止内存泄漏
	cancelMap  map[context.Context]context.CancelFunc
	cancelLock sync.Mutex
}

// NewConnect 创建一个连接对象
func NewConnect(wsURL string) (*Connect, error) {
	if len(wsURL) == 0 {
		return nil, errors.New("devtools WsURL can not be empty")
	}

	// 检测连接是否可用
	resp, err := http.Get(ws2http(wsURL))
	if err != nil {
		return nil, errors.Wrap(err, "ws url not available")
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("ws url not available, status code: %d", resp.StatusCode)
	}
	return &Connect{wsURL: wsURL, cancelMap: map[context.Context]context.CancelFunc{}}, nil
}

// CancelAll 用于在程序退出的时候，取消所有活动的上下文
func (c *Connect) CancelAll() bool {
	c.cancelLock.Lock()
	defer c.cancelLock.Unlock()
	for _, v := range c.cancelMap {
		v()
	}
	return len(c.cancelMap) != 0
}

func (c *Connect) warpContext(ctx context.Context, cancel context.CancelFunc) (context.Context, context.CancelFunc) {
	// 加锁防止panic
	c.cancelLock.Lock()
	defer c.cancelLock.Unlock()
	c.cancelMap[ctx] = cancel
	return ctx, cancel
}

func (c *Connect) warpCancel(ctx context.Context, cancel context.CancelFunc) {
	c.cancelLock.Lock()
	defer c.cancelLock.Unlock()
	delete(c.cancelMap, ctx)
	cancel()
}

// Screenshot 全屏截图，返回图片数据
func (c *Connect) Screenshot(httpCtx context.Context, opts *Options) ([]byte, error) {
	if err := opts.check(); err != nil {
		return nil, errors.Wrap(err, "check options")
	}

	timeout, cancel := c.warpContext(context.WithTimeout(httpCtx, opts.Timeout))
	defer c.warpCancel(timeout, cancel)
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
func (c *Connect) ScreenshotToPath(opts *Options) error {
	if len(opts.Path) == 0 {
		return errors.New("path can not be empty ")
	}
	data, err := c.Screenshot(context.Background(), opts)
	if err != nil {
		return err
	}
	if err := os.WriteFile(opts.Path, data, 0o644); err != nil {
		return errors.Wrapf(err, "write to %s", opts.Path)
	}
	return nil
}

func fullScreenshot(opts *Options, res *[]byte) chromedp.Tasks {
	logger := logWithFields(logx.LogField{Key: "options", Value: fmt.Sprintf("%+v", opts)})
	return chromedp.Tasks{
		network.Enable(),
		runtime.Enable(),
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
			)
		}),
		//chromedp.ActionFunc(func(ctx context.Context) error {
		//	evaluate := func(expression string) error {
		//		// We wait until the evaluation of the expression is true or
		//		// until the context is done.
		//		logger.Debug(fmt.Sprintf("wait until '%s' is true before print", expression))
		//
		//		ticker := time.NewTicker(time.Duration(100) * time.Millisecond)
		//
		//		for {
		//			select {
		//			case <-ctx.Done():
		//				ticker.Stop()
		//
		//				return fmt.Errorf("context done while evaluating '%s': %w", expression, ctx.Err())
		//			case <-ticker.C:
		//				var ok bool
		//
		//				evaluate := chromedp.Evaluate(expression, &ok)
		//				err := evaluate.Do(ctx)
		//
		//				if err != nil {
		//					return err
		//				}
		//
		//				if ok {
		//					ticker.Stop()
		//
		//					return nil
		//				}
		//
		//				continue
		//			}
		//		}
		//	}
		//	return evaluate("window.MAPBOX_OBLOAD == true")
		//}),
		//chromedp.ActionFunc(func(ctx context.Context) error {
		//	const resize = `let event = document.createEvent("HTMLEvents");
		//event.initEvent("resize", true, true);
		//window.dispatchEvent(event);`
		//	ev := chromedp.Evaluate(resize, nil)
		//	return ev.Do(ctx)
		//}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if opts.WaitDelay == 0 {
				return nil
			}
			<-time.After(opts.WaitDelay)
			return nil
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

func ws2http(ws string) string {
	return strings.Replace(ws, "ws", "http", 1)
}
