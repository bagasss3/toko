# Toko

E-commerce platform built with Go microservices.

## Stack

- **Go** 1.26.1
- **gRPC** + Protocol Buffers
- **Echo** v4 (HTTP Gateway)
- **PostgreSQL** + GORM
- **Redis**
- **PASETO** (token authentication)

## Structure

```
toko/
├── packages/          # Shared packages
│   ├── cache/         # Redis wrapper
│   ├── config/        # Viper config
│   ├── database/      # PostgreSQL/GORM
│   └── logger/        # Logrus logger
├── services/          # Microservices
│   ├── auth-service/  # gRPC auth (login/register/token)
│   └── gateway/       # HTTP API Gateway
└── deployments/       # Deployment configs
```

## Services

| Service | Port | Protocol | Description |
|---------|------|----------|-------------|
| auth-service | 8080 | gRPC | User authentication |
| gateway | 8080 | HTTP | REST API router |

## Quick Start

### Prerequisites

- Go 1.26.1+
- Docker & Docker Compose (for PostgreSQL & Redis)
- protoc

### 1. Start PostgreSQL and Redis

```bash
docker-compose up -d
```

This starts:
- PostgreSQL on port `5432`
- Redis on port `6379`

### 2. Run Auth Service

### Run Auth Service

```bash
cd services/auth-service
cp config.yml.example config.yml
# Edit config.yml with your DB/Redis credentials
go run main.go server
```

### Run Gateway

```bash
cd services/gateway
cp config.yml.example config.yml
# Edit config.yml
go run main.go server
```

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | /api/v1/auth/register | No | Register new user |
| POST | /api/v1/auth/login | No | Login user |
| GET | /api/v1/me | Yes | Get current user |

## Regenerate Protobuf

```bash
cd services/auth-service
make proto
```

## Workspace Mode

```bash
# From project root
go run ./services/auth-service/main.go server
go run ./services/gateway/main.go server
```
