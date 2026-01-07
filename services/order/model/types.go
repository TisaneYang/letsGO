package model

import (
	"database/sql"
	"time"
)

// Order represents an order in the database
type Order struct {
	Id          int64        `db:"id"`
	UserId      int64        `db:"user_id"`
	OrderNo     string       `db:"order_no"`
	TotalAmount float64      `db:"total_amount"`
	Status      int          `db:"status"` // 1:pending, 2:paid, 3:shipped, 4:completed, 5:cancelled
	Address     string       `db:"address"`
	Phone       string       `db:"phone"`
	Remark      string       `db:"remark"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
	PaidAt      sql.NullTime `db:"paid_at"`
	ShippedAt   sql.NullTime `db:"shipped_at"`
	CompletedAt sql.NullTime `db:"completed_at"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	Id        int64     `db:"id"`
	OrderId   int64     `db:"order_id"`
	ProductId int64     `db:"product_id"`
	Name      string    `db:"name"`
	Price     float64   `db:"price"`
	Quantity  int       `db:"quantity"`
	Image     string    `db:"image"`
	CreatedAt time.Time `db:"created_at"`
}

// Order Status Constants
const (
	OrderStatusPending   = 1 // 待支付
	OrderStatusPaid      = 2 // 已支付
	OrderStatusShipped   = 3 // 已发货
	OrderStatusCompleted = 4 // 已完成
	OrderStatusCancelled = 5 // 已取消
)
