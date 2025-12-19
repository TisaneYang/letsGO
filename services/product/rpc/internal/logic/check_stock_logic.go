package logic

import (
	"context"

	"letsgo/common/errorx"
	"letsgo/services/product/rpc/internal/svc"
	"letsgo/services/product/rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckStockLogic {
	return &CheckStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Check if products are in stock (batch check)
func (l *CheckStockLogic) CheckStock(in *product.CheckStockRequest) (*product.CheckStockResponse, error) {
	// 1. Validate input
	if len(in.Items) == 0 {
		return nil, errorx.NewCodeError(1001, "No items to check")
	}

	// 2. Collect product IDs
	productIds := make([]int64, 0, len(in.Items))
	requiredQuantities := make(map[int64]int64)

	for _, item := range in.Items {
		if item.ProductId <= 0 {
			return nil, errorx.NewCodeError(1001, "Invalid product ID")
		}
		if item.RequiredQuantity <= 0 {
			return nil, errorx.NewCodeError(1001, "Required quantity must be positive")
		}
		productIds = append(productIds, item.ProductId)
		requiredQuantities[item.ProductId] = item.RequiredQuantity
	}

	// 3. Batch check stock from database
	stockMap, err := l.svcCtx.ProductModel.CheckStock(l.ctx, productIds)
	if err != nil {
		l.Logger.Errorf("Failed to check stock: %v", err)
		return nil, errorx.ErrDatabase
	}

	// 4. Build response and check availability
	allAvailable := true
	resultItems := make([]*product.StockItem, 0, len(in.Items))

	for _, item := range in.Items {
		availableStock, exists := stockMap[item.ProductId]
		if !exists {
			// Product not found
			allAvailable = false
			resultItems = append(resultItems, &product.StockItem{
				ProductId:        item.ProductId,
				RequiredQuantity: item.RequiredQuantity,
				AvailableStock:   0,
			})
			continue
		}

		// Check if stock is sufficient
		if availableStock < item.RequiredQuantity {
			allAvailable = false
		}

		resultItems = append(resultItems, &product.StockItem{
			ProductId:        item.ProductId,
			RequiredQuantity: item.RequiredQuantity,
			AvailableStock:   availableStock,
		})
	}

	l.Logger.Infof("Stock check completed: total_items=%d, all_available=%v",
		len(in.Items), allAvailable)

	return &product.CheckStockResponse{
		Available: allAvailable,
		Items:     resultItems,
	}, nil
}
