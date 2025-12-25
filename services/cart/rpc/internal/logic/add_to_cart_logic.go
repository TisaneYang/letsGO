package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"letsgo/common/errorx"
	"letsgo/services/cart/rpc/cart"
	"letsgo/services/cart/rpc/internal/svc"
	"letsgo/services/product/rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddToCartLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddToCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddToCartLogic {
	return &AddToCartLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CartItemData represents the cart item stored in Redis
type CartItemData struct {
	ProductId int64   `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int64   `json:"quantity"`
	Image     string  `json:"image"`
	AddedAt   int64   `json:"added_at"`
}

// Add product to cart
func (l *AddToCartLogic) AddToCart(in *cart.AddToCartRequest) (*cart.AddToCartResponse, error) {
	// 1. Validate input parameters
	if err := l.validateParams(in); err != nil {
		return nil, err
	}

	// 2. Get product information from Product service
	productInfo, err := l.svcCtx.ProductRpc.GetProduct(l.ctx, &product.GetProductRequest{
		Id: in.ProductId,
	})
	if err != nil {
		l.Logger.Errorf("Failed to get product info: product_id=%d, err=%v", in.ProductId, err)
		return nil, errorx.ErrProductNotFound
	}

	// 3. Prepare cart item data
	cartItem := &CartItemData{
		ProductId: in.ProductId,
		Name:      productInfo.Product.Name,
		Price:     productInfo.Product.Price,
		Quantity:  in.Quantity,
		Image:     getFirstImage(productInfo.Product.Images),
		AddedAt:   time.Now().Unix(),
	}

	itemJSON, err := json.Marshal(cartItem)
	if err != nil {
		l.Logger.Errorf("Failed to marshal cart item: %v", err)
		return nil, errorx.ErrSystem
	}

	// 4. Add to cart using Lua script (atomic operation)
	cartKey := fmt.Sprintf("cart:user:%d", in.UserId)
	productField := fmt.Sprintf("product:%d", in.ProductId)

	// Lua script for atomic add to cart
	script := `
		local cart_key = KEYS[1]
		local product_field = ARGV[1]
		local add_qty = tonumber(ARGV[2])
		local max_qty = tonumber(ARGV[3])
		local max_items = tonumber(ARGV[4])
		local item_data = ARGV[5]
		local expire_time = tonumber(ARGV[6])

		-- Get current item
		local current_item = redis.call('HGET', cart_key, product_field)
		local current_qty = 0

		if current_item then
			-- Item exists, decode and get quantity
			local item = cjson.decode(current_item)
			current_qty = item.quantity
		else
			-- New item, check if cart is full
			local count = redis.call('HLEN', cart_key)
			if count >= max_items then
				return redis.error_reply('CART_FULL')
			end
		end

		-- Check quantity limit
		local new_qty = current_qty + add_qty
		if new_qty > max_qty then
			return redis.error_reply('QUANTITY_LIMIT_EXCEEDED')
		end

		-- Update item data with new quantity
		local new_item = cjson.decode(item_data)
		new_item.quantity = new_qty
		local new_item_json = cjson.encode(new_item)

		-- Save to Redis
		redis.call('HSET', cart_key, product_field, new_item_json)
		redis.call('EXPIRE', cart_key, expire_time)

		return new_qty
	`

	result, err := l.svcCtx.Redis.EvalCtx(l.ctx, script,
		[]string{cartKey},
		productField,
		in.Quantity,
		l.svcCtx.Config.Cart.MaxQuantityPerItem,
		l.svcCtx.Config.Cart.MaxItems,
		string(itemJSON),
		l.svcCtx.Config.Cart.Expire,
	)

	if err != nil {
		errMsg := err.Error()
		if errMsg == "CART_FULL" {
			return nil, errorx.NewCodeError(4002, fmt.Sprintf("Cart is full, maximum %d items allowed", l.svcCtx.Config.Cart.MaxItems))
		}
		if errMsg == "QUANTITY_LIMIT_EXCEEDED" {
			return nil, errorx.NewCodeError(4003, fmt.Sprintf("Quantity limit exceeded, maximum %d per item", l.svcCtx.Config.Cart.MaxQuantityPerItem))
		}
		l.Logger.Errorf("Failed to add to cart: user_id=%d, product_id=%d, err=%v", in.UserId, in.ProductId, err)
		return nil, errorx.ErrCache
	}

	l.Logger.Infof("Added to cart successfully: user_id=%d, product_id=%d, quantity=%d, new_total=%v",
		in.UserId, in.ProductId, in.Quantity, result)

	return &cart.AddToCartResponse{
		Success: true,
	}, nil
}

// validateParams validates input parameters
func (l *AddToCartLogic) validateParams(in *cart.AddToCartRequest) error {
	if in.UserId <= 0 {
		return errorx.NewCodeError(1001, "Invalid user ID")
	}
	if in.ProductId <= 0 {
		return errorx.NewCodeError(1001, "Invalid product ID")
	}
	if in.Quantity <= 0 {
		return errorx.NewCodeError(1001, "Quantity must be greater than 0")
	}
	if in.Quantity > int64(l.svcCtx.Config.Cart.MaxQuantityPerItem) {
		return errorx.NewCodeError(1001, fmt.Sprintf("Quantity cannot exceed %d", l.svcCtx.Config.Cart.MaxQuantityPerItem))
	}
	return nil
}

// getFirstImage returns the first image URL or empty string
func getFirstImage(images []string) string {
	if len(images) > 0 {
		return images[0]
	}
	return ""
}
