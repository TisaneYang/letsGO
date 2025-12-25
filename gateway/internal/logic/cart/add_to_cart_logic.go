package cart

import (
	"context"

	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"
	"letsgo/services/cart/rpc/cart"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddToCartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Add to cart - Add product to shopping cart
func NewAddToCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddToCartLogic {
	return &AddToCartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddToCartLogic) AddToCart(req *types.AddToCartReq) (resp *types.AddToCartResp, err error) {
	// Get user ID from context (set by Auth middleware)
	userId := l.ctx.Value("userId").(int64)

	// Call Cart RPC service
	_, err = l.svcCtx.CartRpc.AddToCart(l.ctx, &cart.AddToCartRequest{
		UserId:    userId,
		ProductId: req.ProductId,
		Quantity:  req.Quantity,
	})
	if err != nil {
		return nil, err
	}

	return &types.AddToCartResp{
		Success: true,
	}, nil
}

