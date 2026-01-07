package logic

import (
	"letsgo/services/order/model"
	"letsgo/services/order/rpc/order"
)

// convertToOrderInfo converts model data to protobuf OrderInfo
func convertToOrderInfo(orderData *model.Order, items []*model.OrderItem) *order.OrderInfo {
	// Convert order items
	orderItems := make([]*order.OrderItem, 0, len(items))
	for _, item := range items {
		orderItems = append(orderItems, &order.OrderItem{
			ProductId: item.ProductId,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  int64(item.Quantity),
			Image:     item.Image,
		})
	}

	// Convert timestamps (handle sql.NullTime)
	var paidAt, shippedAt, completedAt int64
	if orderData.PaidAt.Valid {
		paidAt = orderData.PaidAt.Time.Unix()
	}
	if orderData.ShippedAt.Valid {
		shippedAt = orderData.ShippedAt.Time.Unix()
	}
	if orderData.CompletedAt.Valid {
		completedAt = orderData.CompletedAt.Time.Unix()
	}

	return &order.OrderInfo{
		Id:          orderData.Id,
		UserId:      orderData.UserId,
		OrderNo:     orderData.OrderNo,
		TotalAmount: orderData.TotalAmount,
		Status:      int32(orderData.Status),
		Address:     orderData.Address,
		Phone:       orderData.Phone,
		Remark:      orderData.Remark,
		Items:       orderItems,
		CreatedAt:   orderData.CreatedAt.Unix(),
		UpdatedAt:   orderData.UpdatedAt.Unix(),
		PaidAt:      paidAt,
		ShippedAt:   shippedAt,
		CompletedAt: completedAt,
	}
}
