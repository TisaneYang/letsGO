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

type BatchUpdateStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchUpdateStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchUpdateStockLogic {
	return &BatchUpdateStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Batch update stock for multiple products (transactional)
// All updates succeed or all fail (atomic operation)
func (l *BatchUpdateStockLogic) BatchUpdateStock(in *product.BatchUpdateStockRequest) (*product.BatchUpdateStockResponse, error) {
	// 1. Validate parameters
	if len(in.Items) == 0 {
		return nil, errorx.NewCodeError(1001, "No items to update")
	}

	// 2. Convert protobuf items to model items
	modelItems := make([]model.StockUpdateItem, 0, len(in.Items))
	for _, item := range in.Items {
		if item.ProductId <= 0 {
			return nil, errorx.NewCodeError(1001, "Invalid product ID")
		}
		modelItems = append(modelItems, model.StockUpdateItem{
			ProductId: item.ProductId,
			Quantity:  item.Quantity,
		})
	}

	// 3. Update stock in transaction
	results, err := l.svcCtx.ProductModel.BatchUpdateStock(l.ctx, modelItems)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.ErrProductNotFound
		}
		if err == model.ErrInsufficientStock {
			return nil, errorx.ErrProductOutOfStock
		}
		l.Logger.Errorf("Failed to batch update stock: %v", err)
		return nil, errorx.ErrDatabase
	}

	l.Logger.Infof("Batch stock updated: %d products", len(results))

	// 4. Clear cache for all affected products
	// Note: We don't fail the request if cache clearing fails
	for _, result := range results {
		cacheKey := fmt.Sprintf("product:detail:%d", result.ProductId)
		_, err = l.svcCtx.Redis.DelCtx(l.ctx, cacheKey)
		if err != nil {
			l.Logger.Errorf("Delete product detail cache failed for product %d: %s", result.ProductId, err)
		}
	}

	// Note: For category and global version, we increment once since this is a batch operation
	// We can't determine which categories are affected without additional queries
	// So we just increment the global version
	err = IncGlobalVersion(l.ctx, &l.svcCtx.Redis)
	if err != nil {
		l.Logger.Errorf("Increase global version failed: %s", err)
	}

	// 5. Convert results to protobuf format
	pbResults := make([]*product.StockUpdateResult, 0, len(results))
	for _, result := range results {
		pbResults = append(pbResults, &product.StockUpdateResult{
			ProductId: result.ProductId,
			NewStock:  result.NewStock,
		})
	}

	return &product.BatchUpdateStockResponse{
		Success: true,
		Results: pbResults,
	}, nil
}

