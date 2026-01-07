package order

import (
	"context"

	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"
	"letsgo/services/order/rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get order detail - View single order details
func NewGetOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderDetailLogic {
	return &GetOrderDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderDetailLogic) GetOrderDetail(req *types.OrderDetailReq) (resp *types.OrderDetailResp, err error) {
	// Get user ID from context (set by Auth middleware)
	userId := l.ctx.Value("userId").(int64)

	// Call Order RPC service
	rpcResp, err := l.svcCtx.OrderRpc.GetOrder(l.ctx, &order.GetOrderRequest{
		OrderId: req.Id,
		UserId:  userId,
	})
	if err != nil {
		l.Logger.Errorf("failed to get order detail: %v", err)
		return nil, err
	}

	o := rpcResp.Order

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

	return &types.OrderDetailResp{
		Order: types.Order{
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
		},
	}, nil
}
