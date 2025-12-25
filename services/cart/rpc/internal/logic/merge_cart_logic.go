package logic

import (
	"context"

	"letsgo/services/cart/rpc/cart"
	"letsgo/services/cart/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type MergeCartLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMergeCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MergeCartLogic {
	return &MergeCartLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Merge temporary cart (before login) with user cart (after login)
func (l *MergeCartLogic) MergeCart(in *cart.MergeCartRequest) (*cart.MergeCartResponse, error) {
	// TODO: Implement merge cart logic
	// This feature is for merging guest cart with user cart after login
	// For now, return success (can be implemented later if needed)

	l.Logger.Infof("MergeCart called (not implemented): user_id=%d, temp_cart_id=%s", in.UserId, in.TempCartId)

	return &cart.MergeCartResponse{
		Success: true,
	}, nil
}
