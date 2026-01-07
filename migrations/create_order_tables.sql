-- ========================================
-- Order Tables Migration
-- ========================================
-- This migration creates the tables needed for the order service

-- ========================================
-- 1. Orders Table (Main order information)
-- ========================================
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    order_no VARCHAR(32) UNIQUE NOT NULL,  -- Human-readable order number (e.g., LG20251225143012001)
    total_amount DECIMAL(10,2) NOT NULL CHECK (total_amount >= 0),
    status SMALLINT DEFAULT 1 NOT NULL,    -- 1:pending, 2:paid, 3:shipped, 4:completed, 5:cancelled
    address TEXT NOT NULL,
    phone VARCHAR(20) NOT NULL,
    remark TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    paid_at TIMESTAMP,                     -- Payment time
    shipped_at TIMESTAMP,                  -- Shipping time
    completed_at TIMESTAMP,                -- Completion time

    -- Indexes for better query performance
    CONSTRAINT orders_user_id_idx CHECK (user_id > 0)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_order_no ON orders(order_no);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_user_status ON orders(user_id, status);

-- ========================================
-- 2. Order Items Table (Product snapshot in order)
-- ========================================
CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,            -- Snapshot: product name at time of purchase
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0), -- Snapshot: price at time of purchase
    quantity INT NOT NULL CHECK (quantity > 0),
    image VARCHAR(500) DEFAULT '',         -- Snapshot: product image URL
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key to orders table
    CONSTRAINT fk_order_items_order_id FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);

-- ========================================
-- 3. Comments and Documentation
-- ========================================
COMMENT ON TABLE orders IS 'Main order table storing order information';
COMMENT ON COLUMN orders.order_no IS 'Human-readable unique order number';
COMMENT ON COLUMN orders.status IS '1:pending(待支付), 2:paid(已支付), 3:shipped(已发货), 4:completed(已完成), 5:cancelled(已取消)';
COMMENT ON COLUMN orders.total_amount IS 'Total order amount in decimal format';

COMMENT ON TABLE order_items IS 'Order items table storing product snapshots';
COMMENT ON COLUMN order_items.name IS 'Product name snapshot (in case product is deleted later)';
COMMENT ON COLUMN order_items.price IS 'Product price snapshot (in case price changes later)';
COMMENT ON COLUMN order_items.image IS 'Product image snapshot';

-- ========================================
-- 4. Updated At Trigger (Auto-update timestamp)
-- ========================================
CREATE OR REPLACE FUNCTION update_orders_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION update_orders_updated_at();

-- ========================================
-- 5. Sample Data (Optional - for testing)
-- ========================================
-- Uncomment below to insert test data
/*
INSERT INTO orders (user_id, order_no, total_amount, status, address, phone, remark) VALUES
(1, 'LG20251225143012001', 599.99, 1, '北京市朝阳区xx街道xx号', '13800138000', '请尽快发货'),
(1, 'LG20251225153012002', 299.50, 2, '上海市浦东新区xx路xx号', '13900139000', '');

INSERT INTO order_items (order_id, product_id, name, price, quantity, image) VALUES
(1, 1, 'iPhone 15 Pro', 599.99, 1, 'https://example.com/iphone15.jpg'),
(2, 2, 'AirPods Pro', 299.50, 1, 'https://example.com/airpods.jpg');
*/
