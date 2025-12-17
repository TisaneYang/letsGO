package logic

import (
	"context"

	"letsgo/services/product/rpc/internal/svc"
	"letsgo/services/product/rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductLogic {
	return &GetProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Get product detail
func (l *GetProductLogic) GetProduct(in *product.GetProductRequest) (*product.GetProductResponse, error) {
	// todo: add your logic here and delete this line

	return &product.GetProductResponse{}, nil
}
