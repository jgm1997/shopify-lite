# Shopify Lite

A production-ready multi-tenant e-commerce API built with Go, featuring JWT authentication, atomic order processing, and merchant store management.

## Features

- **Multi-tenant merchant stores** — Support for independent merchant storefronts with isolated product catalogs and orders
- **JWT authentication with role-based access** — Secure token-based auth with merchant and customer roles
- **Atomic order placement with stock management** — Transactional order creation ensuring inventory consistency
- **Order lifecycle management** — Full order status workflow (pending → confirmed → shipped → delivered/cancelled)
- **Paginated product browsing** — Efficient product listing with pagination support
- **Rate limiting on public endpoints** — Per-IP rate limiting to protect against abuse
- **Structured JSON logging** — Production-grade logging with slog for observability
- **Database migrations** — Version-controlled schema management with golang-migrate

## Tech Stack

| Layer                | Technology        | Notes                                        |
| -------------------- | ----------------- | -------------------------------------------- |
| **Runtime**          | Go 1.24           | Cross-platform, statically compiled binaries |
| **Router**           | chi v5            | Lightweight, composable HTTP router          |
| **Database**         | PostgreSQL 16     | ACID-compliant relational database           |
| **Auth**             | JWT (golang-jwt)  | Token-based authentication                   |
| **Migrations**       | golang-migrate v4 | Version control for database schemas         |
| **Query Generation** | sqlc              | Type-safe SQL with auto-generated Go code    |
| **Containerization** | Docker + Compose  | Consistent deployment across environments    |

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.25+ (for local development)
- Git

### Quick Start

```bash
# Clone the repository
git clone https://github.com/yourusername/shopify-lite.git
cd shopify-lite

# Copy and configure environment variables
cp .env.example .env
# Edit .env and fill in:
# - DATABASE_URL: PostgreSQL connection string
# - JWT_SECRET: A strong random secret (32+ chars recommended)

# Start services and build the application
docker compose up --build

# The API will be available at http://localhost:8080
```

### Local Development

```bash
# Set environment variables
export DATABASE_URL="postgres://shopify:secret@localhost:5432/shopify?sslmode=disable"
export JWT_SECRET="your-development-secret-key-here"

# Run migrations
go run main.go  # Migrations run automatically

# Server starts on :8080
```

## API Reference

### Authentication (Public)

| Method | Path                    | Auth | Description                                |
| ------ | ----------------------- | ---- | ------------------------------------------ |
| `POST` | `/api/v1/auth/register` | None | Register a new user (merchant or customer) |
| `POST` | `/api/v1/auth/login`    | None | Login and receive JWT token                |

**Register Request:**

```json
{
  "email": "user@example.com",
  "password": "secure-password-min-8-chars",
  "role": "customer"
}
```

**Login Request:**

```json
{
  "email": "user@example.com",
  "password": "secure-password-min-8-chars"
}
```

**Response (both endpoints):**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

---

### User Management (Protected)

| Method | Path         | Auth         | Description              |
| ------ | ------------ | ------------ | ------------------------ |
| `GET`  | `/api/v1/me` | Bearer Token | Get current user profile |

**Response:**

```json
{
  "id": 1,
  "email": "user@example.com",
  "role": "customer",
  "createdAt": "2025-04-28T10:00:00Z"
}
```

---

### Products (Public + Protected)

| Method   | Path                                 | Auth   | Role     | Description                    |
| -------- | ------------------------------------ | ------ | -------- | ------------------------------ |
| `GET`    | `/api/v1/products`                   | None   | —        | List all products (paginated)  |
| `GET`    | `/api/v1/products/{id}`              | None   | —        | Get product by ID              |
| `GET`    | `/api/v1/merchants/me/products`      | Bearer | merchant | List merchant's own products   |
| `GET`    | `/api/v1/merchants/me/dashboard`     | Bearer | merchant | Get merchant product dashboard |
| `POST`   | `/api/v1/merchants/me/products`      | Bearer | merchant | Create new product             |
| `PUT`    | `/api/v1/merchants/me/products/{id}` | Bearer | merchant | Update product                 |
| `DELETE` | `/api/v1/merchants/me/products/{id}` | Bearer | merchant | Delete product                 |

**Create Product Request:**

```json
{
  "name": "Product Name",
  "description": "Product description",
  "price": 99.99,
  "stock": 50
}
```

**Product Response:**

```json
{
  "id": 1,
  "merchantId": 5,
  "name": "Product Name",
  "description": "Product description",
  "price": 99.99,
  "stock": 50,
  "createdAt": "2025-04-28T10:00:00Z"
}
```

---

### Merchants (Protected)

| Method | Path                   | Auth   | Role     | Description          |
| ------ | ---------------------- | ------ | -------- | -------------------- |
| `GET`  | `/api/v1/merchants/me` | Bearer | merchant | Get merchant profile |

**Response:**

```json
{
  "id": 5,
  "userId": 1,
  "storeName": "My Store",
  "createdAt": "2025-04-28T10:00:00Z"
}
```

---

### Orders (Protected)

| Method  | Path                                      | Auth   | Role     | Description            |
| ------- | ----------------------------------------- | ------ | -------- | ---------------------- |
| `GET`   | `/api/v1/orders`                          | Bearer | customer | List customer's orders |
| `GET`   | `/api/v1/orders/{id}`                     | Bearer | customer | Get order details      |
| `POST`  | `/api/v1/orders`                          | Bearer | customer | Create new order       |
| `PATCH` | `/api/v1/merchants/me/orders/{id}/status` | Bearer | merchant | Update order status    |

**Create Order Request:**

```json
{
  "items": [
    {
      "productId": 1,
      "quantity": 2
    },
    {
      "productId": 3,
      "quantity": 1
    }
  ]
}
```

**Order Response:**

```json
{
  "id": 42,
  "customerId": 1,
  "status": "pending",
  "total": 249.97,
  "createdAt": "2025-04-28T10:00:00Z"
}
```

**Update Order Status Request:**

```json
{
  "status": "shipped"
}
```

Valid statuses: `pending`, `confirmed`, `shipped`, `delivered`, `cancelled`

---

## Architecture

### Package Structure

The application follows a clean architecture with clear separation of concerns. Each domain owns its complete business logic stack:

```plaintext
internal/
├── auth/          # Pure authentication functions (token generation, verification, password hashing)
├── users/         # User management (model, SQL store interface, PostgreSQL implementation, HTTP handlers)
├── merchants/     # Merchant store operations (model, store, implementation, handlers)
├── products/      # Product catalog (model, store, implementation, handlers)
├── orders/        # Order processing (model, store, implementation, handlers)
├── middleware/    # Cross-cutting concerns (JWT auth, rate limiting)
├── db/            # Generated sqlc code (schema, queries) — never edited manually
└── utils/         # Helper functions (JSON encoding/decoding, validation, IP extraction)
```

### Design Principles

**Domain Isolation**: Each module in `internal/` (users, merchants, products, orders) is self-contained and owns:

- **Model**: Domain entity definitions
- **Store Interface**: Abstraction for data access
- **PSQL Implementation**: Concrete PostgreSQL implementation using sqlc-generated queries
- **HTTP Handlers**: Request/response handling and routing logic

**Cross-Cutting Concerns**: Shared functionality lives in dedicated packages:

- **middleware**: Authentication (JWT validation, role-based access) and rate limiting (per-IP token bucket)
- **auth**: Pure functions for token operations and password hashing
- **utils**: JSON serialization, email/password validation, and HTTP utilities

**Database Layer**: The `db/` package contains 100% generated code from sqlc (never edited manually). This ensures strong typing and prevents SQL injection vulnerabilities. Migration files under `db/migrations/` use golang-migrate for version control.

**HTTP Server**: The router (`chi`) is configured in `main.go` with middleware pipeline:

1. JSON logging
2. Panic recovery
3. Rate limiting (public endpoints)
4. Authentication & role-based access control (protected endpoints)

### Request Flow

```plaintext
HTTP Request
    ↓
[Rate Limiter Middleware]
    ↓
[Auth Middleware] (if protected route)
    ↓
[Handler Layer] — validates input, calls business logic
    ↓
[Store Layer] — executes database queries (via sqlc)
    ↓
PostgreSQL
```

### Key Technologies

- **chi**: Router with middleware support for composable HTTP middleware chains
- **sqlc**: Generates type-safe Go code from SQL, ensuring compile-time correctness
- **golang-migrate**: Version-controlled database schema migrations
- **slog**: Structured JSON logging for production observability
- **JWT**: Stateless authentication without server-side sessions

---

## Environment Configuration

Create a `.env` file from `.env.example` and configure the following:

```bash
# Database connection URL (PostgreSQL)
DATABASE_URL=postgres://shopify:secret@localhost:5432/shopify?sslmode=disable

# JWT signing secret (use a strong random value in production)
JWT_SECRET=your-secret-key-32-chars-minimum
```

**Connection String Format:**

```plaintext
postgres://username:password@host:port/database?sslmode=disable
```

### Production Recommendations

- Use a strong, randomly generated JWT_SECRET (at least 32 characters)
- Enable SSL mode: `?sslmode=require` in production
- Use managed PostgreSQL databases (AWS RDS, Google Cloud SQL, etc.)
- Set `POSTGRES_PASSWORD` to a strong value
- Use environment-specific `.env` files (never commit secrets)

---

## Development

### Running Tests

```bash
go test ./...
```

### Database Migrations

Migrations are automatically applied on startup. To create a new migration:

```bash
# Create migration files in db/migrations/
touch db/migrations/000005_your_migration_name.{up,down}.sql
```

### Generating Code from SQL (sqlc)

```bash
sqlc generate
```

Code is generated to `internal/db/` based on `sqlc.yaml` configuration.

---

## Rate Limiting

Public endpoints are rate-limited per IP address:

- **Rate**: 10 requests per second
- **Burst**: 20 requests

Exceeding the limit returns `429 Too Many Requests`.

Implement production-grade rate limiting with:

- Distributed rate limits (Redis-backed)
- Per-user limits (authenticated endpoints)
- TTL-based IP cleanup

---

## License

MIT

---

## Support

For issues and questions:

- Check existing GitHub issues
- Review the Architecture section above
- Examine test files for usage examples
