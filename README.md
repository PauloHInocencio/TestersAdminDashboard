# ThePrice Testers Admin Dashboard

A secure Go backend API for managing beta tester signups with passwordless magic link authentication. This system enables testers to sign up for Android/iOS beta programs and provides administrators with a secure, email-based authentication flow to manage these signups.

## Features

- **Passwordless Authentication** - Magic link email authentication for admins (no passwords to manage)
- **Beta Tester Management** - Handle Android and iOS beta tester signups
- **Session-Based Security** - Secure, time-limited sessions with HttpOnly cookies
- **Email Integration** - Automated magic link and beta invite emails via Resend
- **Type-Safe Database** - SQLC-generated type-safe database queries
- **Containerized Development** - Full Docker Compose setup with PostgreSQL SSL/TLS
- **Request Logging** - Built-in middleware for request/response tracking

## Tech Stack

- **Backend**: Go 1.25.9
- **Database**: PostgreSQL 17 (Alpine) with SSL/TLS
- **SQL Tooling**: SQLC (type-safe queries) + Goose (migrations)
- **Email Service**: Resend
- **Development**: Docker Compose, Air (hot reload)

## Prerequisites

- Docker & Docker Compose
- Go 1.25.9+ (for local development)
- Resend API account (for sending emails)
- Make (optional, for convenience commands)

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/PauloHInocencio/TestersAdminDashboard.git
cd TestersAdminDashboard
```

### 2. Configure Environment Variables

Create a `.env` file in the root directory:

```bash
cp .env.example .env
```

Edit `.env` and configure the following required variables:

```bash
# Email Service - Get your API key from https://resend.com
RESEND_API_KEY=re_your_actual_api_key_here
FROM_EMAIL="YourApp <noreply@yourdomain.com>"

# Application URLs
WEB_BASE_URL=http://localhost:8080
FRONTEND_URL=http://localhost:8081
ALLOWED_ORIGINS=http://localhost:8081

# Beta Testing Links
ANDROID_INVITE_LINK=https://play.google.com/apps/testing/your.app.id
IOS_INVITE_LINK=https://testflight.apple.com/join/YOUR_CODE

# Database (use defaults for development)
DB_PASSWORD=your_secure_password_here
```

### 3. Generate SSL Certificates (PostgreSQL)

```bash
make ssl-certs
# OR
bash scripts/generate-ssl-certs.sh
```

### 4. Start the Application

```bash
make run
# OR
docker-compose up --build
```

The API will be available at `http://localhost:8080`

### 5. Add Your First Admin User

After the database migrations run, add your admin email to the whitelist:

```bash
docker exec -it postgres17-testers psql -U postgres -d testers_admin \
  -c "INSERT INTO admin_whitelist (email) VALUES ('your-email@example.com');"
```

You're now ready to authenticate! Request a magic link at:
```
POST http://localhost:8080/api/v1/admin/request-magic-link
```

## API Endpoints

### Public Endpoints

- `POST /api/v1/testers/signup` - Beta tester signup (Android/iOS)

### Admin Endpoints (Authentication Required)

- `POST /api/v1/admin/request-magic-link` - Request authentication magic link
- `GET /api/v1/admin/callback?token=xxx` - Magic link callback (sets session)
- `GET /api/v1/admin/testers` - List all beta testers
- `GET /api/v1/admin/testers/{id}` - Get single tester details
- `DELETE /api/v1/admin/testers/{id}` - Delete tester
- `POST /api/v1/admin/testers/{id}/invite` - Send beta invite email

## Development

### Project Structure

```
.
├── api/                    # Server setup and routing
├── db/
│   ├── migrations/        # Database migrations (Goose)
│   ├── queries/           # SQL queries (SQLC source)
│   └── database/          # Generated Go code (SQLC)
├── docs/                  # Architecture documentation
├── email/                 # Email service (Resend)
├── middleware/            # HTTP middleware (auth, logging)
├── models/                # Request/response DTOs
├── services/              # Business logic
│   ├── admin/            # Admin authentication & management
│   ├── session/          # Session validation
│   └── tester/           # Beta tester operations
├── utils/                 # Utilities (token generation, hashing)
└── main.go               # Application entry point
```

### Available Commands

```bash
# Start services
make run

# Stop services
make stop

# Restart (removes volumes - fresh database)
make restart

# Run tests
make test

# View logs
docker-compose logs -f api
docker-compose logs -f db
```

### Database Operations

```bash
# Connect to PostgreSQL
docker exec -it postgres17-testers psql -U postgres -d testers_admin

# Run migrations manually
docker exec -it testers-admin-api goose -dir db/migrations postgres \
  "postgres://postgres:yourpass@db:5432/testers_admin?sslmode=require" up

# Check migration status
docker exec -it testers-admin-api goose -dir db/migrations postgres \
  "postgres://postgres:yourpass@db:5432/testers_admin?sslmode=require" status
```

### Regenerate Database Code (After Modifying SQL)

```bash
# Inside API container
docker exec -it testers-admin-api sqlc generate

# OR locally (if sqlc installed)
sqlc generate
```

## Authentication Flow

1. **Request Magic Link**: Admin submits email to `/admin/request-magic-link`
2. **Email Sent**: System validates email against whitelist and sends magic link (2-minute expiry)
3. **Click Link**: Admin clicks link in email (one-time use)
4. **Session Created**: System validates token, creates session (1-hour expiry), sets HttpOnly cookie
5. **Access Protected Routes**: Admin can now access protected endpoints with session cookie

**Security Features**:
- Tokens stored as SHA-256 hashes (never plain text)
- 256-bit cryptographic randomness
- Magic links expire after 2 minutes
- Single-use tokens (marked as used after validation)
- Sessions expire after 1 hour
- HttpOnly cookies prevent XSS attacks

## Environment Variables Reference

| Variable | Description | Example |
|----------|-------------|---------|
| `PORT` | API server port | `8080` |
| `DB_NAME` | Database name | `testers_admin` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `your_secure_password` |
| `DB_HOST` | Database host | `postgres17-testers` |
| `DB_SSLMODE` | SSL mode for PostgreSQL | `require` |
| `RESEND_API_KEY` | Resend API key | `re_xxxxx` |
| `FROM_EMAIL` | Email sender address | `"App <noreply@domain.com>"` |
| `WEB_BASE_URL` | Backend base URL | `http://localhost:8080` |
| `FRONTEND_URL` | Frontend URL for redirects | `http://localhost:8081` |
| `ALLOWED_ORIGINS` | CORS allowed origins (comma-separated) | `http://localhost:8081` |
| `COOKIE_SECURE` | Secure cookie flag (true for HTTPS) | `false` (dev), `true` (prod) |
| `ANDROID_INVITE_LINK` | Google Play beta testing URL | Play Store testing URL |
| `IOS_INVITE_LINK` | TestFlight beta URL | TestFlight join URL |

## Production Deployment

### Pre-Deployment Checklist

- [ ] Set `COOKIE_SECURE=true` for HTTPS environments
- [ ] Update `ALLOWED_ORIGINS` with production frontend domain(s)
- [ ] Set `WEB_BASE_URL` to production API URL (https://)
- [ ] Set `FRONTEND_URL` to production frontend URL
- [ ] Rotate database password from example defaults
- [ ] Verify Resend API key and FROM_EMAIL domain
- [ ] Add production admin email(s) to whitelist
- [ ] Generate fresh SSL certificates for PostgreSQL
- [ ] Review session expiry time (default: 1 hour)
- [ ] Configure proper logging and monitoring

### Production Environment Example

```bash
# Production .env
PORT=8080
DB_SSLMODE=require
RESEND_API_KEY=re_live_xxxxx
FROM_EMAIL="ThePrice <noreply@theprice.app>"
WEB_BASE_URL=https://api.theprice.app
FRONTEND_URL=https://theprice.app
ALLOWED_ORIGINS=https://theprice.app,https://www.theprice.app
COOKIE_SECURE=true
```

## Database Schema

### Tables

- **`tester_signups`** - Beta tester registrations
  - Fields: email, name, status (pending/invited/active/declined/removed), platform (android/ios)
  - Indexed on: status

- **`admin_whitelist`** - Authorized admin emails
  - Fields: email, created_at

- **`magic_links`** - One-time authentication tokens
  - Fields: token_hash, email, expires_at, used_at
  - Indexed on: token_hash

- **`admin_sessions`** - Active admin sessions
  - Fields: token_hash, email, expires_at
  - Indexed on: token_hash

## Contributing

This project is primarily maintained for ThePrice beta testing management. For significant changes, please open an issue first to discuss what you would like to change.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Documentation

Additional documentation can be found in the `docs/` directory:

- `middleware-pattern-explained.md` - Middleware implementation guide
- `session-authentication-explained.md` - Security analysis of auth flow
- `cors-explained.md` - CORS configuration details
- `docker-compose-explained.md` - Development environment setup
- `SSL_CERTIFICATES_EXPLAINED.md` - SSL/TLS setup guide

For Claude Code users, see `CLAUDE.md` for AI-assisted development guidance.

## License

This project is private and proprietary. All rights reserved.

## Support

For issues, questions, or contributions, please open an issue in the GitHub repository.

---

**Built with ❤️ for ThePrice beta testing program**
