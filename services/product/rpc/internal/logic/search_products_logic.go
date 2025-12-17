package logic

import (
	"context"

	"letsgo/services/product/rpc/internal/svc"
	"letsgo/services/product/rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchProductsLogic {
	return &SearchProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Search products by keyword
func (l *SearchProductsLogic) SearchProducts(in *product.SearchProductsRequest) (*product.SearchProductsResponse, error) {
	// todo: add your logic here and delete this line

	return &product.SearchProductsResponse{}, nil
}
