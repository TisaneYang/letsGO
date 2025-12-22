# 管理员账号安全创建指南

## 安全问题说明

**为什么不在注册接口暴露 role 参数？**

如果在公开的注册接口中允许指定 `role` 参数，任何人都可以注册管理员账号：

```bash
# ❌ 危险！任何人都能创建管理员
curl -X POST http://localhost:8888/api/v1/user/register \
  -d '{"username":"hacker","password":"123","email":"hack@evil.com","role":"admin"}'
```

这会导致：
- 🔴 **严重安全漏洞**：攻击者可以获得完全管理权限
- 🔴 **数据泄露风险**：管理员可以访问所有用户数据
- 🔴 **系统破坏**：恶意管理员可以删除或修改任意数据

## 当前安全策略

✅ **注册接口只能创建普通用户**
- 所有通过 `/api/v1/user/register` 注册的用户自动设置为 `role = 'user'`
- 即使请求中包含 `role` 参数也会被忽略
- 前端无法通过 API 创建管理员

✅ **管理员只能通过安全渠道创建**
- 数据库直接操作（需要 DBA 权限）
- 服务器端管理脚本（需要 SSH 访问）
- 未来可添加超级管理员专用的管理员创建接口

## 创建管理员的安全方法

### 方法 1：先注册后提升（推荐）⭐

**步骤 1**：用户通过正常渠道注册
```bash
curl -X POST http://localhost:8888/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "password": "SecurePassword123!",
    "email": "alice@company.com",
    "phone": "13800138000"
  }'
```

**步骤 2**：数据库管理员执行 SQL 提升权限
```bash
# 连接到数据库
psql -h localhost -U postgres -d letsgo

# 提升权限
UPDATE users
SET role = 'admin', updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT
WHERE username = 'alice';

# 验证
SELECT id, username, email, role FROM users WHERE username = 'alice';
```

### 方法 2：使用管理脚本

```bash
# 执行管理员创建脚本
cd /home/damai/letsGO
psql -h localhost -U postgres -d letsgo -f migrations/create_admin_user.sql

# 在 psql 提示符中输入
UPDATE users SET role = 'admin' WHERE username = '目标用户名';
```

### 方法 3：命令行快捷方式

创建一个便捷的 shell 脚本 `scripts/make-admin.sh`：

```bash
#!/bin/bash
# 使用方法: ./scripts/make-admin.sh <username>

if [ -z "$1" ]; then
    echo "Usage: $0 <username>"
    echo "Example: $0 alice"
    exit 1
fi

USERNAME=$1

echo "正在将用户 '$USERNAME' 提升为管理员..."

psql -h localhost -U postgres -d letsgo <<EOF
UPDATE users
SET role = 'admin', updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT
WHERE username = '$USERNAME';

SELECT
    CASE
        WHEN COUNT(*) > 0 THEN '✓ 成功: 用户已提升为管理员'
        ELSE '✗ 失败: 用户不存在'
    END as result
FROM users
WHERE username = '$USERNAME' AND role = 'admin';
EOF

echo "完成！"
```

使用：
```bash
chmod +x scripts/make-admin.sh
./scripts/make-admin.sh alice
```

## 查看和管理管理员

### 列出所有管理员
```sql
SELECT id, username, email, role,
       to_timestamp(created_at) as created_at,
       to_timestamp(updated_at) as updated_at
FROM users
WHERE role = 'admin'
ORDER BY created_at DESC;
```

### 统计角色分布
```sql
SELECT role, COUNT(*) as count
FROM users
GROUP BY role;
```

### 降级管理员为普通用户
```sql
UPDATE users
SET role = 'user', updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT
WHERE username = '要降级的用户名';
```

### 查找可疑的管理员账号
```sql
-- 查找最近创建的管理员
SELECT id, username, email,
       to_timestamp(created_at) as created_at
FROM users
WHERE role = 'admin'
  AND created_at > EXTRACT(EPOCH FROM NOW() - INTERVAL '7 days')::BIGINT
ORDER BY created_at DESC;
```

## 未来改进：超级管理员创建接口

如果需要通过 API 创建管理员，应该：

### 1. 创建专门的管理员管理接口

```go
// gateway.api
@server (
    prefix: /api/v1/admin
    group: admin
    middleware: SuperAdminAuth  // 超级管理员专用
)
service gateway {
    @doc "Create admin user - Super admin only"
    @handler createAdmin
    post /users/create-admin (CreateAdminReq) returns (CreateAdminResp)
}
```

### 2. 实现超级管理员验证

```go
// SuperAdminAuth 中间件
func (m *SuperAdminAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 验证 token
        // 检查角色是否为 "super_admin"
        role, _ := claims["role"].(string)
        if role != "super_admin" {
            http.Error(w, "permission denied: super admin required", http.StatusForbidden)
            return
        }
        next(w, r.WithContext(ctx))
    }
}
```

### 3. 角色层级
```
super_admin (超级管理员) - 可以创建 admin
    └── admin (管理员) - 可以管理商品、订单等
            └── user (普通用户) - 普通权限
```

## 安全最佳实践

### 1. 最小权限原则
- ✅ 只给必要的人管理员权限
- ✅ 定期审计管理员账号
- ✅ 及时撤销离职员工的管理员权限

### 2. 审计日志
考虑添加角色变更日志：

```sql
CREATE TABLE user_role_audit (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    old_role VARCHAR(20),
    new_role VARCHAR(20),
    changed_by VARCHAR(100),
    changed_at BIGINT NOT NULL,
    ip_address VARCHAR(50),
    reason TEXT
);

-- 创建触发器自动记录角色变更
CREATE OR REPLACE FUNCTION log_role_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.role IS DISTINCT FROM NEW.role THEN
        INSERT INTO user_role_audit (user_id, old_role, new_role, changed_at)
        VALUES (NEW.id, OLD.role, NEW.role, EXTRACT(EPOCH FROM NOW())::BIGINT);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER role_change_trigger
AFTER UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION log_role_change();
```

### 3. 双因素认证
对于管理员账号，建议：
- 要求更强的密码策略
- 启用 2FA（双因素认证）
- 限制登录 IP 范围
- 设置更短的 token 过期时间

### 4. 监控和告警
- 监控管理员创建事件
- 监控管理员登录异常（异地登录、频繁失败等）
- 监控敏感操作（批量删除、数据导出等）

## 应急响应

### 如果发现恶意管理员账号

**立即响应步骤：**

1. **立即禁用账号**
```sql
UPDATE users SET status = 2 WHERE username = '恶意用户';
```

2. **撤销管理员权限**
```sql
UPDATE users SET role = 'user' WHERE username = '恶意用户';
```

3. **查看操作日志**
```bash
# 查看该用户的所有操作记录
grep "user_id=<恶意用户ID>" logs/*.log
```

4. **评估影响范围**
- 检查是否有数据被修改或删除
- 检查是否有其他账号被创建
- 检查系统配置是否被篡改

5. **修改密钥**
```yaml
# 修改 JWT Secret
Auth:
  AccessSecret: "新的随机密钥"
  AccessExpire: 7200
```

6. **通知相关人员**
- 通知技术团队
- 通知安全团队
- 必要时通知用户

## 总结

✅ **现在的实现是安全的**
- 注册接口无法创建管理员
- 管理员只能通过数据库创建
- 有明确的管理员创建流程

⚠️ **记住**
- 定期审计管理员账号
- 保护好数据库访问权限
- 记录所有角色变更操作
- 及时响应安全事件

🔐 **安全是一个持续的过程，不是一次性的任务！**
