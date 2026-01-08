package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PaymentModel = (*customPaymentModel)(nil)

type (
	// PaymentModel is an interface for payment operations
	PaymentModel interface {
		// Insert a new payment
		Insert(ctx context.Context, data *Payment) (int64, error)

		// FindOne by payment ID
		FindOne(ctx context.Context, id int64) (*Payment, error)

		// FindOneByPaymentNo finds payment by payment number
		FindOneByPaymentNo(ctx context.Context, paymentNo string) (*Payment, error)

		// FindOneByOrderId finds payment by order ID
		FindOneByOrderId(ctx context.Context, orderId int64) (*Payment, error)

		// FindByUserId finds payments by user ID with pagination
		FindByUserId(ctx context.Context, userId int64, page, pageSize int, status int) ([]*Payment, error)

		// CountByUserId counts total payments for a user
		CountByUserId(ctx context.Context, userId int64, status int) (int64, error)

		// UpdateStatus updates payment status
		UpdateStatus(ctx context.Context, id int64, status int, tradeNo string, paidAt time.Time) error

		// CancelPayment cancels a payment (only if status is pending)
		CancelPayment(ctx context.Context, id int64) error
	}

	customPaymentModel struct {
		conn sqlx.SqlConn
	}
)

// NewPaymentModel returns a PaymentModel instance
func NewPaymentModel(conn sqlx.SqlConn) PaymentModel {
	return &customPaymentModel{
		conn: conn,
	}
}

// Insert inserts a new payment into database
func (m *customPaymentModel) Insert(ctx context.Context, data *Payment) (int64, error) {
	query := `INSERT INTO payments (order_id, user_id, payment_no, amount, payment_type, status, trade_no, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	var id int64
	err := m.conn.QueryRowCtx(ctx, &id, query,
		data.OrderId,
		data.UserId,
		data.PaymentNo,
		data.Amount,
		data.PaymentType,
		data.Status,
		data.TradeNo,
		data.CreatedAt,
		data.UpdatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to insert payment: %w", err)
	}

	return id, nil
}

// FindOne finds a payment by ID
func (m *customPaymentModel) FindOne(ctx context.Context, id int64) (*Payment, error) {
	query := `SELECT id, order_id, user_id, payment_no, amount, payment_type, status, trade_no,
		created_at, updated_at, paid_at
		FROM payments WHERE id = $1`

	var payment Payment
	err := m.conn.QueryRowCtx(ctx, &payment, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return &payment, nil
}

// FindOneByPaymentNo finds a payment by payment number
func (m *customPaymentModel) FindOneByPaymentNo(ctx context.Context, paymentNo string) (*Payment, error) {
	query := `SELECT id, order_id, user_id, payment_no, amount, payment_type, status, trade_no,
		created_at, updated_at, paid_at
		FROM payments WHERE payment_no = $1`

	var payment Payment
	err := m.conn.QueryRowCtx(ctx, &payment, query, paymentNo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return &payment, nil
}

// FindOneByOrderId finds a payment by order ID
func (m *customPaymentModel) FindOneByOrderId(ctx context.Context, orderId int64) (*Payment, error) {
	query := `SELECT id, order_id, user_id, payment_no, amount, payment_type, status, trade_no,
		created_at, updated_at, paid_at
		FROM payments WHERE order_id = $1`

	var payment Payment
	err := m.conn.QueryRowCtx(ctx, &payment, query, orderId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return &payment, nil
}

// FindByUserId finds payments by user ID with pagination and optional status filter
func (m *customPaymentModel) FindByUserId(ctx context.Context, userId int64, page, pageSize int, status int) ([]*Payment, error) {
	offset := (page - 1) * pageSize

	var query string
	var args []interface{}

	if status == 0 {
		// Get all payments
		query = `SELECT id, order_id, user_id, payment_no, amount, payment_type, status, trade_no,
			created_at, updated_at, paid_at
			FROM payments WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`
		args = []interface{}{userId, pageSize, offset}
	} else {
		// Filter by status
		query = `SELECT id, order_id, user_id, payment_no, amount, payment_type, status, trade_no,
			created_at, updated_at, paid_at
			FROM payments WHERE user_id = $1 AND status = $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4`
		args = []interface{}{userId, status, pageSize, offset}
	}

	var payments []*Payment
	err := m.conn.QueryRowsCtx(ctx, &payments, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find payments: %w", err)
	}

	return payments, nil
}

// CountByUserId counts total payments for a user
func (m *customPaymentModel) CountByUserId(ctx context.Context, userId int64, status int) (int64, error) {
	var query string
	var args []interface{}

	if status == 0 {
		query = `SELECT COUNT(*) FROM payments WHERE user_id = $1`
		args = []interface{}{userId}
	} else {
		query = `SELECT COUNT(*) FROM payments WHERE user_id = $1 AND status = $2`
		args = []interface{}{userId, status}
	}

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to count payments: %w", err)
	}

	return count, nil
}

// UpdateStatus updates payment status and related fields
func (m *customPaymentModel) UpdateStatus(ctx context.Context, id int64, status int, tradeNo string, paidAt time.Time) error {
	var query string
	var args []interface{}

	if status == PaymentStatusSuccess {
		// Update status, trade_no, and paid_at for successful payment
		query = `UPDATE payments SET status = $1, trade_no = $2, paid_at = $3, updated_at = $4 WHERE id = $5`
		args = []interface{}{status, tradeNo, paidAt, time.Now(), id}
	} else {
		// Update only status for other cases
		query = `UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3`
		args = []interface{}{status, time.Now(), id}
	}

	_, err := m.conn.ExecCtx(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

// CancelPayment cancels a payment (only if status is pending)
func (m *customPaymentModel) CancelPayment(ctx context.Context, id int64) error {
	query := `UPDATE payments SET status = $1, updated_at = $2
		WHERE id = $3 AND status = $4`

	result, err := m.conn.ExecCtx(ctx, query,
		PaymentStatusCancelled,
		time.Now(),
		id,
		PaymentStatusPending,
	)

	if err != nil {
		return fmt.Errorf("failed to cancel payment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("payment not found or cannot be cancelled")
	}

	return nil
}
