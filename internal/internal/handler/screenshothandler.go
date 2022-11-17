package handler

import (
	"net/http"

	"cdp-screenshot/internal/internal/logic"
	"cdp-screenshot/internal/internal/svc"
	"cdp-screenshot/internal/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ScreenshotHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.Request
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewScreenshotLogic(r.Context(), svcCtx, r, w)
		err := l.Screenshot(&req)
		if err != nil {
			httpx.Error(w, err)
		}
	}
}
