package order

import (
	"context"

	"letsgo/gateway/internal/svc"
	"letsgo/gateway/internal/types"
	"letsgo/services/order/rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryOrderByNoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Query order by number - Search order by order number
func NewQueryOrderByNoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryOrderByNoLogic {
	return &QueryOrderByNoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryOrderByNoLogic) QueryOrderByNo(req *types.QueryOrderByNoReq) (resp *types.QueryOrderByNoResp, err error) {
	// Call Order RPC service
	rpcResp, err := l.svcCtx.OrderRpc.GetOrderByNo(l.ctx, &order.GetOrderByNoRequest{
		OrderNo: req.OrderNo,
	})
	if err != nil {
		l.Logger.Errorf("failed to query order by number: %v", err)
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

	return &types.QueryOrderByNoResp{
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
