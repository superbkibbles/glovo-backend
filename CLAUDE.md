# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Layout

This is a Go monorepo using Go workspaces (`go.work`) with a co-located Next.js frontend:

```
FoodDeliveryBackend/
├── services/          # 9 microservices, each with cmd/main.go + internal/
├── proto/             # Protobuf definitions shared across services
├── pkg/               # Shared Go packages (config, database, logger, middleware, utils)
├── cmd/seed/          # One-time database seeding script
├── food-delivery-dash/# Next.js 16 admin frontend (separate pnpm project)
├── k8s/               # Kubernetes manifests
├── helm/              # Helm charts
├── docker-compose.yml
├── Makefile
└── go.work            # Workspace including pkg, proto, and all 9 services
```

## Backend Commands

```bash
# Install protoc plugins
make setup

# Regenerate Go code from .proto files (required after any proto change)
make proto

# Build all services
make build

# Build a single service
cd services/auth-service && go build -o bin/auth-service ./cmd/main.go

# Run all tests
make test

# Run tests for a single service
cd services/auth-service && go test ./... -v -race

# Local dev (Docker Compose — starts all services + Postgres, MongoDB, Redis)
make docker-up
make docker-down

# Seed the database (requires running Postgres + MongoDB)
make seed            # local
make docker-seed     # in Docker

# Kubernetes dev with hot reload
make tilt-up
make tilt-down

# Helm deploy
make helm-install-dev
make helm-install-prod
```

## Frontend Commands (food-delivery-dash/)

```bash
cd food-delivery-dash

pnpm install
pnpm dev          # Next.js dev server on :3000
pnpm build
pnpm start
pnpm lint

# With nginx reverse proxy (port 80 → 3000)
pnpm dev:nginx    # then run ./start-nginx.sh separately
```

## Architecture: Request Flow

1. **Frontend** (Next.js, :3000) → authenticates with NextAuth using CredentialsProvider
2. **API Gateway** (Gin, :8080) — the only HTTP entry point; handles JWT validation, CORS, rate limiting (100 req/min), and WebSocket
3. Gateway makes **gRPC calls** to the appropriate backend service
4. Each **microservice** owns its own database connections and business logic

```
Browser → Gateway (:8080) → gRPC → auth-service (:50051)
                                  → user-service (:50052)
                                  → notification-service (:50053)
                                  → file-storage-service (:50054)
                                  → restaurant-service (:50055)
                                  → order-service (:50056)
                                  → delivery-service (:50057)
                                  → settings-service
```

All gRPC contracts live in `proto/<service>/` and the generated Go code lives alongside them (same directory). The `proto` module is listed in `go.work` and imported by all services.

## Architecture: Service Internals

Each service follows a layered structure under `internal/`:

```
internal/
├── domain/       # Entities and GORM models
├── repository/   # Database access (Postgres via GORM, MongoDB, Redis)
├── application/  # Business logic / service layer
└── grpc/         # gRPC server implementation (registers with proto-generated interface)
```

The gateway's `internal/` is different — it only has `grpc/` (clients), `handlers/` (Gin HTTP handlers), `middleware/`, and `websocket/`.

## Shared Packages (pkg/)

- `pkg/config` — loads all env vars via `config.Load()`; used by every service's `main.go`
- `pkg/database` — `NewPostgres()`, `NewRedisDB()`, `NewMongoDB()` constructors
- `pkg/logger` — zerolog wrapper; `logger.Init(env)` then `logger.WithService("name")`
- `pkg/middleware` — gRPC auth interceptor, permissions list
- `pkg/utils` — password hashing (bcrypt + pepper + salt)

## Authentication

- Backend issues JWT access tokens (24h) and refresh tokens (168h)
- Gateway's `middleware.Auth(cfg)` validates JWTs on all protected routes
- Each gRPC service also has its own `middleware.NewAuthInterceptor` for internal calls; public methods are explicitly whitelisted in each service's `main.go`
- Frontend uses NextAuth with a custom CredentialsProvider that calls `POST /api/v1/auth/login`; tokens are stored in the JWT session and attached as `Authorization: Bearer <token>` on every Axios request

## Real-time (WebSocket)

The gateway runs a Hub (`internal/websocket/Hub`) that manages connected clients. Notifications broadcast via `POST /internal/notifications/broadcast` are fanned out to connected WebSocket clients at `GET /ws/notifications`. The frontend's `lib/websocket/` connects to this endpoint.

## Frontend Routing

The Next.js app uses the App Router with locale-based routing:
- All pages live under `app/[locale]/`
- `(auth)` group — public pages (login, signup, forgot-password)
- `(dashboard)` group — protected pages; `middleware.ts` redirects unauthenticated users to `/<locale>/login`
- Supported locales: `en`, `ar`, `ku` (configured in `i18n/routing.ts`, translations in `messages/`)

## Environment Setup

Copy `.env.example` to `.env` in the repo root for backend services. Key variables:

| Variable | Purpose |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `MONGO_URI` | MongoDB connection string |
| `REDIS_URI` / `REDIS_PASSWORD` | Redis |
| `JWT_SECRET` | Shared secret for all JWT signing/validation |
| `PASSWORD_PEPPER` | Added to all password hashes |
| `SUPERADMIN_USERNAME` / `SUPERADMIN_PASSWORD` | Auto-created on first boot (default: `admin` / `Admin@123`) |

Frontend: copy `food-delivery-dash/.env.local.example` to `.env.local`:

| Variable | Purpose |
|---|---|
| `NEXT_PUBLIC_API_URL` | Gateway URL, e.g. `http://localhost:8080/api/v1` |
| `NEXTAUTH_SECRET` | Min 32 chars; must match across restarts |
| `NEXTAUTH_URL` | NextAuth callback base URL |

## Proto Regeneration

After editing any `.proto` file, run `make proto` from the repo root. This runs `protoc` over every subdirectory of `proto/` and writes `*.pb.go` / `*_grpc.pb.go` files in-place. Requires `protoc-gen-go` and `protoc-gen-go-grpc` (installed via `make setup`).

## Swagger Docs

The gateway exposes Swagger UI at `http://localhost:8080/swagger/index.html`. Docs are generated from `//go:generate` annotations in handlers. To regenerate after handler changes, install `swag` and run `swag init` from `services/gateway/`.
