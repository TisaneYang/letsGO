package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ProductModel = (*customProductModel)(nil)

type (
	// ProductModel is an interface for product database operations
	ProductModel interface {
		// Insert a new product
		Insert(ctx context.Context, data *Product) (sql.Result, error)

		// FindOne by product ID
		FindOne(ctx context.Context, id int64) (*Product, error)

		// Update product information
		Update(ctx context.Context, data *Product) error

		// Delete product (soft delete by setting status to 2)
		Delete(ctx context.Context, id int64) error

		// List products with pagination and filters
		List(ctx context.Context, page, pageSize int32, category, sortBy, order string) ([]*Product, int64, error)

		// Search products by keyword
		Search(ctx context.Context, keyword string, page, pageSize int32) ([]*Product, int64, error)

		// UpdateStock updates product stock (for order processing)
		UpdateStock(ctx context.Context, productId int64, quantity int64) (int64, string, error)

		// BatchUpdateStock updates multiple products' stock in a transaction
		BatchUpdateStock(ctx context.Context, items []StockUpdateItem) ([]StockUpdateResult, error)

		// CheckStock checks if products are in stock (batch check)
		CheckStock(ctx context.Context, productIds []int64) (map[int64]int64, error)

		// IncrementSales increments product sales count and returns new sales and category
		IncrementSales(ctx context.Context, productId int64, quantity int64) (int64, string, error)
	}

	customProductModel struct {
		conn sqlx.SqlConn
	}
)

// NewProductModel returns a ProductModel instance
func NewProductModel(conn sqlx.SqlConn) ProductModel {
	return &customProductModel{
		conn: conn,
	}
}

// Insert inserts a new product into database
func (m *customProductModel) Insert(ctx context.Context, data *Product) (sql.Result, error) {
	query := `INSERT INTO products (name, description, price, stock, category, images, attributes, sales, status, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			  RETURNING id`

	var id int64
	err := m.conn.QueryRowCtx(ctx, &id, query,
		data.Name,
		data.Description,
		data.Price,
		data.Stock,
		data.Category,
		data.Images,
		data.Attributes,
		data.Sales,
		data.Status,
		data.CreatedAt,
		data.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	data.Id = id
	return &insertResult{lastInsertId: id}, nil
}

// FindOne finds product by ID
func (m *customProductModel) FindOne(ctx context.Context, id int64) (*Product, error) {
	query := `SELECT id, name, description, price, stock, category, images, attributes, sales, status, created_at, updated_at
			  FROM products
			  WHERE id = $1 AND status = 1`

	var product Product
	err := m.conn.QueryRowCtx(ctx, &product, query, id)

	switch err {
	case nil:
		return &product, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Update updates product information
func (m *customProductModel) Update(ctx context.Context, data *Product) error {
	query := `UPDATE products
			  SET name = $1, description = $2, price = $3, stock = $4, category = $5,
			      images = $6, attributes = $7, updated_at = $8
			  WHERE id = $9 AND status = 1`

	result, err := m.conn.ExecCtx(ctx, query,
		data.Name,
		data.Description,
		data.Price,
		data.Stock,
		data.Category,
		data.Images,
		data.Attributes,
		data.UpdatedAt,
		data.Id,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete soft deletes product by setting status to 2
func (m *customProductModel) Delete(ctx context.Context, id int64) error {
	query := `UPDATE products SET status = 2 WHERE id = $1`
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// List returns paginated products with filters
func (m *customProductModel) List(ctx context.Context, page, pageSize int32, category, sortBy, order string) ([]*Product, int64, error) {
	// Build WHERE clause
	whereClause := "WHERE status = 1"
	args := []interface{}{}
	argPos := 1

	if category != "" {
		whereClause += fmt.Sprintf(" AND category = $%d", argPos)
		args = append(args, category)
		argPos++
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", whereClause)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause
	orderByClause := "created_at DESC" // default
	if sortBy != "" {
		validSortFields := map[string]bool{
			"price":      true,
			"created_at": true,
			"sales":      true,
		}
		if validSortFields[sortBy] {
			if order == "asc" {
				orderByClause = fmt.Sprintf("%s ASC", sortBy)
			} else {
				orderByClause = fmt.Sprintf("%s DESC", sortBy)
			}
		}
	}

	// Build pagination
	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)

	query := fmt.Sprintf(`SELECT id, name, description, price, stock, category, images, attributes, sales, status, created_at, updated_at
						  FROM products
						  %s
						  ORDER BY %s
						  LIMIT $%d OFFSET $%d`,
		whereClause, orderByClause, argPos, argPos+1)

	// Use RawDB for queries that need custom scanning
	db, err := m.conn.RawDB()
	if err != nil {
		return nil, 0, err
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		var product Product

		err := rows.Scan(
			&product.Id,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.Category,
			&product.Images,
			&product.Attributes,
			&product.Sales,
			&product.Status,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		products = append(products, &product)
	}

	return products, total, nil
}

// Search searches products by keyword in name and description
func (m *customProductModel) Search(ctx context.Context, keyword string, page, pageSize int32) ([]*Product, int64, error) {
	// Build search condition
	searchPattern := "%" + strings.ToLower(keyword) + "%"

	// Count total matching records
	countQuery := `SELECT COUNT(*) FROM products
				   WHERE status = 1 AND (LOWER(name) LIKE $1 OR LOWER(description) LIKE $1)`

	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, searchPattern)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	query := `SELECT id, name, description, price, stock, category, images, attributes, sales, status, created_at, updated_at
			  FROM products
			  WHERE status = 1 AND (LOWER(name) LIKE $1 OR LOWER(description) LIKE $1)
			  ORDER BY sales DESC, created_at DESC
			  LIMIT $2 OFFSET $3`

	// Use RawDB for queries that need custom scanning
	db, err := m.conn.RawDB()
	if err != nil {
		return nil, 0, err
	}

	rows, err := db.QueryContext(ctx, query, searchPattern, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		var product Product

		err := rows.Scan(
			&product.Id,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.Category,
			&product.Images,
			&product.Attributes,
			&product.Sales,
			&product.Status,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		products = append(products, &product)
	}

	return products, total, nil
}

// UpdateStock updates product stock atomically
// quantity can be positive (increase) or negative (decrease)
func (m *customProductModel) UpdateStock(ctx context.Context, productId int64, quantity int64) (int64, string, error) {
	query := `UPDATE products
			  SET stock = stock + $1, updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT
			  WHERE id = $2 AND status = 1 AND stock + $1 >= 0
			  RETURNING stock, category`
	db, err := m.conn.RawDB()
	if err != nil {
		return 0, "", err
	}

	var newStock int64
	var category string
	err = db.QueryRowContext(ctx, query, quantity, productId).Scan(&newStock, &category)
	if err == sql.ErrNoRows {
		return 0, "", ErrNotFound
	}

	if err != nil {
		return 0, "", err
	}

	return newStock, category, nil
}

// CheckStock checks stock availability for multiple products
func (m *customProductModel) CheckStock(ctx context.Context, productIds []int64) (map[int64]int64, error) {
	if len(productIds) == 0 {
		return map[int64]int64{}, nil
	}

	query := `SELECT id, stock FROM products WHERE id = ANY($1) AND status = 1`

	// Use RawDB for custom query
	db, err := m.conn.RawDB()
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query, pq.Array(productIds))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stockMap := make(map[int64]int64)
	for rows.Next() {
		var id, stock int64
		if err := rows.Scan(&id, &stock); err != nil {
			return nil, err
		}
		stockMap[id] = stock
	}

	return stockMap, nil
}

// IncrementSales increments product sales count and returns new sales and category
func (m *customProductModel) IncrementSales(ctx context.Context, productId int64, quantity int64) (int64, string, error) {
	query := `UPDATE products
			  SET sales = sales + $1, updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT
			  WHERE id = $2 AND status = 1
			  RETURNING sales, category`

	// Use RawDB to get the underlying database connection
	db, err := m.conn.RawDB()
	if err != nil {
		return 0, "", err
	}

	var newSales int64
	var category string

	// QueryRowContext + Scan for multiple return fields
	err = db.QueryRowContext(ctx, query, quantity, productId).Scan(&newSales, &category)

	if err == sql.ErrNoRows {
		return 0, "", ErrNotFound
	}

	if err != nil {
		return 0, "", err
	}

	return newSales, category, nil
}

// StockUpdateItem represents a single stock update operation
type StockUpdateItem struct {
	ProductId int64
	Quantity  int64
}

// StockUpdateResult represents the result of a stock update
type StockUpdateResult struct {
	ProductId int64
	NewStock  int64
}

// BatchUpdateStock updates multiple products' stock in a single transaction
// All updates succeed or all fail (atomic operation)
func (m *customProductModel) BatchUpdateStock(ctx context.Context, items []StockUpdateItem) ([]StockUpdateResult, error) {
	if len(items) == 0 {
		return []StockUpdateResult{}, nil
	}

	// Get raw database connection for transaction
	db, err := m.conn.RawDB()
	if err != nil {
		return nil, err
	}

	// Start transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	results := make([]StockUpdateResult, 0, len(items))

	// Update each product's stock
	query := `UPDATE products
			  SET stock = stock + $1, updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT
			  WHERE id = $2 AND status = 1 AND stock + $1 >= 0
			  RETURNING stock`

	for _, item := range items {
		var newStock int64
		err = tx.QueryRowContext(ctx, query, item.Quantity, item.ProductId).Scan(&newStock)

		if err == sql.ErrNoRows {
			// Product not found or insufficient stock
			return nil, ErrInsufficientStock
		}

		if err != nil {
			return nil, err
		}

		results = append(results, StockUpdateResult{
			ProductId: item.ProductId,
			NewStock:  newStock,
		})
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return results, nil
}

// ErrNotFound is returned when a product is not found
var ErrNotFound = sqlx.ErrNotFound

// ErrInsufficientStock is returned when stock is insufficient
var ErrInsufficientStock = fmt.Errorf("insufficient stock")

// insertResult implements sql.Result for Insert operation
type insertResult struct {
	lastInsertId int64
}

func (r *insertResult) LastInsertId() (int64, error) {
	return r.lastInsertId, nil
}

func (r *insertResult) RowsAffected() (int64, error) {
	return 1, nil
}
