-- ========================================
-- letsGO E-commerce Platform
-- PostgreSQL Database Initialization Script
-- ========================================
-- This script creates all necessary databases
-- and initializes tables for each service

-- ========================================
-- Create Databases
-- ========================================

CREATE DATABASE letsgo_user;
CREATE DATABASE letsgo_product;
CREATE DATABASE letsgo_order;
CREATE DATABASE letsgo_payment;

-- ========================================
-- User Service Database Schema
-- ========================================

\c letsgo_user;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,  -- Hashed password
    salt VARCHAR(50) NOT NULL,        -- Password salt
    email VARCHAR(100) UNIQUE NOT NULL,
    phone VARCHAR(20),
    avatar VARCHAR(255) DEFAULT '',
    status SMALLINT DEFAULT 1,        -- 1:active, 2:disabled
    created_at BIGINT NOT NULL,       -- Unix timestamp
    updated_at BIGINT NOT NULL
);

-- Create indexes for faster queries
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_created_at ON users(created_at);

COMMENT ON TABLE users IS 'User accounts';
COMMENT ON COLUMN users.status IS '1:active, 2:disabled';

-- ========================================
-- Product Service Database Schema
-- ========================================

\c letsgo_product;

-- Products table (core product data)
CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,    -- Product price
    stock BIGINT DEFAULT 0,           -- Inventory quantity
    category VARCHAR(50) NOT NULL,
    images JSONB,                     -- Array of image URLs stored as JSON
    sales BIGINT DEFAULT 0,           -- Total sales count
    status SMALLINT DEFAULT 1,        -- 1:active, 2:inactive, 3:deleted
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

-- Create indexes
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_price ON products(price);
CREATE INDEX idx_products_sales ON products(sales);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_created_at ON products(created_at);

COMMENT ON TABLE products IS 'Product catalog';
COMMENT ON COLUMN products.status IS '1:active, 2:inactive, 3:deleted';

-- ========================================
-- Order Service Database Schema
-- ========================================

\c letsgo_order;

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    order_no VARCHAR(50) UNIQUE NOT NULL,  -- Unique order number
    total_amount DECIMAL(10, 2) NOT NULL,  -- Total order amount
    status SMALLINT DEFAULT 1,             -- Order status
    address TEXT NOT NULL,                 -- Delivery address
    phone VARCHAR(20) NOT NULL,            -- Contact phone
    remark TEXT DEFAULT '',                -- Customer notes
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    paid_at BIGINT DEFAULT 0,              -- Payment timestamp
    shipped_at BIGINT DEFAULT 0,           -- Shipping timestamp
    completed_at BIGINT DEFAULT 0          -- Completion timestamp
);

-- Order items table (products in each order)
CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL,
    name VARCHAR(200) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,         -- Price snapshot at purchase time
    quantity BIGINT NOT NULL,
    image VARCHAR(255) DEFAULT '',
    created_at BIGINT NOT NULL
);

-- Create indexes
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_order_no ON orders(order_no);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);

COMMENT ON TABLE orders IS 'Customer orders';
COMMENT ON TABLE order_items IS 'Order line items';
COMMENT ON COLUMN orders.status IS '1:pending, 2:paid, 3:shipped, 4:completed, 5:cancelled';

-- ========================================
-- Payment Service Database Schema
-- ========================================

\c letsgo_payment;

-- Payments table
CREATE TABLE IF NOT EXISTS payments (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT UNIQUE NOT NULL,      -- One payment per order
    user_id BIGINT NOT NULL,
    payment_no VARCHAR(50) UNIQUE NOT NULL, -- Unique payment number
    amount DECIMAL(10, 2) NOT NULL,
    payment_type SMALLINT NOT NULL,       -- Payment method
    status SMALLINT DEFAULT 1,            -- Payment status
    trade_no VARCHAR(100) DEFAULT '',     -- Third-party transaction ID
    created_at BIGINT NOT NULL,
    paid_at BIGINT DEFAULT 0
);

-- Create indexes
CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_payment_no ON payments(payment_no);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created_at ON payments(created_at);

COMMENT ON TABLE payments IS 'Payment transactions';
COMMENT ON COLUMN payments.payment_type IS '1:Alipay, 2:WeChat Pay, 3:Credit Card';
COMMENT ON COLUMN payments.status IS '1:pending, 2:success, 3:failed, 4:cancelled';

-- ========================================
-- Insert Sample Data (Optional)
-- ========================================

-- Sample user (password: "123456", salt: "test_salt")
-- Hashed password = md5("123456test_salt") = "d7b36de383b7401f8236dbb5f8c33bea"
\c letsgo_user;
INSERT INTO users (username, password, salt, email, phone, created_at, updated_at)
VALUES ('testuser', 'd7b36de383b7401f8236dbb5f8c33bea', 'test_salt', 'test@example.com', '13800138000',
        EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT);

-- Sample products
\c letsgo_product;
INSERT INTO products (name, description, price, stock, category, images, sales, created_at, updated_at) VALUES
('iPhone 14 Pro', 'Latest Apple smartphone with A16 chip', 999.99, 100, 'Electronics',
 '["https://example.com/iphone1.jpg", "https://example.com/iphone2.jpg"]'::jsonb, 0,
 EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
('MacBook Air M2', 'Lightweight laptop with M2 chip', 1199.99, 50, 'Electronics',
 '["https://example.com/macbook1.jpg"]'::jsonb, 0,
 EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
('Nike Air Max', 'Comfortable running shoes', 129.99, 200, 'Fashion',
 '["https://example.com/nike1.jpg"]'::jsonb, 0,
 EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT);

-- Completion message
\c postgres;
SELECT 'letsGO database initialization completed successfully!' as message;
