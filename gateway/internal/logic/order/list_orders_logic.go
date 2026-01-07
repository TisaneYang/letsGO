package order

import (
	"context"

	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"
	"letsgo/services/order/rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get order list - View user's order history
func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListOrdersLogic) ListOrders(req *types.OrderListReq) (resp *types.OrderListResp, err error) {
	// Get user ID from context (set by Auth middleware)
	userId := l.ctx.Value("userId").(int64)

	// Call Order RPC service
	rpcResp, err := l.svcCtx.OrderRpc.ListOrders(l.ctx, &order.ListOrdersRequest{
		UserId:   userId,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		Status:   int32(req.Status),
	})
	if err != nil {
		l.Logger.Errorf("failed to list orders: %v", err)
		return nil, err
	}

	// Convert RPC response to gateway response
	orders := make([]types.Order, 0, len(rpcResp.Orders))
	for _, o := range rpcResp.Orders {
		// Convert order items
		items := make([]types.OrderItem, 0, len(o.Items))
		for _, item := range o.Items {
			items = append(items, types.OrderItem{
				ProductId: item.ProductId,
				Name:      item.Name,
				Price:     item.Price,
				Quantity:  item.Quantity,
				Image:     item.Image,
			})
		}

		orders = append(orders, types.Order{
			Id:          o.Id,
			UserId:      o.UserId,
			OrderNo:     o.OrderNo,
			TotalAmount: o.TotalAmount,
			Status:      int(o.Status),
			Address:     o.Address,
			Phone:       o.Phone,
			Remark:      o.Remark,
			Items:       items,
			CreatedAt:   o.CreatedAt,
			UpdatedAt:   o.UpdatedAt,
			PaidAt:      o.PaidAt,
			ShippedAt:   o.ShippedAt,
			CompletedAt: o.CompletedAt,
		})
	}

	return &types.OrderListResp{
		Total:  rpcResp.Total,
		Orders: orders,
	}, nil
}
