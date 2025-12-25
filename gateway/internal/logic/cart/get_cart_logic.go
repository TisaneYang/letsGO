package cart

import (
	"context"

	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"
	"letsgo/services/cart/rpc/cart"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get cart - Retrieve user's shopping cart
func NewGetCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCartLogic {
	return &GetCartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCartLogic) GetCart() (resp *types.CartResp, err error) {
	// Get user ID from context
	userId := l.ctx.Value("userId").(int64)

	// Call Cart RPC service
	cartResp, err := l.svcCtx.CartRpc.GetCart(l.ctx, &cart.GetCartRequest{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	// Convert RPC response to API response
	items := make([]types.CartItem, 0, len(cartResp.Items))
	for _, item := range cartResp.Items {
		items = append(items, types.CartItem{
			ProductId: item.ProductId,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  item.Quantity,
			Image:     item.Image,
			Stock:     item.Stock,
			Available: item.Available,
		})
	}

	return &types.CartResp{
		Items:      items,
		TotalPrice: cartResp.TotalPrice,
		TotalCount: cartResp.TotalCount,
	}, nil
}

