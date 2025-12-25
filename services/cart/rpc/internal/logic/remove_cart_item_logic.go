package logic

import (
	"context"
	"fmt"

	"letsgo/common/errorx"
	"letsgo/services/cart/rpc/cart"
	"letsgo/services/cart/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveCartItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveCartItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveCartItemLogic {
	return &RemoveCartItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Remove item from cart
func (l *RemoveCartItemLogic) RemoveCartItem(in *cart.RemoveCartItemRequest) (*cart.RemoveCartItemResponse, error) {
	// 1. Validate input
	if in.UserId <= 0 {
		return nil, errorx.NewCodeError(1001, "Invalid user ID")
	}
	if in.ProductId <= 0 {
		return nil, errorx.NewCodeError(1001, "Invalid product ID")
	}

	// 2. Remove item from Redis (HDEL is atomic)
	cartKey := fmt.Sprintf("cart:user:%d", in.UserId)
	productField := fmt.Sprintf("product:%d", in.ProductId)

	deleted, err := l.svcCtx.Redis.HdelCtx(l.ctx, cartKey, productField)
	if err != nil {
		l.Logger.Errorf("Failed to remove cart item: user_id=%d, product_id=%d, err=%v", in.UserId, in.ProductId, err)
		return nil, errorx.ErrCache
	}

	if !deleted {
		return nil, errorx.ErrCartItemNotFound
	}

	// 3. Refresh cart expiration time
	l.svcCtx.Redis.ExpireCtx(l.ctx, cartKey, l.svcCtx.Config.Cart.Expire)

	l.Logger.Infof("Removed cart item successfully: user_id=%d, product_id=%d", in.UserId, in.ProductId)

	return &cart.RemoveCartItemResponse{
		Success: true,
	}, nil
}
