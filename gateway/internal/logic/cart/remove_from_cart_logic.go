// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package cart

import (
	"context"

	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveFromCartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Remove from cart - Delete item from cart
func NewRemoveFromCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveFromCartLogic {
	return &RemoveFromCartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveFromCartLogic) RemoveFromCart(req *types.RemoveCartReq) (resp *types.RemoveCartResp, err error) {
	// todo: add your logic here and delete this line

	return
}
