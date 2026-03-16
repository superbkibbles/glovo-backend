# Tilt configuration for Food Delivery local development

load('ext://restart_process', 'docker_build_with_restart')

allow_k8s_contexts(['docker-desktop', 'minikube', 'kind-kind'])

# Namespace
k8s_yaml('k8s/base/namespace.yaml')

# Infrastructure - secret and configmap first (postgres needs secret)
k8s_yaml('k8s/base/secret.yaml')
k8s_yaml('k8s/base/configmap.yaml')
k8s_yaml('k8s/base/postgres.yaml')
k8s_yaml('k8s/base/mongodb.yaml')
k8s_yaml('k8s/base/redis.yaml')
k8s_yaml('k8s/base/file-storage-pvc.yaml')

# Wait for infrastructure (port-forward for seed to connect from host)
k8s_resource('postgres', labels=['infrastructure'], port_forwards=['5432:5432'])
k8s_resource('mongodb', labels=['infrastructure'], port_forwards=['27017:27017'])
k8s_resource('redis', labels=['infrastructure'], port_forwards=['6379:6379'])

# Services
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

for svc in services:
    name = svc['name']
    port = svc['port']

    docker_build_with_restart(
        'food-delivery-' + name,
        '.',
        dockerfile='services/' + name + '/Dockerfile',
        only=[
            'go.work',
            'proto/',
            'pkg/',
            'services/',
        ],
        entrypoint=['./' + name],
        live_update=[
            sync('./pkg', '/app/pkg'),
            sync('./services/' + name, '/app/services/' + name),
        ],
    )

    k8s_yaml('k8s/services/' + name + '.yaml')

    if name == 'gateway':
        k8s_resource(
            name,
            port_forwards=['8080:8080'],
            labels=['gateway'],
            resource_deps=['postgres', 'mongodb', 'redis', 'auth-service', 'user-service', 'notification-service', 'file-storage-service', 'restaurant-service', 'order-service', 'delivery-service', 'settings-service'],
        )
    else:
        k8s_resource(
            name,
            port_forwards=[str(port) + ':' + str(port)],
            labels=['services'],
            resource_deps=['postgres', 'mongodb', 'redis'],
        )

# Commands
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
