package logic

import (
	"context"

	"letsgo/services/product/rpc/internal/svc"
	"letsgo/services/product/rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckStockLogic {
	return &CheckStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Check if products are in stock (batch check)
func (l *CheckStockLogic) CheckStock(in *product.CheckStockRequest) (*product.CheckStockResponse, error) {
	// todo: add your logic here and delete this line

	return &product.CheckStockResponse{}, nil
}
