package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"

	"letsgo/services/order/rpc/order_client"
	"letsgo/services/payment/model"
	"letsgo/services/payment/rpc/internal/svc"
	"letsgo/services/payment/rpc/internal/utils"
	"letsgo/services/payment/rpc/payment"
)

type PaymentCallbackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPaymentCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PaymentCallbackLogic {
	return &PaymentCallbackLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Process payment callback from payment gateway
func (l *PaymentCallbackLogic) PaymentCallback(in *payment.PaymentCallbackRequest) (*payment.PaymentCallbackResponse, error) {
	// 1. Validate input
	if in.PaymentNo == "" {
		l.Logger.Errorf("invalid payment_no: empty")
		return nil, fmt.Errorf("invalid payment_no")
	}
	if in.OrderId <= 0 {
		l.Logger.Errorf("invalid order_id: %d", in.OrderId)
		return nil, fmt.Errorf("invalid order_id")
	}

	l.Logger.Infof("received payment callback: payment_no=%s, order_id=%d, status=%d, amount=%f, trade_no=%s",
		in.PaymentNo, in.OrderId, in.Status, in.Amount, in.TradeNo)

	// 2. Query payment record by payment_no
	paymentData, err := l.svcCtx.PaymentModel.FindOneByPaymentNo(l.ctx, in.PaymentNo)
	if err != nil {
		l.Logger.Errorf("failed to find payment by payment_no: %v", err)
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// 3. Idempotency check: if status is not pending, return success directly
	if paymentData.Status != model.PaymentStatusPending {
		l.Logger.Infof("payment already processed: payment_id=%d, current_status=%d", paymentData.Id, paymentData.Status)
		return &payment.PaymentCallbackResponse{
			Success: true,
			Message: "payment already processed",
		}, nil
	}

	// 4. Verify amount matches
	if paymentData.Amount != in.Amount {
		l.Logger.Errorf("amount mismatch: expected=%f, got=%f", paymentData.Amount, in.Amount)
		return nil, fmt.Errorf("amount mismatch")
	}

	// 5. Update payment status based on callback status
	now := time.Now()
	var newStatus int

	if in.Status == 2 { // Success
		newStatus = model.PaymentStatusSuccess
		l.Logger.Infof("payment success: payment_id=%d, payment_no=%s", paymentData.Id, paymentData.PaymentNo)
	} else if in.Status == 3 { // Failed
		newStatus = model.PaymentStatusFailed
		l.Logger.Infof("payment failed: payment_id=%d, payment_no=%s", paymentData.Id, paymentData.PaymentNo)
	} else {
		l.Logger.Errorf("invalid payment status: %d", in.Status)
		return nil, fmt.Errorf("invalid payment status")
	}

	// 6. Update payment status in database
	err = l.svcCtx.PaymentModel.UpdateStatus(l.ctx, paymentData.Id, newStatus, in.TradeNo, now)
	if err != nil {
		l.Logger.Errorf("failed to update payment status: %v", err)
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	// 7. If payment success, update order status to paid
	if newStatus == model.PaymentStatusSuccess {
		_, err = l.svcCtx.OrderRpc.UpdateOrderStatus(l.ctx, &order_client.UpdateOrderStatusRequest{
			OrderId: paymentData.OrderId,
			Status:  2, // OrderStatusPaid
		})
		if err != nil {
			l.Logger.Errorf("failed to update order status: %v", err)
			// Don't return error, payment is already successful
			// This will be retried by monitoring system
		} else {
			l.Logger.Infof("order status updated to paid: order_id=%d", paymentData.OrderId)
		}

		// 8. Publish payment success event to Kafka (async)
		go l.publishPaymentSuccessEvent(paymentData.Id, paymentData.PaymentNo, paymentData.OrderId, paymentData.UserId, paymentData.Amount, paymentData.PaymentType, in.TradeNo)
	} else {
		// 9. Publish payment failed event to Kafka (async)
		go l.publishPaymentFailedEvent(paymentData.Id, paymentData.PaymentNo, paymentData.OrderId, paymentData.UserId, paymentData.Amount, paymentData.PaymentType, "payment failed")
	}

	return &payment.PaymentCallbackResponse{
		Success: true,
		Message: "ok",
	}, nil
}

// publishPaymentSuccessEvent publishes payment success event to Kafka
func (l *PaymentCallbackLogic) publishPaymentSuccessEvent(paymentId int64, paymentNo string, orderId, userId int64, amount float64, paymentType int, tradeNo string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	event := utils.PaymentSuccessEvent{
		EventType: "payment.success",
		EventID:   uuid.New().String(),
		Timestamp: time.Now().Unix(),
		Data: utils.PaymentData{
			PaymentID:   paymentId,
			PaymentNo:   paymentNo,
			OrderID:     orderId,
			UserID:      userId,
			Amount:      amount,
			PaymentType: paymentType,
			Status:      model.PaymentStatusSuccess,
			TradeNo:     tradeNo,
		},
	}

	producer := utils.NewKafkaProducer(l.svcCtx.KafkaBrokers)
	err := producer.PublishEvent(ctx, l.svcCtx.KafkaTopics.PaymentSuccess, paymentNo, event)
	if err != nil {
		l.Logger.Errorf("failed to publish payment success event: %v", err)
	} else {
		l.Logger.Infof("published payment success event for payment_no: %s", paymentNo)
	}
}

// publishPaymentFailedEvent publishes payment failed event to Kafka
func (l *PaymentCallbackLogic) publishPaymentFailedEvent(paymentId int64, paymentNo string, orderId, userId int64, amount float64, paymentType int, reason string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	event := utils.PaymentFailedEvent{
		EventType: "payment.failed",
		EventID:   uuid.New().String(),
		Timestamp: time.Now().Unix(),
		Data: utils.PaymentData{
			PaymentID:   paymentId,
			PaymentNo:   paymentNo,
			OrderID:     orderId,
			UserID:      userId,
			Amount:      amount,
			PaymentType: paymentType,
			Status:      model.PaymentStatusFailed,
			Reason:      reason,
		},
	}

	producer := utils.NewKafkaProducer(l.svcCtx.KafkaBrokers)
	err := producer.PublishEvent(ctx, l.svcCtx.KafkaTopics.PaymentFailed, paymentNo, event)
	if err != nil {
		l.Logger.Errorf("failed to publish payment failed event: %v", err)
	} else {
		l.Logger.Infof("published payment failed event for payment_no: %s", paymentNo)
	}
}
