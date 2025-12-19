-- ========================================
-- Product Service Database Schema
-- ========================================
-- This file contains the database schema for the Product service
-- Execute this file in PostgreSQL to create the products table

CREATE DATABASE letsgo_product

-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    description TEXT,
    price       DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    stock       BIGINT NOT NULL DEFAULT 0 CHECK (stock >= 0),
    category    VARCHAR(100) NOT NULL,
    images      TEXT[],                      -- Array of image URLs
    attributes  TEXT,                        -- JSON string of product attributes
    sales       BIGINT NOT NULL DEFAULT 0,   -- Total sales count
    status      INT NOT NULL DEFAULT 1,      -- 1:active, 2:inactive
    created_at  BIGINT NOT NULL,             -- Unix timestamp
    updated_at  BIGINT NOT NULL,             -- Unix timestamp

    -- Indexes for common queries
    CONSTRAINT products_name_not_empty CHECK (LENGTH(TRIM(name)) > 0),
    CONSTRAINT products_category_not_empty CHECK (LENGTH(TRIM(category)) > 0)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_products_sales ON products(sales DESC);
CREATE INDEX IF NOT EXISTS idx_products_price ON products(price);
CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);
CREATE INDEX IF NOT EXISTS idx_products_name ON products USING gin(to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_products_description ON products USING gin(to_tsvector('english', description));

-- Add comments for documentation
COMMENT ON TABLE products IS 'Product catalog table storing core product information';
COMMENT ON COLUMN products.id IS 'Primary key, auto-incrementing product ID';
COMMENT ON COLUMN products.name IS 'Product name';
COMMENT ON COLUMN products.description IS 'Product description';
COMMENT ON COLUMN products.price IS 'Product price in decimal format';
COMMENT ON COLUMN products.stock IS 'Available inventory quantity';
COMMENT ON COLUMN products.category IS 'Product category for filtering';
COMMENT ON COLUMN products.images IS 'Array of product image URLs';
COMMENT ON COLUMN products.attributes IS 'JSON string of product attributes (color, size, etc.)';
COMMENT ON COLUMN products.sales IS 'Total number of units sold';
COMMENT ON COLUMN products.status IS '1=active (available for sale), 2=inactive (hidden)';
COMMENT ON COLUMN products.created_at IS 'Creation timestamp (Unix epoch)';
COMMENT ON COLUMN products.updated_at IS 'Last update timestamp (Unix epoch)';

-- Insert sample data for testing
INSERT INTO products (name, description, price, stock, category, images, attributes, sales, status, created_at, updated_at)
VALUES
    ('iPhone 15 Pro', 'Latest Apple smartphone with A17 Pro chip', 999.99, 100, 'Electronics',
     ARRAY['https://example.com/iphone15-1.jpg', 'https://example.com/iphone15-2.jpg'],
     '{"color": ["Black", "White", "Blue"], "storage": ["128GB", "256GB", "512GB"]}',
     0, 1, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),

    ('MacBook Pro 16"', 'Powerful laptop for professionals', 2499.99, 50, 'Electronics',
     ARRAY['https://example.com/macbook-1.jpg'],
     '{"color": ["Silver", "Space Gray"], "memory": ["16GB", "32GB", "64GB"]}',
     0, 1, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),

    ('Nike Air Max', 'Comfortable running shoes', 129.99, 200, 'Shoes',
     ARRAY['https://example.com/nike-1.jpg', 'https://example.com/nike-2.jpg'],
     '{"size": ["7", "8", "9", "10", "11"], "color": ["Black", "White", "Red"]}',
     0, 1, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),

    ('Levi''s 501 Jeans', 'Classic straight fit jeans', 59.99, 150, 'Clothing',
     ARRAY['https://example.com/levis-1.jpg'],
     '{"size": ["28", "30", "32", "34", "36"], "color": ["Blue", "Black"]}',
     0, 1, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),

    ('Organic Green Tea', 'Premium organic green tea leaves', 19.99, 500, 'Food',
     ARRAY['https://example.com/tea-1.jpg'],
     '{"weight": "250g", "origin": "Japan"}',
     0, 1, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT);

-- Verify table creation
SELECT 'Products table created successfully!' AS status;
SELECT COUNT(*) AS sample_products FROM products;
