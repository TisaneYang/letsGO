-- ================================================
-- 安全的管理员账号创建脚本
-- ================================================
-- 警告：此脚本应只在安全的环境中执行
-- 建议：只有数据库管理员或系统管理员有权限执行

-- 方法1：将现有用户提升为管理员
-- 使用场景：已经注册的普通用户需要升级为管理员
-- ================================================
UPDATE users
SET role = 'admin', updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT
WHERE username = '要提升的用户名';

-- 示例：
-- UPDATE users SET role = 'admin', updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT WHERE username = 'john_doe';


-- 方法2：查看所有管理员账号
-- ================================================
SELECT id, username, email, role, created_at, updated_at
FROM users
WHERE role = 'admin'
ORDER BY created_at DESC;


-- 方法3：将管理员降级为普通用户
-- 使用场景：撤销管理员权限
-- ================================================
UPDATE users
SET role = 'user', updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT
WHERE username = '要降级的管理员用户名';


-- 方法4：创建新的管理员账号（需要先通过注册接口注册）
-- ================================================
-- 步骤1: 先让用户通过正常注册接口注册账号
-- 步骤2: 然后执行以下SQL提升权限

UPDATE users
SET role = 'admin', updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT
WHERE email = '管理员邮箱@example.com';


-- 安全检查：验证管理员数量
-- ================================================
SELECT role, COUNT(*) as count
FROM users
GROUP BY role;


-- 审计：查看最近的角色变更（需要审计日志表）
-- ================================================
-- 注意：这需要你实现审计日志功能
-- CREATE TABLE user_role_audit (
--     id BIGSERIAL PRIMARY KEY,
--     user_id BIGINT NOT NULL,
--     old_role VARCHAR(20),
--     new_role VARCHAR(20),
--     changed_by VARCHAR(100),
--     changed_at BIGINT NOT NULL
-- );
