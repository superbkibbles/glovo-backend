# gRPC Integration Guide

## Overview

This document describes the gRPC implementation for efficient inter-service communication in the Glovo Backend microservices architecture.

## Architecture

The system now supports **dual protocols**:
- **HTTP REST APIs**: For client-facing endpoints and external integrations
- **gRPC**: For internal inter-service communication (more efficient)

## Available gRPC Services

### 1. User Service gRPC
**Port**: 9001 (gRPC), 8001 (HTTP)
**Proto**: `proto/user_service.proto`

**Methods**:
- `GetUser(userID)` - Get user information
- `ValidateUser(userID)` - Validate user exists and is active
- `GetUsersBatch(userIDs[])` - Get multiple users (batch operation)
- `UpdateUserStatus(userID, status)` - Update user status (admin)

### 2. Order Service gRPC
**Port**: 9002 (gRPC), 8002 (HTTP)
**Proto**: `proto/order_service.proto`

**Methods**:
- `GetOrder(orderID)` - Get order details
- `UpdateOrderStatus(orderID, status)` - Update order status
- `CreateOrder(orderData)` - Create new order (internal)
- `GetOrdersByStatus(status)` - Get orders by status

### 3. Location Service gRPC
**Port**: 9008 (gRPC), 8008 (HTTP)
**Proto**: `proto/location_service.proto`

**Methods**:
- `UpdateDriverLocation(driverID, location)` - Real-time location update
- `GetDriverLocation(driverID)` - Get current driver location
- `GetNearbyDrivers(location, radius)` - Find nearby drivers
- `StreamDriverLocation(driverID)` - Stream location updates (real-time)
- `UpdateRouteProgress(orderID, location)` - Update delivery progress

### 4. Payment Service gRPC
**Port**: 9007 (gRPC), 8007 (HTTP)
**Proto**: `proto/payment_service.proto`

**Methods**:
- `ProcessPayment(paymentData)` - Process payment transaction
- `ValidatePaymentMethod(methodID, userID)` - Validate payment method
- `GetWalletBalance(userID)` - Get wallet balance
- `ProcessRefund(transactionID, amount)` - Process refund
- `CalculateCommission(orderData)` - Calculate fees and commissions

## Setup Instructions

### 1. Generate Proto Code

```bash
# Install protoc and Go plugins
make install-proto

# Generate gRPC code from proto files
make proto
```

### 2. Start Services with gRPC

Services automatically start both HTTP and gRPC servers:

```bash
# HTTP + gRPC together
docker-compose up

# Or individual services
make run-user    # HTTP:8001, gRPC:9001
make run-order   # HTTP:8002, gRPC:9002
```

### 3. Using gRPC Clients

Example usage in Go:

```go
import "glovo-backend/shared/grpc"

// Create client
userClient, err := grpc.NewUserGRPCClient("localhost:9001")
if err != nil {
    log.Fatal(err)
}
defer userClient.Close()

// Use client
ctx := context.Background()
user, err := userClient.GetUser(ctx, "user123")
if err != nil {
    log.Printf("Error: %v", err)
    return
}

if user.Success {
    fmt.Printf("User: %+v\n", user.User)
}
```

## Inter-Service Communication Patterns

### 1. Order → User Service (Validation)
```go
// Order service validates customer before creating order
userResp, err := userClient.ValidateUser(ctx, customerID)
if !userResp.IsValid || !userResp.IsActive {
    return errors.New("invalid customer")
}
```

### 2. Delivery → Location Service (Real-time Tracking)
```go
// Delivery service gets driver location for tracking
location, err := locationClient.GetDriverLocation(ctx, driverID)
if err == nil {
    // Update delivery tracking info
    updateDeliveryTracking(location)
}
```

### 3. Admin → User Service (User Management)
```go
// Admin service updates user status
_, err := userClient.UpdateUserStatus(ctx, userID, "suspended", "violation")
```

## Port Configuration

| Service | HTTP Port | gRPC Port |
|---------|-----------|-----------|
| User | 8001 | 9001 |
| Order | 8002 | 9002 |
| Catalog | 8003 | 9003 |
| Delivery | 8004 | 9004 |
| Driver | 8005 | 9005 |
| Payment | 8007 | 9007 |
| Location | 8008 | 9008 |
| Notification | 8009 | 9009 |
| Admin | 8010 | 9010 |
| Analytics | 8011 | 9011 |

## Benefits of gRPC Integration

### Performance
- **Faster**: Binary protocol vs JSON
- **Smaller**: Protobuf serialization
- **HTTP/2**: Multiplexing, server push

### Developer Experience
- **Type Safety**: Generated code with strict types
- **Auto-completion**: IDE support for generated clients
- **Versioning**: Schema evolution with backwards compatibility

### Operations
- **Streaming**: Real-time data streams
- **Load Balancing**: Built-in client-side load balancing
- **Monitoring**: Rich metrics and tracing support

## Development Workflow

### 1. Define Service Contract
```proto
service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
```

### 2. Generate Code
```bash
make proto
```

### 3. Implement Server
```go
func (s *userGRPCServer) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
    // Implementation
}
```

### 4. Use Client
```go
client := grpc.NewUserGRPCClient("user-service:9001")
response, err := client.GetUser(ctx, userID)
```

## Testing gRPC Services

### Using grpcurl
```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:9001 list

# Call method
grpcurl -plaintext -d '{"user_id": "123"}' localhost:9001 user.UserService/GetUser
```

### Health Checks
All gRPC services implement health checks:

```bash
grpcurl -plaintext localhost:9001 grpc.health.v1.Health/Check
```

## Future Enhancements

1. **Authentication**: gRPC middleware for JWT validation
2. **Rate Limiting**: Per-service rate limiting
3. **Circuit Breakers**: Resilience patterns
4. **Service Discovery**: Consul/Kubernetes integration
5. **Load Balancing**: Advanced load balancing strategies

## Monitoring

gRPC services expose metrics for:
- Request/response times
- Error rates
- Active connections
- Throughput

Integrate with Prometheus/Grafana for monitoring dashboards.