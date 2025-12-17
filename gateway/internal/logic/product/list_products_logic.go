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

type ListProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List products - Get paginated list of products
func NewListProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductsLogic {
	return &ListProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListProductsLogic) ListProducts(req *types.ProductListReq) (resp *types.ProductListResp, err error) {
	ProductResp, err := l.svcCtx.ProductRpc.ListProducts(l.ctx, &product_client.ListProductsRequest{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		Category: req.Category,
		SortBy:   req.SortBy,
		Order:    req.Order,
	})
	if err != nil {
		return nil, err
	}

	Products := make([]types.Product, 0, len(ProductResp.Products))
	for _, productInfo := range ProductResp.Products {
		newProduct := types.Product{
			Id:          productInfo.Id,
			Name:        productInfo.Name,
			Description: productInfo.Description,
			Price:       productInfo.Price,
			Stock:       productInfo.Stock,
			Category:    productInfo.Category,
			Images:      productInfo.Images,
			Attributes:  productInfo.Attributes,
			Sales:       productInfo.Sales,
			CreatedAt:   productInfo.CreatedAt,
			UpdatedAt:   productInfo.UpdatedAt,
		}
		Products = append(Products, newProduct)
	}

	return &types.ProductListResp{
		Total:    ProductResp.Total,
		Products: Products,
	}, nil
}
