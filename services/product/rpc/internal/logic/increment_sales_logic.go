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

type IncrementSalesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIncrementSalesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncrementSalesLogic {
	return &IncrementSalesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Increment product sales count (called by order service after order completion)
func (l *IncrementSalesLogic) IncrementSales(in *product.IncrementSalesRequest) (*product.IncrementSalesResponse, error) {
	// 1. Validate parameters
	if in.ProductId <= 0 {
		return nil, errorx.NewCodeError(1001, "Invalid product ID")
	}

	if in.Quantity <= 0 {
		return nil, errorx.NewCodeError(1001, "Quantity must be positive")
	}

	// 2. Increment sales atomically and get category
	newSales, category, err := l.svcCtx.ProductModel.IncrementSales(l.ctx, in.ProductId, in.Quantity)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.ErrProductNotFound
		}
		l.Logger.Errorf("Failed to increment sales: %v", err)
		return nil, errorx.ErrDatabase
	}

	l.Logger.Infof("Sales incremented: product_id=%d, quantity=%d, new_sales=%d, category=%s",
		in.ProductId, in.Quantity, newSales, category)

	// 3. Keep cache data consistant.
	cacheKey := fmt.Sprintf("product:detail:%d", in.ProductId)
	_, err = l.svcCtx.Redis.DelCtx(l.ctx, cacheKey)
	if err != nil {
		logx.Errorf("Delete product detail cache failed! err:%s", err)
	}

	err = IncCategoryVersion(l.ctx, category, &l.svcCtx.Redis)
	if err != nil {
		logx.Errorf("Increase category version failed! err:%s", err)
	}

	err = IncGlobalVersion(l.ctx, &l.svcCtx.Redis)
	if err != nil {
		logx.Errorf("Increase global version failed! err:%s", err)
	}

	return &product.IncrementSalesResponse{
		Success:  true,
		NewSales: newSales,
	}, nil
}
