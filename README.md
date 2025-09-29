# Hackathon Service

Core microservice to manage hackathons for the DatainCube platform:
- CRUD for Hackathon, Participant, Team, Registration
- Keycloak authentication (via keycloak-service)
- Prometheus metrics (/metrics)
- Swagger documentation (/swagger/index.html)

## Features
- Echo HTTP framework
- Generic DB connection (Postgres / MySQL / SQLite) via `pkg/utils/db.go`
- Services / Handlers separation
- Unit tests with sqlmock
- Swagger documentation with swaggo
- Prometheus metrics endpoint

## Getting started

### Prerequisites
- Go 1.20+ installed
- PostgreSQL (or MySQL / SQLite)
- Keycloak (or a keycloak-service)
- `swag` for swagger generation (optional for development): `go install github.com/swaggo/swag/cmd/swag@latest`

# Run locally

1. Clone the repository:
   ```
   git clone https://github.com/Incube/hackathon-service.git
   cd hackathon-service
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Set up environment variables:
   Create a `.env` file in the root directory with the following content:
   ```
   PORT=8081
   DB_DRIVER= ...
   DB_DSN=host=... user=... password=... dbname=hackathondb port=5432 sslmode=disable
   KEYCLOAK_VERIFY_URL=...
   KEYCLOAK_REALM=...
   KEYCLOAK_CLIENT_ID=...
   KEYCLOAK_CLIENT_SECRET=...
   ```

4. Run the application:
   ```
   go run cmd/main.go
   ```

5. Access the application:
   - Swagger UI: http://localhost:8081/swagger/index.html
   - Metrics: http://localhost:8081/metrics
   - Health Check: http://localhost:8081/health
   - Readiness Check: http://localhost:8081/ready

# Tests
To run the tests, use the following command:
```
go test ./... -v
```
# CI
To run the CI pipeline, use the following command:
```
make ci
```
# Project structure
```
.
├── cmd
│   └── main.go
├── docs
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── api
│   ├── handler
│   │   ├── hackathon.go
│   │   ├── participant.go
│   │   ├── registration.go
│   │   └── team.go
│   ├── middleware
│   │   ├── auth.go
│   │   └── metrics.go
│   ├── routes
│   │   ├── router.go
│   ├── service
│   │   ├── hackathon.go
│   │   ├── participant.go
│   │   ├── registration.go
│   │   └── team.go
├── internal
│   ├── models
│   │   ├── hackathon.go
│   │   ├── participant.go
│   │   ├── registration.go
│   │   └── team.go
│   ├── metrics
│   │   └── metrics.go
├── pkg
│   └── utils
│       ├── db.go
│       └── utils.go
│   └── errors
│       └── errors.go
├── Makefile
├── README.md
├── go.mod
├── go.sum
```
