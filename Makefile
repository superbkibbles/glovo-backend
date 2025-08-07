# Glovo Backend Makefile

.PHONY: proto build test clean docker-build docker-up docker-down

# Variables
PROTO_DIR=proto
PROTO_GEN_DIR=proto/gen

# Install protobuf compiler and Go plugins
install-proto:
	@echo "Installing protoc and Go plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate protobuf and gRPC code
proto: install-proto
	@echo "Generating protobuf and gRPC code..."
	@mkdir -p $(PROTO_GEN_DIR)
	
	# Generate Go code for each service separately
	protoc --go_out=. --go_opt=module=glovo-backend \
		--go-grpc_out=. --go-grpc_opt=module=glovo-backend \
		$(PROTO_DIR)/user_service.proto
		
	protoc --go_out=. --go_opt=module=glovo-backend \
		--go-grpc_out=. --go-grpc_opt=module=glovo-backend \
		$(PROTO_DIR)/order_service.proto
		
	protoc --go_out=. --go_opt=module=glovo-backend \
		--go-grpc_out=. --go-grpc_opt=module=glovo-backend \
		$(PROTO_DIR)/location_service.proto
		
	protoc --go_out=. --go_opt=module=glovo-backend \
		--go-grpc_out=. --go-grpc_opt=module=glovo-backend \
		$(PROTO_DIR)/payment_service.proto
	
	@echo "Generated proto files in $(PROTO_GEN_DIR)"

# Build all services
build:
	@echo "Building all services..."
	go build -o bin/user-service ./services/user-service/cmd
	go build -o bin/order-service ./services/order-service/cmd
	go build -o bin/catalog-service ./services/catalog-service/cmd
	go build -o bin/delivery-service ./services/delivery-service/cmd
	go build -o bin/driver-service ./services/driver-service/cmd
	go build -o bin/payment-service ./services/payment-service/cmd
	go build -o bin/location-service ./services/location-service/cmd
	go build -o bin/notification-service ./services/notification-service/cmd
	go build -o bin/admin-service ./services/admin-service/cmd
	go build -o bin/analytics-service ./services/analytics-service/cmd

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf $(PROTO_GEN_DIR)

# Docker operations
docker-build:
	@echo "Building Docker images..."
	docker-compose build

docker-up:
	@echo "Starting all services with Docker Compose..."
	docker-compose up -d

docker-down:
	@echo "Stopping all services..."
	docker-compose down

docker-logs:
	@echo "Showing service logs..."
	docker-compose logs -f

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	cp .env.example .env
	@echo "Please edit .env file with your configuration"

# Generate Swagger docs
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g services/user-service/cmd/main.go -o docs/user-service
	swag init -g services/order-service/cmd/main.go -o docs/order-service
	swag init -g services/catalog-service/cmd/main.go -o docs/catalog-service
	swag init -g services/driver-service/cmd/main.go -o docs/driver-service
	swag init -g services/payment-service/cmd/main.go -o docs/payment-service
	swag init -g services/location-service/cmd/main.go -o docs/location-service
	swag init -g services/notification-service/cmd/main.go -o docs/notification-service
	swag init -g services/admin-service/cmd/main.go -o docs/admin-service
	swag init -g services/analytics-service/cmd/main.go -o docs/analytics-service

# Run individual services (for development)
run-user:
	go run ./services/user-service/cmd

run-order:
	go run ./services/order-service/cmd

run-catalog:
	go run ./services/catalog-service/cmd

run-delivery:
	go run ./services/delivery-service/cmd

run-driver:
	go run ./services/driver-service/cmd

run-payment:
	go run ./services/payment-service/cmd

run-location:
	go run ./services/location-service/cmd

run-notification:
	go run ./services/notification-service/cmd

run-admin:
	go run ./services/admin-service/cmd

run-analytics:
	go run ./services/analytics-service/cmd

# Database operations
db-migrate:
	@echo "Running database migrations..."
	@echo "Migrations are handled automatically by each service"

# Health check all services
health-check:
	@echo "Checking service health..."
	@curl -f http://localhost:8001/health || echo "User Service: DOWN"
	@curl -f http://localhost:8002/health || echo "Order Service: DOWN"
	@curl -f http://localhost:8003/health || echo "Catalog Service: DOWN"
	@curl -f http://localhost:8004/health || echo "Delivery Service: DOWN"
	@curl -f http://localhost:8005/health || echo "Driver Service: DOWN"
	@curl -f http://localhost:8007/health || echo "Payment Service: DOWN"
	@curl -f http://localhost:8008/health || echo "Location Service: DOWN"
	@curl -f http://localhost:8009/health || echo "Notification Service: DOWN"
	@curl -f http://localhost:8010/health || echo "Admin Service: DOWN"
	@curl -f http://localhost:8011/health || echo "Analytics Service: DOWN"