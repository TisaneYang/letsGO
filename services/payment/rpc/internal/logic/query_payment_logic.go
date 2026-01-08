package logic

import (
	"context"
	"fmt"

	"letsgo/services/payment/rpc/internal/svc"
	"letsgo/services/payment/rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryPaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryPaymentLogic {
	return &QueryPaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Query payment status
func (l *QueryPaymentLogic) QueryPayment(in *payment.QueryPaymentRequest) (*payment.QueryPaymentResponse, error) {
	// 1. Validate input
	if in.PaymentId <= 0 {
		l.Logger.Errorf("invalid payment_id: %d", in.PaymentId)
		return nil, fmt.Errorf("invalid payment_id")
	}

	// 2. Query payment from database
	paymentData, err := l.svcCtx.PaymentModel.FindOne(l.ctx, in.PaymentId)
	if err != nil {
		l.Logger.Errorf("failed to find payment: %v", err)
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// 3. Convert to response
	paidAt := int64(0)
	if paymentData.PaidAt.Valid {
		paidAt = paymentData.PaidAt.Time.Unix()
	}

	return &payment.QueryPaymentResponse{
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
