package logic

import (
	"context"
	"fmt"

	"letsgo/services/payment/rpc/internal/svc"
	"letsgo/services/payment/rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelPaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelPaymentLogic {
	return &CancelPaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Cancel payment (timeout or user cancellation)
func (l *CancelPaymentLogic) CancelPayment(in *payment.CancelPaymentRequest) (*payment.CancelPaymentResponse, error) {
	// 1. Validate input
	if in.PaymentId <= 0 {
		l.Logger.Errorf("invalid payment_id: %d", in.PaymentId)
		return nil, fmt.Errorf("invalid payment_id")
	}

	// 2. Cancel payment (only if status is pending)
	err := l.svcCtx.PaymentModel.CancelPayment(l.ctx, in.PaymentId)
	if err != nil {
		l.Logger.Errorf("failed to cancel payment: %v", err)
		return nil, fmt.Errorf("failed to cancel payment: %w", err)
	}

	l.Logger.Infof("payment cancelled successfully: payment_id=%d", in.PaymentId)

	return &payment.CancelPaymentResponse{
		Success: true,
	}, nil
}
