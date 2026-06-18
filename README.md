# Go API Boilerplate

A layered Go REST API boilerplate (Gin + GORM + PostgreSQL) with JWT auth,
role/permission authorization, OWASP-style activity logging, and a sample CRUD
resource (`item`) that demonstrates the full convention end to end.

## Architecture

Each resource flows through four layers, one file per resource per layer:

```
src/entity      structs, request/response shapes, GORM models (+ validation/ tables)
src/repository  data access (GORM), one interface per resource
src/usecase     business logic, validation, activity logging
src/handler     HTTP layer (Gin) — binding + calling usecase + responding
```

Cross-cutting concerns live under `helper/` (db, logger, cryptolib, validator,
async, errors, appcontext, authcontext, types, storage).

The `item` resource (entity/repository/usecase/handler + `validation/item.go`)
is a reference implementation — copy it to scaffold a new resource, then delete
it once you no longer need the example.

## Tech Stack

- **Language**: Go 1.25
- **Database**: PostgreSQL
- **ORM**: GORM
- **HTTP Framework**: Gin
- **API Documentation**: Swagger (swag + gin-swagger)
- **Authentication**: JWT (golang-jwt)
- **Migration Tool**: Goose

## Prerequisites

Make sure the following tools are installed:

- [Go](https://go.dev/dl/) >= 1.25
- [PostgreSQL](https://www.postgresql.org/download/)
- [Goose](https://github.com/pressly/goose) - database migration tool
- [Swag](https://github.com/swaggo/swag) - swagger docs generator

Install Goose and Swag via:
```bash
make install-goose
make install-swag
```

## Use This Template

When you start a new project from this boilerplate, rename the Go module path
(`github.com/yerobalg/wealthpulse-service`) to your own GitHub repository.

1. **Rename the GitHub repository** (on GitHub) and update your local remote:
   ```bash
   git remote set-url origin git@github.com:<your-org>/<your-repo>.git
   ```

2. **Update the module path everywhere.** Run the rename target — it reads the
   current path from `go.mod` and replaces it across `.go`, `.mod`, `.md`, and
   `Makefile` files:
   ```bash
   make rename-module module=github.com/<your-org>/<your-repo>
   ```

3. **Verify** the rename compiles cleanly:
   ```bash
   go build ./...
   ```

4. (Optional) Update the project title/description in [main.go](main.go) swagger
   annotations, [README.md](README.md), and the binary name in the
   [Makefile](Makefile) `build` target.

## How to Run

1. **Set up environment variables**

   Copy `.env example` to `.env` and fill in the required values:
   ```bash
   cp ".env example" .env
   ```
   The seed migration creates an initial admin account — username `admin`,
   password `admin123`. **Change this before any real use.** To set a different
   seed password, generate a bcrypt hash and replace the `password` value in
   [docs/sql/20260616000002_seed_initial_data.sql](docs/sql/20260616000002_seed_initial_data.sql):
   ```bash
   make gen-password password=YourStrongPassword cost=8
   # cost is optional; defaults to PASSWORD_SALT_ROUND (or 10).
   # Or run directly:
   go run ./scripts/genpassword -password=YourStrongPassword -cost=8
   ```
   Use the same `cost` as `PASSWORD_SALT_ROUND` in your `.env` so login works
   consistently.

2. **Create the database**

   Create a PostgreSQL database matching the `DB_DBNAME` value in your `.env` file.

3. **Run database migrations**
   ```bash
   make migrate
   ```
   This will also seed initial data (roles, permissions, and test users).

4. **Generate Swagger docs**
   ```bash
   make swagger
   ```

5. **Run the server**
   ```bash
   make run
   ```

## Using the API Documentation

1. Open the documentation page at `http://{APP_HOST}:{APP_PORT}/docs/api/index.html`
2. Change the scheme from `https` to `http` using the dropdown next to the server URL in the Swagger UI
3. Go to the dummy login page at `http://{APP_HOST}:{APP_PORT}/template/login.html` and log in to get a token
4. Copy the token and click the **Authorize** button in the Swagger UI, then paste it to access protected routes

## Code Conventions

### Import Order

Go imports must be organized in 4 separated groups, in this order:

1. Standard library (e.g. `"context"`, `"time"`)
2. External libraries (e.g. `"github.com/gin-gonic/gin"`)
3. Helper/common utilities (e.g. `"github.com/yerobalg/wealthpulse-service/helper/..."`)
4. Business-related packages (e.g. `"github.com/yerobalg/wealthpulse-service/src/entity"`)

```go
import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/yerobalg/wealthpulse-service/helper/appcontext"
	"github.com/yerobalg/wealthpulse-service/helper/authcontext"
	"github.com/yerobalg/wealthpulse-service/helper/errors"

	"github.com/yerobalg/wealthpulse-service/src/entity"
	"github.com/yerobalg/wealthpulse-service/src/repository"
)
```

### Function Parameters

If a function has more than 3 parameters, wrap them in a struct.

```go
// Bad
func InitUser(userRepo repository.UserInterface, permissionRepo repository.PermissionInterface, password cryptolib.PasswordInterface, jwt cryptolib.JWTInterface, activityLog ActivityLogInterface) UserInterface

// Good
type UserInitParam struct {
	UserRepo       repository.UserInterface
	PermissionRepo repository.PermissionInterface
	Password       cryptolib.PasswordInterface
	JWT            cryptolib.JWTInterface
	ActivityLog    ActivityLogInterface
}

func InitUser(param UserInitParam) UserInterface
```

### Single Responsibility Principle

If a function has complex logic (many loops, branching, or nested conditions), split it into smaller focused functions.

```go
// Bad - Login does too much: builds maps, loops over permissions, constructs JWT claims
func (u *user) Login(ctx context.Context, req entity.UserLoginRequest) (entity.UserLoginResponse, error) {
	// ... authentication logic ...
	userData := map[string]any{
		"id":       user.ID,
		"username": user.Username,
		// ...
	}
	userPerms := make([]map[string]any, len(permissions))
	for i, p := range permissions {
		userPerms[i] = map[string]any{"name": p.Name, "code": p.Code}
	}
	token, err := u.jwt.Encode(userData)
	// ...
}

// Good - Extract mapping logic into its own function
func (u User) ToJWTClaims(permissions []PermissionResponse) map[string]any {
	// ... mapping logic here ...
}

func (u *user) Login(ctx context.Context, req entity.UserLoginRequest) (entity.UserLoginResponse, error) {
	// ... authentication logic ...
	token, err := u.jwt.Encode(user.ToJWTClaims(permissionResponses))
	// ...
}
```
