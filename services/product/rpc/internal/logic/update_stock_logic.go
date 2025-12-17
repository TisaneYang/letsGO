package logic

import (
	"context"

	"letsgo/services/product/rpc/internal/svc"
	"letsgo/services/product/rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateStockLogic {
	return &UpdateStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Update product stock (called by order service)
func (l *UpdateStockLogic) UpdateStock(in *product.UpdateStockRequest) (*product.UpdateStockResponse, error) {
	// todo: add your logic here and delete this line

	return &product.UpdateStockResponse{}, nil
}
