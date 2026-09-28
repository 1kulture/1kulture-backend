1Kulture Backend – Architecture & Development Guide
1. Project Overview
1Kulture is an enterprise event management SaaS backend built with Go and Gin. It provides multi‑role authentication, user profiles, KYC verification, waitlist functionality, and is designed to scale with a clean separation of concerns (Controller → Service → Repository). The system uses PostgreSQL via GORM, Redis for rate limiting (optional), JWT for authentication, and Swagger for API documentation.

2. Directory Structure
text
Copy
Download
1kulture-backend/
├── cmd/
│   └── api/
│       └── main.go               # Application entry point, wiring, server setup
├── internal/
│   ├── config/                   # Configuration loading (Viper)
│   ├── constants/                # (optional) global constants
│   ├── controllers/              # HTTP handlers (Gin)
│   ├── database/                 # Database connection & auto‑migration
│   ├── middleware/               # Auth, CORS, logging, rate limiting, etc.
│   ├── models/                   # GORM models
│   ├── repositories/             # Data access layer
│   │   └── interfaces/           # Repository interfaces
│   ├── requests/                 # Request DTOs (binding/validation)
│   ├── responses/                # Response DTOs (Swagger documented)
│   ├── routes/                   # Router setup
│   ├── services/                 # Business logic
│   │   └── interfaces/           # Service interfaces
│   └── utils/                    # Helper packages
│       ├── email/                # Email sending
│       ├── jwt/                  # JWT token management
│       ├── logger/               # Logging (Logrus)
│       ├── response/             # HTTP response helpers
│       └── validator/            # Custom validation
├── docs/                         # Swagger generated files (committed)
├── templates/email/              # Email templates (if any)
├── .env.example                  # Environment variable template
├── .gitignore
├── Makefile                      # Development automation
├── generate-swagger.sh           # Swagger generation script
└── go.mod / go.sum
3. Module & Dependencies
Module path: github.com/1kulture/1kulture-backend

Go version: 1.26.5 (as in go.mod)

Key libraries:

gin-gonic/gin – HTTP framework

gorm.io/gorm + gorm.io/driver/postgres – ORM

golang-jwt/jwt/v5 – JWT

spf13/viper – Configuration

redis/go-redis/v9 – Redis client (optional)

swaggo/swag + gin-swagger – Swagger docs

sirupsen/logrus + lumberjack – Logging with rotation

go-playground/validator/v10 – Validation

4. Configuration (.env)
All configuration is loaded via Viper from a .env file (not committed). The .env.example lists all required variables. Important sections:

App: APP_NAME, APP_VERSION, ENVIRONMENT, API_URL, WEB_URL

Server: SERVER_PORT, SERVER_HOST

Database: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE, DB_TIMEZONE

JWT: JWT_SECRET, JWT_REFRESH_SECRET, JWT_ACCESS_TOKEN_EXPIRY, JWT_REFRESH_TOKEN_EXPIRY, JWT_ISSUER, JWT_AUDIENCE

Security: BCRYPT_COST, MAX_LOGIN_ATTEMPTS, etc.

Email: SMTP settings

Redis: REDIS_HOST, REDIS_PORT, REDIS_PASSWORD, REDIS_DB (empty host means Redis is skipped)

Rate Limiting: RATE_LIMIT_REQUESTS, RATE_LIMIT_DURATION

Swagger: ENABLE_SWAGGER=true enables Swagger even in production.

5. Database & Models
Connection: initialized in internal/database/database.go

Migrations: Automatic via AutoMigrate in internal/database/migration.go. All models are listed in a slice; new models must be added there.

Base Model: BaseModel includes ID uuid.UUID, CreatedAt, UpdatedAt, DeletedAt (soft delete). Every model embeds BaseModel.

Models reside in internal/models. Each model has GORM tags and JSON tags. Use UUID primary keys (auto‑generated in BeforeCreate). Relationships are defined with GORM tags.

6. Repository Layer
Interfaces in internal/repositories/interfaces/

Implementations in internal/repositories/

Pattern:

go
Copy
Download
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    FindByEmail(ctx context.Context, email string) (*models.User, error)
    // ...
}
Implementation struct holds *gorm.DB. All methods accept context.Context as first argument and use db.WithContext(ctx). Errors are wrapped with context.

Repositories are instantiated in routes.SetupRouter and passed to services via constructors.

7. Service Layer
Interfaces in internal/services/interfaces/

Implementations in internal/services/

Services contain business logic, orchestrate repositories, handle transactions, and call external utilities (email, JWT). They return domain errors (not HTTP‑specific). The controllers map these errors to HTTP responses.

Pattern:

go
Copy
Download
type AuthService interface {
    SignUp(ctx context.Context, req *requests.SignUpRequest) (*responses.AuthResponse, error)
    // ...
}
Implementation struct holds repositories and utilities. All public methods accept context.Context and request DTOs; return response DTOs or errors.

8. Controller Layer
Controllers are in internal/controllers. They:

Receive and bind JSON using ctx.ShouldBindJSON

Validate using validator.Struct(req)

Call the appropriate service method

Map errors to HTTP responses using utils/response helpers

Include Swagger annotations for every endpoint

Crucial: For Swagger to resolve types defined in the internal/responses package, each controller file must include a blank import:

go
Copy
Download
_ "github.com/1kulture/1kulture-backend/internal/responses"
This allows annotations like // @Success 201 {object} responses.AuthResponse to be parsed correctly without importing the package for real.

Controller example:

go
Copy
Download
type AuthController struct {
    authService interfaces.AuthService
}

func NewAuthController(authService interfaces.AuthService) *AuthController {
    return &AuthController{authService: authService}
}

// SignUp godoc
// @Summary Register a new user
// @Tags auth
// @Param request body requests.SignUpRequest true "Sign up request"
// @Success 201 {object} responses.AuthResponse "Account created successfully"
// @Router /auth/signup [post]
func (c *AuthController) SignUp(ctx *gin.Context) { ... }
9. Request & Response DTOs
Requests: in internal/requests. Structs with JSON tags and validate tags. They are used for binding and validation. Add example values for Swagger using example:"..." tags.

Responses: in internal/responses. These are the structures returned to the client and are referenced in Swagger annotations. Do not confuse with internal/utils/response, which contains helper functions to actually send responses.

Every endpoint should have a corresponding request and response DTO, even if it’s just a generic Response type (defined in internal/responses/response.go).

10. Middleware
Located in internal/middleware. Common middleware:

AuthMiddleware(jwtManager) – validates JWT and sets user_id, roles, etc. in context.

RoleMiddleware(...) – checks roles (not heavily used yet).

CORS – handles origins from config.

LoggerMiddleware – logs each request with latency, status, etc.

RecoveryMiddleware – recovers from panics.

RequestIDMiddleware – attaches a request ID.

RateLimitMiddleware – uses Redis if available; otherwise skips.

Middleware are applied globally in routes.SetupRouter.

11. Routes
All routes are defined in internal/routes/router.go.

Public routes are under /api/v1/auth/* and /api/v1/waitlist.

Protected routes require AuthMiddleware and are grouped under /api/v1/users/*.

Health check: GET /health

Swagger: GET /swagger/*any (if enabled)

12. Swagger Documentation
Generation: Run ./generate-swagger.sh. It uses swag init with --parseDependency --parseInternal --parseDepth 5.

Manual fallback: If generation fails, the script creates a minimal docs/docs.go. But with correct annotations and blank imports, automatic generation works.

Commit generated docs: docs/ is committed to the repository.

Access: Swagger UI is available at /swagger/index.html when ENABLE_SWAGGER=true or in non‑production environments.

Important: Always add the blank import _ "github.com/1kulture/1kulture-backend/internal/responses" in controller files that reference responses.* in Swagger annotations. Without it, swag cannot resolve the types.

13. Error Handling & Response Format
Success responses use utils/response helpers: OK, Created, NoContent, etc. They produce a consistent envelope:

json
Copy
Download
{
  "success": true,
  "data": ...,
  "message": "...",
  "timestamp": "...",
  "request_id": "..."
}
Error responses use the same envelope with success: false, an error object containing code, message, and optional details.

HTTP status codes are specific: 400 for bad requests, 401 for auth, 403 for forbidden, 404 for not found, 409 for conflicts, 422 for validation errors, 429 for rate limits, 500 for internal errors.

Validation errors return an array of {field, message, tag, value}.

14. Logging
Uses logrus with JSON formatting in production and text formatting in development.

Logger is initialized in main.go via logger.Init(environment).

Log files are rotated (production) or printed to stdout (development).

Use logger.WithRequest(ctx) to include request ID, path, method, etc.

15. Validation
Custom validator in internal/utils/validator registers additional tags like uuid, phone, password_strength, etc.

Use validator.Struct(req) after binding. It returns a slice of ValidationError that can be passed directly to response.ValidationError.

16. Authentication & JWT
JWT Manager in internal/utils/jwt handles generation and validation.

Access tokens expire after JWT_ACCESS_TOKEN_EXPIRY; refresh tokens after JWT_REFRESH_TOKEN_EXPIRY.

Tokens include claims: user_id, email, roles, token_type, session_id.

On sign‑in or sign‑up, refresh tokens are stored in DB (refresh_tokens table) and revoked on logout/password change.

Email verification is done via a 6‑digit code sent to the user’s email.

17. Email Service
internal/utils/email/email.go provides SendEmail, SendVerificationEmail, SendPasswordResetEmail.

In development, emails are logged/saved to files instead of sending.

In production, SMTP settings from config are used.

18. Redis & Rate Limiting
Redis is optional. If REDIS_HOST is empty, the app logs a warning and sets redis = nil.

Rate limiting middleware checks if Redis client is non‑nil; otherwise it skips.

Redis is also used for other ephemeral data if needed in the future.

19. Deployment (Render)
Build command: chmod +x generate-swagger.sh && ./generate-swagger.sh && go build -o bin/1kulture-api ./cmd/api

Start command: ./bin/1kulture-api

Health check: /health

Environment variables are set in Render dashboard (production values, including ENABLE_SWAGGER=true if Swagger should be visible).

Database migrations run automatically on startup.

Redis instance is optional but recommended for rate limiting.

20. Development Scripts
make start – runs the application.

make setup – copies .env.example to .env and downloads dependencies.

make swagger – generates Swagger docs.

make build – builds binary.

make test – runs tests.

make fmt – formats code.

Use ./generate-swagger.sh to regenerate Swagger after changing annotations.

21. Code Style & Conventions
Packages: Keep layers separate. Controllers import services and requests; services import repositories and models; repositories import models.

Context: Pass context.Context as the first argument to all service and repository methods.

Errors: Return meaningful domain errors from services; map to HTTP in controllers.

Naming: Use clear, descriptive names. Interfaces are named after the entity (e.g., UserRepository) and implementations are unexported structs (userRepository).

Swagger Annotations: Always add full annotations above each handler, including tags, parameters, success/failure responses, and summary. Use responses.Xxx for response types (with blank import).

DTOs: Separate request and response structs. Do not expose model structs directly in responses.

Validation: Add validate tags to request structs; use custom validators when necessary.

Logging: Use logger package consistently. Log errors with context and request details.

This document serves as the definitive reference for maintaining and extending the 1Kulture backend. Always adhere to these patterns to ensure consistency and smooth Swagger generation.