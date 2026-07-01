# Development Guidelines

## Pagination

All paginated list endpoints follow the pattern established in `GetListItem` / [src/repository/item.go](src/repository/item.go) `GetList` (and [src/repository/user.go](src/repository/user.go) `GetListWithRole` for the joined-relation variant).

**Entity helpers** live in [src/entity/rest.go](src/entity/rest.go) and are reusable by every repo:
- `ApplyPagination(tx, req, searchColumn)` — applies WHERE (ILIKE), LIMIT, OFFSET
- `ApplySearchOnly(tx, req, searchColumn)` — same WHERE only, no LIMIT/OFFSET (for count queries)
- `BuildPaginationResponse(req, currentElements, totalElements)` — builds `*PaginationResponse`

**Request struct** — embed `PaginationRequest` directly:
```go
type GetListXxxRequest struct {
    // ... flags ...
    PaginationRequest
}
```

Add a `const XxxSearchByYyy = "yyy"` whitelist constant. Any `search_by` value not in the whitelist → return empty result immediately, no DB call.

**Repository** — two separate queries, never share a builder instance (GORM mutates it):
1. Data query: `Joins("Assoc").Joins("Assoc.Nested")` for optional relations (LEFT JOIN via GORM association shorthand — **not** `Preload`, which issues N+1 queries) + `ApplyPagination`
2. Count query: no Joins, no LIMIT/OFFSET, only `ApplySearchOnly`

Returns `([]entity.Xxx, *entity.PaginationResponse, error)`.

**`searchColumn` must be table-qualified** (e.g. `"items.name"`) when the data query has joins, to avoid ambiguous column errors.

**Usecase** — signature returns `([]XxxResponse, *entity.PaginationResponse, error)`. Extract per-row response mapping into a `toXxxListResponse` helper to keep cognitive complexity ≤ 15.

**Handler** — pass `pagination` from usecase into `SuccessResponse`. Use `Enums(...)` in swagger `@Param` for `search_by`:
```go
// @Param search_by query string false "Field to search by." Enums(name)
```

**Reference implementation:** [src/repository/item.go](src/repository/item.go) `GetList`, [src/usecase/item.go](src/usecase/item.go) `GetList`.

## Database Migrations

All SQL migration files under `docs/sql/` must use the **goose** format with `Up` and `Down` directives:

```sql
-- +goose Up
-- SQL to apply the migration

-- +goose Down
-- SQL to reverse the migration
```

Never write a migration file without these annotations — goose will not recognize or apply it.

## Entity GORM Structs

When defining entity structs that have associations (belongs-to, has-many, etc.), always declare the foreign key field **explicitly** alongside the association field. Do not rely on GORM's convention-based implicit FK inference.

```go
// Correct — FK declared explicitly
type User struct {
    RoleID int64 `gorm:"column:role_id"`
    Role   Role  `gorm:"foreignKey:RoleID"`
}

// Wrong — FK omitted, relying on GORM convention
type User struct {
    Role Role
}
```

This applies to every relation type: `belongsTo`, `hasOne`, `hasMany`, and `many2many`. Explicit FKs prevent silent mis-joins when column names diverge from GORM defaults and make the schema intent clear to readers.

### Zero-value fields must use pointers

When a column allows zero as a valid value (e.g. `default:0`, `NOT NULL DEFAULT 0`, or any nullable column), declare the field as a pointer so GORM can distinguish between "not set" and "intentionally zero".

```go
// Correct — zero is a valid price, use pointer
Price *int64 `json:"price" gorm:"not null;default:0"`

// Wrong — zero value is indistinguishable from "not provided"
Price int64  `json:"price" gorm:"not null;default:0"`
```

Apply this to every field where zero is a meaningful business value, not just an absence of data.

## Import Order

Go imports must be organized in **4 separated groups**, in this exact order:

1. Standard library (`"context"`, `"errors"`, `"strconv"`, …)
2. External libraries (`"gorm.io/gorm"`, `"github.com/gin-gonic/gin"`, …)
3. Helper/common utilities (`"github.com/yerobalg/wealthpulse-service/helper/…"`)
4. Business-related packages (`"github.com/yerobalg/wealthpulse-service/src/…"`)

`gorm.io/gorm` is an **external** library — it belongs in group 2, never in the helper group.

```go
import (
    "context"
    "errors"

    "gorm.io/gorm"

    "github.com/yerobalg/wealthpulse-service/helper/authcontext"
    errorLib "github.com/yerobalg/wealthpulse-service/helper/errors"

    "github.com/yerobalg/wealthpulse-service/src/entity"
)
```

## Handler Parameter Binding

Always use `r.BindParam(c, &req)` for path params and query params — never use `c.Param()` / `strconv.ParseInt` manually.

`BindParam` calls `ShouldBindUri` (path) then `ShouldBindWith(binding.Query)` (query) in one step. Body fields are bound separately via `r.BindBody(c, &req)`.

**Path param + body fields belong in the same request struct.** Declare the path param field with `uri:"field_name"` and `json:"-"` alongside the body fields — do not create a separate param struct.

```go
// entity — one combined struct
type UpdateItemRequest struct {
    ID    int64  `uri:"id" json:"-" validate:"required,gt=0"`
    Name  string `json:"name" validate:"required,max=255"`
    // ...
}

// handler — BindParam first, then BindBody on the same struct
var req entity.UpdateItemRequest
if err := r.BindParam(c, &req); err != nil { ... }
if err := r.BindBody(c, &req); err != nil { ... }
```

**Reference:** `UpdateItemRequest` in [src/entity/item.go](src/entity/item.go), `UpdateItem` handler in [src/handler/item.go](src/handler/item.go).

## Handlers Must Contain No Business Logic

Handlers are responsible only for:
1. Parsing the request (`BindParam`, `BindBody`).
2. Calling the usecase.
3. Returning the response via `SuccessResponse` / `ErrorResponse`.

**Never** construct response structs, combine multiple return values into a response shape, format/transform data, or apply conditional logic in handlers. Any composition of the response payload must happen in the usecase, which should return the final response struct directly.

```go
// Wrong — handler composes response struct
item, related, err := r.usecase.Item.CreateWithRelated(ctx, req)
// ...
res := entity.CreateItemResponse{Item: item, Related: related}
r.SuccessResponse(c, "...", res, nil)

// Correct — usecase returns the final response struct
res, err := r.usecase.Item.CreateWithRelated(ctx, req)
// ...
r.SuccessResponse(c, "...", res, nil)
```

If a usecase produces multiple entities, define a `XxxResponse` struct in the entity package and have the usecase return that single response type.

## FK Existence Validation on Insert (Post-check)

For insert operations, do **not** pre-fetch a record to validate whether a foreign key target exists. Instead, attempt the insert directly and handle any `IsForeignKeyViolation` returned by the DB in a dedicated `handleXxxCreateError` function.

```go
// Wrong — pre-check GetByID before insert
existing, err := repo.GetByID(ctx, id)
if errors.Is(err, gorm.ErrRecordNotFound) { ... }

// Correct — insert and handle FK violation
if err := repo.Create(ctx, &m); err != nil {
    return handleXxxCreateError(err)
}

func handleXxxCreateError(err error) error {
    switch {
    case helperDb.IsForeignKeyViolation(err, "fk_xxx_yyy"):
        return errorLib.NotFound("Yyy")
    default:
        return errorLib.InternalServerError("")
    }
}
```

No FK-bearing resource ships in this boilerplate (the sample `item` has no foreign keys), so there is no in-repo example yet — apply the pattern above when you add a resource that references another table, naming the handler `handleXxxCreateError` and matching the DB constraint name passed to `helperDb.IsForeignKeyViolation`.

## No Panic or Fatal in Business Code

Never call `panic(...)` or `log.Fatal(...)` in any function invoked from business code (entity, usecase, repository, handler, or any helper they call). Always return an `error` instead.

```go
// Wrong
func GenerateShortUID() string {
    if _, err := rand.Read(b); err != nil {
        panic("failed: " + err.Error())
    }
}

// Correct
func GenerateShortUID() (string, error) {
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
}
```

The usecase layer wraps the error as `errorLib.InternalServerError("")` and the handler returns a 500 — the server never crashes.

## Function Parameter Limit

When a function requires more than **4 parameters**, define a dedicated param struct and pass that instead of positional arguments.

```go
// Wrong — 5 positional params
func buildRows(req entity.Foo, byCode map[string]entity.Bar, statusID int64, now int64, userID int64) []entity.Row

// Correct — use a struct
type buildRowsParam struct {
    Req      entity.Foo
    ByCode   map[string]entity.Bar
    StatusID int64
    Now      int64
    UserID   int64
}

func buildRows(p buildRowsParam) []entity.Row
```

This applies to all functions — helper, usecase, repository, and handler alike.

## No Pointer Fields in Response Structs

Never use pointer fields in response structs unless explicitly asked. When a field is conditionally populated (e.g. sensitive data that may be withheld), return an empty zero value (`""`, `0`, empty struct) instead of `nil`.

```go
// Wrong — pointer + omitempty to hide conditionally-populated fields
Description *string             `json:"description,omitempty"`
Detail      *ItemDetailResponse `json:"detail,omitempty"`

// Correct — zero value when not populated
Description string             `json:"description"`
Detail      ItemDetailResponse `json:"detail"`
```

This keeps response shapes predictable for clients — every field is always present.

## No Re-fetch After Create or Update

After a create or update operation, **do not issue a follow-up GET query to build the response**. Return only the data already in memory (the struct that was just written, or the relevant IDs/fields). A post-write SELECT adds an unnecessary round-trip and may load stale data under concurrent writes.

```go
// Wrong — re-fetches the row after insert/update
if err := repo.Create(ctx, &m); err != nil { ... }
updated, err := repo.Get(ctx, entity.XxxRequest{ID: m.ID})
return entity.XxxResponse{Data: updated}, nil

// Correct — return what was already built
if err := repo.Create(ctx, &m); err != nil { ... }
return entity.XxxResponse{Data: m}, nil
```

Only add a post-write GET when explicitly asked for by product (e.g. the response must include server-computed fields not available in the written struct).

## Validation Messages

Validation message tables live in `src/entity/validation/`, one file per resource (e.g. [src/entity/validation/item.go](src/entity/validation/item.go)). They must **not** live inline inside usecase functions or as vars/funcs inside usecase files.

**In the usecase**, validation is always a single line using `validator.Bind`:

```go
if err := i.validator.Bind(req, validation.ItemCreate); err != nil {
    return zero, err
}
```

**In `src/entity/validation/`**, build tables using the single-tag helpers from [helper/validator/validator.go](helper/validator/validator.go). Each helper returns one `validator.Response`:

```go
var ItemCreate = []validator.Response{
    validator.Required("name", "Nama item"),
    validator.Max("name", "Nama item", 255),
    validator.Max("description", "Deskripsi", 1000),
    validator.GTE("price", "Harga"),
    // ...
}
```

Available single-tag helpers: `Required`, `GT`, `GTE`, `Max`, `MaxDigit`, `Min`, `Len`, `LTE`, `Numeric`, `PrintASCII`, `ContainsAny`, `Email`, `Regex`, `IsColor`, `RequiredIf`.

Use a raw `validator.Response{Tag: "...", ...}` literal only when the message is intentionally non-standard (e.g. the `required` tag on an ID field that should say "Item tidak valid" rather than the standard "wajib diisi").

Compose multi-tag shapes with `validator.Concat` or the compound builders (`RequiredID`, `RequiredString`, `RequiredDigits`, `RequiredEmail`, `RequiredUsername`). Shared sub-tables (e.g. `userCommon`, `itemCommon`) may be unexported package-level vars within the same `validation/` file.

Password validation is handled by `validation.Password(field, label string) []validator.Response` in [src/entity/validation/password.go](src/entity/validation/password.go).

**Reference implementation:** [src/entity/validation/item.go](src/entity/validation/item.go), [src/entity/validation/user.go](src/entity/validation/user.go).

## Cognitive Complexity

Warn when any usecase function exceeds a **cognitive complexity of 15**. After writing or reviewing any usecase function, calculate its cognitive complexity and surface a warning with the measured score if it is exceeded.

## No Magic Literals Without a TODO

Never emit a hardcoded business-value literal (placeholder numbers, fake balances, stub IDs, dummy timestamps) without an adjacent `// TODO:` comment that names what should replace it. If the user asks for a feature and the real value source is not yet available, **flag it explicitly in chat before writing the code** — do not silently accept a placeholder.

The rule applies to any literal that represents a business quantity the system is supposed to compute (amounts, counts, statuses, durations). It does not apply to genuine constants (HTTP status codes, validation limits, pagination defaults).

The TODO marker must be written as an **inline trailing comment** on the same line as the literal, not on a separate line above it.

```go
// Wrong — silent placeholder, no marker, no flag in chat
RemainingLoan: 10000000,

// Wrong — TODO on a separate line above the literal
// TODO: replace placeholder with real remaining-loan computation once the loan ledger is wired in.
RemainingLoan: 10000000,

// Correct — flagged in chat AND marked inline on the same line
RemainingLoan: 10000000, // TODO: replace placeholder with real remaining-loan computation once the loan ledger is wired in.
```

When reviewing changes, grep for `TODO(` and `TODO:` and surface a list before declaring the work done.

## PR Files (`docs/pr/`)

Every PR file under `docs/pr/` must include a **Test Plan** section written in **English**. Place it between the Changes section and the Checklist section.

The test plan is a markdown checklist covering all meaningful scenarios: happy path, error paths (not found, unauthorized, conflict/duplicate), edge cases (optional fields omitted), and any side-effects.

`docs/pr/` ships empty in the boilerplate — create the first PR file when you land your first feature, following the structure above.

## Commits per Implementation Step

After **each** implementation step has been executed and builds, create a commit using the [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/) format. Do **not** bundle the whole feature into one commit — commit step by step as the work progresses (specs → entity → repository → usecase → handler/routes → PR doc), matching the layered structure of the codebase.

**Format:** `<type>[optional scope]: <description>`
- `description` is a lowercase, imperative-mood summary (e.g. "add", not "added"/"adds").
- Common `type`s: `feat` (new behavior), `fix` (bug fix), `docs` (specs/PR/markdown), `refactor`, `test`, `chore`.
- Use a `!` after the type/scope (or a `BREAKING CHANGE:` footer) only for breaking API changes.

**One commit per layer/step.** A typical feature lands as a sequence such as:

```
docs: add specs for GetListItem
feat: add GetListItem request response struct
feat: add GetListItem repository
feat: add new Item usecase and GetListItem usecase
feat: add GetListItem handler and register to routes
docs: create pr for GetListItem
```

Each commit must build on its own (`go build ./...` succeeds) so the history stays bisectable. Only commit when the user has asked for the work — follow the harness rule of branching off the default branch first if needed.

## Pointer Dereferencing

Always use `types.SafelyDereference` from [helper/types/types.go](helper/types/types.go) to dereference pointer fields. Never write a manual nil check + zero-value fallback pattern.

```go
// Wrong — manual nil check
var count int64
if d.DependantCount != nil {
    count = *d.DependantCount
}

// Correct
count := types.SafelyDereference(d.DependantCount)
```

This applies wherever a `*T` field must be read as `T` with a zero fallback — entity mapping, response building, any helper.
