.PHONY: all proto build test clean docker-up docker-down seed docker-seed tilt-up tilt-down

# Variables
PROTO_DIR := proto
SERVICES := auth-service user-service notification-service file-storage-service restaurant-service order-service delivery-service settings-service gateway

# Default target
all: proto build

# Generate protobuf files
proto:
	@echo "Generating protobuf files..."
	@export PATH="$$(go env GOPATH)/bin:$$PATH"; \
	for dir in $(PROTO_DIR)/*/; do \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			$$dir*.proto; \
	done

# Build all services
build:
	@echo "Building all services..."
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		cd services/$$svc && go build -o bin/$$svc ./cmd/main.go && cd ../..; \
	done

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v -race

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@for svc in $(SERVICES); do \
		rm -rf services/$$svc/bin; \
	done

# Docker compose up
docker-up:
	docker-compose up -d

# Docker compose down
docker-down:
	docker-compose down

# Seed database (run locally with Postgres/MongoDB running)
seed:
	@echo "Seeding database..."
	@DATABASE_URL="$${DATABASE_URL:-postgres://postgres:postgres@localhost:5432/food_delivery?sslmode=disable}" \
	MONGO_URI="$${MONGO_URI:-mongodb://localhost:27017}" \
	MONGO_DB="$${MONGO_DB:-food_delivery}" \
	PASSWORD_PEPPER="$${PASSWORD_PEPPER:-default-pepper-change-me}" \
	SUPERADMIN_USERNAME="$${SUPERADMIN_USERNAME:-admin}" \
	SUPERADMIN_PASSWORD="$${SUPERADMIN_PASSWORD:-Admin@123}" \
	go run ./cmd/seed

# Seed in Docker (runs once)
docker-seed:
	@echo "Seeding database in Docker..."
	@docker-compose up seed

# Tilt up for Kubernetes development
tilt-up:
	@command -v tilt >/dev/null 2>&1 || { \
		echo "Error: tilt is not installed."; \
		echo "Install it with: brew install tilt-dev/tap/tilt"; \
		exit 1; \
	}
	tilt up

# Tilt down
tilt-down:
	@command -v tilt >/dev/null 2>&1 || { \
		echo "Error: tilt is not installed."; \
		exit 1; \
	}
	tilt down

# Helm install (default)
helm-install:
	@command -v helm >/dev/null 2>&1 || { echo "Error: helm is not installed."; exit 1; }
	helm install food-delivery-system ./helm/food-delivery-system --namespace food-delivery-system-helm --create-namespace

# Helm install with dev values
helm-install-dev:
	@command -v helm >/dev/null 2>&1 || { echo "Error: helm is not installed."; exit 1; }
	helm install food-delivery-system ./helm/food-delivery-system -f helm/food-delivery-system/values-dev.yaml --namespace food-delivery-system-helm --create-namespace

# Helm install with prod values
helm-install-prod:
	@command -v helm >/dev/null 2>&1 || { echo "Error: helm is not installed."; exit 1; }
	helm install food-delivery-system ./helm/food-delivery-system -f helm/food-delivery-system/values-prod.yaml --namespace food-delivery-system-helm --create-namespace

# Helm uninstall
helm-uninstall:
	@command -v helm >/dev/null 2>&1 || { echo "Error: helm is not installed."; exit 1; }
	helm uninstall food-delivery-system --namespace food-delivery-system-helm

# Install development dependencies
setup:
	@echo "Installing development dependencies..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo ""
	@echo "For Tilt: brew install tilt-dev/tap/tilt"
	@echo "For Helm: brew install helm"
