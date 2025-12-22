# letsGO Platform - Architecture Documentation

## Table of Contents
1. [System Overview](#system-overview)
2. [Service Descriptions](#service-descriptions)
3. [Data Models](#data-models)
4. [Communication Patterns](#communication-patterns)
5. [Technology Decisions](#technology-decisions)

---

## System Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Client Layer                         │
│   (Web Browser, Mobile App, Third-party Integrations)       │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP/REST
                         │
┌────────────────────────▼────────────────────────────────────┐
│                     API Gateway (Port 8888)                 │
│  • Request routing                                          │
│  • JWT authentication                                       │
│  • Rate limiting                                            │
│  • Request/Response transformation                          │
└────────────────────────┬────────────────────────────────────┘
                         │ gRPC (via etcd)
         ┌───────────────┼───────────────────┬───────────────┐
         │               │                   │               │
    ┌────▼────┐    ┌─────▼────┐     ┌────────▼──┐    ┌───────▼──────┐
    │  User   │    │ Product  │     │   Cart    │    │    Order     │
    │ Service │    │ Service  │     │  Service  │    │   Service    │
    │  :9001  │    │  :9002   │     │   :9003   │    │    :9004     │
    └────┬────┘    └─────┬────┘     └────┬──────┘    └───────┬──────┘
         │               │               │                   │
         │               │               │              ┌────▼──────┐
         │               │               │              │  Payment  │
         │               │               │              │  Service  │
         │               │               │              │   :9005   │
         │               │               │              └────┬──────┘
         │               │               │                   │
┌────────▼───────────────▼───────────────▼───────────────────▼───────┐
│                        Data Layer                                  │
│                                                                    │
│   ┌──────────────┐  ┌──────────────┐  ┌─────────┐  ┌──────────┐    │
│   │  PostgreSQL  │  │   MongoDB    │  │  Redis  │  │  Kafka   │    │
│   │              │  │              │  │         │  │          │    │
│   │ • Users      │  │ • Product    │  │ • Carts │  │ • Events │    │
│   │ • Products   │  │   Attributes │  │ • Cache │  │ • Logs   │    │
│   │ • Orders     │  │ • Reviews    │  │ • Sess. │  │          │    │
│   │ • Payments   │  │              │  │         │  │          │    │
│   └──────────────┘  └──────────────┘  └─────────┘  └──────────┘    │
│                                                                    │
│  ┌──────────────┐                                                  │
│  │     etcd     │  ← Service Discovery & Configuration             │
│  └──────────────┘                                                  │
└────────────────────────────────────────────────────────────────────┘
```

---

## Service Descriptions

### 1. API Gateway

**Purpose**: Single entry point for all client requests

**Responsibilities**:
- Route incoming HTTP requests to appropriate microservices
- JWT token validation and authentication
- Request/response format transformation (HTTP ↔ gRPC)
- Rate limiting and DDoS protection
- API versioning management

**Technology**: go-zero API framework

**Port**: 8888

**Dependencies**: All RPC services (via etcd)

---

### 2. User Service

**Purpose**: User account management and authentication

**Responsibilities**:
- User registration and login
- Password hashing and verification
- JWT token generation and validation
- User profile management
- Session management

**Technology**: go-zero RPC

**Port**: 9001

**Data Storage**:
- **PostgreSQL** (`letsgo_user` database)
  - `users` table: User accounts and credentials
- **Redis** (DB 1): User session cache

**Key Operations**:
- `Register(username, password, email)` → Creates new user
- `Login(username, password)` → Returns JWT token
- `GetUserInfo(userId)` → Returns user profile
- `VerifyToken(token)` → Validates JWT token

---

### 3. Product Service

**Purpose**: Product catalog and inventory management

**Responsibilities**:
- Product CRUD operations
- Inventory tracking and stock management
- Product search and filtering
- Category management
- Product recommendations (future)

**Technology**: go-zero RPC

**Port**: 9002

**Data Storage**:
- **PostgreSQL** (`letsgo_product` database)
  - `products` table: Core product data (id, name, price, stock)
- **MongoDB** (`letsgo_product` database)
  - Product extended data: descriptions, attributes, specifications, reviews
- **Redis** (DB 2): Hot product cache, search results cache

**Why Two Databases?**
- PostgreSQL: Structured data requiring ACID transactions (price, stock)
- MongoDB: Flexible schema for varying product attributes

**Key Operations**:
- `AddProduct(name, price, stock, ...)` → Creates product
- `GetProduct(productId)` → Returns product details
- `ListProducts(page, category, sort)` → Returns product list
- `SearchProducts(keyword)` → Full-text search
- `UpdateStock(productId, quantity)` → Adjusts inventory

---

### 4. Cart Service

**Purpose**: Shopping cart management

**Responsibilities**:
- Add/remove items from cart
- Update item quantities
- Calculate cart totals
- Merge guest and user carts after login
- Cart persistence and expiration

**Technology**: go-zero RPC

**Port**: 9003

**Data Storage**:
- **Redis** (DB 3): Primary storage
  - Key pattern: `cart:{userId}`
  - Data structure: Hash (field=productId, value=quantity)
  - TTL: 7 days (configurable)

**Why Redis?**
- Extremely fast read/write operations
- Shopping carts are temporary by nature
- Easy to set expiration times
- Atomic operations for quantity updates

**Key Operations**:
- `AddToCart(userId, productId, quantity)` → Adds item
- `GetCart(userId)` → Returns all cart items
- `UpdateCartItem(userId, productId, quantity)` → Updates quantity
- `RemoveCartItem(userId, productId)` → Removes item
- `ClearCart(userId)` → Empties cart

---

### 5. Order Service

**Purpose**: Order lifecycle management

**Responsibilities**:
- Create orders from cart items
- Order status tracking (pending → paid → shipped → completed)
- Order history and details
- Order cancellation (if not paid)
- Inventory reservation
- Event publishing to Kafka

**Technology**: go-zero RPC

**Port**: 9004

**Data Storage**:
- **PostgreSQL** (`letsgo_order` database)
  - `orders` table: Order metadata
  - `order_items` table: Products in each order
- **Kafka**: Order events (created, paid, shipped, completed, cancelled)

**Order Status Flow**:
```
1. Pending (待支付)
    ↓ Payment successful
2. Paid (已支付)
    ↓ Warehouse ships
3. Shipped (已发货)
    ↓ Customer receives & confirms
4. Completed (已完成)

Alternative flows:
1. Pending → 5. Cancelled (用户取消 or 超时)
```

**Key Operations**:
- `CreateOrder(userId, items, address)` → Creates order, reserves stock
- `GetOrder(orderId)` → Returns order details
- `ListOrders(userId, page, status)` → Returns order history
- `CancelOrder(orderId)` → Cancels unpaid order, restores stock
- `UpdateOrderStatus(orderId, status)` → Internal status update

**Event Publishing**:
```
Order Created → Kafka (order.created)
    ↓ Triggers: Email notification, inventory update, analytics
Order Paid → Kafka (order.paid)
    ↓ Triggers: Shipping notification, warehouse system
Order Shipped → Kafka (order.shipped)
    ↓ Triggers: SMS notification, delivery tracking
```

---

### 6. Payment Service

**Purpose**: Payment processing and transaction management

**Responsibilities**:
- Create payment transactions
- Integration with payment gateways (Alipay, WeChat Pay, Credit Card)
- Payment status tracking
- Handle payment callbacks/webhooks
- Payment reconciliation
- Notify Order Service on success

**Technology**: go-zero RPC

**Port**: 9005

**Data Storage**:
- **PostgreSQL** (`letsgo_payment` database)
  - `payments` table: Payment records
- **Kafka**: Payment events (success, failed)

**Payment Flow**:
```
1. User clicks "Pay Now"
   ↓
2. Order Service calls Payment Service → CreatePayment()
   ↓
3. Payment Service generates payment_no, returns pay_url
   ↓
4. User redirected to payment gateway (Alipay/WeChat)
   ↓
5. User completes payment on gateway
   ↓
6. Payment gateway calls PaymentCallback() webhook
   ↓
7. Payment Service verifies callback signature
   ↓
8. Payment Service updates status to "success"
   ↓
9. Payment Service calls Order Service → UpdateOrderStatus("paid")
   ↓
10. Payment Service publishes event to Kafka (payment.success)
```

**Key Operations**:
- `CreatePayment(orderId, amount, paymentType)` → Initiates payment
- `QueryPayment(paymentId)` → Checks payment status
- `PaymentCallback(paymentNo, status, tradeNo)` → Processes webhook
- `CancelPayment(paymentId)` → Cancels pending payment

**Security**:
- Signature verification for callbacks
- Idempotency to prevent duplicate processing
- Amount verification (payment amount must match order amount)

---

## Data Models

### User Model (PostgreSQL)

```sql
CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    username      VARCHAR(50) UNIQUE NOT NULL,
    password      VARCHAR(255) NOT NULL,    -- Hashed
    salt          VARCHAR(50) NOT NULL,
    email         VARCHAR(100) UNIQUE NOT NULL,
    phone         VARCHAR(20),
    avatar        VARCHAR(255) DEFAULT '',
    status        SMALLINT DEFAULT 1,       -- 1:active, 2:disabled
    created_at    BIGINT NOT NULL,
    updated_at    BIGINT NOT NULL
);
```

### Product Model (PostgreSQL)

```sql
CREATE TABLE products (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(200) NOT NULL,
    description   TEXT,
    price         DECIMAL(10, 2) NOT NULL,
    stock         BIGINT DEFAULT 0,
    category      VARCHAR(50) NOT NULL,
    images        JSONB,                    -- ["url1", "url2"]
    sales         BIGINT DEFAULT 0,
    status        SMALLINT DEFAULT 1,       -- 1:active, 2:inactive
    created_at    BIGINT NOT NULL,
    updated_at    BIGINT NOT NULL
);
```

### Cart Model (Redis Hash)

```
Key: cart:123 (userId = 123)
Fields:
    1001 → 2     (productId: 1001, quantity: 2)
    1005 → 1     (productId: 1005, quantity: 1)
    1008 → 3     (productId: 1008, quantity: 3)
```

### Order Model (PostgreSQL)

```sql
CREATE TABLE orders (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL,
    order_no      VARCHAR(50) UNIQUE NOT NULL,
    total_amount  DECIMAL(10, 2) NOT NULL,
    status        SMALLINT DEFAULT 1,       -- 1-5
    address       TEXT NOT NULL,
    phone         VARCHAR(20) NOT NULL,
    remark        TEXT DEFAULT '',
    created_at    BIGINT NOT NULL,
    updated_at    BIGINT NOT NULL,
    paid_at       BIGINT DEFAULT 0,
    shipped_at    BIGINT DEFAULT 0,
    completed_at  BIGINT DEFAULT 0
);

CREATE TABLE order_items (
    id            BIGSERIAL PRIMARY KEY,
    order_id      BIGINT NOT NULL REFERENCES orders(id),
    product_id    BIGINT NOT NULL,
    name          VARCHAR(200) NOT NULL,
    price         DECIMAL(10, 2) NOT NULL,  -- Snapshot at purchase
    quantity      BIGINT NOT NULL,
    image         VARCHAR(255),
    created_at    BIGINT NOT NULL
);
```

### Payment Model (PostgreSQL)

```sql
CREATE TABLE payments (
    id            BIGSERIAL PRIMARY KEY,
    order_id      BIGINT UNIQUE NOT NULL,
    user_id       BIGINT NOT NULL,
    payment_no    VARCHAR(50) UNIQUE NOT NULL,
    amount        DECIMAL(10, 2) NOT NULL,
    payment_type  SMALLINT NOT NULL,        -- 1:Alipay, 2:WeChat, 3:Card
    status        SMALLINT DEFAULT 1,       -- 1:pending, 2:success, 3:failed
    trade_no      VARCHAR(100),             -- Third-party transaction ID
    created_at    BIGINT NOT NULL,
    paid_at       BIGINT DEFAULT 0
);
```

---

## Communication Patterns

### 1. Synchronous Communication (gRPC)

**Used for**: Immediate responses required

**Example**: Gateway → User Service (verify token)

```
Gateway receives request with JWT token
    ↓ [gRPC call]
User Service validates token
    ↓ [gRPC response]
Gateway receives user_id
```

**Advantages**:
- Fast (binary protocol)
- Type-safe (protobuf)
- Built-in load balancing
- Streaming support

### 2. Asynchronous Communication (Kafka)

**Used for**: Fire-and-forget operations, event-driven actions

**Example**: Order created → Send email

```
Order Service creates order
    ↓ [Publish to Kafka: order.created]
Order Service returns response immediately
    ↓ (async)
Email Service consumes event
Email Service sends confirmation email
```

**Advantages**:
- Decoupling (services don't depend on each other)
- Reliability (messages persisted in Kafka)
- Scalability (multiple consumers)
- Fault tolerance (can retry)

**Kafka Topics**:
- `order.created`: New order created
- `order.paid`: Order payment successful
- `order.shipped`: Order shipped
- `order.completed`: Order completed
- `order.cancelled`: Order cancelled
- `payment.success`: Payment successful
- `payment.failed`: Payment failed

---

## Technology Decisions

### Why go-zero?

1. **Built-in Service Discovery**: Automatic via etcd
2. **Code Generation**: Reduces boilerplate
3. **Performance**: Written in Go, excellent concurrency
4. **Microservices Ready**: RPC, load balancing, circuit breaker out-of-the-box
5. **Active Community**: Well-documented, maintained

### Why PostgreSQL?

1. **ACID Compliance**: Critical for financial data (orders, payments)
2. **Mature & Stable**: Battle-tested, reliable
3. **Rich Feature Set**: JSON support, full-text search, transactions
4. **Open Source**: No licensing costs

### Why MongoDB?

1. **Flexible Schema**: Product attributes vary widely
2. **Document Storage**: Natural fit for nested data
3. **Horizontal Scaling**: Sharding support
4. **Fast Writes**: Good for reviews, logs

### Why Redis?

1. **Speed**: In-memory, microsecond latency
2. **Data Structures**: Hashes, lists, sets, sorted sets
3. **TTL Support**: Auto-expire carts, cache
4. **Pub/Sub**: Real-time features (future)

### Why Kafka?

1. **High Throughput**: Millions of messages/second
2. **Durability**: Messages persisted to disk
3. **Scalability**: Horizontal scaling via partitions
4. **Replayability**: Can re-consume old messages

### Why etcd?

1. **Service Discovery**: Services auto-register
2. **Configuration**: Distributed config storage
3. **High Availability**: Raft consensus algorithm
4. **Watch API**: Real-time updates

---

## Scalability Considerations

### Horizontal Scaling

Each service can scale independently:

```
Gateway (3 replicas) → Load Balancer
User Service (2 replicas) → etcd manages
Product Service (5 replicas) → etcd manages
Cart Service (3 replicas) → etcd manages
Order Service (4 replicas) → etcd manages
Payment Service (2 replicas) → etcd manages
```

### Database Scaling

- **PostgreSQL**: Read replicas for queries, master for writes
- **MongoDB**: Sharding by product_id
- **Redis**: Redis Cluster or Sentinel for HA
- **Kafka**: Increase partitions for parallelism

### Caching Strategy

1. **Hot Products**: Cache in Redis with 1-hour TTL
2. **User Sessions**: Redis with 24-hour TTL
3. **Search Results**: Redis with 5-minute TTL
4. **Cart Data**: Primary storage in Redis

---

## Security Measures

1. **Authentication**: JWT tokens with expiration
2. **Authorization**: Middleware checks user permissions
3. **Password Security**: Salted + hashed (MD5 → bcrypt in production)
4. **SQL Injection**: Parameterized queries
5. **Rate Limiting**: Prevent abuse and DDoS
6. **HTTPS**: TLS encryption (in production)
7. **Input Validation**: Validate all user inputs

---

**This architecture supports:**
- ✅ High availability
- ✅ Horizontal scaling
- ✅ Fault isolation
- ✅ Independent deployments
- ✅ Technology diversity
- ✅ Future extensibility
