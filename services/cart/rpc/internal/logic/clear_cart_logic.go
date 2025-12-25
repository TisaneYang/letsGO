package logic

import (
	"context"
	"fmt"

	"letsgo/common/errorx"
	"letsgo/services/cart/rpc/cart"
	"letsgo/services/cart/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearCartLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearCartLogic {
	return &ClearCartLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Clear entire cart
func (l *ClearCartLogic) ClearCart(in *cart.ClearCartRequest) (*cart.ClearCartResponse, error) {
	// 1. Validate input
	if in.UserId <= 0 {
		return nil, errorx.NewCodeError(1001, "Invalid user ID")
	}

	// 2. Delete entire cart from Redis (DEL is atomic)
	cartKey := fmt.Sprintf("cart:user:%d", in.UserId)

	deleted, err := l.svcCtx.Redis.DelCtx(l.ctx, cartKey)
	if err != nil {
		l.Logger.Errorf("Failed to clear cart: user_id=%d, err=%v", in.UserId, err)
		return nil, errorx.ErrCache
	}

	l.Logger.Infof("Cleared cart successfully: user_id=%d, deleted=%d", in.UserId, deleted)

	return &cart.ClearCartResponse{
		Success: true,
	}, nil
}
