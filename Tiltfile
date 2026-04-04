# Tilt configuration for Food Delivery local development
# Uses local Go compilation for fast rebuilds (~1-3s per service instead of ~30-60s)

load('ext://restart_process', 'docker_build_with_restart')

allow_k8s_contexts(['docker-desktop', 'minikube', 'kind-kind'])

# ──────────────────────────────────────────────────
# Infrastructure
# ──────────────────────────────────────────────────
k8s_yaml('k8s/base/namespace.yaml')
k8s_yaml('k8s/base/secret.yaml')
k8s_yaml('k8s/base/configmap.yaml')
k8s_yaml('k8s/base/postgres.yaml')
k8s_yaml('k8s/base/mongodb.yaml')
k8s_yaml('k8s/base/redis.yaml')
k8s_yaml('k8s/base/file-storage-pvc.yaml')

k8s_resource('postgres', labels=['infrastructure'], port_forwards=['5432:5432'])
k8s_resource('mongodb', labels=['infrastructure'], port_forwards=['27017:27017'])
k8s_resource('redis', labels=['infrastructure'], port_forwards=['6379:6379'])

# ──────────────────────────────────────────────────
# Services
# ──────────────────────────────────────────────────
services = [
    {'name': 'auth-service', 'port': 50051},
    {'name': 'user-service', 'port': 50052},
    {'name': 'notification-service', 'port': 50053},
    {'name': 'file-storage-service', 'port': 50054},
    {'name': 'restaurant-service', 'port': 50055},
    {'name': 'order-service', 'port': 50056},
    {'name': 'delivery-service', 'port': 50057},
    {'name': 'settings-service', 'port': 50058},
    {'name': 'gateway', 'port': 8080},
]

common_deps = ['pkg/', 'proto/']

for svc in services:
    name = svc['name']
    port = svc['port']

    svc_deps = [
        'services/' + name + '/cmd',
        'services/' + name + '/internal',
        'services/' + name + '/go.mod',
        'services/' + name + '/go.sum',
    ]

    if name == 'gateway':
        svc_deps.append('services/gateway/docs')

    # Compile on the host — uses local Go build cache for ~1-3s incremental rebuilds
    local_resource(
        name + '-compile',
        cmd='CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o .build/' + name + ' ./services/' + name + '/cmd/main.go',
        deps=svc_deps + common_deps,
        labels=['compilers'],
        allow_parallel=True,
    )

    # Minimal Docker image: just alpine + the pre-compiled binary
    # live_update syncs the new binary and restarts the process — no image rebuild needed
    docker_build_with_restart(
        'food-delivery-' + name,
        '.build',
        dockerfile='Dockerfile.tilt',
        build_args={'SERVICE_NAME': name},
        only=[name],
        entrypoint=['/app/' + name],
        live_update=[
            sync('.build/' + name, '/app/' + name),
        ],
    )

    k8s_yaml('k8s/services/' + name + '.yaml')

    if name == 'gateway':
        k8s_resource(
            name,
            port_forwards=['8080:8080'],
            labels=['gateway'],
            resource_deps=['postgres', 'mongodb', 'redis', name + '-compile'],
        )
    else:
        k8s_resource(
            name,
            port_forwards=[str(port) + ':' + str(port)],
            labels=['services'],
            resource_deps=['postgres', 'mongodb', 'redis', name + '-compile'],
        )

# ──────────────────────────────────────────────────
# Utility commands
# ──────────────────────────────────────────────────
local_resource(
    'seed',
    cmd='DATABASE_URL="postgres://postgres:postgres@localhost:5432/food_delivery?sslmode=disable" MONGO_URI="mongodb://localhost:27017" MONGO_DB="food_delivery" PASSWORD_PEPPER="default-pepper-change-me" SUPERADMIN_USERNAME="admin" SUPERADMIN_PASSWORD="Admin@123" go run ./cmd/seed',
    labels=['commands'],
    auto_init=False,
)

local_resource(
    'generate-proto',
    cmd='make proto',
    labels=['commands'],
    auto_init=False,
)
