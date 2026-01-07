package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ OrderModel = (*customOrderModel)(nil)

type (
	// OrderModel is an interface for order operations
	OrderModel interface {
		// Insert a new order (with transaction support)
		Insert(ctx context.Context, tx *sql.Tx, data *Order) (int64, error)

		// FindOne by order ID
		FindOne(ctx context.Context, id int64) (*Order, error)

		// FindOneByOrderNo finds order by order number
		FindOneByOrderNo(ctx context.Context, orderNo string) (*Order, error)

		// FindByUserId finds orders by user ID with pagination
		FindByUserId(ctx context.Context, userId int64, page, pageSize int, status int) ([]*Order, error)

		// CountByUserId counts total orders for a user
		CountByUserId(ctx context.Context, userId int64, status int) (int64, error)

		// UpdateStatus updates order status
		UpdateStatus(ctx context.Context, id int64, status int, timestamp time.Time) error

		// CancelOrder cancels an order (only if status is pending)
		CancelOrder(ctx context.Context, id int64, userId int64) error

		// BeginTrans starts a transaction
		BeginTrans(ctx context.Context) (*sql.Tx, error)
	}

	customOrderModel struct {
		conn sqlx.SqlConn
	}
)

// NewOrderModel returns an OrderModel instance
func NewOrderModel(conn sqlx.SqlConn) OrderModel {
	return &customOrderModel{
		conn: conn,
	}
}

// Insert inserts a new order into database (with transaction)
func (m *customOrderModel) Insert(ctx context.Context, tx *sql.Tx, data *Order) (int64, error) {
	query := `INSERT INTO orders (user_id, order_no, total_amount, status, address, phone, remark, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	var id int64
	err := tx.QueryRowContext(ctx, query,
		data.UserId,
		data.OrderNo,
		data.TotalAmount,
		data.Status,
		data.Address,
		data.Phone,
		data.Remark,
		data.CreatedAt,
		data.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert order: %w", err)
	}

	return id, nil
}

// FindOne finds an order by ID
func (m *customOrderModel) FindOne(ctx context.Context, id int64) (*Order, error) {
	query := `SELECT id, user_id, order_no, total_amount, status, address, phone, remark,
		created_at, updated_at, paid_at, shipped_at, completed_at
		FROM orders WHERE id = $1`

	var order Order
	err := m.conn.QueryRowCtx(ctx, &order, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to find order: %w", err)
	}

	return &order, nil
}

// FindOneByOrderNo finds an order by order number
func (m *customOrderModel) FindOneByOrderNo(ctx context.Context, orderNo string) (*Order, error) {
	query := `SELECT id, user_id, order_no, total_amount, status, address, phone, remark,
		created_at, updated_at, paid_at, shipped_at, completed_at
		FROM orders WHERE order_no = $1`

	var order Order
	err := m.conn.QueryRowCtx(ctx, &order, query, orderNo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to find order: %w", err)
	}

	return &order, nil
}

// FindByUserId finds orders by user ID with pagination and optional status filter
func (m *customOrderModel) FindByUserId(ctx context.Context, userId int64, page, pageSize int, status int) ([]*Order, error) {
	offset := (page - 1) * pageSize

	var query string
	var args []interface{}

	if status == 0 {
		// Get all orders
		query = `SELECT id, user_id, order_no, total_amount, status, address, phone, remark,
			created_at, updated_at, paid_at, shipped_at, completed_at
			FROM orders WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`
		args = []interface{}{userId, pageSize, offset}
	} else {
		// Filter by status
		query = `SELECT id, user_id, order_no, total_amount, status, address, phone, remark,
			created_at, updated_at, paid_at, shipped_at, completed_at
			FROM orders WHERE user_id = $1 AND status = $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4`
		args = []interface{}{userId, status, pageSize, offset}
	}

	var orders []*Order
	err := m.conn.QueryRowsCtx(ctx, &orders, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders: %w", err)
	}

	return orders, nil
}

// CountByUserId counts total orders for a user
func (m *customOrderModel) CountByUserId(ctx context.Context, userId int64, status int) (int64, error) {
	var query string
	var args []interface{}

	if status == 0 {
		query = `SELECT COUNT(*) FROM orders WHERE user_id = $1`
		args = []interface{}{userId}
	} else {
		query = `SELECT COUNT(*) FROM orders WHERE user_id = $1 AND status = $2`
		args = []interface{}{userId, status}
	}

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to count orders: %w", err)
	}

	return count, nil
}

// UpdateStatus updates order status and corresponding timestamp
func (m *customOrderModel) UpdateStatus(ctx context.Context, id int64, status int, timestamp time.Time) error {
	var query string

	switch status {
	case OrderStatusPaid:
		query = `UPDATE orders SET status = $1, paid_at = $2, updated_at = $3 WHERE id = $4`
	case OrderStatusShipped:
		query = `UPDATE orders SET status = $1, shipped_at = $2, updated_at = $3 WHERE id = $4`
	case OrderStatusCompleted:
		query = `UPDATE orders SET status = $1, completed_at = $2, updated_at = $3 WHERE id = $4`
	case OrderStatusCancelled:
		query = `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`
		_, err := m.conn.ExecCtx(ctx, query, status, timestamp, id)
		return err
	default:
		query = `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`
		_, err := m.conn.ExecCtx(ctx, query, status, timestamp, id)
		return err
	}

	_, err := m.conn.ExecCtx(ctx, query, status, timestamp, timestamp, id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

// CancelOrder cancels an order (only if status is pending)
func (m *customOrderModel) CancelOrder(ctx context.Context, id int64, userId int64) error {
	query := `UPDATE orders SET status = $1, updated_at = $2
		WHERE id = $3 AND user_id = $4 AND status = $5`

	result, err := m.conn.ExecCtx(ctx, query,
		OrderStatusCancelled,
		time.Now(),
		id,
		userId,
		OrderStatusPending,
	)

	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found or cannot be cancelled")
	}

	return nil
}

// BeginTrans starts a database transaction
func (m *customOrderModel) BeginTrans(ctx context.Context) (*sql.Tx, error) {
	rawdb, err := m.conn.RawDB()
	if err != nil {
		return nil, err
	}
	return rawdb.BeginTx(ctx, nil)
}
