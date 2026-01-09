# 超时控制中间件 (Timeout Middleware)

## 概述

超时控制中间件用于限制每个 HTTP 请求的最长响应时间。如果请求处理时间超过配置的超时时间，中间件会自动终止请求并返回 `504 Gateway Timeout` 错误。

## 功能特性

- ✅ 可配置的超时时间
- ✅ 自动终止超时请求
- ✅ 返回标准的 HTTP 504 状态码
- ✅ 支持 panic 恢复
- ✅ 使用 Go context 进行超时控制

## 配置

### 1. 配置文件设置

在 `gateway/gateway.yaml` 中配置超时时间（单位：毫秒）：

```yaml
# Request Timeout - Maximum response time for each request
RequestTimeout: 5000     # 5秒超时
```

### 2. 应用到路由

超时中间件已经在 `ServiceContext` 中初始化，可以在路由中使用：

#### 方式一：应用到所有路由（全局）

在 `gateway/gateway.go` 中添加全局中间件：

```go
func main() {
    // ... 其他代码 ...

    server := rest.MustNewServer(c.RestConf)
    defer server.Stop()

    // 添加全局超时中间件
    server.Use(ctx.Timeout)

    ctx := svc.NewServiceContext(c)
    handler.RegisterHandlers(server, ctx)

    // ... 其他代码 ...
}
```

#### 方式二：应用到特定路由组

在 `gateway/internal/handler/routes.go` 中为特定路由组添加：

```go
server.AddRoutes(
    rest.WithMiddlewares(
        []rest.Middleware{serverCtx.Timeout, serverCtx.Auth},  // 添加 Timeout 中间件
        []rest.Route{
            {
                Method:  http.MethodGet,
                Path:    "/",
                Handler: cart.GetCartHandler(serverCtx),
            },
            // ... 其他路由 ...
        }...,
    ),
    rest.WithPrefix("/api/v1/cart"),
)
```

## 工作原理

1. **创建超时上下文**：中间件使用 `context.WithTimeout` 创建一个带超时的上下文
2. **并发执行**：在 goroutine 中执行实际的请求处理
3. **监听完成信号**：使用 `select` 语句监听三个通道：
   - `done`：请求正常完成
   - `panicChan`：请求发生 panic
   - `ctx.Done()`：超时发生
4. **超时处理**：如果超时发生，返回 504 状态码和错误消息

## 代码示例

### 中间件实现

```go
type TimeoutMiddleware struct {
    timeout time.Duration
}

func NewTimeoutMiddleware(timeout time.Duration) *TimeoutMiddleware {
    return &TimeoutMiddleware{
        timeout: timeout,
    }
}

func (m *TimeoutMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), m.timeout)
        defer cancel()

        done := make(chan struct{})
        panicChan := make(chan interface{}, 1)

        go func() {
            defer func() {
                if p := recover(); p != nil {
                    panicChan <- p
                }
            }()

            next(w, r.WithContext(ctx))
            close(done)
        }()

        select {
        case p := <-panicChan:
            panic(p)
        case <-done:
            return
        case <-ctx.Done():
            w.WriteHeader(http.StatusGatewayTimeout)
            w.Write([]byte("request timeout"))
            return
        }
    }
}
```

## 测试超时中间件

### 创建测试端点

可以创建一个测试端点来验证超时功能：

```go
// 在某个 handler 中添加延迟
func TestTimeoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 模拟长时间处理（10秒）
        time.Sleep(10 * time.Second)

        w.WriteHeader(http.StatusOK)
        w.Write([]byte("success"))
    }
}
```

### 使用 curl 测试

```bash
# 如果 RequestTimeout 设置为 5000ms (5秒)
# 这个请求会在 5 秒后超时
curl -X GET http://localhost:8888/api/v1/test/timeout

# 预期响应：
# HTTP/1.1 504 Gateway Timeout
# request timeout
```

## 最佳实践

1. **合理设置超时时间**
   - 根据业务需求设置合适的超时时间
   - 考虑数据库查询、RPC 调用等操作的平均响应时间
   - 建议：API 网关 5-10 秒，内部服务 2-5 秒

2. **不同路由不同超时**
   - 对于快速查询接口，使用较短超时（如 2-3 秒）
   - 对于复杂计算或批量操作，使用较长超时（如 30-60 秒）
   - 可以为不同路由组创建不同的超时中间件实例

3. **监控和日志**
   - 记录超时事件，用于性能分析
   - 监控超时率，及时发现性能问题

4. **客户端配置**
   - 确保客户端的超时时间大于服务端
   - 建议客户端超时 = 服务端超时 + 2 秒

## 配置示例

### 开发环境（较长超时）
```yaml
RequestTimeout: 30000  # 30秒，方便调试
```

### 生产环境（较短超时）
```yaml
RequestTimeout: 5000   # 5秒，快速失败
```

### 不同路由不同超时
```go
// 在 service_context.go 中
type ServiceContext struct {
    Config        config.Config
    Auth          rest.Middleware
    AdminAuth     rest.Middleware
    TimeoutShort  rest.Middleware  // 短超时：2秒
    TimeoutMedium rest.Middleware  // 中等超时：5秒
    TimeoutLong   rest.Middleware  // 长超时：30秒
    // ... 其他字段 ...
}

func NewServiceContext(c config.Config) *ServiceContext {
    return &ServiceContext{
        Config:        c,
        TimeoutShort:  middleware.NewTimeoutMiddleware(2 * time.Second).Handle,
        TimeoutMedium: middleware.NewTimeoutMiddleware(5 * time.Second).Handle,
        TimeoutLong:   middleware.NewTimeoutMiddleware(30 * time.Second).Handle,
        // ... 其他初始化 ...
    }
}
```

## 注意事项

1. **响应已发送的情况**
   - 如果 handler 已经开始写入响应，超时中间件无法撤销已发送的数据
   - 建议在 handler 中也检查 context 状态

2. **资源清理**
   - 超时后，handler 中的 goroutine 仍会继续执行直到完成
   - 确保 handler 中正确处理 context 取消信号

3. **数据库连接**
   - 使用支持 context 的数据库驱动
   - 在数据库查询中传递 context，确保超时时能正确取消查询

## 相关文件

- 中间件实现：[gateway/internal/middleware/timeout_middleware.go](../gateway/internal/middleware/timeout_middleware.go)
- 配置定义：[gateway/internal/config/config.go](../gateway/internal/config/config.go)
- 服务上下文：[gateway/internal/svc/service_context.go](../gateway/internal/svc/service_context.go)
- 配置文件：[gateway/gateway.yaml](../gateway/gateway.yaml)
