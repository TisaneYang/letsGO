package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ OrderItemModel = (*customOrderItemModel)(nil)

type (
	// OrderItemModel is an interface for order item operations
	OrderItemModel interface {
		// BatchInsert inserts multiple order items (with transaction)
		BatchInsert(ctx context.Context, tx *sql.Tx, items []*OrderItem) error

		// FindByOrderId finds all items for an order
		FindByOrderId(ctx context.Context, orderId int64) ([]*OrderItem, error)

		// DeleteByOrderId deletes all items for an order
		DeleteByOrderId(ctx context.Context, orderId int64) error
	}

	customOrderItemModel struct {
		conn sqlx.SqlConn
	}
)

// NewOrderItemModel returns an OrderItemModel instance
func NewOrderItemModel(conn sqlx.SqlConn) OrderItemModel {
	return &customOrderItemModel{
		conn: conn,
	}
}

// BatchInsert inserts multiple order items in a transaction
func (m *customOrderItemModel) BatchInsert(ctx context.Context, tx *sql.Tx, items []*OrderItem) error {
	if len(items) == 0 {
		return nil
	}

	query := `INSERT INTO order_items (order_id, product_id, name, price, quantity, image, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		_, err := stmt.ExecContext(ctx,
			item.OrderId,
			item.ProductId,
			item.Name,
			item.Price,
			item.Quantity,
			item.Image,
			item.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	return nil
}

// FindByOrderId finds all items for a specific order
func (m *customOrderItemModel) FindByOrderId(ctx context.Context, orderId int64) ([]*OrderItem, error) {
	query := `SELECT id, order_id, product_id, name, price, quantity, image, created_at
		FROM order_items WHERE order_id = $1
		ORDER BY id ASC`

	var items []*OrderItem
	err := m.conn.QueryRowsCtx(ctx, &items, query, orderId)
	if err != nil {
		return nil, fmt.Errorf("failed to find order items: %w", err)
	}

	return items, nil
}

// DeleteByOrderId deletes all items for an order
func (m *customOrderItemModel) DeleteByOrderId(ctx context.Context, orderId int64) error {
	query := `DELETE FROM order_items WHERE order_id = $1`

	_, err := m.conn.ExecCtx(ctx, query, orderId)
	if err != nil {
		return fmt.Errorf("failed to delete order items: %w", err)
	}

	return nil
}
