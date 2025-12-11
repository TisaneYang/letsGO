# gRPC ClientConn 连接生命周期详解

## 1. 连接建立（Gateway 启动时）

```go
// gateway/internal/svc/service_context.go:25
UserRpc: user_client.NewUser(zrpc.MustNewClient(c.UserRpc))
```

### 内部流程：

```
Step 1: 创建 gRPC ClientConn
├─> grpc.Dial("etcd://127.0.0.1:2379/user.rpc")
│
Step 2: 解析 Target
├─> Resolver 连接 etcd
├─> 查询 key="user.rpc"
├─> 获取实例列表: [192.168.1.10:9001, 192.168.1.11:9001, 192.168.1.12:9001]
│
Step 3: 建立 TCP 连接（为每个实例）
├─> net.Dial("tcp", "192.168.1.10:9001")
│   └─> TCP 三次握手
│   └─> HTTP/2 连接协商
│   └─> 创建 SubConn 1
│
├─> net.Dial("tcp", "192.168.1.11:9001")
│   └─> TCP 三次握手
│   └─> HTTP/2 连接协商
│   └─> 创建 SubConn 2
│
└─> net.Dial("tcp", "192.168.1.12:9001")
    └─> TCP 三次握手
    └─> HTTP/2 连接协商
    └─> 创建 SubConn 3

Step 4: 初始化负载均衡器
└─> Balancer 接收实例列表
    └─> 准备就绪，可以处理 RPC 请求
```

## 2. 连接维护（运行时）

### HTTP/2 Keep-Alive

```go
// gRPC 默认配置
keepalive.ClientParameters{
    Time:                10 * time.Second,  // 每 10 秒发送 ping
    Timeout:             20 * time.Second,  // ping 超时时间
    PermitWithoutStream: true,              // 即使没有活跃请求也发送 ping
}
```

### 连接状态监控

```
每个 SubConn 的状态机:

IDLE (空闲)
  ↓
CONNECTING (连接中)
  ↓
READY (就绪) ←──────────┐
  ↓                     │
TRANSIENT_FAILURE       │ 自动重连
  ↓                     │
CONNECTING ─────────────┘
  ↓
SHUTDOWN (关闭)
```

### 自动重连机制

```go
// 伪代码
func (sc *SubConn) monitorConnection() {
    for {
        if sc.state == TRANSIENT_FAILURE {
            // 指数退避重连
            backoff := calculateBackoff(sc.retryCount)
            time.Sleep(backoff)

            // 尝试重新连接
            conn, err := net.Dial("tcp", sc.address)
            if err == nil {
                sc.conn = conn
                sc.state = READY
                sc.retryCount = 0
            } else {
                sc.retryCount++
            }
        }
        time.Sleep(1 * time.Second)
    }
}
```

## 3. RPC 调用（复用连接）

### 单次 RPC 调用流程

```
请求 1: UserRpc.Login()
  ↓
1. 负载均衡器选择连接
   Balancer.Pick() → SubConn 1 (192.168.1.10:9001)
   ↓
2. 复用已有的 TCP 连接
   使用 SubConn 1 的 HTTP/2 连接
   ↓
3. 创建 HTTP/2 Stream
   Stream ID: 1
   Method: /user.User/Login
   ↓
4. 发送请求数据
   通过 Stream 1 发送 Protobuf 数据
   ↓
5. 等待响应
   从 Stream 1 接收响应
   ↓
6. 关闭 Stream
   Stream 1 关闭，但 TCP 连接保持
   ↓
7. 连接返回连接池
   SubConn 1 可以处理下一个请求


请求 2: UserRpc.GetUserInfo()
  ↓
1. 负载均衡器可能选择不同的连接
   Balancer.Pick() → SubConn 2 (192.168.1.11:9001)
   ↓
2. 复用 SubConn 2 的 TCP 连接
   创建新的 HTTP/2 Stream ID: 1
   ↓
3. 发送请求...


并发请求: 同时调用多个 RPC
  ↓
请求 A: UserRpc.Login()      → SubConn 1, Stream 1
请求 B: UserRpc.Register()   → SubConn 1, Stream 3
请求 C: UserRpc.GetUserInfo() → SubConn 2, Stream 1

注意: HTTP/2 支持多路复用，同一个 TCP 连接可以同时处理多个请求！
```

## 4. 连接池管理

### 连接数量

```
默认情况:
- 每个服务实例: 1 个 TCP 连接
- 3 个实例: 3 个 TCP 连接

可配置:
- 可以为每个实例建立多个连接
- 通过 grpc.WithDefaultServiceConfig() 配置
```

### 连接复用

```go
// HTTP/2 多路复用示例
SubConn 1 (192.168.1.10:9001)
  └─> TCP 连接
      ├─> Stream 1: Login 请求
      ├─> Stream 3: Register 请求
      ├─> Stream 5: GetUserInfo 请求
      └─> Stream 7: UpdateProfile 请求

所有请求共享同一个 TCP 连接！
```

## 5. 实例变化处理

### 新实例上线

```
etcd Watch 事件: PUT user.rpc/instance-4 = "192.168.1.13:9001"
  ↓
Resolver 收到通知
  ↓
更新实例列表: [10:9001, 11:9001, 12:9001, 13:9001]
  ↓
通知 Balancer
  ↓
Balancer 创建新的 SubConn
  ├─> net.Dial("tcp", "192.168.1.13:9001")
  └─> SubConn 4 加入连接池
  ↓
新实例立即可用于负载均衡
```

### 实例下线

```
etcd Watch 事件: DELETE user.rpc/instance-2
  ↓
Resolver 收到通知
  ↓
更新实例列表: [10:9001, 12:9001, 13:9001]
  ↓
通知 Balancer
  ↓
Balancer 移除 SubConn 2
  ├─> 等待正在进行的请求完成
  ├─> 关闭 TCP 连接
  └─> 从连接池移除
  ↓
后续请求不会再路由到该实例
```

## 6. 连接关闭（Gateway 关闭时）

```go
// gateway/gateway.go
defer server.Stop()
```

### 内部流程：

```
Step 1: 停止接收新请求
  ↓
Step 2: 等待正在进行的 RPC 完成
  ├─> 等待所有 Stream 关闭
  └─> 超时时间: 默认 30 秒
  ↓
Step 3: 关闭所有 SubConn
  ├─> SubConn 1.Close()
  │   └─> 关闭 TCP 连接到 192.168.1.10:9001
  ├─> SubConn 2.Close()
  │   └─> 关闭 TCP 连接到 192.168.1.11:9001
  └─> SubConn 3.Close()
      └─> 关闭 TCP 连接到 192.168.1.12:9001
  ↓
Step 4: 关闭 Resolver
  └─> 断开与 etcd 的连接
  ↓
Step 5: 释放资源
```

## 7. 性能优势

### 长连接 vs 短连接

```
短连接（每次请求都建立连接）:
Request 1:
  ├─> TCP 三次握手 (1.5 RTT)
  ├─> TLS 握手 (2 RTT)
  ├─> 发送请求 (1 RTT)
  ├─> 接收响应
  └─> TCP 四次挥手 (2 RTT)
  总计: ~6.5 RTT

Request 2:
  └─> 重复上述过程 (~6.5 RTT)

总延迟: 13 RTT


长连接（复用连接）:
初始化:
  ├─> TCP 三次握手 (1.5 RTT)
  └─> TLS 握手 (2 RTT)

Request 1:
  ├─> 发送请求 (1 RTT)
  └─> 接收响应

Request 2:
  ├─> 发送请求 (1 RTT)
  └─> 接收响应

总延迟: 5.5 RTT

性能提升: ~57%！
```

### HTTP/2 多路复用优势

```
HTTP/1.1 (需要多个连接):
Connection 1: Request A ────────────> Response A
Connection 2: Request B ────────────> Response B
Connection 3: Request C ────────────> Response C

需要 3 个 TCP 连接


HTTP/2 (单个连接):
Connection 1:
  Stream 1: Request A ────────────> Response A
  Stream 3: Request B ────────────> Response B
  Stream 5: Request C ────────────> Response C

只需 1 个 TCP 连接！

优势:
- 减少连接数
- 降低服务器负载
- 减少内存占用
- 避免队头阻塞
```

## 8. 配置示例

### 在 gateway.yaml 中配置连接参数

```yaml
UserRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: user.rpc

  # 连接超时
  Timeout: 5000  # 5 秒

  # Keep-Alive 配置
  KeepaliveTime: 10s  # 每 10 秒发送 ping

  # 非阻塞模式
  NonBlock: false  # false = 启动时必须连接成功
```

### 在代码中配置

```go
// 自定义连接配置
client := zrpc.MustNewClient(c.UserRpc,
    zrpc.WithDialOption(grpc.WithKeepaliveParams(keepalive.ClientParameters{
        Time:                10 * time.Second,
        Timeout:             3 * time.Second,
        PermitWithoutStream: true,
    })),
    zrpc.WithTimeout(5 * time.Second),
)
```

## 总结

**gRPC ClientConn 的连接特性：**

1. ✅ **长连接**: 启动时建立，持续存在
2. ✅ **连接池**: 为每个实例维护独立连接
3. ✅ **自动维护**: Keep-Alive 心跳保持活跃
4. ✅ **自动重连**: 连接断开时自动恢复
5. ✅ **多路复用**: HTTP/2 支持并发请求
6. ✅ **动态更新**: 实例变化时自动调整连接池
7. ✅ **负载均衡**: 在连接池之上实现

**这就是为什么 gRPC 性能如此优秀的原因！**
