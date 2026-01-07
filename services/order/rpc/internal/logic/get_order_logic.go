package logic

import (
	"context"
	"fmt"

	"letsgo/services/order/rpc/internal/svc"
	"letsgo/services/order/rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Get order detail
func (l *GetOrderLogic) GetOrder(in *order.GetOrderRequest) (*order.GetOrderResponse, error) {
	// 1. Find order by ID
	orderData, err := l.svcCtx.OrderModel.FindOne(l.ctx, in.OrderId)
	if err != nil {
		l.Logger.Errorf("failed to find order %d: %v", in.OrderId, err)
		return nil, err
	}

	// 2. Verify ownership
	if orderData.UserId != in.UserId {
		l.Logger.Errorf("user %d attempted to access order %d owned by user %d", in.UserId, in.OrderId, orderData.UserId)
		return nil, fmt.Errorf("order not found")
	}

	// 3. Get order items
	items, err := l.svcCtx.OrderItemModel.FindByOrderId(l.ctx, in.OrderId)
	if err != nil {
		l.Logger.Errorf("failed to find order items for order %d: %v", in.OrderId, err)
		return nil, fmt.Errorf("failed to get order items")
	}

	// 4. Convert to response format
	orderInfo := convertToOrderInfo(orderData, items)

	return &order.GetOrderResponse{
		Order: orderInfo,
	}, nil
}
