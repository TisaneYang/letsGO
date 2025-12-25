package cart

import (
	"context"

	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"
	"letsgo/services/cart/rpc/cart"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearCartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Clear cart - Remove all items from cart
func NewClearCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearCartLogic {
	return &ClearCartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ClearCartLogic) ClearCart() (resp *types.ClearCartResp, err error) {
	// Get user ID from context
	userId := l.ctx.Value("userId").(int64)

	// Call Cart RPC service
	_, err = l.svcCtx.CartRpc.ClearCart(l.ctx, &cart.ClearCartRequest{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	return &types.ClearCartResp{
		Success: true,
	}, nil
}

