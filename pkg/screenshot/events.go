package screenshot

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/errgroup"
)

func waitEventIdle(cancel func(), logger logx.Logger) func() {
	const idleTime = 10 * time.Second
	timer := time.NewTimer(idleTime)
	go func() {
		<-timer.C
		cancel()
		logger.Debugf("There is no event triggering for %v, and now it is judged that the loading is complete", idleTime)
	}()
	return func() { timer.Reset(idleTime) }
}

// waitForEventNetworkIdle waits until the event networkIdle is fired or the
// context timeout.
func waitForEventNetworkIdle(ctx context.Context, logger logx.Logger) func() error {
	// 此处除了监听networkIdle事件以外还会判断无事件触发的持续时间，当大于10s时判定界面加载成功
	return func() error {
		ch := make(chan struct{})
		cctx, cancel := context.WithCancel(ctx)
		idleReset := waitEventIdle(func() {
			select {
			case <-ch: // 通道已经关闭
			default:
				cancel()
				close(ch)
			}
		}, logger)
		chromedp.ListenTarget(cctx, func(ev interface{}) {
			idleReset()
			switch e := ev.(type) {
			case *page.EventLifecycleEvent:
				if e.Name == "networkIdle" {
					cancel()
					close(ch)
				}
			}
		})

		select {
		case <-ch:
			logger.Debug("event networkIdle fired")
			return nil
		case <-ctx.Done():
			return fmt.Errorf("wait for event networkIdle: %w", ctx.Err())
		}
	}
}

// waitForEventLoadingFinished waits until the event LoadingFinished is fired
// or the context timeout.
func waitForEventLoadingFinished(ctx context.Context, logger logx.Logger) func() error {
	return func() error {
		ch := make(chan struct{})
		cctx, cancel := context.WithCancel(ctx)
		chromedp.ListenTarget(cctx, func(ev interface{}) {
			switch ev.(type) {
			case *network.EventLoadingFinished:
				cancel()
				close(ch)
			}
		})

		select {
		case <-ch:
			logger.Debug("event LoadingFinished fired")
			return nil
		case <-ctx.Done():
			return fmt.Errorf("wait for event LoadingFinished: %w", ctx.Err())
		}
	}
}

// runBatch runs all functions simultaneously and waits until all of them are
// completed or an error is encountered.
func runBatch(ctx context.Context, fn ...func() error) error {
	eg, _ := errgroup.WithContext(ctx)
	for _, f := range fn {
		eg.Go(f)
	}
	return eg.Wait()
}
