package logic

import (
	"context"
	"fmt"

	"letsgo/common/errorx"
	"letsgo/services/cart/rpc/cart"
	"letsgo/services/cart/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCartItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCartItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCartItemLogic {
	return &UpdateCartItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Update cart item quantity
func (l *UpdateCartItemLogic) UpdateCartItem(in *cart.UpdateCartItemRequest) (*cart.UpdateCartItemResponse, error) {
	// 1. Validate input
	if in.UserId <= 0 {
		return nil, errorx.NewCodeError(1001, "Invalid user ID")
	}
	if in.ProductId <= 0 {
		return nil, errorx.NewCodeError(1001, "Invalid product ID")
	}
	if in.Quantity <= 0 || in.Quantity > int64(l.svcCtx.Config.Cart.MaxQuantityPerItem) {
		return nil, errorx.NewCodeError(1001, fmt.Sprintf("Quantity must be between 1 and %d", l.svcCtx.Config.Cart.MaxQuantityPerItem))
	}

	// 2. Update quantity using Lua script (atomic operation)
	cartKey := fmt.Sprintf("cart:user:%d", in.UserId)
	productField := fmt.Sprintf("product:%d", in.ProductId)

	// Lua script for atomic update
	script := `
		local cart_key = KEYS[1]
		local product_field = ARGV[1]
		local new_qty = tonumber(ARGV[2])
		local max_qty = tonumber(ARGV[3])
		local expire_time = tonumber(ARGV[4])

		-- Check if item exists
		local current_item = redis.call('HGET', cart_key, product_field)
		if not current_item then
			return redis.error_reply('ITEM_NOT_FOUND')
		end

		-- Check quantity limit
		if new_qty > max_qty then
			return redis.error_reply('QUANTITY_LIMIT_EXCEEDED')
		end

		-- Update item data with new quantity
		local item = cjson.decode(current_item)
		item.quantity = new_qty
		local new_item_json = cjson.encode(item)

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
		l.svcCtx.Config.Cart.Expire,
	)

	if err != nil {
		errMsg := err.Error()
		if errMsg == "ITEM_NOT_FOUND" {
			return nil, errorx.ErrCartItemNotFound
		}
		if errMsg == "QUANTITY_LIMIT_EXCEEDED" {
			return nil, errorx.NewCodeError(4003, fmt.Sprintf("Quantity limit exceeded, maximum %d per item", l.svcCtx.Config.Cart.MaxQuantityPerItem))
		}
		l.Logger.Errorf("Failed to update cart item: user_id=%d, product_id=%d, err=%v", in.UserId, in.ProductId, err)
		return nil, errorx.ErrCache
	}

	l.Logger.Infof("Updated cart item successfully: user_id=%d, product_id=%d, new_quantity=%v",
		in.UserId, in.ProductId, result)

	return &cart.UpdateCartItemResponse{
		Success: true,
	}, nil
}
