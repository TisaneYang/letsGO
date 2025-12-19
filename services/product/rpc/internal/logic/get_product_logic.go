package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"letsgo/common/errorx"
	"letsgo/services/product/model"
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
	// 1. Validate product ID
	if in.Id <= 0 {
		return nil, errorx.NewCodeError(1001, "Invalid product ID")
	}

	// 2. Try to get Product ID from cache
	cacheKey := fmt.Sprintf("product:detail:%d", in.Id)
	cacheData, err := l.svcCtx.Redis.GetCtx(l.ctx, cacheKey)
	if err == nil && cacheData != "" {
		if cacheData == "null" {
			return nil, errorx.ErrProductNotFound
		}

		var cachedProduct product.GetProductResponse
		if err := json.Unmarshal([]byte(cacheData), &cachedProduct); err == nil {
			l.Logger.Infof("Product info retrieved from cache: product_id=%d", in.Id)
			return &cachedProduct, nil
		}
	}

	// 3. Query product from database
	productData, err := l.svcCtx.ProductModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			// Cache "not found" to prevent cache penetration
			l.svcCtx.Redis.SetexCtx(l.ctx, cacheKey, "null", 60) // 缓存60秒
			return nil, errorx.ErrProductNotFound
		}
		l.Logger.Errorf("Failed to get product: %v", err)
		return nil, errorx.ErrDatabase
	}

	result := &product.GetProductResponse{
		Product: &product.ProductInfo{
			Id:          productData.Id,
			Name:        productData.Name,
			Description: productData.Description,
			Price:       productData.Price,
			Stock:       productData.Stock,
			Category:    productData.Category,
			Images:      productData.Images,
			Attributes:  productData.Attributes,
			Sales:       productData.Sales,
			CreatedAt:   productData.CreatedAt,
			UpdatedAt:   productData.UpdatedAt,
		},
	}

	JsonResult, err := json.Marshal(result)
	if err == nil {
		err := l.svcCtx.Redis.SetexCtx(l.ctx, cacheKey, string(JsonResult), l.svcCtx.Config.Cache.ProductExpire)
		if err != nil {
			l.Logger.Errorf("Failed to cache product info: %v", err)
		}
	}

	l.Logger.Infof("Product info retrieved from database: product_id=%d", in.Id)

	// 4. Build response
	return result, nil
}
