package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"letsgo/common/errorx"
	"letsgo/services/product/rpc/internal/svc"
	"letsgo/services/product/rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductsLogic {
	return &ListProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// List products with pagination
func (l *ListProductsLogic) ListProducts(in *product.ListProductsRequest) (*product.ListProductsResponse, error) {
	// 1. Validate pagination parameters
	page := in.Page
	pageSize := in.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // Max 100 items per page
	}

	// 2. Try to get list from cache first
	categoryVersion := GetCategoryVersion(l.ctx, in.Category, &l.svcCtx.Redis)

	cacheKey := fmt.Sprintf("product:list:v%d:%d:%d:%s:%s:%s", categoryVersion, in.Page, in.PageSize, in.Category, in.SortBy, in.Order)
	cacheData, err := l.svcCtx.Redis.GetCtx(l.ctx, cacheKey)
	if err == nil && cacheData != "" {
		var cachedResponse product.ListProductsResponse
		if err := json.Unmarshal([]byte(cacheData), &cachedResponse); err == nil {
			l.Logger.Infof("Product list retrieved from cache: Total = %d", cachedResponse.Total)
			return &cachedResponse, nil
		}
	}

	// 3. Query products from database
	products, total, err := l.svcCtx.ProductModel.List(l.ctx, page, pageSize, in.Category, in.SortBy, in.Order)
	if err != nil {
		l.Logger.Errorf("Failed to list products: %v", err)
		return nil, errorx.ErrDatabase
	}

	// 4. Build response
	productList := make([]*product.ProductInfo, 0, len(products))
	for _, p := range products {
		productList = append(productList, &product.ProductInfo{
			Id:          p.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
			Category:    p.Category,
			Images:      p.Images,
			Attributes:  p.Attributes,
			Sales:       p.Sales,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
	}

	result := &product.ListProductsResponse{
		Total:    total,
		Products: productList,
	}
	JsonResult, err := json.Marshal(&result)
	if err == nil {
		err := l.svcCtx.Redis.SetexCtx(l.ctx, cacheKey, string(JsonResult), l.svcCtx.Config.Cache.ListExpire)
		if err != nil {
			l.Logger.Errorf("Failed to cache product list info: %v", err)
		}
	}

	return result, nil
}
