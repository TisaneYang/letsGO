# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

letsGO is a production-ready microservices e-commerce platform built with the go-zero framework. It demonstrates microservices architecture with 5 core services (User, Product, Cart, Order, Payment) communicating via gRPC, backed by PostgreSQL, MongoDB, Redis, Kafka, and etcd.

## Development Commands

### Initial Setup
```bash
# Install go-zero tools (goctl) and dependencies
make init

# Start all middleware (PostgreSQL, MongoDB, Redis, Kafka, etcd)
make docker-up

# Generate code from .api and .proto files
make generate

# Build all services
make build

# Run all services
make run
```

### Common Development Tasks
```bash
# Build services after code changes
make build

# Stop and restart services
make stop && make run

# View service logs
tail -f logs/gateway.log
tail -f logs/user-rpc.log

# Run tests
make test

# Clean build artifacts
make clean
```

### Code Generation (after modifying .api or .proto files)
```bash
# Regenerate all services
make generate

# Or generate specific services
make gen-gateway
make gen-user
make gen-product
make gen-cart
make gen-order
make gen-payment
```

### Docker Management
```bash
# Start middleware
make docker-up

# Stop middleware
make docker-down

# Remove all data volumes (⚠️ deletes all data)
make docker-clean

# Check middleware status
docker-compose ps
```

### Service Status
```bash
# Check all services status
make status

# Check specific middleware logs
docker-compose logs postgres
docker-compose logs redis
docker-compose logs kafka
```

## Architecture

### Service Communication Pattern
- **Client → Gateway**: HTTP/REST on port 8888
- **Gateway → Services**: gRPC via etcd service discovery
- **Services → Databases**: Direct connections
- **Services → Kafka**: Async event publishing/consuming

### Service Ports
- API Gateway: 8888
- User Service: 9001
- Product Service: 9002
- Cart Service: 9003
- Order Service: 9004
- Payment Service: 9005

### Middleware
- PostgreSQL: localhost:5432 (user: postgres, password: postgres)
- MongoDB: localhost:27017 (user: admin, password: admin123)
- Redis: localhost:6379
- Kafka: localhost:9092
- etcd: localhost:2379

### Management UIs
- Adminer (PostgreSQL): http://localhost:8080
- Mongo Express (MongoDB): http://localhost:8081

## Code Structure

### Service Organization
Each service follows this structure:
```
services/{service-name}/
├── rpc/                    # gRPC service
│   ├── {service}.proto    # Proto definition
│   ├── {service}.go       # Main entry point
│   ├── etc/              # Configuration
│   │   └── {service}.yaml
│   ├── internal/
│   │   ├── config/       # Config structs
│   │   ├── logic/        # Business logic (IMPLEMENT HERE)
│   │   ├── server/       # gRPC server
│   │   └── svc/          # Service context
│   └── pb/               # Generated protobuf code
└── model/                 # Database models (optional)
```

### Gateway Structure
```
gateway/
├── gateway.api           # API definition (HTTP endpoints)
├── gateway.go           # Main entry point
├── gateway.yaml         # Configuration
└── internal/
    ├── config/          # Config structs
    ├── handler/         # HTTP handlers (generated)
    ├── logic/           # Business logic (IMPLEMENT HERE)
    ├── middleware/      # Auth, logging, etc.
    ├── svc/            # Service context (RPC clients)
    └── types/          # Request/response types (generated)
```

### Common Shared Code
```
common/
├── errorx/              # Custom error types with error codes
├── response/            # Standard API response format
└── utils/              # Shared utilities
```

## Key Implementation Patterns

### Adding a New API Endpoint

1. **Define in .api file** (gateway/gateway.api):
```go
@server (
    prefix: /api/v1/user
    group: user
    middleware: Auth  // If authentication required
)
service gateway {
    @doc "Get user by ID"
    @handler getUserById
    get /info/:id (GetUserByIdReq) returns (GetUserByIdResp)
}
```

2. **Regenerate gateway code**:
```bash
make gen-gateway
```

3. **Implement logic** in `gateway/internal/logic/user/getuserbyidlogic.go`:
```go
func (l *GetUserByIdLogic) GetUserById(req *types.GetUserByIdReq) (*types.GetUserByIdResp, error) {
    // Call RPC service
    user, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &user.GetUserInfoRequest{
        UserId: req.Id,
    })
    if err != nil {
        return nil, err
    }

    return &types.GetUserByIdResp{
        UserId:   user.UserId,
        Username: user.Username,
        Email:    user.Email,
    }, nil
}
```

4. **Rebuild and restart**:
```bash
make build
make stop && make run
```

### Adding a New RPC Method

1. **Define in .proto file** (services/user/rpc/user.proto):
```protobuf
service User {
    rpc GetUserInfo(GetUserInfoRequest) returns (GetUserInfoResponse);
    rpc NewMethod(NewMethodRequest) returns (NewMethodResponse);  // Add this
}

message NewMethodRequest {
    int64 userId = 1;
}

message NewMethodResponse {
    string result = 1;
}
```

2. **Regenerate RPC code**:
```bash
make gen-user  # Or the specific service
```

3. **Implement logic** in `services/user/rpc/internal/logic/newmethodlogic.go`

### Error Handling

Always use structured errors from `common/errorx`:
```go
import "letsgo/common/errorx"

// Return predefined errors
return nil, errorx.ErrUserNotFound
return nil, errorx.ErrProductOutOfStock
return nil, errorx.ErrTokenInvalid

// Or create custom errors
return nil, errorx.NewCodeError(2100, "Custom error message")
```

### Standard Response Format

All API responses follow this format:
```json
{
  "code": 0,
  "msg": "success",
  "data": { ... }
}
```

Use `common/response` helpers in gateway handlers:
```go
import "letsgo/common/response"

// Success response
return response.Success(data), nil

// Error response
return response.Error(code, message), nil
```

### Error Code Ranges
- 0: Success
- 1000-1999: System errors (invalid params, database error, RPC error)
- 2000-2999: User errors (not found, wrong password, invalid token)
- 3000-3999: Product errors (not found, out of stock)
- 4000-4999: Cart errors (empty cart, item not found)
- 5000-5999: Order errors (not found, cannot cancel)
- 6000-6999: Payment errors (not found, payment failed)

### Database Access Patterns

**PostgreSQL** (structured data requiring ACID):
- User accounts, credentials
- Product core data (price, stock)
- Orders and order items
- Payment records

**MongoDB** (flexible schema):
- Product extended attributes
- Product reviews
- Flexible document data

**Redis** (cache and temporary data):
- Shopping cart storage (primary, not cache)
- User session cache
- Hot product cache
- TTL-based expiration

### Kafka Event Publishing

Services publish events for async processing:
```go
// After creating order
message := map[string]interface{}{
    "order_id": orderId,
    "user_id": userId,
    "total_amount": totalAmount,
}

data, _ := json.Marshal(message)
l.svcCtx.KafkaProducer.SendMessage("order.created", data)
```

**Key Topics**:
- order.created, order.paid, order.shipped, order.completed, order.cancelled
- payment.success, payment.failed

## Important Notes

### Service Discovery
Services auto-register with etcd. The gateway discovers services automatically. Service addresses are configured in YAML files under `RpcClientConf` sections.

### Configuration Files
Each service has a YAML config file:
- Gateway: `gateway/gateway.yaml`
- Services: `services/{service}/rpc/etc/{service}.yaml`

Key config sections:
- `Name`: Service name
- `ListenOn`: Port binding
- `Etcd`: Service discovery config
- `DataSource`: Database connection (PostgreSQL)
- `Redis`: Redis connection
- `Mongo`: MongoDB connection (if applicable)
- `Kafka`: Kafka broker config (if applicable)

### Code Generation Workflow
go-zero uses code generation to minimize boilerplate:
1. Define APIs in `.api` files (HTTP) or `.proto` files (gRPC)
2. Run `make generate` to create handlers, types, and service stubs
3. Implement business logic in generated `*logic.go` files
4. **Never modify generated files** except `*logic.go` and `*model.go`

### Testing
Write tests in `*_test.go` files. Run with `make test` or `go test ./...`

### Building for Production
```bash
# Build all services
make build

# Binaries are created in ./bin/
# - gateway
# - user-rpc, product-rpc, cart-rpc, order-rpc, payment-rpc

# Run in production with explicit config
./bin/gateway -f gateway/gateway.yaml
./bin/user-rpc -f services/user/rpc/etc/user.yaml
```

## Data Flow Examples

### User Registration Flow
1. Client → Gateway `/api/v1/user/register` (HTTP POST)
2. Gateway → User Service `Register()` (gRPC)
3. User Service validates, hashes password, saves to PostgreSQL
4. User Service returns success
5. Gateway returns JWT token to client

### Order Creation Flow
1. Client → Gateway `/api/v1/order/create` (HTTP POST)
2. Gateway → Order Service `CreateOrder()` (gRPC)
3. Order Service:
   - Fetches cart items from Cart Service
   - Validates product stock from Product Service
   - Creates order in PostgreSQL
   - Reserves inventory
   - Publishes "order.created" event to Kafka
   - Returns order details
4. Gateway returns order to client
5. Email/Notification service (async) consumes Kafka event

### Shopping Cart Storage
Cart data is stored in Redis as hashes:
```
Key: cart:{userId}
Fields:
  {productId} → {quantity}
  {productId} → {quantity}
```

Operations are atomic and fast. TTL is set to 7 days.
