// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"
	"fmt"

	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"
	"letsgo/services/payment/rpc/payment_client"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryPaymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Query payment status - Check payment result
func NewQueryPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryPaymentLogic {
	return &QueryPaymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryPaymentLogic) QueryPayment(req *types.QueryPaymentReq) (resp *types.QueryPaymentResp, err error) {
	// Get payment by order ID
	paymentResp, err := l.svcCtx.PaymentRpc.GetPaymentByOrderId(l.ctx, &payment_client.GetPaymentByOrderIdRequest{
		OrderId: req.OrderId,
	})
	if err != nil {
		l.Logger.Errorf("failed to get payment: %v", err)
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	payment := paymentResp.Payment
	return &types.QueryPaymentResp{
		PaymentId:   payment.Id,
		PaymentNo:   payment.PaymentNo,
		OrderId:     payment.OrderId,
		Status:      int(payment.Status),
		Amount:      payment.Amount,
		PaymentType: int(payment.PaymentType),
		PaidAt:      payment.PaidAt,
	}, nil
}
