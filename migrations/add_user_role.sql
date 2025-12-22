-- Migration: Add role field to users table
-- Date: 2025-12-22
-- Description: Add role field to support admin/user distinction

-- Add role column (default to 'user' for existing users)
ALTER TABLE users
ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user';

-- Create index for role queries
CREATE INDEX idx_users_role ON users(role);

-- Add comment
COMMENT ON COLUMN users.role IS 'User role: user, admin';

-- Optional: Create a sample admin user
-- Password: admin123 (you should change this in production)
-- Note: You'll need to generate the actual hashed password
INSERT INTO users (username, password, salt, email, phone, avatar, status, role, created_at, updated_at)
VALUES (
    'admin',
    -- This is a placeholder hash, replace with actual bcrypt hash
    '$2a$10$placeholder',
    'placeholder_salt',
    'admin@letsgo.com',
    '13800000000',
    '',
    1,
    'admin',
    EXTRACT(EPOCH FROM NOW())::BIGINT,
    EXTRACT(EPOCH FROM NOW())::BIGINT
)
ON CONFLICT (username) DO NOTHING;
