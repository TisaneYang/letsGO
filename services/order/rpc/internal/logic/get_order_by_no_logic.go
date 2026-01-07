package logic

import (
	"context"

	"letsgo/services/order/rpc/internal/svc"
	"letsgo/services/order/rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderByNoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderByNoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderByNoLogic {
	return &GetOrderByNoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Get order by order number
func (l *GetOrderByNoLogic) GetOrderByNo(in *order.GetOrderByNoRequest) (*order.GetOrderByNoResponse, error) {
	// 1. Find order by order number
	orderData, err := l.svcCtx.OrderModel.FindOneByOrderNo(l.ctx, in.OrderNo)
	if err != nil {
		l.Logger.Errorf("failed to find order by order_no %s: %v", in.OrderNo, err)
		return nil, err
	}

	// 2. Get order items
	items, err := l.svcCtx.OrderItemModel.FindByOrderId(l.ctx, orderData.Id)
	if err != nil {
		l.Logger.Errorf("failed to find order items for order %d: %v", orderData.Id, err)
		return nil, err
	}

	// 3. Convert to response format (reuse the helper from get_order_logic)
	orderInfo := convertToOrderInfo(orderData, items)

	return &order.GetOrderByNoResponse{
		Order: orderInfo,
	}, nil
}
