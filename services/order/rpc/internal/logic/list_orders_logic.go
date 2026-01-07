package logic

import (
	"context"

	"letsgo/services/order/rpc/internal/svc"
	"letsgo/services/order/rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// List user's orders with pagination
func (l *ListOrdersLogic) ListOrders(in *order.ListOrdersRequest) (*order.ListOrdersResponse, error) {
	// 1. Set default pagination values
	page := in.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // Max page size
	}

	// 2. Get total count
	total, err := l.svcCtx.OrderModel.CountByUserId(l.ctx, in.UserId, int(in.Status))
	if err != nil {
		l.Logger.Errorf("failed to count orders for user %d: %v", in.UserId, err)
		return nil, err
	}

	// 3. Get orders
	orders, err := l.svcCtx.OrderModel.FindByUserId(l.ctx, in.UserId, int(page), int(pageSize), int(in.Status))
	if err != nil {
		l.Logger.Errorf("failed to find orders for user %d: %v", in.UserId, err)
		return nil, err
	}

	// 4. Convert to response format
	orderInfos := make([]*order.OrderInfo, 0, len(orders))
	for _, orderData := range orders {
		// Get order items for each order
		items, err := l.svcCtx.OrderItemModel.FindByOrderId(l.ctx, orderData.Id)
		if err != nil {
			l.Logger.Errorf("failed to find order items for order %d: %v", orderData.Id, err)
			continue // Skip this order if items can't be loaded
		}

		orderInfo := convertToOrderInfo(orderData, items)
		orderInfos = append(orderInfos, orderInfo)
	}

	return &order.ListOrdersResponse{
		Total:  total,
		Orders: orderInfos,
	}, nil
}
