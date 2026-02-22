# XYZ Football API

REST API for managing football teams, players, match schedules, results, and reports for Perusahaan XYZ. Built with Go, GIN Framework, PostgreSQL, and GORM.

## Table of Contents

- [Key Features](#key-features)
- [Tech Stack](#tech-stack)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
  - [Option A: Docker (Recommended)](#option-a-docker-recommended)
  - [Option B: Local Development](#option-b-local-development)
- [Architecture](#architecture)
  - [Directory Structure](#directory-structure)
  - [Clean Architecture Layers](#clean-architecture-layers)
  - [Request Lifecycle](#request-lifecycle)
  - [Database Schema](#database-schema)
- [Environment Variables](#environment-variables)
- [API Endpoints](#api-endpoints)
  - [Authentication](#authentication)
  - [Teams](#teams)
  - [Players](#players)
  - [Matches](#matches)
  - [Reports](#reports)
  - [Response Format](#response-format)
- [Swagger Documentation](#swagger-documentation)
- [Postman Collection](#postman-collection)
- [Testing](#testing)
- [Deployment](#deployment)
- [Troubleshooting](#troubleshooting)

---

## Key Features

- **Team Management** -- Full CRUD for football teams with logo URL, founded year, city, and address
- **Player Management** -- CRUD for players nested under teams, with position validation and jersey number uniqueness per team
- **Match Scheduling** -- Create and manage match schedules between teams with date/time tracking
- **Match Results & Goals** -- Submit and update match results with individual goal tracking (scorer, minute, team); scores computed automatically
- **Reports** -- Match report generation with result classification (Home Win / Away Win / Draw), top scorer per match, and accumulated total wins across all matches
- **JWT Authentication** -- Access token (15 min) + Refresh token (7 days) with DB-stored rotation and secure logout
- **Admin Seeding** -- No registration endpoint; admin credentials are seeded from environment variables at startup
- **Swagger API Docs** -- Interactive API documentation at `/swagger/index.html` (disabled in production)
- **Docker Ready** -- Multi-stage Dockerfile + Docker Compose for one-command startup

---

## Tech Stack

- **Language**: Go 1.25+
- **Framework**: [GIN](https://github.com/gin-gonic/gin) v1.11
- **Database**: PostgreSQL 17
- **ORM**: [GORM](https://gorm.io/) v1.31 with AutoMigrate
- **Authentication**: JWT ([golang-jwt](https://github.com/golang-jwt/jwt) v5) with access + refresh tokens
- **Config**: [Viper](https://github.com/spf13/viper) v1.21 with environment variable binding
- **UUIDs**: UUID v7 (time-ordered) via [google/uuid](https://github.com/google/uuid)
- **Docs**: [Swaggo](https://github.com/swaggo/swag) for Swagger/OpenAPI 2.0 generation
- **Testing**: [testify](https://github.com/stretchr/testify) + [mockery](https://github.com/vektra/mockery)
- **Containerization**: Docker multi-stage build + Docker Compose

---

## Prerequisites

### For Docker (Recommended)

- [Docker](https://docs.docker.com/get-docker/) 20.10+
- [Docker Compose](https://docs.docker.com/compose/install/) v2+

### For Local Development

- [Go](https://go.dev/dl/) 1.25+
- [PostgreSQL](https://www.postgresql.org/download/) 15+ (running locally or via Docker)
- [Swaggo CLI](https://github.com/swaggo/swag) (optional, for regenerating API docs)
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```
- [Mockery](https://github.com/vektra/mockery) v2 (optional, for regenerating mocks)
  ```bash
  go install github.com/vektra/mockery/v2@v2.53.5
  ```

---

## Getting Started

### Option A: Docker (Recommended)

Start both the API server and PostgreSQL with a single command:

```bash
# Clone the repository
git clone https://github.com/mhakimsaputra17/xyz-football-api.git
cd xyz-football-api

# Start all services (builds the Go app + starts PostgreSQL)
docker compose up --build -d
```

The API will be available at `http://localhost:8080`. Docker Compose handles:
- Building the Go binary in a multi-stage Docker image
- Starting PostgreSQL 17 with a named volume for data persistence
- Waiting for PostgreSQL to be healthy before starting the app
- Running GORM AutoMigrate to create/update tables
- Seeding a default admin user (username: `admin`, password: `password123`)

To stop:

```bash
docker compose down          # Stop containers (data persists in volume)
docker compose down -v       # Stop containers AND delete database volume
```

To view logs:

```bash
docker compose logs -f app   # Application logs
docker compose logs -f db    # Database logs
```

### Option B: Local Development

#### 1. Clone the Repository

```bash
git clone https://github.com/mhakimsaputra17/xyz-football-api.git
cd xyz-football-api
```

#### 2. Install Dependencies

```bash
go mod download
```

#### 3. Set Up PostgreSQL

Start PostgreSQL locally (or use Docker for just the database):

```bash
# Using Docker for PostgreSQL only
docker run --name xyz-football-db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=xyz_football \
  -p 5432:5432 \
  -d postgres:17-alpine
```

Or use an existing local PostgreSQL installation and create the database:

```sql
CREATE DATABASE xyz_football;
```

#### 4. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` with your database credentials. The defaults work with the Docker PostgreSQL command above:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=xyz_football
```

#### 5. Run the Application

```bash
go run ./cmd/api
```

The server starts at `http://localhost:8080`. On first run it will:
1. Connect to PostgreSQL
2. Run AutoMigrate (create all tables)
3. Seed the default admin (username: `admin`, password: `password123`)
4. Start listening on port 8080

#### 6. Verify It Works

```bash
# Health check
curl http://localhost:8080/health
# Expected: {"status":"ok"}

# Login to get JWT tokens
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}'
```

---

## Architecture

### Directory Structure

```
xyz-football-api/
├── cmd/
│   └── api/
│       └── main.go              # Entry point: config, DB, migration, seed, DI, server
├── internal/
│   ├── config/
│   │   └── config.go            # Viper-based config loader (env vars → struct)
│   ├── model/                   # GORM models (database entities)
│   │   ├── base.go              # UUID v7 base model with soft delete
│   │   ├── admin.go
│   │   ├── team.go
│   │   ├── player.go
│   │   ├── match.go
│   │   ├── goal.go
│   │   └── refresh_token.go
│   ├── dto/                     # Data Transfer Objects (request/response)
│   │   ├── auth_dto.go
│   │   ├── team_dto.go
│   │   ├── player_dto.go
│   │   ├── match_dto.go
│   │   ├── report_dto.go
│   │   └── pagination_dto.go
│   ├── repository/              # Data access layer (interfaces + GORM implementations)
│   │   ├── admin_repository.go
│   │   ├── team_repository.go
│   │   ├── player_repository.go
│   │   ├── match_repository.go
│   │   ├── goal_repository.go
│   │   └── refresh_token_repository.go
│   ├── service/                 # Business logic layer (interfaces + implementations)
│   │   ├── auth_service.go      + auth_service_test.go
│   │   ├── team_service.go      + team_service_test.go
│   │   ├── player_service.go    + player_service_test.go
│   │   ├── match_service.go     + match_service_test.go
│   │   └── report_service.go    + report_service_test.go
│   ├── mocks/                   # Auto-generated mocks (mockery v2)
│   ├── handler/                 # HTTP handlers (GIN handlers with Swagger annotations)
│   │   ├── helper.go            # Shared handler utilities
│   │   ├── auth_handler.go
│   │   ├── team_handler.go
│   │   ├── player_handler.go
│   │   ├── match_handler.go
│   │   └── report_handler.go
│   ├── middleware/
│   │   ├── auth.go              # JWT authentication middleware
│   │   └── cors.go              # CORS configuration
│   └── router/
│       └── router.go            # Route definitions and middleware wiring
├── pkg/                         # Shared packages (usable outside internal)
│   ├── errs/
│   │   └── errors.go            # AppError type with HTTP status codes
│   ├── jwt/
│   │   └── jwt.go               # JWT service (generate/validate access + refresh tokens)
│   └── response/
│       └── response.go          # Standard envelope response helpers
├── docs/                        # Auto-generated Swagger docs
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── Dockerfile                   # 3-stage multi-stage build
├── docker-compose.yml           # App + PostgreSQL
├── .dockerignore
├── .env.example                 # Environment variable template
├── .gitignore
├── .mockery.yaml                # Mockery configuration
├── go.mod
└── go.sum
```

### Clean Architecture Layers

The application follows clean architecture with three layers:

```
HTTP Request
    │
    ▼
┌──────────┐     Parses request, validates input, calls service,
│  Handler  │     formats response. No business logic here.
└────┬─────┘
     │
     ▼
┌──────────┐     Contains all business logic and validation.
│  Service  │     Orchestrates repositories. Depends on interfaces only.
└────┬─────┘
     │
     ▼
┌──────────────┐  Data access via GORM. Pagination, sorting,
│  Repository  │  filtering. No business logic here.
└──────────────┘
     │
     ▼
  PostgreSQL
```

Each layer communicates through **interfaces**, making the service layer fully unit-testable with mocks.

### Request Lifecycle

1. HTTP request hits GIN router (`internal/router/router.go`)
2. Global middleware runs (CORS)
3. For protected routes, `AuthMiddleware` validates JWT access token
4. Handler parses request body/params, calls the appropriate service method
5. Service executes business logic, calls one or more repositories
6. Repository performs database operations via GORM
7. Response flows back up: Repository → Service → Handler → JSON response

### Database Schema

6 tables with UUID v7 primary keys and GORM soft delete:

```
admins                    refresh_tokens
├── id (uuid, PK)         ├── id (uuid, PK)
├── username (text)       ├── admin_id (uuid, FK → admins)
├── password (text)       ├── token (text, unique)
├── created_at            ├── expires_at (timestamptz)
├── updated_at            ├── created_at
└── deleted_at            └── updated_at

teams                     players
├── id (uuid, PK)         ├── id (uuid, PK)
├── name (text)           ├── team_id (uuid, FK → teams)
├── logo_url (text)       ├── name (text)
├── founded_year (int)    ├── height (int, cm)
├── address (text)        ├── weight (int, kg)
├── city (text)           ├── position (text)
├── created_at            ├── jersey_number (int)
├── updated_at            ├── created_at
└── deleted_at            ├── updated_at
                          └── deleted_at

matches                   goals
├── id (uuid, PK)         ├── id (uuid, PK)
├── home_team_id (FK)     ├── match_id (uuid, FK → matches)
├── away_team_id (FK)     ├── player_id (uuid, FK → players)
├── match_date (text)     ├── team_id (uuid, FK → teams)
├── match_time (text)     ├── minute (int)
├── home_score (int)      ├── created_at
├── away_score (int)      ├── updated_at
├── status (text)         └── deleted_at
├── created_at
├── updated_at
└── deleted_at
```

Key design decisions:
- **UUID v7** for all PKs (time-ordered, better index performance than UUID v4)
- **TEXT** columns over VARCHAR (PostgreSQL best practice -- no performance difference)
- **TIMESTAMPTZ** for all timestamps
- **Soft delete** via GORM `DeletedAt` for all entities; refresh tokens use hard delete
- **Jersey number uniqueness** per team enforced at service layer (not DB constraint) so soft-deleted players free up their numbers
- **Match scores** (`home_score`, `away_score`) computed automatically from the `goals` table

---

## Environment Variables

### Required (Production)

| Variable | Description | Example |
|---|---|---|
| `ADMIN_USERNAME` | Admin username for initial seed | `admin` |
| `ADMIN_PASSWORD` | Admin password for initial seed | `strong-password-here` |
| `JWT_SECRET` | Secret key for JWT signing (min 256 bits) | `your-super-secret-key...` |
| `DB_HOST` | PostgreSQL host | `db` (Docker) or `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL username | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | `your-db-password` |
| `DB_NAME` | PostgreSQL database name | `xyz_football` |

### Optional

| Variable | Description | Default |
|---|---|---|
| `APP_NAME` | Application name | `xyz-football-api` |
| `APP_ENV` | Environment (`development` / `production`) | `development` |
| `DB_SSLMODE` | PostgreSQL SSL mode | `disable` |
| `DB_TIMEZONE` | PostgreSQL timezone | `UTC` |
| `JWT_ACCESS_EXPIRATION_MINUTES` | Access token TTL in minutes | `15` |
| `JWT_REFRESH_EXPIRATION_DAYS` | Refresh token TTL in days | `7` |
| `SERVER_PORT` | HTTP server port | `8080` |
| `SERVER_READ_TIMEOUT_SECONDS` | HTTP read timeout | `10` |
| `SERVER_WRITE_TIMEOUT_SECONDS` | HTTP write timeout | `10` |

### Environment-Specific Behavior

| Behavior | Development | Production |
|---|---|---|
| Admin credentials | Defaults to `admin`/`password123` if unset | **Required** -- app refuses to start without them |
| Swagger UI | Enabled at `/swagger/index.html` | Disabled |
| GIN mode | Debug (verbose logging) | Release |
| GORM log level | Info (logs all SQL) | Silent |

---

## API Endpoints

Base URL: `http://localhost:8080/api/v1`

All protected endpoints require the `Authorization: Bearer <access_token>` header.

### Authentication

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `POST` | `/auth/login` | No | Login with username/password, returns access + refresh tokens |
| `POST` | `/auth/refresh` | No | Exchange refresh token for new access + refresh tokens (rotation) |
| `POST` | `/auth/logout` | Yes | Invalidate refresh token (hard delete from DB) |

### Teams

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `GET` | `/teams` | Yes | List all teams (paginated, sortable) |
| `GET` | `/teams/:id` | Yes | Get team by ID |
| `POST` | `/teams` | Yes | Create a new team |
| `PUT` | `/teams/:id` | Yes | Update a team |
| `DELETE` | `/teams/:id` | Yes | Soft delete a team |

### Players

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `GET` | `/teams/:id/players` | Yes | List players for a team (paginated, sortable) |
| `POST` | `/teams/:id/players` | Yes | Create a player under a team |
| `GET` | `/players/:id` | Yes | Get player by ID |
| `PUT` | `/players/:id` | Yes | Update a player |
| `DELETE` | `/players/:id` | Yes | Soft delete a player |

### Matches

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `GET` | `/matches` | Yes | List all matches (paginated, sortable) |
| `GET` | `/matches/:id` | Yes | Get match by ID (includes teams and goals) |
| `POST` | `/matches` | Yes | Create a match schedule |
| `PUT` | `/matches/:id` | Yes | Update match schedule |
| `DELETE` | `/matches/:id` | Yes | Soft delete a match |
| `POST` | `/matches/:id/result` | Yes | Submit match result with goals |
| `PUT` | `/matches/:id/result` | Yes | Update match result (replace goals) |

### Reports

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `GET` | `/reports/matches` | Yes | List all match reports (paginated) |
| `GET` | `/reports/matches/:id` | Yes | Detailed match report |

Report data includes:
- Match result classification: **Home Win**, **Away Win**, or **Draw**
- Top scorer for the match (player with most goals)
- Accumulated total wins for both teams across all completed matches

### Utility

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `GET` | `/health` | No | Health check (returns `{"status":"ok"}`) |
| `GET` | `/swagger/*any` | No | Swagger UI (non-production only) |

### Response Format

All endpoints return a standard envelope:

```json
{
  "status": "success",
  "message": "teams fetched successfully",
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 25,
    "total_pages": 3
  }
}
```

Error responses:

```json
{
  "status": "error",
  "message": "validation failed",
  "errors": [
    {
      "field": "name",
      "message": "name is required"
    }
  ]
}
```

The `meta` field is only present on paginated list endpoints. The `errors` field is only present on validation errors.

---

## Swagger Documentation

Interactive Swagger UI is available at:

```
http://localhost:8080/swagger/index.html
```

Swagger is **disabled in production** (`APP_ENV=production`) to prevent API spec leakage.

To regenerate Swagger docs after changing handler annotations:

```bash
swag init -g cmd/api/main.go --parseDependency --parseInternal
```

This regenerates the `docs/` directory (`docs.go`, `swagger.json`, `swagger.yaml`).

---

## Postman Collection

A ready-to-use Postman collection is available in the `docs/` folder:

```
docs/XYZ_Football_API.postman_collection.json
```

### How to Import

1. Open Postman
2. Click **Import** (top-left)
3. Drag and drop the file or browse to `docs/XYZ_Football_API.postman_collection.json`
4. The collection will appear in your sidebar

### Collection Features

- **All 22 API endpoints** organized by category (Auth, Teams, Players, Matches, Reports)
- **Collection-level Bearer auth** -- set the token once, all protected endpoints inherit it automatically
- **Public endpoints** (login, refresh) override with `noauth` so they work without a token
- **Cleanup folder** at the end for sequential Collection Runner compatibility (delete test data in order)
- Pre-configured request bodies with example data for every POST/PUT endpoint

### Quick Start with Postman

1. Import the collection
2. Send the **Login** request (`POST /api/v1/auth/login`) -- copy the `access_token` from the response
3. Go to the collection's **Authorization** tab → set the token value
4. All other requests will work automatically with the inherited token

---

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run all tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./internal/service/...

# Run a specific test function
go test -v -run TestAuthService_Login ./internal/service/...

# Run with coverage report
go test -coverprofile=coverage.out ./internal/service/...
go tool cover -html=coverage.out -o coverage.html
```

### Test Summary

- **46 test cases** across 5 test files
- **70.5% coverage** on the service layer
- All tests pass

### Test Structure

Tests are co-located with the code they test:

```
internal/service/
├── auth_service.go           # Implementation
├── auth_service_test.go      # Tests (login, refresh, logout)
├── team_service.go
├── team_service_test.go      # Tests (CRUD operations)
├── player_service.go
├── player_service_test.go    # Tests (CRUD, jersey uniqueness)
├── match_service.go
├── match_service_test.go     # Tests (CRUD, submit/update result)
├── report_service.go
└── report_service_test.go    # Tests (match reports)
```

### Mocks

Mocks are generated with [mockery](https://github.com/vektra/mockery) and configured in `.mockery.yaml`. To regenerate:

```bash
mockery
```

This generates mock implementations in `internal/mocks/` for all repository and service interfaces.

---

## Deployment

### Docker (Recommended)

The application ships with a production-ready Docker setup:

**Dockerfile** -- 3-stage multi-stage build:
1. **deps** -- Downloads Go modules (cached layer)
2. **builder** -- Compiles the Go binary with `-ldflags="-s -w"` (stripped, no debug symbols)
3. **runtime** -- Minimal `alpine:3.21` image with only the compiled binary, running as non-root user

**docker-compose.yml** -- Orchestrates:
- Go application container (256MB memory limit)
- PostgreSQL 17 container (512MB memory limit)
- Named volume for database persistence
- Health checks for both services
- Private bridge network

#### Production Deployment

```bash
# Set production environment variables
export APP_ENV=production
export ADMIN_USERNAME=your-admin-username
export ADMIN_PASSWORD=your-secure-password
export JWT_SECRET=your-production-jwt-secret-min-256-bits
export DB_PASSWORD=your-db-password

# Build and start
docker compose up --build -d

# Verify
curl http://localhost:8080/health
```

In production:
- `ADMIN_USERNAME` and `ADMIN_PASSWORD` are **required** (app refuses to start without them)
- Swagger UI is **disabled**
- GIN runs in **release mode** (no debug logging)
- GORM logger is set to **silent**

#### Docker Image Details

| Property | Value |
|---|---|
| Base image | `alpine:3.21` |
| User | Non-root (`appuser:appgroup`, UID 1001) |
| Exposed port | `8080` |
| Health check | `wget --spider http://localhost:8080/health` (every 30s) |
| Binary size | ~20-25 MB (stripped, statically linked) |

---

## Troubleshooting

### Database Connection Issues

**Error:** `failed to connect to database: failed to open database connection`

**Solution:**
1. Verify PostgreSQL is running:
   ```bash
   docker ps                              # If using Docker
   pg_isready -h localhost -p 5432        # If running locally
   ```
2. Check your database credentials in `.env` match the PostgreSQL configuration
3. Ensure the database exists:
   ```bash
   psql -U postgres -c "SELECT 1 FROM pg_database WHERE datname = 'xyz_football'"
   ```

### Docker Compose: App Exits Immediately

**Error:** `xyz-football-api` container exits with code 1

**Solution:**
1. Check application logs:
   ```bash
   docker compose logs app
   ```
2. Verify PostgreSQL is healthy:
   ```bash
   docker compose ps
   ```
3. If database is not ready, try restarting just the app:
   ```bash
   docker compose restart app
   ```

### JWT Token Expired

**Error:** `{"status":"error","message":"token has expired"}`

**Solution:**
Use the refresh endpoint to get new tokens:
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"your-refresh-token-here"}'
```

Access tokens expire every 15 minutes by default. Refresh tokens expire every 7 days.

### Swagger UI Not Loading

**Error:** Navigating to `/swagger/index.html` returns 404

**Possible causes:**
1. `APP_ENV` is set to `production` (Swagger is intentionally disabled)
2. Swagger docs were not generated. Regenerate with:
   ```bash
   swag init -g cmd/api/main.go --parseDependency --parseInternal
   ```

### Port Already in Use

**Error:** `failed to start server: listen tcp :8080: bind: address already in use`

**Solution:**
1. Find the process using port 8080:
   ```bash
   # Linux/macOS
   lsof -i :8080
   # Windows
   netstat -ano | findstr :8080
   ```
2. Kill the process or change `SERVER_PORT` in `.env`

### Admin Already Exists (Seed Skipped)

If you need to reset the admin credentials, delete the existing admin from the database:

```sql
-- Connect to the database
DELETE FROM admins;
```

Then restart the application. It will re-seed with the current `ADMIN_USERNAME`/`ADMIN_PASSWORD` values.

---

## License

MIT
