package order

import (
	"context"

	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"
	"letsgo/services/order/rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create order - Convert cart to order
func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderReq) (resp *types.CreateOrderResp, err error) {
	// Get user ID from context (set by Auth middleware)
	userId := l.ctx.Value("userId").(int64)

	// Convert request items to RPC format
	items := make([]*order.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, &order.OrderItem{
			ProductId: item.ProductId,
			Quantity:  item.Quantity,
		})
	}

	// Call Order RPC service
	rpcResp, err := l.svcCtx.OrderRpc.CreateOrder(l.ctx, &order.CreateOrderRequest{
		UserId:  userId,
		Items:   items,
		Address: req.Address,
		Phone:   req.Phone,
		Remark:  req.Remark,
	})
	if err != nil {
		l.Logger.Errorf("failed to create order: %v", err)
		return nil, err
	}

	return &types.CreateOrderResp{
		OrderId:     rpcResp.OrderId,
		OrderNo:     rpcResp.OrderNo,
		TotalAmount: rpcResp.TotalAmount,
	}, nil
}
