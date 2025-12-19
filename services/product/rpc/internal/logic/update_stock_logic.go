package logic

import (
	"context"
	"fmt"

	"letsgo/common/errorx"
	"letsgo/services/product/model"
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
// quantity: positive = increase, negative = decrease
func (l *UpdateStockLogic) UpdateStock(in *product.UpdateStockRequest) (*product.UpdateStockResponse, error) {
	// 1. Validate parameters
	if in.ProductId <= 0 {
		return nil, errorx.NewCodeError(1001, "Invalid product ID")
	}

	// 2. Update stock atomically
	newStock, category, err := l.svcCtx.ProductModel.UpdateStock(l.ctx, in.ProductId, in.Quantity)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.ErrProductNotFound
		}
		if err == model.ErrInsufficientStock {
			return nil, errorx.ErrProductOutOfStock
		}
		l.Logger.Errorf("Failed to update stock: %v", err)
		return nil, errorx.ErrDatabase
	}

	l.Logger.Infof("Stock updated: product_id=%d, quantity=%d, new_stock=%d",
		in.ProductId, in.Quantity, newStock)

	// 3. Keep cache data consistant.
	cacheKey := fmt.Sprintf("product:detail:%d", in.ProductId)
	_, err = l.svcCtx.Redis.DelCtx(l.ctx, cacheKey)
	if err != nil {
		l.Logger.Errorf("Delete product detail cache failed! err:%s", err)
	}

	err = IncCategoryVersion(l.ctx, category, &l.svcCtx.Redis)
	if err != nil {
		l.Logger.Errorf("Increase category version failed! err:%s", err)
	}

	err = IncGlobalVersion(l.ctx, &l.svcCtx.Redis)
	if err != nil {
		l.Logger.Errorf("Increase global version failed! err:%s", err)
	}

	return &product.UpdateStockResponse{
		Success:  true,
		NewStock: newStock,
	}, nil
}
