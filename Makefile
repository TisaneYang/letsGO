# ========================================
# letsGO E-commerce Platform - Makefile
# ========================================
# This Makefile provides convenient commands for development

.PHONY: help init generate build run stop clean docker-up docker-down

# Default target - show help
help:
	@echo "letsGO E-commerce Platform - Available Commands:"
	@echo ""
	@echo "  make init          - Initialize project (install dependencies)"
	@echo "  make generate      - Generate code from .api and .proto files"
	@echo "  make build         - Build all services"
	@echo "  make run           - Run all services"
	@echo "  make stop          - Stop all running services"
	@echo "  make clean         - Clean build artifacts"
	@echo ""
	@echo "  make docker-up     - Start all middleware (PostgreSQL, Redis, Kafka, etc.)"
	@echo "  make docker-down   - Stop all middleware"
	@echo "  make docker-clean  - Remove all middleware data volumes"
	@echo ""
	@echo "  make gen-gateway   - Generate gateway API code"
	@echo "  make gen-user      - Generate user service code"
	@echo "  make gen-product   - Generate product service code"
	@echo "  make gen-cart      - Generate cart service code"
	@echo "  make gen-order     - Generate order service code"
	@echo "  make gen-payment   - Generate payment service code"
	@echo ""
	@echo "  make test          - Run all tests"
	@echo ""

# ========================================
# Project Initialization
# ========================================

# Initialize project: install go-zero tools and dependencies
init:
	@echo "Installing go-zero toolchain..."
	go install github.com/zeromicro/go-zero/tools/goctl@latest
	@echo "Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Initializing go modules..."
	go mod init letsgo || true
	go mod tidy
	@echo ""
	@echo "✅ Initialization complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run 'make docker-up' to start middleware"
	@echo "  2. Run 'make generate' to generate code"
	@echo "  3. Run 'make build' to build services"
	@echo ""

# ========================================
# Code Generation
# ========================================

# Generate all code from .api and .proto files
generate: gen-gateway gen-user gen-product gen-cart gen-order gen-payment
	@echo ""
	@echo "✅ All code generated successfully!"
	@echo ""

# Generate gateway API code
gen-gateway:
	@echo "Generating gateway API code..."
	cd gateway && goctl api go -api gateway.api -dir . -style go_zero

# Generate user service code
gen-user:
	@echo "Generating user service code..."
	cd services/user/rpc && goctl rpc protoc user.proto --go_out=. --go-grpc_out=. --zrpc_out=. -style go_zero

# Generate product service code
gen-product:
	@echo "Generating product service code..."
	cd services/product/rpc && goctl rpc protoc product.proto --go_out=. --go-grpc_out=. --zrpc_out=. -style go_zero

# Generate cart service code
gen-cart:
	@echo "Generating cart service code..."
	cd services/cart/rpc && goctl rpc protoc cart.proto --go_out=. --go-grpc_out=. --zrpc_out=. -style go_zero

# Generate order service code
gen-order:
	@echo "Generating order service code..."
	cd services/order/rpc && goctl rpc protoc order.proto --go_out=. --go-grpc_out=. --zrpc_out=. -style go_zero

# Generate payment service code
gen-payment:
	@echo "Generating payment service code..."
	cd services/payment/rpc && goctl rpc protoc payment.proto --go_out=. --go-grpc_out=. --zrpc_out=. -style go_zero

# ========================================
# Docker Middleware Management
# ========================================

# Start all middleware services
docker-up:
	@echo "Starting middleware services..."
	docker-compose up -d
	@echo ""
	@echo "✅ Middleware started!"
	@echo ""
	@echo "Services available at:"
	@echo "  PostgreSQL:    localhost:5432 (user: postgres, password: postgres)"
	@echo "  MongoDB:       localhost:27017 (user: admin, password: admin123)"
	@echo "  Redis:         localhost:6379"
	@echo "  Kafka:         localhost:9092"
	@echo "  etcd:          localhost:2379"
	@echo ""
	@echo "Management UIs:"
	@echo "  Adminer (PostgreSQL): http://localhost:8080"
	@echo "  Mongo Express:        http://localhost:8081"
	@echo ""

# Stop all middleware services
docker-down:
	@echo "Stopping middleware services..."
	docker-compose down
	@echo "✅ Middleware stopped!"

# Stop and remove all data volumes
docker-clean:
	@echo "⚠️  WARNING: This will delete all data!"
	@echo "Press Ctrl+C to cancel, or wait 5 seconds to continue..."
	@sleep 5
	docker-compose down -v
	@echo "✅ Middleware and data removed!"

# ========================================
# Build & Run
# ========================================

# Build all services
build:
	@echo "Building all services..."
	go build -o bin/gateway gateway/gateway.go
	go build -o bin/user-rpc services/user/rpc/user.go
	go build -o bin/product-rpc services/product/rpc/product.go
	go build -o bin/cart-rpc services/cart/rpc/cart.go
	go build -o bin/order-rpc services/order/rpc/order.go
	go build -o bin/payment-rpc services/payment/rpc/payment.go
	@echo "✅ Build complete! Binaries in ./bin/"

# Run all services (in background)
run:
	@echo "Starting all services..."
	@echo "⚠️  Note: Make sure middleware is running (make docker-up)"
	@echo ""
	nohup ./bin/user-rpc -f services/user/rpc/etc/user.yaml > logs/user-rpc.log 2>&1 &
	nohup ./bin/product-rpc -f services/product/rpc/etc/product.yaml > logs/product-rpc.log 2>&1 &
	nohup ./bin/cart-rpc -f services/cart/rpc/etc/cart.yaml > logs/cart-rpc.log 2>&1 &
	nohup ./bin/order-rpc -f services/order/rpc/etc/order.yaml > logs/order-rpc.log 2>&1 &
	nohup ./bin/payment-rpc -f services/payment/rpc/etc/payment.yaml > logs/payment-rpc.log 2>&1 &
	@sleep 2
	nohup ./bin/gateway -f gateway/gateway.yaml > logs/gateway.log 2>&1 &
	@echo ""
	@echo "✅ All services started!"
	@echo ""
	@echo "API Gateway: http://localhost:8888"
	@echo ""
	@echo "View logs:"
	@echo "  tail -f logs/gateway.log"
	@echo "  tail -f logs/user-rpc.log"
	@echo ""

# Stop all running services
stop:
	@echo "Stopping all services..."
	pkill -f "bin/gateway" || true
	pkill -f "bin/user-rpc" || true
	pkill -f "bin/product-rpc" || true
	pkill -f "bin/cart-rpc" || true
	pkill -f "bin/order-rpc" || true
	pkill -f "bin/payment-rpc" || true
	@echo "✅ All services stopped!"

# ========================================
# Cleanup
# ========================================

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf logs/*.log
	@echo "✅ Clean complete!"

# ========================================
# Testing
# ========================================

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...
	@echo "✅ Tests complete!"

# ========================================
# Development Helpers
# ========================================

# Check service status
status:
	@echo "Checking service status..."
	@echo ""
	@echo "Middleware (Docker):"
	@docker-compose ps
	@echo ""
	@echo "Go Services:"
	@ps aux | grep "bin/" | grep -v grep || echo "No services running"
	@echo ""

# View logs for all services
logs:
	tail -f logs/*.log

# View gateway logs
logs-gateway:
	tail -f logs/gateway.log

# Create necessary directories
dirs:
	mkdir -p bin logs services/{user,product,cart,order,payment}/{api,rpc,model}
