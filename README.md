# letsGO - E-commerce Platform

A complete microservices-based e-commerce platform built with go-zero framework.

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Technology Stack](#technology-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Development Guide](#development-guide)
- [API Documentation](#api-documentation)
- [Deployment](#deployment)
- [Troubleshooting](#troubleshooting)

---

## 🎯 Overview

**letsGO** is a production-ready e-commerce platform that demonstrates microservices architecture best practices using the go-zero framework. It includes:

- **User Management**: Registration, login, profile management with JWT authentication
- **Product Catalog**: Product listing, search, inventory management
- **Shopping Cart**: Fast cart operations using Redis
- **Order Management**: Order creation, status tracking, cancellation
- **Payment Processing**: Multiple payment methods with mock implementation

---

## 🏗️ Architecture

### System Architecture Diagram

```
┌─────────────┐
│   Client    │
│  (Frontend) │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────────────┐
│         API Gateway (:8888)                 │
│  - Route requests to services               │
│  - JWT authentication                       │
│  - Rate limiting                            │
└──────┬──────────────────────────────────────┘
       │
       ├─────────────┬──────────────┬──────────────┬──────────────┐
       ▼             ▼              ▼              ▼              ▼
  ┌──────────┐  ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
  │   User   │  │ Product  │   │   Cart   │   │  Order   │   │ Payment  │
  │ Service  │  │ Service  │   │ Service  │   │ Service  │   │ Service  │
  │  :9001   │  │  :9002   │   │  :9003   │   │  :9004   │   │  :9005   │
  └────┬─────┘  └────┬─────┘   └────┬─────┘   └────┬─────┘   └────┬─────┘
       │             │              │              │              │
       ▼             ▼              ▼              ▼              ▼
┌────────────────────────────────────────────────────────────────────┐
│                          Middleware Layer                          │
│    ┌──────────┐  ┌──────────┐  ┌───────┐  ┌───────┐  ┌──────┐      │
│    │PostgreSQL│  │ MongoDB  │  │ Redis │  │ Kafka │  │ etcd │      │
│    │  :5432   │  │  :27017  │  │ :6379 │  │ :9092 │  │:2379 │      │
│    └──────────┘  └──────────┘  └───────┘  └───────┘  └──────┘      │
└────────────────────────────────────────────────────────────────────┘
```

### Service Communication

- **Client ↔ Gateway**: HTTP/REST
- **Gateway ↔ Services**: gRPC (via etcd service discovery)
- **Services ↔ Database**: Direct connection
- **Services ↔ Kafka**: Async event publishing/consuming

---

## 🛠️ Technology Stack

### Core Framework
- **go-zero**: Microservices framework with built-in service discovery, load balancing, circuit breaker

### Databases
- **PostgreSQL**: Relational database for structured data (users, orders, payments, products)
- **MongoDB**: Document database for flexible data (product attributes, reviews)

### Cache & Session
- **Redis**: In-memory cache for shopping carts, session management, hot data caching

### Message Queue
- **Kafka**: Async event processing (order notifications, inventory updates, analytics)

### Service Discovery
- **etcd**: Service registration and discovery

### Tools
- **Docker**: Container platform for middleware
- **goctl**: go-zero CLI tool for code generation
- **protoc**: Protocol Buffers compiler

---

## 📁 Project Structure

```
letsGO/
├── gateway/                    # API Gateway
│   ├── gateway.api            # API definition
│   └── gateway.yaml           # Configuration
│
├── services/                   # Microservices
│   ├── user/                  # User service
│   │   ├── api/              # HTTP API layer
│   │   ├── rpc/              # gRPC service
│   │   │   ├── user.proto   # Proto definition
│   │   │   └── etc/         # Configuration
│   │   └── model/            # Database models
│   │
│   ├── product/              # Product service
│   ├── cart/                 # Cart service
│   ├── order/                # Order service
│   └── payment/              # Payment service
│
├── common/                     # Shared code
│   ├── errorx/               # Error handling
│   ├── response/             # Standard responses
│   ├── middleware/           # Middleware (Auth, etc.)
│   └── utils/                # Utility functions
│
├── deploy/                     # Deployment files
│   ├── docker/               # Docker configs
│   │   └── init-db.sql      # Database initialization
│   └── k8s/                  # Kubernetes manifests (optional)
│
├── docker-compose.yml          # Middleware services
├── Makefile                    # Build automation
├── go.mod                      # Go modules
└── README.md                   # This file
```

---

## 🚀 Getting Started

### Prerequisites

Before starting, ensure you have the following installed:

1. **Go** (version 1.21 or higher)
   ```bash
   go version
   ```

2. **Docker & Docker Compose**
   ```bash
   docker --version
   docker-compose --version
   ```

3. **Make** (usually pre-installed on Linux/Mac, install on Windows via chocolatey)
   ```bash
   make --version
   ```

### Installation Steps

#### Step 1: Initialize Project

```bash
# Install go-zero toolchain and dependencies
make init
```

This will:
- Install `goctl` (go-zero code generator)
- Install Protocol Buffers plugins
- Initialize Go modules

#### Step 2: Start Middleware

```bash
# Start PostgreSQL, MongoDB, Redis, Kafka, etcd
make docker-up
```

Wait about 30 seconds for all services to be ready. Check status:
```bash
docker-compose ps
```

You should see all services with status "Up".

**Access Management UIs:**
- Adminer (PostgreSQL): http://localhost:8080
- Mongo Express (MongoDB): http://localhost:8081

#### Step 3: Generate Code

```bash
# Generate all service code from .api and .proto files
make generate
```

This creates:
- HTTP handlers from `.api` files
- gRPC stubs from `.proto` files
- Service skeletons with all boilerplate

#### Step 4: Build Services

```bash
# Build all microservices
make build
```

Binaries will be created in the `./bin/` directory.

#### Step 5: Run Services

```bash
# Start all services
make run
```

Services will start in the background:
- User Service: :9001
- Product Service: :9002
- Cart Service: :9003
- Order Service: :9004
- Payment Service: :9005
- **API Gateway: :8888** ← Your main endpoint

#### Step 6: Test the Platform

The API Gateway is now running at `http://localhost:8888`. Test it:

```bash
# Check if gateway is running
curl http://localhost:8888/health

# Register a new user
curl -X POST http://localhost:8888/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "password": "password123",
    "email": "john@example.com",
    "phone": "13800138000"
  }'

# Login
curl -X POST http://localhost:8888/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "password": "password123"
  }'

# List products (public endpoint)
curl http://localhost:8888/api/v1/product/list?page=1&pageSize=10
```

---

## 💻 Development Guide

### Code Generation Workflow

go-zero uses **code generation** to create boilerplate code:

1. **Define API** in `.api` files (for HTTP endpoints)
2. **Define Services** in `.proto` files (for gRPC services)
3. **Generate code** using `make generate`
4. **Implement business logic** in generated handler/logic files

### Example: Adding a New Endpoint

Let's add a "Get User by ID" endpoint:

#### 1. Update API Definition

Edit `gateway/gateway.api`:

```go
@server (
    prefix:     /api/v1/user
    group:      user
    middleware: Auth
)
service gateway {
    @doc "Get user by ID"
    @handler getUserById
    get /info/:id (GetUserByIdReq) returns (GetUserByIdResp)
}

type (
    GetUserByIdReq {
        Id int64 `path:"id"`
    }

    GetUserByIdResp {
        UserId   int64  `json:"userId"`
        Username string `json:"username"`
        Email    string `json:"email"`
    }
)
```

#### 2. Regenerate Code

```bash
make gen-gateway
```

#### 3. Implement Logic

The generator creates `gateway/internal/logic/user/getuserbyidlogic.go`. Implement:

```go
func (l *GetUserByIdLogic) GetUserById(req *types.GetUserByIdReq) (*types.GetUserByIdResp, error) {
    // Call user RPC service
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

#### 4. Test

```bash
make build
make stop && make run

curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8888/api/v1/user/info/1
```

### Working with Databases

#### PostgreSQL Example

Database models are in `services/*/model/` directories. Example:

```go
// services/user/model/user_model.go

type User struct {
    Id        int64  `db:"id"`
    Username  string `db:"username"`
    Password  string `db:"password"`
    Email     string `db:"email"`
    CreatedAt int64  `db:"created_at"`
}

func (m *UserModel) FindOneByUsername(username string) (*User, error) {
    query := `SELECT * FROM users WHERE username = $1 LIMIT 1`
    var user User
    err := m.conn.QueryRow(&user, query, username)
    return &user, err
}
```

#### Redis Example (Cart Service)

```go
// Store cart in Redis
func (l *AddToCartLogic) AddToCart(req *cart.AddToCartRequest) error {
    key := fmt.Sprintf("cart:%d", req.UserId)

    // Store as hash: field=productId, value=quantity
    return l.svcCtx.Redis.Hset(key,
        strconv.FormatInt(req.ProductId, 10),
        strconv.FormatInt(req.Quantity, 10))
}
```

### Using Kafka

#### Publishing Events

```go
// services/order/internal/logic/create_order_logic.go

// After order is created, publish event to Kafka
message := map[string]interface{}{
    "order_id": orderId,
    "user_id": userId,
    "total_amount": totalAmount,
}

data, _ := json.Marshal(message)
l.svcCtx.KafkaProducer.SendMessage("order.created", data)
```

#### Consuming Events

```go
// Create a Kafka consumer service
consumer, err := kafka.NewConsumer([]string{"localhost:9092"}, "order-consumer-group")
consumer.Subscribe("order.created", func(message []byte) {
    // Process order created event
    // Send email, update analytics, etc.
})
```

---

## 📚 API Documentation

### Authentication

Most endpoints require JWT authentication. Include token in header:

```
Authorization: Bearer YOUR_JWT_TOKEN
```

Get token by calling `/api/v1/user/login`.

### User APIs

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/user/register` | Register new user | No |
| POST | `/api/v1/user/login` | User login | No |
| GET | `/api/v1/user/profile` | Get user profile | Yes |
| PUT | `/api/v1/user/profile` | Update profile | Yes |

### Product APIs

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/product/list` | List products | No |
| GET | `/api/v1/product/detail/:id` | Get product detail | No |
| GET | `/api/v1/product/search` | Search products | No |
| POST | `/api/v1/product/add` | Add product (admin) | Yes |
| PUT | `/api/v1/product/update` | Update product (admin) | Yes |

### Cart APIs (All require authentication)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/cart/` | Get cart |
| POST | `/api/v1/cart/add` | Add to cart |
| PUT | `/api/v1/cart/update` | Update cart item |
| DELETE | `/api/v1/cart/remove/:itemId` | Remove item |
| DELETE | `/api/v1/cart/clear` | Clear cart |

### Order APIs (All require authentication)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/order/create` | Create order |
| GET | `/api/v1/order/list` | List orders |
| GET | `/api/v1/order/detail/:id` | Get order detail |
| PUT | `/api/v1/order/cancel/:id` | Cancel order |

### Payment APIs (All require authentication)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/payment/create` | Create payment |
| GET | `/api/v1/payment/query/:orderId` | Query payment status |
| POST | `/api/v1/payment/callback` | Payment callback (webhook) |

### Standard Response Format

All APIs return responses in this format:

**Success:**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    // Response data here
  }
}
```

**Error:**
```json
{
  "code": 2000,
  "msg": "User not found",
  "data": null
}
```

### Error Codes

| Code Range | Category | Examples |
|------------|----------|----------|
| 0 | Success | - |
| 1000-1999 | System Errors | 1001: Invalid params, 1002: Database error |
| 2000-2999 | User Errors | 2000: User not found, 2003: Invalid token |
| 3000-3999 | Product Errors | 3000: Product not found, 3001: Out of stock |
| 4000-4999 | Cart Errors | 4000: Cart empty |
| 5000-5999 | Order Errors | 5000: Order not found, 5002: Cannot cancel |
| 6000-6999 | Payment Errors | 6001: Payment failed |

---

## 🚢 Deployment

### Local Development
Already covered in [Getting Started](#getting-started).

### Production Deployment

#### Option 1: Docker Deployment

Build Docker images for each service:

```bash
# Example Dockerfile for user service
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o user-rpc services/user/rpc/user.go

FROM alpine:latest
COPY --from=builder /app/user-rpc /app/
ENTRYPOINT ["/app/user-rpc"]
```

#### Option 2: Kubernetes Deployment

Deploy to Kubernetes cluster (manifests in `deploy/k8s/`):

```bash
kubectl apply -f deploy/k8s/
```

#### Environment Variables

Configure services via environment variables in production:

- `DB_HOST`: PostgreSQL host
- `REDIS_HOST`: Redis host
- `KAFKA_BROKERS`: Kafka broker list
- `ETCD_HOSTS`: etcd cluster hosts
- `JWT_SECRET`: JWT signing key

---

## 🔧 Troubleshooting

### Services Not Starting

**Problem**: Services fail to start or can't connect to databases.

**Solution**:
```bash
# Check if middleware is running
docker-compose ps

# Check logs
docker-compose logs postgres
docker-compose logs redis

# Restart middleware
make docker-down
make docker-up
```

### Port Already in Use

**Problem**: Error like "address already in use".

**Solution**:
```bash
# Find process using port 8888
netstat -ano | findstr :8888  # Windows
lsof -i :8888                 # Linux/Mac

# Kill the process or change port in config
```

### Code Generation Fails

**Problem**: `make generate` fails.

**Solution**:
```bash
# Reinstall goctl
go install github.com/zeromicro/go-zero/tools/goctl@latest

# Check PATH
goctl --version

# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Database Connection Errors

**Problem**: Services can't connect to PostgreSQL.

**Solution**:
```bash
# Check if PostgreSQL is accepting connections
docker exec -it letsgo-postgres psql -U postgres -c "SELECT 1"

# Check configuration in services/*/rpc/etc/*.yaml
# Make sure DataSource URL is correct
```

### View Service Logs

```bash
# All logs
tail -f logs/*.log

# Specific service
tail -f logs/gateway.log
tail -f logs/user-rpc.log
```

---

## 📝 Useful Commands

```bash
# Start everything from scratch
make docker-up && make generate && make build && make run

# Stop everything
make stop && make docker-down

# Clean and rebuild
make clean && make build

# Check service status
make status

# View all logs
make logs

# Run tests
make test
```

---

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Write tests
5. Submit a pull request

---

## 📄 License

This project is licensed under the MIT License.

---

## 📧 Contact

For questions or support, please open an issue on GitHub.

---

## 🎓 Learning Resources

- [go-zero Documentation](https://go-zero.dev/)
- [gRPC Documentation](https://grpc.io/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Redis Documentation](https://redis.io/documentation)
- [Kafka Documentation](https://kafka.apache.org/documentation/)

---

**Happy Coding! 🚀**
