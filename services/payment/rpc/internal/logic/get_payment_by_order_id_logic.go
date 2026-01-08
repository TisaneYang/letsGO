package logic

import (
	"context"
	"fmt"

	"letsgo/services/payment/rpc/internal/svc"
	"letsgo/services/payment/rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaymentByOrderIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPaymentByOrderIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaymentByOrderIdLogic {
	return &GetPaymentByOrderIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Get payment by order ID
func (l *GetPaymentByOrderIdLogic) GetPaymentByOrderId(in *payment.GetPaymentByOrderIdRequest) (*payment.GetPaymentByOrderIdResponse, error) {
	// 1. Validate input
	if in.OrderId <= 0 {
		l.Logger.Errorf("invalid order_id: %d", in.OrderId)
		return nil, fmt.Errorf("invalid order_id")
	}

	// 2. Query payment by order_id from database
	paymentData, err := l.svcCtx.PaymentModel.FindOneByOrderId(l.ctx, in.OrderId)
	if err != nil {
		l.Logger.Errorf("failed to find payment by order_id: %v", err)
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// 3. Convert to response
	paidAt := int64(0)
	if paymentData.PaidAt.Valid {
		paidAt = paymentData.PaidAt.Time.Unix()
	}

	return &payment.GetPaymentByOrderIdResponse{
		Payment: &payment.PaymentInfo{
			Id:          paymentData.Id,
			OrderId:     paymentData.OrderId,
			UserId:      paymentData.UserId,
			PaymentNo:   paymentData.PaymentNo,
			Amount:      paymentData.Amount,
			PaymentType: int32(paymentData.PaymentType),
			Status:      int32(paymentData.Status),
			TradeNo:     paymentData.TradeNo,
			CreatedAt:   paymentData.CreatedAt.Unix(),
			PaidAt:      paidAt,
		},
	}, nil
}
