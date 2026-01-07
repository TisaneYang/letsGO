// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"letsgo/gateway/internal/logic/order"
	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"
)

// Query order by number - Search order by order number
func QueryOrderByNoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QueryOrderByNoReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := order.NewQueryOrderByNoLogic(r.Context(), svcCtx)
		resp, err := l.QueryOrderByNo(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
