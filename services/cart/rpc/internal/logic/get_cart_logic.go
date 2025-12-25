package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"letsgo/common/errorx"
	"letsgo/services/cart/rpc/cart"
	"letsgo/services/cart/rpc/internal/svc"
	"letsgo/services/product/rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCartLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCartLogic {
	return &GetCartLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Get user's cart
func (l *GetCartLogic) GetCart(in *cart.GetCartRequest) (*cart.GetCartResponse, error) {
	// 1. Validate input
	if in.UserId <= 0 {
		return nil, errorx.NewCodeError(1001, "Invalid user ID")
	}

	// 2. Get all cart items from Redis
	cartKey := fmt.Sprintf("cart:user:%d", in.UserId)
	items, err := l.svcCtx.Redis.HgetallCtx(l.ctx, cartKey)
	if err != nil {
		l.Logger.Errorf("Failed to get cart: user_id=%d, err=%v", in.UserId, err)
		return nil, errorx.ErrCache
	}

	// 3. If cart is empty, return empty response
	if len(items) == 0 {
		return &cart.GetCartResponse{
			Items:      []*cart.CartItem{},
			TotalPrice: 0,
			TotalCount: 0,
		}, nil
	}

	// 4. Parse cart items and collect product IDs
	var cartItems []*cart.CartItem
	var totalPrice float64
	var totalCount int64
	productIds := make([]int64, 0, len(items))

	for _, itemJSON := range items {
		var item CartItemData
		if err := json.Unmarshal([]byte(itemJSON), &item); err != nil {
			l.Logger.Errorf("Failed to unmarshal cart item: %v", err)
			continue
		}

		productIds = append(productIds, item.ProductId)

		cartItems = append(cartItems, &cart.CartItem{
			ProductId: item.ProductId,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  item.Quantity,
			Image:     item.Image,
			Stock:     0,    // Will be filled later
			Available: true, // Will be filled later
		})

		totalPrice += item.Price * float64(item.Quantity)
		totalCount += item.Quantity
	}

	// 5. Batch check product stock and availability
	if len(productIds) > 0 {
		l.updateCartItemsStock(cartItems, productIds)
	}

	// 6. Refresh cart expiration time
	l.svcCtx.Redis.ExpireCtx(l.ctx, cartKey, l.svcCtx.Config.Cart.Expire)

	l.Logger.Infof("Get cart successfully: user_id=%d, items=%d, total_price=%.2f",
		in.UserId, len(cartItems), totalPrice)

	return &cart.GetCartResponse{
		Items:      cartItems,
		TotalPrice: totalPrice,
		TotalCount: totalCount,
	}, nil
}

// updateCartItemsStock updates stock and availability for cart items
func (l *GetCartLogic) updateCartItemsStock(cartItems []*cart.CartItem, productIds []int64) {
	// Build stock check request
	stockItems := make([]*product.StockItem, len(productIds))
	for i, pid := range productIds {
		stockItems[i] = &product.StockItem{
			ProductId:        pid,
			RequiredQuantity: 0, // We just want to know current stock, not checking if available
		}
	}

	// Call Product RPC to check stock
	stockResp, err := l.svcCtx.ProductRpc.CheckStock(l.ctx, &product.CheckStockRequest{
		Items: stockItems,
	})
	if err != nil {
		l.Logger.Errorf("Failed to check stock: err=%v", err)
		// If failed, just return without stock info
		return
	}

	// Create a map for quick lookup
	stockMap := make(map[int64]int64)
	for _, item := range stockResp.Items {
		stockMap[item.ProductId] = item.AvailableStock
	}

	// Update cart items with stock info
	for _, cartItem := range cartItems {
		if stock, exists := stockMap[cartItem.ProductId]; exists {
			cartItem.Stock = stock
			cartItem.Available = stock > 0
		} else {
			cartItem.Stock = 0
			cartItem.Available = false
		}
	}
}
