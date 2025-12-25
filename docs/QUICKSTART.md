# Quick Start Guide - letsGO Platform

This is a simplified guide to get you up and running in 5 minutes!

## 🎯 What You'll Build

By the end of this guide, you'll have:
- ✅ All middleware running (PostgreSQL, Redis, Kafka, MongoDB, etcd)
- ✅ 5 microservices running
- ✅ API Gateway accepting requests
- ✅ Sample data loaded

---

## ⚡ 5-Minute Setup

### Step 1: Install Prerequisites (5 minutes)

**Windows:**
```powershell
# Install Go from: https://go.dev/dl/
# Install Docker Desktop from: https://www.docker.com/products/docker-desktop

# Install Make (via Chocolatey)
choco install make
```

**Linux:**
```bash
# Install Go
sudo apt install golang-go

# Install Docker
sudo apt install docker.io docker-compose

# Install Make
sudo apt install make
```

**Mac:**
```bash
# Install Go
brew install go

# Install Docker Desktop from: https://www.docker.com/products/docker-desktop

# Make is already installed
```

### Step 2: Initialize Project (2 minutes)

```bash
cd D:\Projects\letsGO

# This installs all tools and dependencies
make init
```

Wait for it to complete. You'll see "✅ Initialization complete!"

### Step 3: Start Middleware (2 minutes)

```bash
# Start PostgreSQL, MongoDB, Redis, Kafka, etcd
make docker-up
```

Wait ~30 seconds. Verify everything is running:
```bash
docker-compose ps
```

All services should show status "Up".

### Step 4: Generate Code (1 minute)

```bash
# Generate all service code from .api and .proto files
make generate
```

This creates all the boilerplate code for your microservices.

### Step 5: Build & Run (2 minutes)

```bash
# Build all services
make build

# Start all services
make run
```

---

## 🎉 You're Done!

Your e-commerce platform is now running!

**API Gateway:** http://localhost:8888

**Database UIs:**
- PostgreSQL Admin: http://localhost:8080 (System: PostgreSQL, Server: postgres, Username: postgres, Password: postgres)
- MongoDB Admin: http://localhost:8081

---

## 🧪 Test It Out

### 1. Register a User

```bash
curl -X POST http://localhost:8888/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "password": "password123",
    "email": "alice@example.com",
    "phone": "13900139000"
  }'
```

**Expected response:**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "userId": 1,
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

**Save the token!** You'll need it for authenticated requests.

### 2. Login

```bash
curl -X POST http://localhost:8888/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "password": "password123"
  }'
```

### 3. View Products

```bash
curl http://localhost:8888/api/v1/product/list?page=1&pageSize=10
```

You should see sample products (iPhone, MacBook, Nike shoes) that were loaded during database initialization.

### 4. Get Your Profile (Requires Token)

```bash
# Replace YOUR_TOKEN with the token from login
curl http://localhost:8888/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 5. Add Product to Cart (Requires Token)

```bash
curl -X POST http://localhost:8888/api/v1/cart/add \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "productId": 1,
    "quantity": 2
  }'
```

### 6. View Cart

```bash
curl http://localhost:8888/api/v1/cart/ \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 7. Create Order

```bash
curl -X POST http://localhost:8888/api/v1/order/create \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {"productId": 1, "quantity": 2}
    ],
    "address": "123 Main Street, New York, NY 10001",
    "phone": "13900139000",
    "remark": "Please deliver in the morning"
  }'
```

**Expected response:**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "orderId": 1,
    "orderNo": "LG20250104123456789012",
    "totalAmount": 1999.98
  }
}
```

### 8. Create Payment

```bash
curl -X POST http://localhost:8888/api/v1/payment/create \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "orderId": 1,
    "paymentType": 1,
    "amount": 1999.98
  }'
```

### 9. Check Order Status

```bash
curl http://localhost:8888/api/v1/order/list?page=1&pageSize=10 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 🛠️ Useful Commands

```bash
# View logs (gateway)
tail -f logs/gateway.log

# View all service logs
tail -f logs/*.log

# Stop all services
make stop

# Stop middleware
make docker-down

# Restart everything
make stop && make docker-down
make docker-up && make run

# Check service status
make status
```

---

## 🐛 Common Issues

### "Cannot connect to database"

```bash
# Make sure middleware is running
docker-compose ps

# Restart middleware
make docker-down
make docker-up
```

### "Port already in use"

```bash
# Find what's using the port
netstat -ano | findstr :8888   # Windows
lsof -i :8888                  # Linux/Mac

# Stop your services first
make stop

# Then restart
make run
```

### "goctl: command not found"

```bash
# Reinstall
go install github.com/zeromicro/go-zero/tools/goctl@latest

# Check installation
goctl --version
```

---

## 📖 Next Steps

Now that your platform is running:

1. **Explore the code structure** - Look at `services/user/rpc/` to understand how services work
2. **Read the full README.md** - Learn about architecture and development workflow
3. **Try adding a new endpoint** - Follow the development guide in README.md
4. **Customize the platform** - Add your own features!

---

## 🎓 Understanding the Platform

### What Just Happened?

When you ran `make run`, the following started:

1. **User Service (Port 9001)**: Handles authentication and user management
2. **Product Service (Port 9002)**: Manages product catalog
3. **Cart Service (Port 9003)**: Manages shopping carts in Redis
4. **Order Service (Port 9004)**: Processes orders
5. **Payment Service (Port 9005)**: Handles payments
6. **API Gateway (Port 8888)**: Routes requests to appropriate services

### Data Flow Example: Creating an Order

```
1. Client sends POST /api/v1/order/create to Gateway
   ↓
2. Gateway validates JWT token (calls User Service)
   ↓
3. Gateway forwards request to Order Service via gRPC
   ↓
4. Order Service checks stock (calls Product Service)
   ↓
5. Order Service creates order in PostgreSQL
   ↓
6. Order Service publishes "order.created" event to Kafka
   ↓
7. Order Service clears cart (calls Cart Service)
   ↓
8. Order Service returns order details to Gateway
   ↓
9. Gateway returns JSON response to Client
```

### Where is Data Stored?

- **User accounts**: PostgreSQL (`letsgo_user` database)
- **Products**: PostgreSQL (`letsgo_product` database) + MongoDB (extended data)
- **Shopping carts**: Redis (key: `cart:{userId}`)
- **Orders**: PostgreSQL (`letsgo_order` database)
- **Payments**: PostgreSQL (`letsgo_payment` database)
- **Events**: Kafka topics (order.created, payment.success, etc.)

---

## 💡 Tips for Beginners

1. **Use the Management UIs**: http://localhost:8080 lets you see the database tables visually
2. **Check logs frequently**: `tail -f logs/gateway.log` shows what's happening
3. **Start simple**: Test one service at a time before connecting them all
4. **Use Postman**: Import the API endpoints for easier testing
5. **Read go-zero docs**: https://go-zero.dev/en/ is your friend

---

**You're all set! Happy coding! 🚀**
