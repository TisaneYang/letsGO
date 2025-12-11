# RPC 调用流程详解

## gRPC ClientConn 内部结构

```go
type ClientConn struct {
    // 服务发现 Resolver
    resolver resolver.Resolver  // 从 etcd 获取实例列表

    // 负载均衡器
    balancer balancer.Balancer   // P2C 算法

    // 连接池（SubConn = Sub Connection）
    subConns map[string]*SubConn // 每个实例一个 TCP 连接
    // {
    //   "192.168.1.10:9001": SubConn{conn: *net.TCPConn},
    //   "192.168.1.11:9001": SubConn{conn: *net.TCPConn},
    //   "192.168.1.12:9001": SubConn{conn: *net.TCPConn},
    // }
}
```

## 调用流程

### 步骤 1: Invoke 被调用
```go
// user_grpc.pb.go:68
c.cc.Invoke(ctx, "/user.User/Login", in, out, opts...)
```

### 步骤 2: gRPC 内部处理（伪代码）
```go
func (cc *ClientConn) Invoke(ctx, method, args, reply, opts) error {
    // 1. 获取负载均衡器的 Picker
    picker := cc.balancer.Pick()

    // 2. Picker 根据算法选择一个实例
    result := picker.Pick(PickInfo{
        FullMethodName: "/user.User/Login",
        Ctx:           ctx,
    })
    // result.SubConn = 指向 192.168.1.10:9001 的连接

    // 3. 使用选中的连接发送请求
    err := result.SubConn.Invoke(ctx, method, args, reply, opts)

    // 4. 如果失败，可能重试其他实例
    if err != nil && shouldRetry(err) {
        // 重新 Pick 另一个实例
        result2 := picker.Pick(...)
        err = result2.SubConn.Invoke(...)
    }

    return err
}
```

### 步骤 3: TCP 层发送
```go
func (sc *SubConn) Invoke(ctx, method, args, reply, opts) error {
    // 1. 序列化请求（Protobuf）
    data := proto.Marshal(args)

    // 2. 构造 HTTP/2 帧
    frame := http2.DataFrame{
        StreamID: newStreamID(),
        Data:     data,
    }

    // 3. 通过 TCP 连接发送
    // 直接发送到 192.168.1.10:9001，没有中间代理！
    _, err := sc.conn.Write(frame.Serialize())

    // 4. 等待响应
    response := sc.conn.Read()

    // 5. 反序列化响应
    proto.Unmarshal(response, reply)

    return nil
}
```

## 网络层视图

```
应用层:  Gateway Logic
         ↓
传输层:  gRPC (HTTP/2)
         ↓
网络层:  TCP 连接
         ↓
         直接连接到目标实例
         ↓
物理层:  192.168.1.10:9001
```

## 与传统代理模式的对比

### 传统代理模式（如 Nginx）
```
Client → Nginx (代理) → Backend Server 1
                      → Backend Server 2
                      → Backend Server 3

特点:
- 有中间代理
- 代理负责负载均衡
- 客户端只知道代理地址
- 多一跳网络延迟
```

### gRPC 客户端负载均衡
```
Client → Backend Server 1
      → Backend Server 2
      → Backend Server 3

特点:
- 无中间代理
- 客户端自己负载均衡
- 客户端知道所有实例地址
- 少一跳网络延迟
- 更高性能
```

## 服务发现的作用

etcd 的作用是**告诉客户端有哪些实例**，而不是转发请求：

```
1. 服务注册（User RPC 启动时）
   User RPC (192.168.1.10:9001) → etcd
   写入: key="user.rpc", value="192.168.1.10:9001"

2. 服务发现（Gateway 启动时）
   Gateway → etcd
   查询: key="user.rpc"
   返回: ["192.168.1.10:9001", "192.168.1.11:9001", "192.168.1.12:9001"]

3. RPC 调用（运行时）
   Gateway → 直接连接 → 192.168.1.10:9001

   etcd 不参与！只是在启动时和实例变化时通知客户端。
```

## 连接管理

### 连接建立（启动时）
```go
// Gateway 启动时
for _, addr := range instances {
    // 为每个实例建立 TCP 连接
    conn, _ := net.Dial("tcp", addr)
    subConns[addr] = &SubConn{conn: conn}
}
```

### 连接复用（运行时）
```go
// 每次 RPC 调用
subConn := balancer.Pick()  // 选择一个已有连接
subConn.Invoke(...)         // 复用 TCP 连接
```

### 连接更新（实例变化时）
```go
// etcd Watch 到变化
OnAdd("192.168.1.13:9001") {
    // 新实例上线，建立新连接
    conn := net.Dial("tcp", "192.168.1.13:9001")
    subConns["192.168.1.13:9001"] = &SubConn{conn: conn}
}

OnDelete("192.168.1.10:9001") {
    // 实例下线，关闭连接
    subConns["192.168.1.10:9001"].Close()
    delete(subConns, "192.168.1.10:9001")
}
```

## 总结

**RPC 调用过程：**
1. ✅ 客户端维护服务实例列表（从 etcd 获取）
2. ✅ 客户端维护到每个实例的 TCP 连接池
3. ✅ 客户端使用负载均衡算法选择实例
4. ✅ 客户端直接发送请求到选中的实例
5. ❌ 没有中间代理转发

**优势：**
- 性能更高（少一跳）
- 延迟更低
- 无单点故障（没有中心化的代理）
- 更灵活的负载均衡策略
