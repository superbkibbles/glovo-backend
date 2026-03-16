# Food Delivery Backend – Documentation

This document describes the backend architecture, services, APIs, data models, and deployment for the Food Delivery system.

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Services](#2-services)
3. [API Gateway](#3-api-gateway)
4. [Protocol Buffers (gRPC)](#4-protocol-buffers-grpc)
5. [Data Storage](#5-data-storage)
6. [Authentication & Authorization](#6-authentication--authorization)
7. [Configuration](#7-configuration)
8. [Deployment](#8-deployment)

---

## 1. Architecture Overview

The backend is a **microservices** architecture with:

- **API Gateway** – HTTP REST API on port 8080, routes requests to gRPC services
- **gRPC services** – Internal services communicating via gRPC
- **Databases** – PostgreSQL (relational), MongoDB (documents), Redis (cache/sessions)

```
┌─────────────┐     HTTP/REST      ┌──────────────┐     gRPC      ┌─────────────────────┐
│   Client    │ ◄────────────────► │   Gateway    │ ◄───────────► │  auth-service       │
│  (Frontend) │                    │   :8080      │               │  user-service       │
└─────────────┘                    └──────────────┘               │  restaurant-service  │
                                        │                         │  order-service      │
                                        │ WebSocket               │  delivery-service   │
                                        │                         │  settings-service   │
                                        ▼                         │  notification-svc   │
                                 ┌──────────────┐                 │  file-storage-svc   │
                                 │  WebSocket   │                 └─────────────────────┘
                                 │   Hub        │
                                 └──────────────┘
```

---

## 2. Services

| Service | Port | Purpose | Database |
|---------|------|---------|----------|
| **auth-service** | 50051 | Login, signup/signin (phone OTP, email OTP, Google), JWT, refresh, logout, password change | Postgres, Redis |
| **user-service** | 50052 | User and role management | Postgres |
| **notification-service** | 50053 | Notifications, Firebase push, device tokens | MongoDB |
| **file-storage-service** | 50054 | File upload/download | MongoDB, local storage |
| **restaurant-service** | 50055 | Restaurants, categories, menu items | Postgres |
| **order-service** | 50056 | Order lifecycle, commission from settings | Postgres |
| **delivery-service** | 50057 | Driver assignment, tracking, location updates | Postgres, Redis |
| **settings-service** | 50058 | Commission, operating areas | MongoDB |
| **gateway** | 8080 | HTTP API, routes to gRPC, WebSocket | - |

### Infrastructure (Docker Compose)

- **PostgreSQL** – 5432
- **MongoDB** – 27017
- **Redis** – 6379

---

## 3. API Gateway

Base path: `/api/v1`

### Public Routes (no auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/auth/login` | Admin login (username/password) |
| POST | `/auth/refresh-token` | Refresh JWT |
| POST | `/auth/signup/phone` | Customer signup (phone + OTP) |
| POST | `/auth/signup/email` | Customer signup (email + OTP) |
| POST | `/auth/signup/google` | Customer signup (Google ID token) |
| POST | `/auth/signin/phone` | Customer signin (phone + OTP) |
| POST | `/auth/signin/email` | Customer signin (email + OTP) |
| POST | `/auth/signin/google` | Customer signin (Google ID token) |
| POST | `/auth/send-otp` | Send OTP to phone or email |
| GET | `/ws/notifications` | WebSocket for real-time notifications |

### Internal Routes (no auth)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/internal/notifications/broadcast` | Broadcast notification |
| GET | `/internal/notifications/check-connection` | Check WebSocket connection |

### Protected Routes (JWT required)

#### Auth

| Method | Path | Description |
|--------|------|-------------|
| GET | `/auth/me` | Current user info |
| POST | `/auth/change-password` | Change password |
| POST | `/auth/logout` | Logout |

#### Users

| Method | Path | Description |
|--------|------|-------------|
| GET | `/users` | List users |
| POST | `/users` | Create user |
| GET | `/users/:id` | Get user |
| PUT | `/users/:id` | Update user |
| DELETE | `/users/:id` | Delete user |

#### Roles

| Method | Path | Description |
|--------|------|-------------|
| GET | `/roles` | List roles |
| GET | `/roles/:id` | Get role |

#### Restaurants

| Method | Path | Description |
|--------|------|-------------|
| GET | `/restaurants` | List restaurants |
| POST | `/restaurants` | Create restaurant |
| GET | `/restaurants/:id` | Get restaurant |
| PUT | `/restaurants/:id` | Update restaurant |
| DELETE | `/restaurants/:id` | Delete restaurant |
| GET | `/restaurants/:id/categories` | List categories |
| POST | `/restaurants/categories` | Create category |
| GET | `/restaurants/menu-items` | List menu items |
| POST | `/restaurants/menu-items` | Create menu item |

#### Orders

| Method | Path | Description |
|--------|------|-------------|
| GET | `/orders` | List orders |
| POST | `/orders` | Create order |
| GET | `/orders/:id` | Get order |
| PUT | `/orders/:id/status` | Update order status |
| POST | `/orders/:id/cancel` | Cancel order |

#### Delivery

| Method | Path | Description |
|--------|------|-------------|
| GET | `/delivery/assignments` | List assignments |
| GET | `/delivery/assignments/by` | Get assignment (by order_id) |
| POST | `/delivery/assign` | Assign driver |
| POST | `/delivery/location` | Update driver location |
| POST | `/delivery/assignments/:id/complete` | Complete delivery |
| GET | `/delivery/assignments/:id/tracking-history` | Get tracking history |
| GET | `/delivery/driver-locations` | List all driver locations |

#### Settings

| Method | Path | Description |
|--------|------|-------------|
| GET | `/settings/commission` | Get commission |
| PUT | `/settings/commission` | Update commission |
| GET | `/settings/operating-areas` | List operating areas |
| POST | `/settings/operating-areas` | Create operating area |
| PUT | `/settings/operating-areas/:id` | Update operating area |
| DELETE | `/settings/operating-areas/:id` | Delete operating area |

#### WebSocket Admin

| Method | Path | Description |
|--------|------|-------------|
| GET | `/websocket/connections` | List WebSocket connections |

### Response Format

- Success: `{ "success": true, "data": <payload> }`
- Error: `{ "error": "<message>" }`

---

## 4. Protocol Buffers (gRPC)

### Proto Files

| Proto | Path | Services |
|-------|------|----------|
| **auth** | `proto/auth/auth.proto` | AuthService |
| **common** | `proto/common/common.proto` | Shared types |
| **user** | `proto/user/user.proto` | UserService, RoleService, PermissionService |
| **restaurant** | `proto/restaurant/restaurant.proto` | RestaurantService |
| **order** | `proto/order/order.proto` | OrderService |
| **delivery** | `proto/delivery/delivery.proto` | DeliveryService |
| **settings** | `proto/settings/settings.proto` | SettingsService |
| **notification** | `proto/notification/notification.proto` | NotificationService |
| **file** | `proto/file/file.proto` | FileService |

### Auth Service

- `Login` – Admin login (username/password), returns JWT, refresh token, user info
- `SignUpWithPhone`, `SignInWithPhone` – Customer signup/signin via phone + OTP
- `SignUpWithEmail`, `SignInWithEmail` – Customer signup/signin via email + OTP
- `SignUpWithGoogle`, `SignInWithGoogle` – Customer signup/signin via Google ID token
- `SendOTP` – Send OTP to phone or email (stored in Redis)
- `ValidateToken` – Validates JWT
- `ChangePassword` – Change user password
- `ResetPassword` – Admin reset password
- `ForgotPassword` – Request password reset
- `RefreshToken` – Refresh JWT
- `Logout` – Invalidate session

### User Service

- `CreateUser`, `GetUser`, `GetUserByUsername`, `ListUsers`, `UpdateUser`, `DeleteUser`
- `CreateRole`, `GetRole`, `ListRoles`, `UpdateRole`, `DeleteRole`
- `ListPermissions`

### Restaurant Service

- Restaurant CRUD
- Category CRUD
- MenuItem CRUD

### Order Service

- `CreateOrder` – Creates order, applies commission from settings service
- `GetOrder`, `ListOrders`
- `UpdateOrderStatus`, `CancelOrder`

**Order statuses:** `pending`, `accepted`, `preparing`, `ready`, `assigned`, `picked_up`, `delivered`, `cancelled`

### Delivery Service

- `AssignDriver` – Assign driver to order
- `GetAssignment`, `ListAssignments`
- `UpdateLocation` – Update driver location (Redis + Postgres history)
- `CompleteDelivery` – Mark delivery complete
- `GetTrackingHistory` – Location history for assignment
- `ListDriverLocations` – Current locations of all drivers

**Assignment statuses:** `assigned`, `picked_up`, `delivered`

### Settings Service

- `GetCommission`, `UpdateCommission` – Platform commission (0–100%)
- `ListOperatingAreas`, `CreateOperatingArea`, `UpdateOperatingArea`, `DeleteOperatingArea` – Delivery areas (polygons)

### Common Types

- `PaginationRequest`, `PaginationMeta`
- `Address`, `AuditInfo`, `Empty`, `IDRequest`, `SuccessResponse`

---

## 5. Data Storage

### PostgreSQL (GORM)

| Table | Entity | Service |
|-------|--------|---------|
| **users** | User (username, password_hash, salt, email, phone_number, google_id, role_id, user_type, is_superadmin, active). password_hash/salt nullable for OTP/Google users | auth, user |
| **roles** | Role (name, icon, permissions JSON) | auth, user |
| **restaurants** | Restaurant (name, address, lat, lng, owner_id, status) | restaurant |
| **categories** | Category (name, restaurant_id, sort_order) | restaurant |
| **menu_items** | MenuItem (name, price, category_id, available, options JSON) | restaurant |
| **orders** | Order (customer_id, restaurant_id, status, total, delivery_address, driver_id) | order |
| **order_items** | OrderItem (order_id, menu_item_id, quantity, unit_price, options) | order |
| **delivery_assignments** | DeliveryAssignment (order_id, driver_id, status, assigned_at, picked_up_at, delivered_at) | delivery |
| **delivery_location_history** | DeliveryLocationHistory (assignment_id, driver_id, lat, lng, recorded_at) | delivery |

**Migrations:** GORM AutoMigrate in `cmd/seed/main.go` and each service's `cmd/main.go`.

### MongoDB Collections

| Collection | Purpose | Indexes |
|------------|---------|---------|
| **notifications** | User notifications | user_id, created_at, read |
| **device_tokens** | FCM device tokens | user_id, token |
| **files** | File metadata | (default) |
| **shared_drive_folders** | Shared folders | (default) |
| **settings** | Key-value (e.g. commission) | key |
| **operating_areas** | Delivery areas (polygons) | active |

### Redis Keys

| Prefix | Purpose | TTL |
|--------|---------|-----|
| `session:` | JWT sessions (auth-service) | JWT expiry |
| `otp:phone:` | OTP codes for phone (auth-service) | 5 min |
| `otp:email:` | OTP codes for email (auth-service) | 5 min |
| `driver:location:` | Driver lat/lng (delivery-service) | 5 min |
| `ratelimit:` | Rate limiting (reserved) | - |
| `cache:` | Generic cache (reserved) | - |
| `lock:` | Distributed locks (reserved) | - |

---

## 6. Authentication & Authorization

### Auth Flows

- **Admin** – Username + password via `Login`
- **Customer** – OTP-only (no password):
  - **Phone** – `SendOTP` → `SignUpWithPhone` or `SignInWithPhone`
  - **Email** – `SendOTP` → `SignUpWithEmail` or `SignInWithEmail`
  - **Google** – `SignUpWithGoogle` / `SignInWithGoogle` (Firebase ID token verification)

### JWT

- **Access token** – Short-lived (default 24h)
- **Refresh token** – Longer-lived (default 7 days)
- **Header:** `Authorization: Bearer <token>`

### Gateway Middleware

| Middleware | Purpose |
|------------|---------|
| **Logger** | Request logging |
| **CORS** | CORS headers |
| **RateLimit** | 100 req/min per IP |
| **Auth** | JWT validation, sets user_id, username, company_id, role_id, is_superadmin, permissions, token |

### gRPC Auth Interceptor

- **Path:** `pkg/middleware/auth.go`
- Validates JWT from `Authorization` header
- Stores claims in context: `user_id`, `username`, `company_id`, `role_id`, `is_superadmin`, `permissions`

### Permissions

- **Path:** `pkg/middleware/permission.go`
- **Admin roles:** `users.view`, `users.create`, `roles.*`, `restaurants.*`, `orders.*`, `delivery.view`, `delivery.manage`, `files.*`, `notifications.view`, `websocket.view_send`, `shared_drive.*`
- **Customer role:** `CustomerPermissions()` – limited to customer-facing actions (e.g. place orders, view own profile)

---

## 7. Configuration

### Config struct (`pkg/config/config.go`)

Configuration is loaded from environment variables via `config.Load()`.

### Environment Variables

| Category | Variable | Default |
|----------|----------|---------|
| **Postgres** | `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/food_delivery?sslmode=disable` |
| **MongoDB** | `MONGO_URI` | `mongodb://localhost:27017` |
| **MongoDB** | `MONGO_DB` | `food_delivery` |
| **Redis** | `REDIS_URI` | `localhost:6379` |
| **Redis** | `REDIS_PASSWORD` | `""` |
| **JWT** | `JWT_SECRET` | `default-secret-change-me` |
| **JWT** | `JWT_EXPIRY` | `24h` |
| **JWT** | `JWT_REFRESH_EXPIRY` | `168h` (7 days) |
| **Password** | `PASSWORD_PEPPER` | `default-pepper-change-me` |
| **Superadmin** | `SUPERADMIN_USERNAME` | `admin` |
| **Superadmin** | `SUPERADMIN_PASSWORD` | `Admin@123` |
| **Gateway** | `GATEWAY_HTTP_PORT` | `8080` |
| **Gateway** | `GATEWAY_BASE_URL` | `http://localhost:8080` |
| **File storage** | `UPLOAD_PATH` | `./uploads` |
| **File storage** | `MAX_UPLOAD_SIZE` | `52428800` (50MB) |
| **Firebase** | `FIREBASE_PROJECT_ID` | `food-delivery-4479f` |
| **Firebase** | `FIREBASE_CREDENTIALS_PATH` | `./food-delivery-4479f-firebase-adminsdk-fbsvc-e96b9be0ce.json` |
| **Environment** | `ENV` | `development` |

### Service Addresses

| Variable | Default |
|----------|---------|
| `AUTH_SERVICE_ADDR` | `localhost:50051` |
| `USER_SERVICE_ADDR` | `localhost:50052` |
| `NOTIFICATION_SERVICE_ADDR` | `localhost:50053` |
| `FILE_STORAGE_SERVICE_ADDR` | `localhost:50054` |
| `RESTAURANT_SERVICE_ADDR` | `localhost:50055` |
| `ORDER_SERVICE_ADDR` | `localhost:50056` |
| `DELIVERY_SERVICE_ADDR` | `localhost:50057` |
| `SETTINGS_SERVICE_ADDR` | `localhost:50058` |

---

## 8. Deployment

### Local Development

```bash
# Generate protobuf
make proto

# Build all services
make build

# Start infrastructure
make docker-up

# Seed database
make seed

# Run services (each in its own terminal or via Tilt)
cd services/gateway && go run ./cmd/main.go
cd services/auth-service && go run ./cmd/main.go
# ... etc
```

### Tilt (Kubernetes)

```bash
make tilt-up
```

- Tilt runs all services in Kubernetes
- `Tiltfile` defines services and dependencies
- Gateway depends on settings-service

### Helm

```bash
# Install
make helm-install

# Install with dev values
make helm-install-dev

# Install with prod values
make helm-install-prod

# Uninstall
make helm-uninstall
```

### Key Files

| Component | Path |
|-----------|------|
| Config | `pkg/config/config.go` |
| Postgres | `pkg/database/postgres.go` |
| MongoDB | `pkg/database/mongodb.go` |
| Redis | `pkg/database/redis.go` |
| Gateway main | `services/gateway/cmd/main.go` |
| gRPC clients | `services/gateway/internal/grpc/clients.go` |
| Handlers | `services/gateway/internal/handlers/*.go` |
| Seed/migrations | `cmd/seed/main.go` |
| Makefile | `Makefile` |
| Docker Compose | `docker-compose.yml` |
| Helm values | `helm/food-delivery-system/values.yaml` |

---

## Business Logic

### Commission

- Stored in MongoDB `settings` collection with key `app_commission`
- Order service calls `GetCommission` when creating orders
- `total = subtotal * (1 + commissionPercent/100)`
- Falls back to 0% if settings service is unavailable

### Tracking History

- `UpdateLocation` appends to `delivery_location_history` for active assignments
- Current driver location stored in Redis `driver:location:{driver_id}` (TTL 5 min)
- `GetTrackingHistory` returns points from Postgres
- `ListDriverLocations` returns current locations from Redis

### Operating Areas

- Stored in MongoDB `operating_areas` collection
- Each area has `name`, `coordinates` (polygon), `active`
- Used for defining delivery service areas on the map

### OTP Delivery

- OTP codes stored in Redis (`otp:phone:`, `otp:email:`), TTL 5 min
- **Development:** OTP logged to auth-service console
- **Production:** Configure Twilio (SMS) and SendGrid/SMTP (email) for delivery
