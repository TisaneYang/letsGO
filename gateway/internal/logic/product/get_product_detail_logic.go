// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package product

import (
	"context"

	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"
	"letsgo/services/product/rpc/product_client"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get product detail - View single product information
func NewGetProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductDetailLogic {
	return &GetProductDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProductDetailLogic) GetProductDetail(req *types.ProductDetailReq) (resp *types.ProductDetailResp, err error) {
	ProductResp, err := l.svcCtx.ProductRpc.GetProduct(l.ctx, &product_client.GetProductRequest{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}

	ProductInfo := ProductResp.Product
	return &types.ProductDetailResp{
		Product: types.Product{
			Id:          ProductInfo.Id,
			Name:        ProductInfo.Name,
			Description: ProductInfo.Description,
			Price:       ProductInfo.Price,
			Stock:       ProductInfo.Stock,
			Category:    ProductInfo.Category,
			Images:      ProductInfo.Images,
			Attributes:  ProductInfo.Attributes,
			Sales:       ProductInfo.Sales,
			CreatedAt:   ProductInfo.CreatedAt,
			UpdatedAt:   ProductInfo.UpdatedAt,
		},
	}, nil
}
