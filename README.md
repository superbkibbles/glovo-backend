# Glovo-Style Delivery Platform Backend

A comprehensive microservices backend built with Go, implementing a delivery platform using Hexagonal Architecture (Ports & Adapters pattern).

## 🏗️ Architecture Overview

This platform follows the **Hexagonal Architecture** pattern, ensuring clean separation of concerns and high testability. Each service is completely independent with its own database and can be deployed separately.

### Tech Stack

- **Language**: Go 1.21
- **Architecture**: Hexagonal (Ports & Adapters)
- **REST API**: Gin-gonic framework
- **Inter-service Communication**: gRPC
- **Authentication**: JWT with phone number + OTP
- **Databases**: PostgreSQL, Redis, MongoDB
- **Message Queue**: NATS
- **Documentation**: Swagger/OpenAPI
- **Containerization**: Docker & Docker Compose

## 🧱 Microservices

| Service | Port | Database | Description |
|---------|------|----------|-------------|
| **User Service** | 8001 | PostgreSQL + Redis | Authentication, user management, OTP verification |
| **Order Service** | 8002 | PostgreSQL | Order management, tracking, status updates |
| **Catalog Service** | 8003 | PostgreSQL | Stores, products, categories management |
| **Delivery Service** | 8004 | PostgreSQL | Driver assignment, delivery coordination |
| **Driver Service** | 8005 | PostgreSQL | Driver profiles, performance tracking |
| **Location Service** | 8008 | MongoDB | Real-time GPS tracking, location history |
| **Payment Service** | 8007 | PostgreSQL | Wallets, transactions, payments |
| **Notification Service** | 8009 | PostgreSQL + Redis | SMS, email, push notifications |
| **Admin Service** | 8010 | PostgreSQL | Platform administration, analytics |
| **Analytics Service** | 8011 | PostgreSQL | Metrics, reporting, insights |

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for local development)
- Make (optional, for convenience commands)

### 1. Clone & Setup

```bash
git clone <repository-url>
cd GlovoBackend

# Copy environment configuration
cp env.example .env
```

### 2. Configure Environment

Edit `.env` file with your configuration:

```bash
# Essential configurations to change:
JWT_SECRET=your-super-secret-jwt-key
TWILIO_ACCOUNT_SID=your-twilio-sid
TWILIO_AUTH_TOKEN=your-twilio-token
TWILIO_PHONE_NUMBER=your-twilio-number
```

### 3. Start Infrastructure

```bash
# Start databases and message queue
docker-compose up postgres redis mongodb nats -d

# Or start everything
docker-compose up -d
```

### 4. Run Services Locally (Development)

```bash
# User Service
cd services/user-service
go run cmd/main.go

# Order Service
cd services/order-service
go run cmd/main.go

# Continue for other services...
```

### 5. Access Services

- **User Service API**: http://localhost:8001/swagger/index.html
- **Order Service API**: http://localhost:8002/swagger/index.html
- **Health Checks**: http://localhost:800X/health

## 📱 Client Applications

### 1. Customer App Flow

```bash
# 1. Send OTP
curl -X POST http://localhost:8001/api/v1/auth/send-otp \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+1234567890"}'

# 2. Verify OTP & Login
curl -X POST http://localhost:8001/api/v1/auth/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+1234567890", "otp_code": "123456"}'

# 3. Create Order
curl -X POST http://localhost:8002/api/v1/orders \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "merchant_id": "merchant-123",
    "items": [{"product_id": "product-1", "quantity": 2}],
    "delivery_info": {
      "address": "123 Main St",
      "latitude": 40.7128,
      "longitude": -74.0060,
      "phone": "+1234567890"
    },
    "payment_info": {"method": "card", "status": "pending"}
  }'
```

### 2. Merchant Panel Flow

```bash
# 1. Login (same OTP flow)
# 2. Manage Products
curl -X POST http://localhost:8003/api/v1/products \
  -H "Authorization: Bearer <merchant-jwt-token>" \
  -d '{"name": "Pizza Margherita", "price": 12.99, "category": "Pizza"}'

# 3. Update Order Status
curl -X PUT http://localhost:8002/api/v1/orders/{order-id}/status \
  -H "Authorization: Bearer <merchant-jwt-token>" \
  -d '{"status": "confirmed", "estimated_time": 30}'
```

### 3. Driver App Flow

```bash
# 1. Login & Update Status
curl -X PUT http://localhost:8005/api/v1/driver/status \
  -H "Authorization: Bearer <driver-jwt-token>" \
  -d '{"status": "online"}'

# 2. Accept Delivery
curl -X PUT http://localhost:8004/api/v1/deliveries/{delivery-id}/accept \
  -H "Authorization: Bearer <driver-jwt-token>"

# 3. Update Location
curl -X POST http://localhost:8006/api/v1/location \
  -H "Authorization: Bearer <driver-jwt-token>" \
  -d '{"latitude": 40.7128, "longitude": -74.0060}'
```

## 🔐 Authentication & Authorization

### JWT Token Structure

```json
{
  "sub": "user-id",
  "role": "customer|merchant|driver|admin",
  "exp": 1234567890,
  "iat": 1234567890,
  "iss": "glovo-backend"
}
```

### Role-Based Access Control

- **Customer**: Can create orders, view own orders, update profile
- **Merchant**: Can manage products, update order status, view merchant orders
- **Driver**: Can accept deliveries, update location, update delivery status
- **Admin**: Full access to all resources and platform management

## 🏗️ Service Architecture

Each service follows the same hexagonal architecture pattern:

```
services/[service-name]/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── domain/              # Business entities and interfaces
│   ├── app/                 # Use cases (business logic)
│   └── adapters/
│       ├── http/            # REST API handlers
│       ├── grpc/            # gRPC servers
│       ├── db/              # Database repositories
│       └── client/          # External service clients
├── proto/                   # gRPC protocol definitions
├── docs/                    # Generated Swagger docs
├── configs/                 # Configuration files
└── Dockerfile              # Container definition
```

## 🔌 Inter-Service Communication

### REST API (External)
- Client apps communicate with services via REST APIs
- Each service exposes Swagger documentation
- Authentication via JWT Bearer tokens

### gRPC (Internal)
- Services communicate internally via gRPC
- High-performance, type-safe communication
- Service discovery and load balancing ready

### Event-Driven (Async)
- NATS for publishing/subscribing to events
- Order events, notification triggers
- Eventual consistency between services

## 📊 Monitoring & Health Checks

### Health Check Endpoints

Each service exposes a health check endpoint:

```bash
curl http://localhost:800X/health
```

Response:
```json
{
  "status": "healthy",
  "service": "service-name",
  "timestamp": "2023-01-01T00:00:00Z"
}
```

### Swagger Documentation

Access API documentation for each service:
- User Service: http://localhost:8001/swagger/index.html
- Order Service: http://localhost:8002/swagger/index.html
- [Continue for all services...]

## 🧪 Testing

### Unit Tests

```bash
# Run tests for a specific service
cd services/user-service
go test ./...

# Run tests with coverage
go test -cover ./...
```

### Integration Tests

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test -tags=integration ./...
```

### API Testing

Use the provided Postman collection or curl commands to test APIs.

## 🚀 Deployment

### Docker Compose (Development)

```bash
docker-compose up -d
```

### Kubernetes (Production)

```bash
# Apply Kubernetes manifests
kubectl apply -f k8s/

# Or use Helm charts
helm install glovo-backend ./helm/glovo-backend
```

### CI/CD Pipeline

The project includes GitHub Actions workflows for:
- Running tests
- Building Docker images
- Deploying to staging/production

## 🔧 Development

### Code Generation

```bash
# Generate Swagger docs
go install github.com/swaggo/swag/cmd/swag@latest
cd services/user-service
swag init -g cmd/main.go

# Generate gRPC code
protoc --go_out=. --go-grpc_out=. proto/*.proto
```

### Adding a New Service

1. Create service directory structure
2. Implement domain models and interfaces
3. Add business logic in app layer
4. Create adapters (HTTP, gRPC, DB)
5. Add to docker-compose.yml
6. Update documentation

## 📈 Performance Considerations

### Database Optimization
- Each service has its own database (data isolation)
- Connection pooling configured
- Database indexes on frequently queried fields

### Caching Strategy
- Redis for session storage and OTP verification
- Application-level caching for frequently accessed data
- HTTP caching headers for static data

### Scaling
- Stateless services (horizontally scalable)
- Database read replicas for read-heavy services
- Load balancing ready architecture

## 🛡️ Security

### Authentication
- JWT tokens with expiration
- Phone number + OTP verification
- Role-based access control

### Data Protection
- Environment variables for sensitive configuration
- Database password encryption
- HTTPS/TLS in production

### Rate Limiting
- API rate limiting per user/IP
- OTP request rate limiting
- Circuit breaker pattern for external services

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Implement changes following the architecture patterns
4. Add tests for new functionality
5. Update documentation
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For questions and support:
- Create an issue in the repository
- Check the API documentation
- Review the architecture diagrams

---

**Happy coding! 🚀** 