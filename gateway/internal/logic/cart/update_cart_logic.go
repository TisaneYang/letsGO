package cart

import (
	"context"

	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"
	"letsgo/services/cart/rpc/cart"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update cart item - Change quantity of item in cart
func NewUpdateCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCartLogic {
	return &UpdateCartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateCartLogic) UpdateCart(req *types.UpdateCartReq) (resp *types.UpdateCartResp, err error) {
	// Get user ID from context
	userId := l.ctx.Value("userId").(int64)

	// Call Cart RPC service
	_, err = l.svcCtx.CartRpc.UpdateCartItem(l.ctx, &cart.UpdateCartItemRequest{
		UserId:    userId,
		ProductId: req.ProductId,
		Quantity:  req.Quantity,
	})
	if err != nil {
		return nil, err
	}

	return &types.UpdateCartResp{
		Success: true,
	}, nil
}

