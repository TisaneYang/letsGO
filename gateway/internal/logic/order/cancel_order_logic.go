package order

import (
	"context"

	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"
	"letsgo/services/order/rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Cancel order - Cancel pending order
func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelOrderLogic) CancelOrder(req *types.CancelOrderReq) (resp *types.CancelOrderResp, err error) {
	// Get user ID from context (set by Auth middleware)
	userId := l.ctx.Value("userId").(int64)

	// Call Order RPC service
	rpcResp, err := l.svcCtx.OrderRpc.CancelOrder(l.ctx, &order.CancelOrderRequest{
		OrderId: req.Id,
		UserId:  userId,
	})
	if err != nil {
		l.Logger.Errorf("failed to cancel order: %v", err)
		return nil, err
	}

	return &types.CancelOrderResp{
		Success: rpcResp.Success,
	}, nil
}
