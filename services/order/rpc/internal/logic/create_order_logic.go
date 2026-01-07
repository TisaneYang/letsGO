package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"

	"letsgo/services/cart/rpc/cart"
	"letsgo/services/order/model"
	"letsgo/services/order/rpc/internal/svc"
	"letsgo/services/order/rpc/internal/utils"
	"letsgo/services/order/rpc/order"
	"letsgo/services/product/rpc/product"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Create new order from cart
func (l *CreateOrderLogic) CreateOrder(in *order.CreateOrderRequest) (*order.CreateOrderResponse, error) {
	// 1. Validate input
	if len(in.Items) == 0 {
		return nil, fmt.Errorf("order must have at least one item")
	}

	// 2. Generate unique order number
	orderNo := utils.GenerateOrderNo()

	// 3. Verify products and calculate total amount
	var totalAmount float64
	orderItems := make([]*model.OrderItem, 0, len(in.Items))

	for _, item := range in.Items {
		// Get product info from Product Service
		productResp, err := l.svcCtx.ProductRpc.GetProduct(l.ctx, &product.GetProductRequest{
			Id: item.ProductId,
		})
		if err != nil {
			l.Logger.Errorf("failed to get product %d: %v", item.ProductId, err)
			return nil, fmt.Errorf("product %d not found", item.ProductId)
		}

		// Check stock availability
		// if productResp.Product.Stock < item.Quantity {
		// 	return nil, fmt.Errorf("product %s is out of stock (available: %d, requested: %d)",
		// 		productResp.Product.Name, productResp.Product.Stock, item.Quantity)
		// }

		// Use real-time price from product service (防止前端篡改价格)
		itemPrice := productResp.Product.Price
		itemTotal := itemPrice * float64(item.Quantity)
		totalAmount += itemTotal

		// Get first image or empty string
		var imageUrl string
		if len(productResp.Product.Images) > 0 {
			imageUrl = productResp.Product.Images[0]
		}

		// Prepare order item
		orderItems = append(orderItems, &model.OrderItem{
			ProductId: item.ProductId,
			Name:      productResp.Product.Name,
			Price:     itemPrice,
			Quantity:  int(item.Quantity),
			Image:     imageUrl,
			CreatedAt: time.Now(),
		})
	}

	// 4. Start database transaction
	tx, err := l.svcCtx.OrderModel.BeginTrans(l.ctx)
	if err != nil {
		l.Logger.Errorf("failed to begin transaction: %v", err)
		return nil, fmt.Errorf("failed to create order")
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			l.Logger.Infof("transaction rolled back for order %s", orderNo)
		}
	}()

	// 5. Insert order record
	orderData := &model.Order{
		UserId:      in.UserId,
		OrderNo:     orderNo,
		TotalAmount: totalAmount,
		Status:      model.OrderStatusPending,
		Address:     in.Address,
		Phone:       in.Phone,
		Remark:      in.Remark,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	orderId, err := l.svcCtx.OrderModel.Insert(l.ctx, tx, orderData)
	if err != nil {
		l.Logger.Errorf("failed to insert order: %v", err)
		return nil, fmt.Errorf("failed to create order")
	}

	// 6. Insert order items
	for _, item := range orderItems {
		item.OrderId = orderId
	}

	err = l.svcCtx.OrderItemModel.BatchInsert(l.ctx, tx, orderItems)
	if err != nil {
		l.Logger.Errorf("failed to insert order items: %v", err)
		return nil, fmt.Errorf("failed to create order items")
	}

	// 7. Deduct stock using batch update (transactional)
	stockItems := make([]*product.StockUpdateItem, 0, len(in.Items))
	for _, item := range in.Items {
		stockItems = append(stockItems, &product.StockUpdateItem{
			ProductId: item.ProductId,
			Quantity:  -item.Quantity, // Negative to deduct
		})
	}

	_, err = l.svcCtx.ProductRpc.BatchUpdateStock(l.ctx, &product.BatchUpdateStockRequest{
		Items: stockItems,
	})
	if err != nil {
		l.Logger.Errorf("failed to batch deduct stock: %v", err)
		return nil, fmt.Errorf("failed to deduct stock")
	}

	// 8. Commit transaction
	if err = tx.Commit(); err != nil {
		l.Logger.Errorf("failed to commit transaction: %v", err)

		// Compensate the stock that was deducted, as it won't be rollback by tx
		for i := range stockItems {
			stockItems[i].Quantity = -stockItems[i].Quantity // positive to add back
		}
		_, compensateErr := l.svcCtx.ProductRpc.BatchUpdateStock(l.ctx, &product.BatchUpdateStockRequest{
			Items: stockItems,
		})
		if compensateErr != nil {
			// Critical: Stock compensation failed, need manual intervention
			l.Logger.Errorf("CRITICAL_STOCK_COMPENSATION_FAILED order=%s original_error=%v compensation_error=%v stock_items=%+v",
				orderNo, err, compensateErr, stockItems)

			// Publish compensation failure event to Kafka for async retry
			go l.publishStockCompensationFailedEvent(orderNo, in.UserId, stockItems, err.Error(), compensateErr.Error())
		} else {
			l.Logger.Infof("Stock compensation succeeded for order %s", orderNo)
		}

		return nil, fmt.Errorf("failed to create order")
	}

	l.Logger.Infof("order created successfully: %s (id: %d)", orderNo, orderId)

	// 9. Publish order created event to Kafka (异步，不影响订单创建)
	go l.publishOrderCreatedEvent(orderId, orderNo, in.UserId, totalAmount, in.Items)

	// 10. Clear user's cart (异步，不影响订单创建)
	go l.clearUserCart(in.UserId)

	return &order.CreateOrderResponse{
		OrderId:     orderId,
		OrderNo:     orderNo,
		TotalAmount: totalAmount,
	}, nil
}

// publishOrderCreatedEvent publishes order created event to Kafka
func (l *CreateOrderLogic) publishOrderCreatedEvent(orderId int64, orderNo string, userId int64, totalAmount float64, items []*order.OrderItem) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare event data
	eventItems := make([]utils.OrderItem, 0, len(items))
	for _, item := range items {
		eventItems = append(eventItems, utils.OrderItem{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
		})
	}

	event := utils.OrderCreatedEvent{
		EventType: "order.created",
		EventID:   uuid.New().String(),
		Timestamp: time.Now().Unix(),
		Data: utils.OrderData{
			OrderID:     orderId,
			OrderNo:     orderNo,
			UserID:      userId,
			TotalAmount: totalAmount,
			Items:       eventItems,
		},
	}

	// Publish to Kafka
	producer := utils.NewKafkaProducer(l.svcCtx.KafkaBrokers)
	err := producer.PublishEvent(ctx, l.svcCtx.KafkaTopics.OrderCreated, orderNo, event)
	if err != nil {
		l.Logger.Errorf("failed to publish order created event: %v", err)
		// Don't return error, just log it
	} else {
		l.Logger.Infof("published order created event for order %s", orderNo)
	}
}

// clearUserCart clears user's shopping cart
func (l *CreateOrderLogic) clearUserCart(userId int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := l.svcCtx.CartRpc.ClearCart(ctx, &cart.ClearCartRequest{
		UserId: userId,
	})
	if err != nil {
		l.Logger.Errorf("failed to clear cart for user %d: %v", userId, err)
		// Don't return error, user can manually clear cart
	} else {
		l.Logger.Infof("cleared cart for user %d", userId)
	}
}

// publishStockCompensationFailedEvent publishes stock compensation failure event to Kafka
func (l *CreateOrderLogic) publishStockCompensationFailedEvent(orderNo string, userId int64, stockItems []*product.StockUpdateItem, originalError, compensationError string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare event data
	items := make([]utils.OrderItem, 0, len(stockItems))
	for _, item := range stockItems {
		items = append(items, utils.OrderItem{
			ProductID: item.ProductId,
			Quantity:  -item.Quantity, // Convert back to original quantity (positive)
		})
	}

	event := map[string]interface{}{
		"event_type":         "stock.compensation.failed",
		"event_id":           uuid.New().String(),
		"timestamp":          time.Now().Unix(),
		"order_no":           orderNo,
		"user_id":            userId,
		"items":              items,
		"original_error":     originalError,
		"compensation_error": compensationError,
		"retry_count":        0,
		"max_retries":        10,
	}

	// Publish to Kafka (use a dedicated topic for compensation failures)
	producer := utils.NewKafkaProducer(l.svcCtx.KafkaBrokers)
	err := producer.PublishEvent(ctx, "order.stock.compensation.failed", orderNo, event)
	if err != nil {
		l.Logger.Errorf("EMERGENCY: Failed to publish stock compensation failure event for order %s: %v", orderNo, err)
		// If even Kafka fails, this is logged with EMERGENCY tag for monitoring
	} else {
		l.Logger.Infof("Published stock compensation failure event for order %s", orderNo)
	}
}
