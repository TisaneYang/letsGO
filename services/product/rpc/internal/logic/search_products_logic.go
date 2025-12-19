package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"letsgo/common/errorx"
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
	// 1. Validate search parameters
	if len(strings.TrimSpace(in.Keyword)) == 0 {
		return nil, errorx.NewCodeError(1001, "Search keyword cannot be empty")
	}

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

	// 2. Try to get response from cache first.
	globalVersion := GetGlobalVersion(l.ctx, &l.svcCtx.Redis)

	cacheKey := fmt.Sprintf("product:search:v%d:%s:%d:%d", globalVersion, in.Keyword, in.Page, in.PageSize)
	cacheData, err := l.svcCtx.Redis.GetCtx(l.ctx, cacheKey)
	if err == nil && cacheData != "" {
		var cachedResponse product.SearchProductsResponse
		if err := json.Unmarshal([]byte(cacheData), &cachedResponse); err == nil {
			l.Logger.Infof("Search info retrieved from cache: Total = %d", cachedResponse.Total)
			return &cachedResponse, nil
		}
	}

	// 3. Search products from database
	products, total, err := l.svcCtx.ProductModel.Search(l.ctx, in.Keyword, page, pageSize)
	if err != nil {
		l.Logger.Errorf("Failed to search products: %v", err)
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

	l.Logger.Infof("Search completed: keyword=%s, total=%d", in.Keyword, total)

	result := &product.SearchProductsResponse{
		Total:    total,
		Products: productList,
	}

	JsonResult, err := json.Marshal(result)
	if err == nil {
		err := l.svcCtx.Redis.SetexCtx(l.ctx, cacheKey, string(JsonResult), l.svcCtx.Config.Cache.SearchExpire)
		if err != nil {
			l.Logger.Errorf("Failed to cache product search info: %v", err)
		}
	}

	return result, nil
}
