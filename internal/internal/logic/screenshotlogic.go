package logic

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cdp-screenshot/internal/internal/svc"
	"cdp-screenshot/internal/internal/types"
	"cdp-screenshot/pkg/screenshot"

	"github.com/zeromicro/go-zero/core/logx"
)

type ScreenshotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext

	r *http.Request
	w http.ResponseWriter
}

func NewScreenshotLogic(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request, w http.ResponseWriter) *ScreenshotLogic {
	return &ScreenshotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
		w:      w,
	}
}

func (l *ScreenshotLogic) Screenshot(req *types.Request) error {
	opts := getOptions(req)
	data, err := l.svcCtx.Connect.Screenshot(opts)
	if err != nil {
		return err
	}
	filename, contentType := getFilename(opts)
	logx.Debugf("detected file name as %s", filename)
	l.w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	l.w.Header().Set("Content-Type", contentType)
	_, err = l.w.Write(data)
	return err
}

func getOptions(req *types.Request) *screenshot.Options {
	opts := screenshot.DefaultOptions().WithURL(req.URL)
	if req.Clarity != 0 {
		opts.WithClarity(float64(req.Clarity) / 10)
	}
	if req.Quality != 0 {
		opts.WithQuality(req.Quality)
	}
	if req.ViewportWidth != 0 && req.ViewportHeight != 0 {
		opts.WithViewport(int64(req.ViewportWidth), int64(req.ViewportHeight))
	}
	if req.Timeout != 0 {
		opts.WithTimeout(time.Second * time.Duration(req.Timeout))
	}
	return opts
}

// 查找URL的域名作为文件名，当quality=100时为png，其他为jpg
func getFilename(opts *screenshot.Options) (string, string) {
	var suffix, contentType = "png", "image/png"
	if opts.Quality != 100 {
		suffix, contentType = "jpg", "image/jpeg"
	}

	var filename = opts.URL
	if i := strings.Index(filename, "//"); i >= 0 {
		filename = filename[i+2:]
	}
	b := base64.StdEncoding.EncodeToString([]byte(filename))
	if i := strings.Index(filename, "/"); i >= 0 {
		filename = filename[:i]
	}

	if len(b) > 10 {
		b = b[len(b)-10 : len(b)-2]
	}
	return fmt.Sprintf("%s-%s.%s", filename, b, suffix), contentType
}
