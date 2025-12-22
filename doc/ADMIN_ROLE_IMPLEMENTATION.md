# 管理员角色实现总结

## 实现方案

选择了**方案1：在现有用户表中添加 role 字段**，这是最适合当前项目需求的方案。

### 为什么选择这个方案？

1. **简单直接**：不需要额外的表和复杂的 JOIN 查询
2. **性能好**：查询效率高，无需额外关联
3. **足够用**：对于电商系统的用户/管理员区分完全够用
4. **易维护**：代码简单，后期维护成本低

## 安全设计 🔐

### 重要安全措施

✅ **注册接口不暴露 role 参数**
- 所有通过 `/api/v1/user/register` 注册的用户都是普通用户（`role = 'user'`）
- Proto 定义中**已移除** `RegisterRequest.role` 字段
- 代码层面**强制设置** `role = "user"`，无法通过 API 指定角色

✅ **管理员只能通过安全渠道创建**
- 数据库直接操作（需要 DBA 权限）
- 先注册为普通用户，再通过 SQL 提升权限

### 为什么这样设计？

**❌ 如果在注册接口暴露 role 参数（危险）：**
```bash
# 任何人都能创建管理员！严重安全漏洞！
curl -X POST http://localhost:8888/api/v1/user/register \
  -d '{"username":"hacker","role":"admin",...}'
```

**✅ 当前安全设计：**
```bash
# 1. 只能创建普通用户
curl -X POST http://localhost:8888/api/v1/user/register \
  -d '{"username":"alice",...}'  # 自动设置 role='user'

# 2. 管理员需要 DBA 权限提升
psql -d letsgo -c "UPDATE users SET role='admin' WHERE username='alice';"
```

**参考文档：** `docs/ADMIN_SECURITY_GUIDE.md` - 详细的管理员安全创建指南

## 已完成的修改

### 1. 数据库层 (Database)

**文件位置：** `migrations/add_user_role.sql`

- 添加 `role` 字段到 `users` 表
- 默认值为 `'user'`
- 创建索引 `idx_users_role`
- 包含示例管理员用户插入语句

**执行迁移：**
```bash
# 连接到 PostgreSQL 数据库
psql -h localhost -U postgres -d letsgo -f migrations/add_user_role.sql
```

### 2. Model 层

**修改文件：**
- `services/user/model/types.go` - 添加 `Role` 字段到 `User` 结构体
- `services/user/model/user_model.go` - 更新所有 SQL 查询以包含 `role` 字段

**改动点：**
- `Insert()`: 插入时包含 role
- `FindOne()`: 查询时返回 role
- `FindOneByUsername()`: 查询时返回 role
- `FindOneByEmail()`: 查询时返回 role

### 3. Proto 定义

**文件：** `services/user/rpc/user.proto`

修改内容：
- ❌ **已移除** `RegisterRequest.role` 字段（安全考虑，防止通过 API 注册管理员）
- ✅ **添加** `LoginResponse.role` 字段 - 登录时返回用户角色
- ✅ **添加** `GetUserInfoResponse.role` 字段 - 获取用户信息时返回角色
- ✅ **添加** `VerifyTokenResponse.role` 字段 - 验证 token 时返回角色

### 4. RPC Logic 层

**修改文件：**

#### `register_logic.go`
```go
// SECURITY: Force all registrations to 'user' role
// Admin users must be created through secure channels
role := "user"
```
- 强制所有注册用户为普通用户
- JWT token 中包含 `role` 字段
- 日志中记录用户角色

#### `login_logic.go`
- 登录时从数据库读取用户角色
- JWT token 中包含 `role` 字段
- 返回响应中包含角色信息
- 日志中记录用户角色

#### `get_user_info_logic.go`
- 返回用户信息时包含角色

#### `verify_token_logic.go`
- 验证 token 时提取并返回角色信息
- 兼容旧 token（没有 role 字段时默认为 'user'）

### 5. Gateway 中间件

**新增文件：** `gateway/internal/middleware/adminauth_middleware.go`

功能：
- 验证 JWT token 有效性
- 检查 token 中的 `role` 字段是否为 `"admin"`
- 如果不是管理员，返回 403 Forbidden
- 将 `userId` 和 `role` 存入 context

**修改文件：**
- `gateway/gateway.api` - Product 的 `/add` 和 `/update` 接口使用 `AdminAuth` 中间件
- `gateway/internal/svc/service_context.go` - 注册 AdminAuth 中间件

## JWT Token 结构

### 旧版本（不包含角色）
```json
{
  "userId": 123,
  "exp": 1234567890,
  "iat": 1234567800
}
```

### 新版本（包含角色）
```json
{
  "userId": 123,
  "role": "admin",
  "exp": 1234567890,
  "iat": 1234567800
}
```

## API 权限说明

### 普通用户可访问
- `/api/v1/user/*` - 用户相关接口
- `/api/v1/product/list` - 查看商品列表
- `/api/v1/product/detail/:id` - 查看商品详情
- `/api/v1/product/search` - 搜索商品
- `/api/v1/cart/*` - 购物车管理
- `/api/v1/order/*` - 订单管理
- `/api/v1/payment/*` - 支付相关

### 仅管理员可访问
- `/api/v1/product/add` - 添加商品 ⚠️ AdminAuth
- `/api/v1/product/update` - 更新商品 ⚠️ AdminAuth

## 部署步骤

### 1. 运行数据库迁移
```bash
cd /home/damai/letsGO
psql -h localhost -U postgres -d letsgo -f migrations/add_user_role.sql
```

### 2. 创建管理员账户（推荐方式）

**方法1：先注册后提升（推荐）⭐**
```bash
# 步骤1: 正常注册账号
curl -X POST http://localhost:8888/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "password": "SecurePassword123!",
    "email": "alice@company.com"
  }'

# 步骤2: 数据库提升权限
psql -h localhost -U postgres -d letsgo -c \
  "UPDATE users SET role='admin' WHERE username='alice';"
```

**方法2：直接 SQL 操作**
```bash
# 查看所有用户
psql -d letsgo -c "SELECT id, username, email, role FROM users;"

# 提升现有用户为管理员
psql -d letsgo -c "UPDATE users SET role='admin' WHERE username='目标用户';"

# 查看所有管理员
psql -d letsgo -c "SELECT * FROM users WHERE role='admin';"
```

**方法3：使用管理脚本**
```bash
psql -h localhost -U postgres -d letsgo -f migrations/create_admin_user.sql
# 然后按照脚本中的指引执行相应的 SQL 命令
```

**详细文档：** 参见 `docs/ADMIN_SECURITY_GUIDE.md`

### 3. 重新构建和启动服务
```bash
make build
make run
```

## 测试验证

### 1. 注册管理员账户
```bash
curl -X POST http://localhost:8888/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123",
    "email": "admin@letsgo.com",
    "phone": "13800000000",
    "role": "admin"
  }'
```

响应包含 token：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "userId": 1,
    "token": "eyJhbGci..."
  }
}
```

### 2. 注册普通用户
```bash
curl -X POST http://localhost:8888/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user001",
    "password": "password123",
    "email": "user001@example.com"
  }'
```

### 3. 测试管理员添加商品（应成功）
```bash
curl -X POST http://localhost:8888/api/v1/product/add \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <管理员token>" \
  -d '{
    "name": "测试商品",
    "description": "这是一个测试商品",
    "price": 99.99,
    "stock": 100,
    "category": "电子产品",
    "images": ["https://example.com/image.jpg"]
  }'
```

预期：成功返回商品ID

### 4. 测试普通用户添加商品（应失败）
```bash
curl -X POST http://localhost:8888/api/v1/product/add \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <普通用户token>" \
  -d '{
    "name": "测试商品",
    "description": "这是一个测试商品",
    "price": 99.99,
    "stock": 100,
    "category": "电子产品",
    "images": ["https://example.com/image.jpg"]
  }'
```

预期：返回 403 Forbidden
```
permission denied: admin access required
```

## 错误码说明

- `401 Unauthorized` - Token 无效或缺失
- `403 Forbidden` - Token 有效但不是管理员角色

## 兼容性说明

### 向后兼容
- 旧的 JWT token（不包含 role）仍然可以使用
- `VerifyToken` 接口会为旧 token 返回默认角色 `"user"`
- 现有用户在数据库迁移后会自动获得 `"user"` 角色

### 注意事项
1. **现有用户**：迁移脚本会为所有现有用户设置 `role = 'user'`
2. **旧 Token**：用户需要重新登录以获取包含角色的新 token
3. **缓存清理**：建议清理 Redis 中的用户信息缓存，或等待过期（1小时）

## 安全建议

1. **限制管理员注册**：
   - 建议在生产环境中移除注册接口的 `role` 参数
   - 仅允许通过数据库直接修改用户角色

2. **审计日志**：
   - 所有管理员操作都会记录在日志中
   - 包含 user_id, username, role 等信息

3. **Token 安全**：
   - 使用 HTTPS 传输 token
   - 定期更换 JWT secret
   - 设置合理的 token 过期时间

## 未来扩展

如果需要更复杂的权限系统，可以考虑：

1. **添加更多角色**：
   - `merchant` - 商家
   - `operator` - 运营人员
   - `super_admin` - 超级管理员

2. **细粒度权限**：
   - 创建 `permissions` 表
   - 创建 `role_permissions` 关联表
   - 实现 RBAC (Role-Based Access Control)

3. **权限组**：
   - 用户可以拥有多个角色
   - 创建 `user_roles` 关联表

## 相关文件清单

### 新增文件
- `migrations/add_user_role.sql` - 数据库迁移脚本
- `gateway/internal/middleware/adminauth_middleware.go` - 管理员鉴权中间件
- `ADMIN_ROLE_IMPLEMENTATION.md` - 本文档

### 修改文件
- `gateway/gateway.api`
- `gateway/internal/svc/service_context.go`
- `services/user/model/types.go`
- `services/user/model/user_model.go`
- `services/user/rpc/user.proto`
- `services/user/rpc/internal/logic/register_logic.go`
- `services/user/rpc/internal/logic/login_logic.go`
- `services/user/rpc/internal/logic/get_user_info_logic.go`
- `services/user/rpc/internal/logic/verify_token_logic.go`

## 总结

✅ 已成功实现管理员角色功能
✅ 所有代码已编译通过
✅ JWT token 包含角色信息
✅ Product 的 add/update 接口受管理员权限保护
✅ 完全向后兼容现有系统

接下来只需要运行数据库迁移并重启服务即可生效！
