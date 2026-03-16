# Food Delivery Backend and Dashboard

A Glovo/Talabat-style food delivery platform built with Go microservices and a Next.js dashboard.

## Architecture

- **API Gateway** (Gin) - HTTP REST on port 8080
- **Microservices** (gRPC): auth, user, notification, file-storage, restaurant, order, delivery
- **Databases**: Postgres (GORM) for relational data, MongoDB for flexible data, Redis for sessions

## Quick Start

### Prerequisites

- Go 1.23+
- Docker & Docker Compose
- Node.js 18+ (for dashboard)

### 1. Start infrastructure and services

```bash
docker-compose up -d
```

### 2. Seed the database

```bash
make docker-seed
```

Or run seed locally (with Postgres/MongoDB/Redis running):

```bash
make seed
```

### 3. Default credentials

- **Username:** admin
- **Password:** Admin@123

Change these after first login!

### 4. Run the dashboard

```bash
cd food-delivery-dash
npm install
npm run dev
```

Dashboard: http://localhost:3000  
API: http://localhost:8080

## Development

### Generate protobuf

```bash
make proto
```

### Build all services

```bash
make build
```

### Environment

Copy `.env.example` to `.env` and adjust as needed.

## Kubernetes

### Tilt (local dev with hot reload)

```bash
make tilt-up
```

### Helm

```bash
# Build images first (e.g. docker build -t food-delivery-auth-service:latest -f services/auth-service/Dockerfile .)
make helm-install-dev
```

## Project Structure

```
├── cmd/seed/           # Database seed script
├── proto/              # gRPC definitions
├── pkg/                # Shared packages
├── services/
│   ├── gateway/        # API Gateway
│   ├── auth-service/
│   ├── user-service/
│   ├── notification-service/
│   ├── file-storage-service/
│   ├── restaurant-service/
│   ├── order-service/
│   └── delivery-service/
├── food-delivery-dash/ # Next.js dashboard
├── k8s/                # Kubernetes manifests
└── helm/               # Helm charts
```
