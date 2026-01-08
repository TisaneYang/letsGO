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

type PaymentCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Payment callback - Webhook for payment gateway (for testing)
func NewPaymentCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PaymentCallbackLogic {
	return &PaymentCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PaymentCallbackLogic) PaymentCallback(req *types.PaymentCallbackReq) (resp *types.PaymentCallbackResp, err error) {
	// Call Payment RPC to process callback
	callbackResp, err := l.svcCtx.PaymentRpc.PaymentCallback(l.ctx, &payment_client.PaymentCallbackRequest{
		PaymentNo: req.PaymentNo,
		OrderId:   req.OrderId,
		Status:    int32(req.Status),
		Amount:    req.Amount,
		TradeNo:   req.TradeNo,
		Sign:      "", // Signature verification would go here
	})
	if err != nil {
		l.Logger.Errorf("failed to process payment callback: %v", err)
		return nil, fmt.Errorf("failed to process payment callback: %w", err)
	}

	return &types.PaymentCallbackResp{
		Success: callbackResp.Success,
		Message: callbackResp.Message,
	}, nil
}
