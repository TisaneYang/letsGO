package logic

import (
	"context"
	"fmt"
	"time"

	"letsgo/common/errorx"
	"letsgo/services/product/model"
	"letsgo/services/product/rpc/internal/svc"
	"letsgo/services/product/rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductLogic {
	return &UpdateProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Update product information (admin)
func (l *UpdateProductLogic) UpdateProduct(in *product.UpdateProductRequest) (*product.UpdateProductResponse, error) {
	// 1. Validate product ID
	if in.Id <= 0 {
		return nil, errorx.NewCodeError(1001, "Invalid product ID")
	}

	// 2. Get existing product to check if it exists
	existingProduct, err := l.svcCtx.ProductModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.ErrProductNotFound
		}
		l.Logger.Errorf("Failed to get product: %v", err)
		return nil, errorx.ErrDatabase
	}

	// 3. Update fields (only update non-empty/non-zero values)
	// Note: stock and sales are managed by dedicated RPCs (UpdateStock, IncrementSales)
	updatedProduct := &model.Product{
		Id:          in.Id,
		Name:        in.Name,
		Description: in.Description,
		Price:       in.Price,
		Stock:       existingProduct.Stock, // Keep existing stock - use UpdateStock RPC instead
		Category:    in.Category,
		Images:      in.Images,
		Attributes:  in.Attributes,
		Sales:       existingProduct.Sales,  // Keep existing sales - use IncrementSales RPC instead
		Status:      existingProduct.Status, // Keep existing status
		CreatedAt:   existingProduct.CreatedAt,
		UpdatedAt:   time.Now().Unix(),
	}

	// Use existing values if new values are empty
	if in.Name == "" {
		updatedProduct.Name = existingProduct.Name
	}
	if in.Description == "" {
		updatedProduct.Description = existingProduct.Description
	}
	if in.Price == 0 {
		updatedProduct.Price = existingProduct.Price
	}
	if in.Category == "" {
		updatedProduct.Category = existingProduct.Category
	}
	if len(in.Images) == 0 {
		updatedProduct.Images = existingProduct.Images
	}
	if in.Attributes == "" {
		updatedProduct.Attributes = existingProduct.Attributes
	}

	// 4. Update product in database
	err = l.svcCtx.ProductModel.Update(l.ctx, updatedProduct)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.ErrProductNotFound
		}
		l.Logger.Errorf("Failed to update product: %v", err)
		return nil, errorx.ErrDatabase
	}

	l.Logger.Infof("Product updated successfully: product_id=%d", in.Id)

	// 5. Keep cache data consistant.
	cacheKey := fmt.Sprintf("product:detail:%d", in.Id)
	_, err = l.svcCtx.Redis.DelCtx(l.ctx, cacheKey)
	if err != nil {
		l.Logger.Errorf("Delete product detail cache failed! err:%s", err)
	}

	err = IncCategoryVersion(l.ctx, existingProduct.Category, &l.svcCtx.Redis)
	if err != nil {
		l.Logger.Errorf("Increse category version failed! err:%s", err)
	}

	if in.Category != "" && in.Category != existingProduct.Category {
		err = IncCategoryVersion(l.ctx, in.Category, &l.svcCtx.Redis)
		if err != nil {
			l.Logger.Errorf("Failed to increment new category version: %v", err)
		}
	}

	err = IncGlobalVersion(l.ctx, &l.svcCtx.Redis)
	if err != nil {
		l.Logger.Errorf("Increse global version failed! err:%s", err)
	}

	return &product.UpdateProductResponse{
		Success: true,
	}, nil
}
