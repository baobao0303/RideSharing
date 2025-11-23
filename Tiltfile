# Load the restart_process extension (kept, but not used in custom_build mode)
load('ext://restart_process', 'docker_build_with_restart')

# Configure default Docker registry from environment
# Set DOCKERHUB_USER to your Docker Hub username before running `tilt up`
#   export DOCKERHUB_USER=your_dockerhub_user
if 'DOCKERHUB_USER' in os.environ:
  default_registry('docker.io/' + os.environ['DOCKERHUB_USER'])

### K8s Config ###

# Uncomment to use secrets
# k8s_yaml('./infra/development/k8s/secrets.yaml')

k8s_yaml('./infra/development/k8s/app-config.yaml')
k8s_yaml('./infra/development/k8s/postgres-deployment.yaml')
k8s_yaml('./infra/development/k8s/mongo-deployment.yaml')

### End of K8s Config ###
### API Gateway ###

gateway_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/api-gateway ./services/api-gateway'
if os.name == 'nt':
  gateway_compile_cmd = './infra/development/docker/api-gateway-build.bat'

local_resource(
  'api-gateway-compile',
  gateway_compile_cmd,
  deps=['./services/api-gateway', './shared'], labels='compiles')

# Push to registry using custom_build (publish=True forces push)
custom_build(
  'ride-sharing/api-gateway',
  'docker build -t $EXPECTED_REF -f infra/development/docker/api-gateway.Dockerfile .',
  deps=[
    'infra/development/docker/api-gateway.Dockerfile',
    'build/api-gateway',
    'shared',
  ],
  publish=True,
)

k8s_yaml('./infra/development/k8s/api-gateway-deployment.yaml')
k8s_resource('api-gateway', port_forwards=8081,
             resource_deps=['api-gateway-compile'], labels='services')
### End of API Gateway ###
### Auth Service ###

auth_compile_cmd = 'cd services/auth && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../../build/auth ./cmd/api'
if os.name == 'nt':
  auth_compile_cmd = './infra/development/docker/auth-build.bat'

local_resource(
  'auth-compile',
  auth_compile_cmd,
  deps=['./services/auth', './shared'], labels='compiles')

# Push to registry
custom_build(
  'ride-sharing/auth',
  'docker build -t $EXPECTED_REF -f infra/development/docker/auth.Dockerfile .',
  deps=[
    'infra/development/docker/auth.Dockerfile',
    'build/auth',
    'shared',
    'services/auth/migrations',
  ],
  publish=True,
)

k8s_yaml('./infra/development/k8s/auth-deployment.yaml')
k8s_resource('auth', port_forwards=8080,
             resource_deps=['auth-compile'], labels='services')
# Postgres deployment is loaded but not shown in UI
# k8s_resource('postgres', labels='services')
### End of Auth Service ###
### Logger Service ###

logger_compile_cmd = 'cd services/logger-service && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../../build/logger-service ./cmd/api'
if os.name == 'nt':
  logger_compile_cmd = './infra/development/docker/logger-service-build.bat'

local_resource(
  'logger-service-compile',
  logger_compile_cmd,
  deps=['./services/logger-service'], labels='compiles')

# Push to registry
custom_build(
  'ride-sharing/logger-service',
  'docker build -t $EXPECTED_REF -f infra/development/docker/logger-service.Dockerfile .',
  deps=[
    'infra/development/docker/logger-service.Dockerfile',
    'build/logger-service',
  ],
  publish=True,
)

k8s_yaml('./infra/development/k8s/logger-service-deployment.yaml')
k8s_resource('logger-service', port_forwards=8082,
             resource_deps=['logger-service-compile'], labels='services')
# Mongo deployment is loaded but not shown in UI
# k8s_resource('mongo', labels='services')
### End of Logger Service ###
### Mail Service ###

mail_compile_cmd = 'cd services/mail-service && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../../build/mail-service ./cmd/api'
if os.name == 'nt':
  mail_compile_cmd = './infra/development/docker/mail-service-build.bat'

local_resource(
  'mail-service-compile',
  mail_compile_cmd,
  deps=['./services/mail-service'], labels='compiles')

# Push to registry
custom_build(
  'ride-sharing/mail-service',
  'docker build -t $EXPECTED_REF -f infra/development/docker/mail-service.Dockerfile .',
  deps=[
    'infra/development/docker/mail-service.Dockerfile',
    'build/mail-service',
    'services/mail-service/templates',
  ],
  publish=True,
)

k8s_yaml('./infra/development/k8s/mail-service-deployment.yaml')
k8s_resource('mail-service', port_forwards=8083,
             resource_deps=['mail-service-compile'], labels='services')
### End of Mail Service ###
### Trip Service ###

# Uncomment once we have a trip service

#trip_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/trip-service ./services/trip-service/cmd/main.go'
#if os.name == 'nt':
#  trip_compile_cmd = './infra/development/docker/trip-build.bat'

# local_resource(
#   'trip-service-compile',
#   trip_compile_cmd,
#   deps=['./services/trip-service', './shared'], labels='compiles')

# custom_build(
#   'ride-sharing/trip-service',
#   'docker build -t $EXPECTED_REF -f infra/development/docker/trip-service.Dockerfile .',
#   deps=[
#     'infra/development/docker/trip-service.Dockerfile',
#     'build/trip-service',
#     'shared',
#   ],
#   publish=True,
# )

# k8s_yaml('./infra/development/k8s/trip-service-deployment.yaml')
# k8s_resource('trip-service', resource_deps=['trip-service-compile'], labels='services')

### End of Trip Service ###
### Web Frontend ###

# Optional: push to registry if needed; for now keep as local build
docker_build(
  'ride-sharing/web',
  '.',
  dockerfile='./infra/development/docker/web.Dockerfile',
)

k8s_yaml('./infra/development/k8s/web-deployment.yaml')
k8s_resource('web-trip', port_forwards=3000, labels='frontend')

### End of Web Frontend ###
### Portal API Gateway ###

# Temporarily disabled - uncomment to enable
# docker_build(
#   'ride-sharing/portal-api-gateway',
#   '.',
#   dockerfile='./infra/development/docker/portal-api-gateway.Dockerfile',
# )
# 
# k8s_yaml('./infra/development/k8s/portal-api-gateway-deployment.yaml')
# k8s_resource('portal-api-gateway', port_forwards=5050, labels='services')

### End of Portal API Gateway ###
### API .NET Gateway ###

# DISABLED - Run locally with: cd services/api-net-gateway/RideSharing.Api && dotnet run --urls "http://localhost:8084"
# This is disabled in Tilt to avoid port conflicts when running locally
# docker_build(
#   'ride-sharing/api-net-gateway',
#   '.',
#   dockerfile='./infra/development/docker/api-net-gateway.Dockerfile',
# )
# 
# k8s_yaml('./infra/development/k8s/api-net-gateway-deployment.yaml')
# k8s_resource('api-net-gateway', port_forwards=8084, labels='services')

### End of API .NET Gateway ###
### Word Search App ###

# Temporarily disabled - uncomment to enable
# docker_build(
#   'ride-sharing/web-ng-word-search',
#   '.',
#   dockerfile='./infra/development/docker/web-ng-word-search.Dockerfile',
# )
# 
# k8s_yaml('./infra/development/k8s/web-ng-word-search-deployment.yaml')
# k8s_resource('web-ng-word-search', port_forwards=6001, labels='frontend')

### End of Word Search App ###
### Flutter Mobile App ###

# Temporarily disabled - uncomment to enable
# docker_build(
#   'ride-sharing/ecom-flutter',
#   '.',
#   dockerfile='./infra/development/docker/ecom-flutter.Dockerfile',
# )
# 
# k8s_yaml('./infra/development/k8s/ecom-flutter-deployment.yaml')
# k8s_resource('ecom-flutter', port_forwards=5000, labels='mobile')

### End of Flutter Mobile App ###
