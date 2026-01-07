package logic

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"

	"letsgo/services/order/model"
	"letsgo/services/order/rpc/internal/svc"
	"letsgo/services/order/rpc/internal/utils"
	"letsgo/services/order/rpc/order"
	"letsgo/services/product/rpc/product"
)

type CancelOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Cancel order (only if not paid)
func (l *CancelOrderLogic) CancelOrder(in *order.CancelOrderRequest) (*order.CancelOrderResponse, error) {
	// 1. Find order first to check status
	orderData, err := l.svcCtx.OrderModel.FindOne(l.ctx, in.OrderId)
	if err != nil {
		l.Logger.Errorf("failed to find order %d: %v", in.OrderId, err)
		return &order.CancelOrderResponse{
			Success: false,
			Message: "Order not found",
		}, nil
	}

	// 2. Verify ownership
	if orderData.UserId != in.UserId {
		l.Logger.Errorf("user %d attempted to cancel order %d owned by user %d", in.UserId, in.OrderId, orderData.UserId)
		return &order.CancelOrderResponse{
			Success: false,
			Message: "Order not found",
		}, nil
	}

	// 3. Check if order can be cancelled (only pending orders)
	if orderData.Status != model.OrderStatusPending {
		statusMsg := map[int]string{
			model.OrderStatusPaid:      "Order has been paid and cannot be cancelled",
			model.OrderStatusShipped:   "Order has been shipped and cannot be cancelled",
			model.OrderStatusCompleted: "Order has been completed and cannot be cancelled",
			model.OrderStatusCancelled: "Order has already been cancelled",
		}
		msg := statusMsg[orderData.Status]
		if msg == "" {
			msg = "Order cannot be cancelled"
		}
		l.Logger.Infof("user %d attempted to cancel order %d with status %d", in.UserId, in.OrderId, orderData.Status)
		return &order.CancelOrderResponse{
			Success: false,
			Message: msg,
		}, nil
	}

	// 4. Get order items to restore stock
	items, err := l.svcCtx.OrderItemModel.FindByOrderId(l.ctx, in.OrderId)
	if err != nil {
		l.Logger.Errorf("failed to find order items for order %d: %v", in.OrderId, err)
		return &order.CancelOrderResponse{
			Success: false,
			Message: "Failed to cancel order",
		}, nil
	}

	// 5. Cancel order in database
	err = l.svcCtx.OrderModel.CancelOrder(l.ctx, in.OrderId, in.UserId)
	if err != nil {
		l.Logger.Errorf("failed to cancel order %d: %v", in.OrderId, err)
		return &order.CancelOrderResponse{
			Success: false,
			Message: "Failed to cancel order",
		}, nil
	}

	// 6. Restore stock (add back the quantities)
	stockItems := make([]*product.StockUpdateItem, 0, len(items))
	for _, item := range items {
		stockItems = append(stockItems, &product.StockUpdateItem{
			ProductId: item.ProductId,
			Quantity:  int64(item.Quantity), // Positive to add back
		})
	}

	_, err = l.svcCtx.ProductRpc.BatchUpdateStock(l.ctx, &product.BatchUpdateStockRequest{
		Items: stockItems,
	})
	if err != nil {
		l.Logger.Errorf("failed to restore stock for cancelled order %d: %v", in.OrderId, err)
		// Don't fail the cancellation, but log it for manual intervention
		// The order is already cancelled in DB
	} else {
		l.Logger.Infof("restored stock for cancelled order %d", in.OrderId)
	}

	l.Logger.Infof("order %d cancelled successfully by user %d", in.OrderId, in.UserId)

	// 7. Publish order cancelled event to Kafka (异步，不影响取消操作)
	go l.publishOrderCancelledEvent(in.OrderId, orderData.OrderNo, in.UserId, orderData.TotalAmount, items)

	return &order.CancelOrderResponse{
		Success: true,
		Message: "Order cancelled successfully",
	}, nil
}

// publishOrderCancelledEvent publishes order cancelled event to Kafka
func (l *CancelOrderLogic) publishOrderCancelledEvent(orderId int64, orderNo string, userId int64, totalAmount float64, items []*model.OrderItem) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare event data
	eventItems := make([]utils.OrderItem, 0, len(items))
	for _, item := range items {
		eventItems = append(eventItems, utils.OrderItem{
			ProductID: item.ProductId,
			Quantity:  int64(item.Quantity),
		})
	}

	event := map[string]interface{}{
		"event_type": "order.cancelled",
		"event_id":   uuid.New().String(),
		"timestamp":  time.Now().Unix(),
		"data": map[string]interface{}{
			"order_id":     orderId,
			"order_no":     orderNo,
			"user_id":      userId,
			"total_amount": totalAmount,
			"items":        eventItems,
			"cancelled_at": time.Now().Unix(),
		},
	}

	// Publish to Kafka
	producer := utils.NewKafkaProducer(l.svcCtx.KafkaBrokers)
	err := producer.PublishEvent(ctx, l.svcCtx.KafkaTopics.OrderCancelled, orderNo, event)
	if err != nil {
		l.Logger.Errorf("failed to publish order cancelled event: %v", err)
		// Don't return error, just log it
	} else {
		l.Logger.Infof("published order cancelled event for order %s", orderNo)
	}
}
