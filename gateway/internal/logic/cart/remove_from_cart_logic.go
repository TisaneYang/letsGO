package cart

import (
	"context"

	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"
	"letsgo/services/cart/rpc/cart"

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
	// Get user ID from context
	userId := l.ctx.Value("userId").(int64)

	// Call Cart RPC service
	_, err = l.svcCtx.CartRpc.RemoveCartItem(l.ctx, &cart.RemoveCartItemRequest{
		UserId:    userId,
		ProductId: req.ProductId,
	})
	if err != nil {
		return nil, err
	}

	return &types.RemoveCartResp{
		Success: true,
	}, nil
}

