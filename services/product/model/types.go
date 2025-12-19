package model

import "github.com/lib/pq"

// Product represents the products table in PostgreSQL
// All product data including images and attributes are stored in PostgreSQL
type Product struct {
	Id          int64           `db:"id"`
	Name        string          `db:"name"`
	Description string          `db:"description"`
	Price       float64         `db:"price"`
	Stock       int64           `db:"stock"`
	Category    string          `db:"category"`
	Images      pq.StringArray  `db:"images"`      // PostgreSQL array
	Attributes  string          `db:"attributes"`  // JSON string
	Sales       int64           `db:"sales"`       // Total sales count
	Status      int64           `db:"status"`      // 1:active, 2:inactive
	CreatedAt   int64           `db:"created_at"`  // Unix timestamp
	UpdatedAt   int64           `db:"updated_at"`
}
