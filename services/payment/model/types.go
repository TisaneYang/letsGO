package model

import (
	"database/sql"
	"time"
)

// Payment status constants
const (
	PaymentStatusPending   = 1 // 待支付
	PaymentStatusSuccess   = 2 // 成功
	PaymentStatusFailed    = 3 // 失败
	PaymentStatusCancelled = 4 // 已取消
)

// Payment type constants
const (
	PaymentTypeAlipay     = 1 // 支付宝
	PaymentTypeWechat     = 2 // 微信支付
	PaymentTypeCreditCard = 3 // 信用卡
)

// Payment represents a payment record
type Payment struct {
	Id          int64        `db:"id"`
	OrderId     int64        `db:"order_id"`
	UserId      int64        `db:"user_id"`
	PaymentNo   string       `db:"payment_no"`
	Amount      float64      `db:"amount"`
	PaymentType int          `db:"payment_type"`
	Status      int          `db:"status"`
	TradeNo     string       `db:"trade_no"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
	PaidAt      sql.NullTime `db:"paid_at"`
}
