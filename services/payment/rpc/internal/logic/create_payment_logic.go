package logic

import (
	"context"
	"fmt"
	"time"

	"letsgo/services/order/rpc/order_client"
	"letsgo/services/payment/model"
	"letsgo/services/payment/rpc/internal/svc"
	"letsgo/services/payment/rpc/internal/utils"
	"letsgo/services/payment/rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePaymentLogic {
	return &CreatePaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Create payment for order
func (l *CreatePaymentLogic) CreatePayment(in *payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	// 1. Validate input
	if in.OrderId <= 0 {
		l.Logger.Errorf("invalid order_id: %d", in.OrderId)
		return nil, fmt.Errorf("invalid order_id")
	}
	if in.UserId <= 0 {
		l.Logger.Errorf("invalid user_id: %d", in.UserId)
		return nil, fmt.Errorf("invalid user_id")
	}
	if in.Amount <= 0 {
		l.Logger.Errorf("invalid amount: %f", in.Amount)
		return nil, fmt.Errorf("invalid amount")
	}

	// 2. Verify order exists and status is pending (call Order RPC)
	orderResp, err := l.svcCtx.OrderRpc.GetOrder(l.ctx, &order_client.GetOrderRequest{
		OrderId: in.OrderId,
		UserId:  in.UserId,
	})
	if err != nil {
		l.Logger.Errorf("failed to get order: %v", err)
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Check order status is pending (1)
	if orderResp.Order.Status != 1 {
		l.Logger.Errorf("order status is not pending: %d", orderResp.Order.Status)
		return nil, fmt.Errorf("order status is not pending, cannot create payment")
	}

	// 3. Check if payment already exists for this order
	existingPayment, err := l.svcCtx.PaymentModel.FindOneByOrderId(l.ctx, in.OrderId)
	if err == nil && existingPayment != nil {
		l.Logger.Infof("payment already exists for order_id: %d, payment_id: %d", in.OrderId, existingPayment.Id)

		// Return existing payment info
		payUrl := ""
		if l.svcCtx.Config.MockPayment.Enabled {
			payUrl = fmt.Sprintf("http://mock-payment.com/pay?payment_no=%s", existingPayment.PaymentNo)
		}

		return &payment.CreatePaymentResponse{
			PaymentId: existingPayment.Id,
			PaymentNo: existingPayment.PaymentNo,
			PayUrl:    payUrl,
			QrCode:    "",
		}, nil
	}

	// 4. Generate unique payment transaction number
	paymentNo := utils.GeneratePaymentNo()
	l.Logger.Infof("generated payment_no: %s for order_id: %d", paymentNo, in.OrderId)

	// 5. Create payment record
	now := time.Now()
	paymentData := &model.Payment{
		OrderId:     in.OrderId,
		UserId:      in.UserId,
		PaymentNo:   paymentNo,
		Amount:      in.Amount,
		PaymentType: int(in.PaymentType),
		Status:      model.PaymentStatusPending,
		TradeNo:     "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	paymentId, err := l.svcCtx.PaymentModel.Insert(l.ctx, paymentData)
	if err != nil {
		l.Logger.Errorf("failed to insert payment: %v", err)
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	l.Logger.Infof("created payment: payment_id=%d, payment_no=%s, order_id=%d", paymentId, paymentNo, in.OrderId)

	// 6. Cache to Redis (key: payment:order:{orderId}, value: paymentId, TTL: 15 minutes)
	cacheKey := fmt.Sprintf("payment:order:%d", in.OrderId)
	err = l.svcCtx.Redis.SetexCtx(l.ctx, cacheKey, fmt.Sprintf("%d", paymentId), 900) // 15 minutes = 900 seconds
	if err != nil {
		l.Logger.Errorf("failed to cache payment to redis: %v", err)
		// Don't fail the request if cache fails
	}

	// 7. Generate payment URL or QR code
	payUrl := ""
	qrCode := ""

	if l.svcCtx.Config.MockPayment.Enabled {
		// Mock payment mode
		payUrl = fmt.Sprintf("http://mock-payment.com/pay?payment_no=%s", paymentNo)
		l.Logger.Infof("mock payment enabled, pay_url: %s", payUrl)
	} else {
		// Real payment gateway integration would go here
		// For now, return empty
		l.Logger.Info("real payment gateway not implemented yet")
	}

	return &payment.CreatePaymentResponse{
		PaymentId: paymentId,
		PaymentNo: paymentNo,
		PayUrl:    payUrl,
		QrCode:    qrCode,
	}, nil
}
