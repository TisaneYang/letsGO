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

type CreatePaymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create payment - Initiate payment for order
func NewCreatePaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePaymentLogic {
	return &CreatePaymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreatePaymentLogic) CreatePayment(req *types.CreatePaymentReq) (resp *types.CreatePaymentResp, err error) {
	// Get user ID from context (set by Auth middleware)
	userId, ok := l.ctx.Value("userId").(int64)
	if !ok {
		l.Logger.Error("failed to get userId from context")
		return nil, fmt.Errorf("unauthorized")
	}

	// Call Payment RPC to create payment
	paymentResp, err := l.svcCtx.PaymentRpc.CreatePayment(l.ctx, &payment_client.CreatePaymentRequest{
		OrderId:     req.OrderId,
		UserId:      userId,
		Amount:      req.Amount,
		PaymentType: int32(req.PaymentType),
	})
	if err != nil {
		l.Logger.Errorf("failed to create payment: %v", err)
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return &types.CreatePaymentResp{
		PaymentId: paymentResp.PaymentId,
		PaymentNo: paymentResp.PaymentNo,
		PayUrl:    paymentResp.PayUrl,
		QrCode:    paymentResp.QrCode,
	}, nil
}
