package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"

	"letsgo/services/order/model"
	"letsgo/services/order/rpc/internal/svc"
	"letsgo/services/order/rpc/internal/utils"
	"letsgo/services/order/rpc/order"
)

type UpdateOrderStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateOrderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrderStatusLogic {
	return &UpdateOrderStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Update order status (called by payment service, shipping system)
func (l *UpdateOrderStatusLogic) UpdateOrderStatus(in *order.UpdateOrderStatusRequest) (*order.UpdateOrderStatusResponse, error) {
	// 1. Validate status (不允许通过此接口设置为 cancelled)
	validStatuses := map[int32]bool{
		int32(model.OrderStatusPaid):      true,
		int32(model.OrderStatusShipped):   true,
		int32(model.OrderStatusCompleted): true,
	}
	if !validStatuses[in.Status] {
		l.Logger.Errorf("invalid status %d for order %d (use CancelOrder for cancellation)", in.Status, in.OrderId)
		return &order.UpdateOrderStatusResponse{
			Success: false,
		}, fmt.Errorf("invalid status (use CancelOrder for cancellation)")
	}

	// 2. Find order to get current status
	orderData, err := l.svcCtx.OrderModel.FindOne(l.ctx, in.OrderId)
	if err != nil {
		l.Logger.Errorf("failed to find order %d: %v", in.OrderId, err)
		return &order.UpdateOrderStatusResponse{
			Success: false,
		}, err
	}

	// 3. Validate status transition
	if !isValidStatusTransition(orderData.Status, int(in.Status)) {
		l.Logger.Errorf("invalid status transition for order %d: %d -> %d", in.OrderId, orderData.Status, in.Status)
		return &order.UpdateOrderStatusResponse{
			Success: false,
		}, fmt.Errorf("invalid status transition")
	}

	// 4. Update status with timestamp
	timestamp := time.Unix(in.OperatedAt, 0)
	if in.OperatedAt == 0 {
		timestamp = time.Now()
	}

	err = l.svcCtx.OrderModel.UpdateStatus(l.ctx, in.OrderId, int(in.Status), timestamp)
	if err != nil {
		l.Logger.Errorf("failed to update order %d status to %d: %v", in.OrderId, in.Status, err)
		return &order.UpdateOrderStatusResponse{
			Success: false,
		}, err
	}

	l.Logger.Infof("order %d status updated from %d to %d", in.OrderId, orderData.Status, in.Status)

	// 5. Publish status change event to Kafka (异步)
	go l.publishOrderStatusChangedEvent(in.OrderId, orderData.OrderNo, orderData.UserId, orderData.Status, int(in.Status), timestamp)

	return &order.UpdateOrderStatusResponse{
		Success: true,
	}, nil
}

// isValidStatusTransition checks if status transition is valid
// Note: Cancellation is not allowed through this interface, use CancelOrder instead
func isValidStatusTransition(currentStatus int, newStatus int) bool {
	// Define valid transitions (excluding cancellation)
	validTransitions := map[int][]int{
		model.OrderStatusPending: {
			model.OrderStatusPaid,
		},
		model.OrderStatusPaid: {
			model.OrderStatusShipped,
		},
		model.OrderStatusShipped: {
			model.OrderStatusCompleted,
		},
		model.OrderStatusCompleted: {
			// No transitions from completed
		},
		model.OrderStatusCancelled: {
			// No transitions from cancelled
		},
	}

	allowedStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	for _, allowed := range allowedStatuses {
		if allowed == newStatus {
			return true
		}
	}

	return false
}

// publishOrderStatusChangedEvent publishes order status changed event to Kafka
func (l *UpdateOrderStatusLogic) publishOrderStatusChangedEvent(orderId int64, orderNo string, userId int64, oldStatus, newStatus int, timestamp time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	statusNames := map[int]string{
		model.OrderStatusPending:   "pending",
		model.OrderStatusPaid:      "paid",
		model.OrderStatusShipped:   "shipped",
		model.OrderStatusCompleted: "completed",
		model.OrderStatusCancelled: "cancelled",
	}

	event := map[string]interface{}{
		"event_type": "order.status.changed",
		"event_id":   uuid.New().String(),
		"timestamp":  time.Now().Unix(),
		"data": map[string]interface{}{
			"order_id":        orderId,
			"order_no":        orderNo,
			"user_id":         userId,
			"old_status":      oldStatus,
			"new_status":      newStatus,
			"old_status_name": statusNames[oldStatus],
			"new_status_name": statusNames[newStatus],
			"operated_at":     timestamp.Unix(),
		},
	}

	// Publish to Kafka
	producer := utils.NewKafkaProducer(l.svcCtx.KafkaBrokers)
	err := producer.PublishEvent(ctx, l.svcCtx.KafkaTopics.OrderStatusChanged, orderNo, event)
	if err != nil {
		l.Logger.Errorf("failed to publish order status changed event: %v", err)
		// Don't return error, just log it
	} else {
		l.Logger.Infof("published order status changed event for order %s: %s -> %s", orderNo, statusNames[oldStatus], statusNames[newStatus])
	}
}
