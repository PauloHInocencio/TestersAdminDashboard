# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ThePrice Testers Admin Dashboard - A Go backend API for managing beta tester signups with magic link authentication for administrators. The system allows testers to sign up for Android/iOS beta testing and provides admins with a secure authentication flow to manage these signups.

**Tech Stack:**
- Go 1.25.9
- PostgreSQL 17 (Alpine) with SSL/TLS
- SQLC for type-safe SQL queries
- Goose for database migrations
- Resend for email delivery
- Docker Compose for development

## Development Commands

### Docker Development (Recommended)

```bash
# Start services (API + PostgreSQL)
make run
# OR
docker-compose up --build

# Stop services
make stop

# Restart (removes volumes - fresh database)
make restart
# OR
docker-compose down -v

# View logs
docker-compose logs -f api
docker-compose logs -f db
```

### Local Go Development

```bash
# Run tests
make test
# OR
go test -v ./...

# Build binary locally
go build -o bin/testers-admin-api

# Run locally (requires PostgreSQL)
PORT=8080 DB_HOST=localhost go run main.go
```

### Database Operations

```bash
# Connect to PostgreSQL
docker exec -it postgres17-testers psql -U postgres -d testers_admin

# Run migrations manually
docker exec -it testers-admin-api goose -dir db/migrations postgres "postgres://user:pass@db:5432/testers_admin?sslmode=require" up

# Rollback migration
docker exec -it testers-admin-api goose -dir db/migrations postgres "connection-string" down

# Check migration status
docker exec -it testers-admin-api goose -dir db/migrations postgres "connection-string" status

# Add admin user to whitelist
docker exec -it postgres17-testers psql -U postgres -d testers_admin -c "INSERT INTO admin_whitelist (email) VALUES ('your-email@example.com');"
```

**First-Time Setup**: After running migrations, you must manually add at least one admin email to the whitelist to enable authentication:

```sql
INSERT INTO admin_whitelist (email) VALUES ('your-email@example.com');
```

This ensures each deployment uses the correct admin email(s) without hardcoding them in migrations.

### SQLC Code Generation

After modifying SQL files in `db/queries/*.sql`:

```bash
# Inside API container
docker exec -it testers-admin-api sqlc generate

# OR locally (if sqlc installed)
sqlc generate
```

This regenerates `db/database/*.go` files with type-safe Go code.

### SSL Certificates

```bash
# Generate SSL certificates for PostgreSQL
make ssl-certs
# OR
bash scripts/generate-ssl-certs.sh
```

## Architecture

### Service-Oriented Architecture

The API is organized into isolated services under `services/`:

```
services/
├── admin/          # Admin authentication & tester management
│   ├── route.go    # HTTP handlers (magic link, callback, protected routes)
│   └── store.go    # Database operations (whitelist, sessions, magic links)
├── tester/         # Beta tester signup
│   ├── route.go    # Public signup endpoint
│   └── store.go    # Tester CRUD operations
└── session/        # Session validation (used by middleware)
    └── store.go    # Session lookup queries
```

**Key Pattern**: Each service has:
- `route.go` - HTTP handlers and route registration
- `store.go` - Database access layer (wraps SQLC queries)

### Magic Link Authentication Flow

**Critical Security Pattern** - No passwords, uses time-limited magic links:

1. **Request** (`POST /admin/request-magic-link`):
   - Validates email against `admin_whitelist` table
   - Generates 32-byte cryptographic token
   - Stores SHA-256 hash in `magic_links` table (2-minute expiry)
   - Sends email via Resend with callback URL

2. **Callback** (`GET /admin/callback?token=xxx`):
   - Validates token hash (one-time use, transactional with row lock)
   - Creates session in `admin_sessions` table (1-hour expiry)
   - Sets `admin_session` HttpOnly cookie (SameSite=Lax)
   - Redirects to frontend dashboard

3. **Protected Routes** (via `middleware.RequireAdmin`):
   - Extracts cookie, hashes value
   - Validates hash exists and not expired in database
   - Adds admin email to request context
   - Continues chain or returns 401

**Token Security**:
- Plain tokens never stored (only SHA-256 hashes)
- 32 bytes = 256-bit entropy from `crypto/rand`
- Magic links: 2-minute expiry, single-use
- Sessions: 1-hour expiry, validated per-request

### Middleware Pattern

Uses three-layer function closure pattern for dependency injection. See `docs/middleware-pattern-explained.md` for detailed explanation.

```go
// Layer 1: Accept dependencies (called once at setup)
func RequireAdmin(store session.SessionStore) func(http.HandlerFunc) http.HandlerFunc {
    // Layer 2: Wrap next handler
    return func(next http.HandlerFunc) http.HandlerFunc {
        // Layer 3: Actual middleware logic (runs per request)
        return func(w http.ResponseWriter, r *http.Request) {
            // Validate session...
            next(w, r)  // Continue chain
        }
    }
}

// Usage
requireAdmin := middleware.RequireAdmin(sessionStore)
router.HandleFunc("GET /admin/testers", requireAdmin(handler))
```

**Key Points**:
- Uses `http.HandlerFunc` instead of `http.Handler` for simpler code
- Context used to pass data between middleware and handlers
- Stop chain by returning early (don't call `next()`)

**Available Middleware**:
- `middleware/admin_auth.go` - Session authentication (`RequireAdmin()`)
- `middleware/logging.go` - Request/response logging (`GetLoggingMiddleware()`)
  - Logs request method, path, and remote address
  - Tracks and logs response time
  - Uses same three-layer closure pattern

### Database Layer (SQLC)

**Pattern**: SQL-first with type-safe generated Go code

```
db/
├── migrations/                    # Goose migrations
│   └── 001_create_tester_admin_tables.sql
├── queries/                       # Source SQL queries
│   ├── admin.sql                 # Admin operations
│   └── testers.sql               # Tester operations
├── database/                      # Generated by SQLC (don't edit)
│   ├── admin.sql.go
│   ├── testers.sql.go
│   ├── models.go
│   └── querier.go
└── storage.go                     # Database connection pool
```

**Workflow**:
1. Write SQL in `db/queries/*.sql` with SQLC annotations
2. Run `sqlc generate` to create Go code
3. Import generated code in service stores

**Example Query**:
```sql
-- name: GetTester :one
SELECT * FROM tester_signups WHERE id = $1;
```

Generates:
```go
func (q *Queries) GetTester(ctx context.Context, id uuid.UUID) (TesterSignup, error)
```

### Email Service

`email/email.go` - Resend integration with three templates:
- `SendMagicLink()` - Admin authentication link (2-min expiry notice)
- `SendAndroidInvite()` - Beta invite with Play Store instructions
- `SendIOSInvite()` - Beta invite with TestFlight instructions

**Configuration**: Requires `RESEND_API_KEY` and `FROM_EMAIL` environment variables.

### CORS Configuration

**Location**: `api/server.go:49-59`

CORS is configured via the `ALLOWED_ORIGINS` environment variable:
- Accepts comma-separated list of origins (e.g., `https://theprice.app,https://www.theprice.app`)
- Defaults to `http://localhost:8081` if not set or empty
- `AllowCredentials: true` enabled (required for cookie-based authentication)
- Allowed methods: GET, POST, DELETE
- See `docs/cors-explained.md` for detailed configuration guide

## Database Schema

**Tables**:
- `tester_signups` - Beta tester registrations (email, name, status, platform)
- `admin_whitelist` - Authorized admin emails
- `magic_links` - One-time authentication tokens (2-min expiry, tracks used_at)
- `admin_sessions` - Active admin sessions (1-hour expiry)

**Key Indexes**:
- `idx_tester_signups_status` - Fast filtering by status
- `idx_magic_links_token_hash` - Fast token lookup
- `idx_admin_sessions_token_hash` - Fast session validation

## Environment Variables

Required in `.env` file (see `.env.example`):

```bash
# Server
PORT=8080

# Database
DB_NAME=testers_admin
DB_USER=postgres
DB_PASSWORD=your_secure_password_here
DB_HOST=postgres17-testers
DB_SSLMODE=require

# Email Service
RESEND_API_KEY=re_your_api_key_here
FROM_EMAIL="ThePrice <noreply@theprice.app>"

# Application URLs
WEB_BASE_URL=https://theprice.app
ANDROID_INVITE_LINK=https://play.google.com/apps/testing/br.com.noartcode.theprice
IOS_INVITE_LINK=https://testflight.apple.com/join/YOUR_TESTFLIGHT_CODE

# Security (Production)
COOKIE_SECURE=true  # MUST be "true" in production with HTTPS
FRONTEND_URL=https://theprice.app
ALLOWED_ORIGINS=https://theprice.app,https://www.theprice.app  # Comma-separated list
```

**Development Example**:
```bash
WEB_BASE_URL=http://localhost:8080
FRONTEND_URL=http://localhost:8081
ALLOWED_ORIGINS=http://localhost:8081
COOKIE_SECURE=false  # Only for local development
```

## Common Patterns

### Adding a New Protected Endpoint

1. Define handler in service (e.g., `services/admin/route.go`):
```go
func (h *Handler) myHandler(w http.ResponseWriter, r *http.Request) {
    email := r.Context().Value(middleware.AdminEmailKey).(string)
    // Handler logic
}
```

2. Register with middleware in `RegisterRoutes()`:
```go
router.HandleFunc("GET /admin/myroute", requireAdmin(h.myHandler))
```

### Adding Database Queries

1. Write SQL in `db/queries/[service].sql`:
```sql
-- name: MyQuery :one
SELECT * FROM table WHERE id = $1;
```

2. Run `sqlc generate` (inside container or locally)

3. Use in store:
```go
result, err := s.queries.MyQuery(ctx, id)
```

### Database Transactions

When multiple operations must be atomic (see `admin/store.go:53-76`):

```go
func (s *Store) AtomicOperation(ctx context.Context) error {
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    qtx := s.queries.WithTx(tx)

    // Multiple operations with qtx
    err = qtx.FirstQuery(ctx, ...)
    err = qtx.SecondQuery(ctx, ...)

    return tx.Commit()
}
```

**Critical**: Use `FOR UPDATE` in SELECT queries within transactions to prevent race conditions (see `FindValidMagicLinkForUpdate`).

## Frontend Integration Notes

**Current State**: Backend API only. Frontend being built with Kobweb (Kotlin Compose for Web).

**Integration Pattern**:
- Go backend: `localhost:8080` (API + auth)
- Kobweb frontend: `localhost:8081` (UI)
- Cookie-based auth across origins
- Callback redirect: `http://localhost:8081/admin/testers-list`

**Required Changes for Integration**:
1. Update redirect URL in `services/admin/route.go:167`
2. Set `ALLOWED_ORIGINS=http://localhost:8081` in `.env`
3. Ensure CORS `AllowCredentials: true` (line 52 in `api/server.go`)

See `docs/session-authentication-explained.md` for security analysis.

## Security Best Practices

**Token Handling**:
- Always use `utils.GenerateToken()` for cryptographic randomness
- Always hash tokens with `utils.HashToken()` before database storage
- Never log or expose plain tokens

**Context Keys**:
- Use typed context keys (see `middleware/admin_auth.go:11-13`)
- Extract with type assertion: `email := r.Context().Value(AdminEmailKey).(string)`

**HTTPS Requirements**:
- `Secure: true` cookie flag requires HTTPS in production
- PostgreSQL SSL/TLS enabled by default (see `docker-compose.yml`)

**Production Checklist**:
- [ ] Set `COOKIE_SECURE=true` for HTTPS environments
- [ ] Verify `ALLOWED_ORIGINS` includes all frontend domains (comma-separated)
- [ ] Set `FRONTEND_URL` to match actual frontend domain
- [ ] Verify `WEB_BASE_URL` uses https:// protocol
- [ ] Rotate database passwords from `.env.example` defaults
- [ ] Validate Resend API key and FROM_EMAIL domain
- [ ] Review session expiry times (currently 1 hour)

## File Structure Reference

```
.
├── api/                    # Server setup and routing
├── db/
│   ├── migrations/        # Goose SQL migrations
│   │   └── 001_create_tester_admin_tables.sql
│   ├── queries/           # SQLC source queries
│   ├── database/          # SQLC generated code (don't edit)
│   └── storage.go         # Connection pool
├── docs/                  # Architecture documentation
├── email/                 # Resend integration
├── middleware/            # HTTP middleware
│   ├── admin_auth.go     # Session authentication
│   └── logging.go        # Request/response logging
├── models/                # Request/response DTOs
├── services/              # Business logic by domain
│   ├── admin/            # Admin auth & management
│   ├── session/          # Session validation
│   └── tester/           # Tester signups
├── utils/                 # Token generation, hashing
├── certs/                 # SSL certificates (PostgreSQL)
├── docker-compose.yml     # Development environment
├── Dockerfile             # Multi-stage build (dev + prod)
├── .air.toml             # Hot-reload configuration
└── main.go               # Application entry point
```

## Related Documentation

### Core Documentation
- `docs/middleware-pattern-explained.md` - Detailed middleware implementation guide
- `docs/session-authentication-explained.md` - Security analysis of auth flow
- `docs/workflow.md` - Development workflow guide
- `.env.example` - Required environment variables

### Infrastructure Documentation
- `docs/docker-compose-explained.md` - Development environment setup
- `docs/dockerfile-security-explained.md` - Docker security patterns
- `docs/SSL_CERTIFICATES_EXPLAINED.md` - SSL/TLS setup guide
- `docs/pg_hba_conf_explained.md` - PostgreSQL authentication config
- `docs/db-init-script-explained.md` - Database initialization

### Configuration Guides
- `docs/cors-explained.md` - CORS configuration details
- `db/migrations/001_create_tester_admin_tables.sql` - Complete schema