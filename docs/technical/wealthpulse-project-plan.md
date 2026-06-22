# WealthPulse — Backend Project Plan

## Go · Gin · GORM · SQLite · Layered Architecture · Single Container

---

## 1. Scope

This plan covers the **backend service** (`wealthpulse-service`) and its self-hosted packaging. The
React UI in [wealthpulse-prototype.jsx](wealthpulse-prototype.jsx) is the product reference — every
screen and interaction there is the source of truth for what the API must expose. The frontend is
built and **embedded into the Go binary** so the whole product ships as a single container (§16).

WealthPulse is a personal, transaction-ledger portfolio tracker. Every buy/sell is recorded; all
portfolio state (holdings, average cost, P&L, allocation) is **computed from the full transaction
history** — never stored as denormalized truth. Live prices come from third-party market-data APIs,
snapshots are taken on a schedule (an in-process loop), and Telegram alerts fire when
price/allocation thresholds are breached.

**Asset categories:** Crypto · IDX Stocks · US Stocks/ETFs · Precious Metals · IDN Bonds · Cash (IDR/USD).

---

## 2. Tech Stack (as built)

| Concern | Choice |
|---|---|
| Language | Go 1.25 |
| HTTP framework | Gin (`github.com/gin-gonic/gin`) |
| ORM | GORM (`gorm.io/gorm`) |
| Database | **SQLite**, single file on a mounted volume |
| SQLite driver | **`github.com/glebarez/sqlite`** (pure-Go, wraps `modernc.org/sqlite`) — no CGO, so `CGO_ENABLED=0` static binary |
| Decimal/money | **`github.com/shopspring/decimal`** stored as `TEXT` (never `float`) |
| Migrations | **goose** SQL files under `docs/sql/` (`-- +goose Up` / `-- +goose Down`), `sqlite3` dialect |
| Auth | JWT (`golang-jwt/jwt/v5`) + RBAC (roles/permissions) + revoked-token table |
| Validation | `go-playground/validator/v10` via `helper/validator` tables |
| Logging | logrus + OWASP audit-log vocabulary (`helper/logger`) |
| HTTP client | shared `helper/httpclient` (all outbound 3rd-party calls go through it) |
| Scheduler | **in-process loop** started by the REST app at boot (§8) |
| Frontend packaging | React `dist/` embedded via `embed.FS`, served by Gin |
| Config | environment variables (`os.Getenv`, `joho/godotenv` for local) |
| API docs | swaggo / gin-swagger (`docs/api`) |

> **Migration from the boilerplate.** The boilerplate ships Postgres (pgx, `gorm.io/driver/postgres`,
> Postgres-specific migrations, Postgres FK-violation parsing). Slice 0 converts it to SQLite — see
> §13 and the caveats in §6.4. The only net-new packages are `helper/httpclient`,
> `github.com/shopspring/decimal`, `github.com/glebarez/sqlite`, and the React embed.

---

## 3. Architecture & Layering

A **single deployable** (`main.go`): the Gin HTTP server, the embedded React SPA, and the in-process
scheduler loop all run in one process.

```
  ┌──────────────────────────── one process (main.go) ───────────────────────────┐
  │  Gin HTTP server  ──▶  handler/   parse req, call usecase, SuccessResponse     │
  │       │                usecase/   validation, portfolio computation, activity  │
  │       │                repository/ DB (GORM/SQLite)  +  3rd-party API clients  │
  │       │                entity/    GORM structs, requests, responses, validation│
  │       ├─ serves embedded React SPA at /                                        │
  │       └─ in-process scheduler loop  ──▶ usecase (price fetch / snapshot / alert)│
  └───────────────────────────────────────────────────────────────────────────────┘
                                  │
                    helper/httpclient ──▶ CoinGecko / Yahoo / MetalPrice /
                                          ExchangeRate / Telegram
                                  │
                          SQLite file  (./data/wealthpulse.db, on a volume)
```

**Hard rules carried from CLAUDE.md (apply to every resource below):**

- **All third-party API calls go through `helper/httpclient` and live in the `repository/` layer**,
  behind an interface, exactly like DB repositories. No layer other than `helper/httpclient` ever
  touches `net/http`. These repos are wired into `repository.Repository` in
  [src/repository/repository.go](../../src/repository/repository.go).
- Handlers contain **no business logic** — bind, call usecase, respond.
- Every insert/update/delete/sensitive-read usecase ends with `activityLog.Log(...)` using
  `context.WithoutCancel(ctx)`, after any transaction commits, with OWASP action constants.
- Paginated lists follow the `GetListItem` pattern (request embeds `PaginationRequest`, whitelist
  `search_by`, two-query data+count, `toXxxListResponse` helper).
- No re-fetch after create/update; no magic literals without an inline `// TODO:`; no
  panic/fatal in business code; pointer fields for zero-meaningful columns; explicit FKs;
  `types.SafelyDereference` for pointer reads; 4-group import order.
- Commit per layer (specs → entity → repository → usecase → handler/routes → PR doc).

---

## 4. Auth Model — Single Seeded Superuser

The boilerplate's full auth/RBAC stack (login, JWT, roles, permissions, revoked tokens, activity
log) is **kept as-is**. WealthPulse is single-owner, so we do **not** expose user self-registration
or user management as a product feature — instead **exactly one superuser is seeded from
environment variables** via a goose migration (§6.3, §12).

- A goose seed migration inserts one `admin`-role user whose username, bcrypt password hash, and
  name come from the environment using goose's `ENVSUB` substitution. The boilerplate's hardcoded
  `admin` user seed is removed so this env-seeded user is the only account.
- No second user is ever created through the API for this product.
- `POST /user/login`, JWT issuance, `Authorization()` middleware, and `AuthorizePermission(...)`
  remain the gate for every portfolio endpoint.
- All portfolio rows carry `created_by` = the superuser id (standard on every entity), so the schema
  stays future-proof if multi-user is ever wanted, without per-user scoping work now.
- The legacy `manageUser` / `manageItem` permissions and the sample `item` resource are removed once
  the real resources land.

> **Decision flag:** to add true multi-tenancy later, scope transaction/alert queries by `created_by`
> and drop the single-seed restriction. The schema already supports it.

---

## 5. Domain Model

| Entity | Kind | Ownership | Notes |
|---|---|---|---|
| `AssetType` | reference | global | crypto, idx, us, precious_metal, bonds, cash + `target_allocation` |
| `Asset` | reference | global | a tradable instrument (BTC, BBCA, AAPL, ORI024, IDR…) belongs to an `AssetType` |
| `AssetPrice` | time-series | global | latest + historical prices per asset (USD/IDR), written by the scheduler |
| `Transaction` | ledger | owner | buy/sell, qty, price, bonds coupon rate, notes — **source of truth** |
| `Alert` | config | owner | upper / lower / pct threshold + active flag + last-triggered |
| `AssetValue` | snapshot | owner | total value per asset type over time, for portfolio growth charts |
| **Portfolio** | **computed** | owner | **not a table** — derived in a usecase from transactions + latest prices |

Relationships (explicit FKs per CLAUDE.md):

```
asset_types 1───∞ assets 1───∞ asset_prices
                       │
                       └──∞ transactions
                       └──∞ alerts
asset_types 1───∞ asset_value
```

---

## 6. Database Schema (goose / SQLite)

New migration files under `docs/sql/`, in the SQLite dialect: `INTEGER PRIMARY KEY AUTOINCREMENT`
ids, `INTEGER` epoch `created_at/updated_at`, `DATETIME` `deleted_at` (GORM soft delete),
`created_by/updated_by/deleted_by`, named FK constraints, indexes on `deleted_at`.

### 6.1 Money, quantity & percentages — `decimal` stored as `TEXT`

Monetary amounts, prices, and quantities use **`github.com/shopspring/decimal`** in entity structs
and are stored as **`TEXT`** columns — **never `float64`** (loses precision) and **not SQLite
`NUMERIC`** (numeric affinity coerces high-precision values to 8-byte `REAL`). `TEXT` preserves the
exact decimal string; `decimal.Decimal` scans from / serializes to it directly via GORM. Arithmetic
stays exact in Go (`qty.Mul(price)`). Zero is a meaningful value for most of these, so the entity
fields are **pointers** (`*decimal.Decimal`) per CLAUDE.md.

### 6.2 `<ts>_create_portfolio_tables.sql` (sketch)

```sql
-- +goose Up
CREATE TABLE asset_types (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at        INTEGER NOT NULL DEFAULT 0,
    updated_at        INTEGER NOT NULL DEFAULT 0,
    deleted_at        DATETIME NULL,
    created_by        INTEGER NULL,
    updated_by        INTEGER NULL,
    deleted_by        INTEGER NULL,
    name              TEXT NOT NULL,
    code              TEXT NOT NULL,            -- crypto, idx, us, precious_metal, bonds, cash
    target_allocation TEXT NOT NULL DEFAULT '0' -- percent as decimal string, e.g. '20'
);
CREATE UNIQUE INDEX idx_asset_types_code ON asset_types (code);
CREATE INDEX idx_asset_types_deleted_at ON asset_types (deleted_at);

CREATE TABLE assets (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at    INTEGER NOT NULL DEFAULT 0,
    updated_at    INTEGER NOT NULL DEFAULT 0,
    deleted_at    DATETIME NULL,
    created_by    INTEGER NULL,
    updated_by    INTEGER NULL,
    deleted_by    INTEGER NULL,
    asset_type_id INTEGER NOT NULL,
    name          TEXT NOT NULL,
    ticker        TEXT NOT NULL,
    external_id   TEXT NULL,                    -- provider id (e.g. coingecko "bitcoin")
    CONSTRAINT fk_assets_asset_type FOREIGN KEY (asset_type_id) REFERENCES asset_types (id)
);
CREATE UNIQUE INDEX idx_assets_ticker ON assets (ticker);
CREATE INDEX idx_assets_asset_type_id ON assets (asset_type_id);
CREATE INDEX idx_assets_deleted_at ON assets (deleted_at);

CREATE TABLE asset_prices (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at INTEGER NOT NULL DEFAULT 0,
    updated_at INTEGER NOT NULL DEFAULT 0,
    deleted_at DATETIME NULL,
    asset_id   INTEGER NOT NULL,
    price_usd  TEXT NULL,                       -- decimal string
    price_idr  TEXT NULL,
    CONSTRAINT fk_asset_prices_asset FOREIGN KEY (asset_id) REFERENCES assets (id)
);
CREATE INDEX idx_asset_prices_asset_created ON asset_prices (asset_id, created_at DESC);

CREATE TABLE transactions (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at          INTEGER NOT NULL DEFAULT 0,
    updated_at          INTEGER NOT NULL DEFAULT 0,
    deleted_at          DATETIME NULL,
    created_by          INTEGER NULL,
    updated_by          INTEGER NULL,
    deleted_by          INTEGER NULL,
    asset_id            INTEGER NOT NULL,
    type                TEXT NOT NULL,          -- 'buy' | 'sell'
    quantity            TEXT NOT NULL,          -- decimal string
    price_usd           TEXT NULL,
    price_idr           TEXT NULL,
    annual_return_bonds TEXT NULL,              -- bonds coupon rate, percent decimal string
    transaction_date    INTEGER NOT NULL,       -- epoch; the user-entered trade date
    notes               TEXT NULL,
    CONSTRAINT fk_transactions_asset FOREIGN KEY (asset_id) REFERENCES assets (id)
);
CREATE INDEX idx_transactions_asset_id ON transactions (asset_id);
CREATE INDEX idx_transactions_date ON transactions (transaction_date DESC);
CREATE INDEX idx_transactions_deleted_at ON transactions (deleted_at);

CREATE TABLE alerts (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at        INTEGER NOT NULL DEFAULT 0,
    updated_at        INTEGER NOT NULL DEFAULT 0,
    deleted_at        DATETIME NULL,
    created_by        INTEGER NULL,
    updated_by        INTEGER NULL,
    deleted_by        INTEGER NULL,
    asset_id          INTEGER NULL,             -- NULL for portfolio-wide pct alerts
    type              TEXT NOT NULL,            -- 'upper' | 'lower' | 'pct'
    threshold         TEXT NOT NULL,            -- decimal string
    is_active         BOOLEAN NOT NULL DEFAULT 1,
    last_triggered_at INTEGER NULL,
    CONSTRAINT fk_alerts_asset FOREIGN KEY (asset_id) REFERENCES assets (id)
);
CREATE INDEX idx_alerts_asset_id ON alerts (asset_id);
CREATE INDEX idx_alerts_is_active ON alerts (is_active);
CREATE INDEX idx_alerts_deleted_at ON alerts (deleted_at);

CREATE TABLE asset_value (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at      INTEGER NOT NULL DEFAULT 0,
    updated_at      INTEGER NOT NULL DEFAULT 0,
    deleted_at      DATETIME NULL,
    asset_type_id   INTEGER NOT NULL,
    total_value_usd TEXT NULL,                  -- decimal string
    total_value_idr TEXT NULL,
    CONSTRAINT fk_asset_value_asset_type FOREIGN KEY (asset_type_id) REFERENCES asset_types (id)
);
CREATE INDEX idx_asset_value_type_created ON asset_value (asset_type_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS asset_value;
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS asset_prices;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS asset_types;
```

### 6.3 Seed migrations

- **`<ts>_seed_portfolio_data.sql`** — the six `asset_types` with default `target_allocation`
  (crypto 20%, idx 30%, us 20%, precious_metal 15%, bonds 15%, cash 0%) and the new permissions (§11).
- **`<ts>_seed_superuser.sql`** — the single owner, from env via goose `ENVSUB` (§12). Created at
  [docs/sql/20260616000003_seed_superuser.sql](../sql/20260616000003_seed_superuser.sql).

### 6.4 SQLite caveats (handle in Slice 0)

- **Enable foreign keys.** SQLite ignores FK constraints unless `PRAGMA foreign_keys = ON` is set
  per connection. With glebarez/sqlite, set it in the DSN:
  `file:./data/wealthpulse.db?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)`.
- **FK-violation post-check (CLAUDE.md).** SQLite's error does **not** include the constraint name —
  only `SQLITE_CONSTRAINT_FOREIGNKEY` ("FOREIGN KEY constraint failed"). So `IsForeignKeyViolation`
  can detect *that* an FK failed but not *which*. Rework `helper/db` to map the SQLite extended error
  code; when several FKs exist on one insert, distinguish by the single most likely target or accept
  a generic "related record not found".
- **Single writer.** SQLite serializes writes. Use WAL mode + `busy_timeout`; the in-process
  scheduler and HTTP handlers share one DB, so keep write transactions short. Connection pool max-open
  is effectively 1 for writes.
- **Booleans** are `0/1` (numeric affinity); **timestamps** in seeds use
  `CAST(strftime('%s','now') AS INTEGER)` rather than Postgres `EXTRACT(EPOCH …)`.
- **Convert the boilerplate migrations** (`…_create_auth_tables.sql`, `…_create_items_table.sql`,
  `…_seed_initial_data.sql`) from Postgres to SQLite syntax, and swap the driver + `helper/db`
  connection from pgx/postgres to glebarez/sqlite.

---

## 7. External Market-Data Integrations (repository layer)

Each provider is a repository with its own interface, wired into `repository.Repository`. They take
the shared **`helper/httpclient`** + config (API key, base URL, timeout) rather than the DB handle.
**Every outbound call goes through `helper/httpclient`** (a thin wrapper over `net/http` with
timeout, JSON encode/decode, and error mapping). Provider errors are **returned**, never panicked,
and usecases wrap them as `errorLib.InternalServerError("")`.

| Provider repo | File | Used for | Returns |
|---|---|---|---|
| `CoinGeckoInterface` | `repository/coingecko.go` | crypto prices + ticker search | price USD, provider id |
| `YahooFinanceInterface` | `repository/yahoo_finance.go` | IDX (`.JK`), US stocks/ETFs | price (native ccy) |
| `MetalPriceInterface` | `repository/metal_price.go` | precious metals (XAU, XAG) | price USD |
| `ExchangeRateInterface` | `repository/exchange_rate.go` | USD↔IDR rate | rate |
| `TelegramInterface` | `repository/telegram.go` | alert delivery (`sendMessage`) | error |

Suggested interface shape (mirrors DB repo style):

```go
type CoinGeckoInterface interface {
    GetPrices(ctx context.Context, externalIDs []string) (map[string]entity.ProviderPrice, error)
    Search(ctx context.Context, query string) ([]entity.ProviderSearchResult, error)
}
```

> Manual-priced classes (Precious Metals, e.g. Antam gold; IDN Bonds) may fall back to the last
> user-entered transaction price when no live provider is configured — flag any such fallback with an
> inline `// TODO:` if it stands in for a value the system should eventually compute.

---

## 8. Scheduler — In-Process Loop

The periodic jobs run **inside the REST process**, started by `main.go` at boot. There is no
separate scheduler binary.

- After `usecase.Init` and before `handler.Run()`, `main.go` starts the scheduler in a goroutine.
  Each job **runs once immediately on startup**, then repeats on its interval (a `time.Ticker` /
  `time.After` loop, or `robfig/cron` with an explicit initial run).
- The loop honors the same graceful-shutdown context as the HTTP server (SIGINT/SIGTERM in
  [src/handler/rest.go](../../src/handler/rest.go)) — on shutdown the ticker stops and any in-flight
  job is allowed to finish/cancel cleanly.
- **Each job is a usecase method**, not logic living in the scheduler. The loop only triggers; the
  usecase orchestrates repository calls. Jobs run with `context.Background()` (no HTTP request).

| Job | Interval (env) | Steps |
|---|---|---|
| Price fetch | `PRICE_FETCH_INTERVAL` (5 min) | for each asset → provider repo → insert `asset_prices` → recompute & store actual allocation → run alert engine |
| Value snapshot | `SNAPSHOT_INTERVAL` (30 min) | compute total value per `asset_type` from holdings × latest price → insert `asset_value` |

Alert engine (end of each price fetch): load active alerts, compare latest price / portfolio
day-change against `threshold`, and on breach call `TelegramInterface.SendMessage`, then update
`last_triggered_at` (with a re-arm/cooldown rule to avoid spamming).

> Because the loop lives in the API process, alerts only fire while the container is running. For
> 24/7 alerts, keep the container up (VPS / always-on host).

---

## 9. Portfolio Computation (usecase)

`PortfolioInterface` computes everything the dashboard needs, with **no portfolio table**, using
`decimal.Decimal` throughout:

1. Load all (non-deleted) transactions joined to asset + asset_type.
2. Aggregate per asset: `totalQty = Σbuy − Σsell`, `totalCost = Σ(buyQty × buyPrice)`, drop
   zero/negative positions.
3. `avgPrice = totalCost / totalQty`; `currentValue = totalQty × latestPrice`;
   `pnl = currentValue − totalCost`; `pnlPct = pnl / totalCost`.
4. Convert USD-priced assets to IDR via the latest exchange rate; expose both currencies so the
   IDR/USD toggle is a pure read.
5. Roll up per `asset_type` for the donut + actual-vs-target allocation; compute totals
   (value, cost, P&L, day-change, YoY progress vs `YOY_TARGET`).

Keep each function ≤ cognitive complexity 15 (extract row-mapping/rollup helpers). This is a
**sensitive read** — the summary/detail endpoints emit an activity log; the read-only **list**
endpoints do not.

---

## 10. API Endpoints

All under JWT auth (`Authorization()` group); write endpoints additionally gated by
`AuthorizePermission(...)`. Paths/verbs and the swagger annotation style follow
[src/handler/item.go](../../src/handler/item.go) and [src/handler/rest.go](../../src/handler/rest.go).

```
# Asset types (reference + target allocation)
GET    /asset-type                     # list with target & computed actual allocation
PUT    /asset-type/:id                 # update target_allocation        [managePortfolio]

# Assets (instruments + ticker autocomplete)
GET    /asset                          # paginated list (search_by=ticker,name)
GET    /asset/search                   # provider-backed ticker autocomplete (q, type)
POST   /asset                          # create instrument               [managePortfolio]

# Transactions (the ledger)
POST   /transaction                    # record buy/sell                 [manageTransaction]
GET    /transaction                    # paginated list + date range filter
GET    /transaction/:id
DELETE /transaction/:id                # soft delete                     [manageTransaction]

# Portfolio (computed)
GET    /portfolio                      # holdings + allocation + P&L (per category)
GET    /portfolio/summary              # totals: value, cost, P&L, day change, YoY
GET    /portfolio/history              # asset_value snapshots for growth chart

# Prices
GET    /price                          # latest price per asset
GET    /price/history                  # price series for sparklines (asset_id, from, to)

# Alerts
POST   /alert                          # create                          [manageAlert]
GET    /alert                          # list
PATCH  /alert/:id                      # toggle / update threshold       [manageAlert]
DELETE /alert/:id                                                        [manageAlert]

# Frontend (embedded SPA)
GET    /*                              # embedded React app (served from embed.FS)
```

Pagination params (`limit`, `page`, `search_by`, `search_value`) and the `entity.HTTPResponse`
envelope are reused unchanged. Date-range filtering on `/transaction` adds `from`/`to` (epoch) query
params bound via `BindParam`. API routes are registered before the SPA catch-all.

---

## 11. Permissions

Replace the boilerplate's `manageUser` / `manageItem` with portfolio permissions (seeded, granted to
the `admin` role; the single superuser has all of them):

```go
// src/entity/permission.go
const (
    PermissionManagePortfolio   = "managePortfolio"   // asset types, assets
    PermissionManageTransaction = "manageTransaction" // record/delete transactions
    PermissionManageAlert       = "manageAlert"       // alert CRUD
)
```

Read endpoints (portfolio, prices, lists) require only a valid token, matching how `GET /item`
needs auth but no specific permission.

---

## 12. Configuration (env)

Replace the Postgres DB_* vars with a single SQLite path, and extend with provider/scheduler config:

```env
# Database (SQLite file on the mounted volume)
DB_PATH=./data/wealthpulse.db

# Single seeded superuser (consumed by the goose ENVSUB seed migration)
SUPERUSER_USERNAME=
SUPERUSER_PASSWORD_HASH=        # bcrypt hash from `make gen-password` — see note below
SUPERUSER_NAME=

# Market-data providers
COINGECKO_API_KEY=
COINGECKO_BASE_URL=
YAHOO_FINANCE_BASE_URL=
METALPRICE_API_KEY=
EXCHANGERATE_API_KEY=

# Telegram alerts
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=

# Scheduler & targets
PRICE_FETCH_INTERVAL=5          # minutes
SNAPSHOT_INTERVAL=30            # minutes
PREFERRED_CURRENCY=IDR          # IDR | USD (display default)
YOY_TARGET=15                   # annual return target, percent
```

**Superuser seeding mechanism:** the migration
[docs/sql/20260616000003_seed_superuser.sql](../sql/20260616000003_seed_superuser.sql) uses goose's
`-- +goose ENVSUB ON` directive to substitute `${SUPERUSER_USERNAME}`, `${SUPERUSER_PASSWORD_HASH}`,
and `${SUPERUSER_NAME}` from the process environment at `goose up` time. The Makefile `migrate`
target `export`s these three variables so goose sees them.

> **bcrypt `$` caveat:** a bcrypt hash contains `$` (`$2a$08$...`). Because the Makefile loads
> `.env` via `-include`, `make` will interpret `$` as a variable reference. Write the hash in `.env`
> with **doubled dollars** (`$$2a$$08$$...`) so `make` expands it back to a single `$` before
> exporting. (Only `make`/goose consume this value; the running Go app never reads it.)

---

## 13. Build Roadmap — Vertical Slices

Build one resource end-to-end before the next; **commit per layer** (per CLAUDE.md).

### Slice 0 — Foundation (incl. Postgres → SQLite conversion)
- [ ] Swap driver: `gorm.io/driver/postgres` + pgx → `github.com/glebarez/sqlite`; update
      `helper/db` connection (DSN with `foreign_keys`, `WAL`, `busy_timeout`) and `main.go` config
      (DB_PATH instead of DB_HOST/PORT/USER/PASS/NAME)
- [ ] Convert existing boilerplate migrations to SQLite; Makefile `migrate` → `sqlite3` dialect
- [ ] Rework `helper/db.IsForeignKeyViolation` for SQLite extended error codes (§6.4)
- [ ] Add `github.com/shopspring/decimal`; `helper/httpclient` shared HTTP wrapper
- [ ] `docs/sql`: portfolio tables + asset_type/permission seed + **env superuser seed**;
      remove the boilerplate hardcoded `admin` user
- [ ] `entity/permission.go`: replace item/user perms with portfolio perms
- [ ] Remove sample `item` resource (entity/repo/usecase/handler/routes/validation/migration)

### Slice 1 — Asset types & assets (reference data)
- [ ] entity → repository → usecase → handler/routes; `GET/PUT /asset-type`, `GET/POST /asset`
- [ ] PR doc with English Test Plan

### Slice 2 — Transactions (the ledger)
- [ ] CRUD + paginated list with date-range filter; activity log on create/delete
- [ ] **Milestone:** can record a real portfolio via API

### Slice 3 — Portfolio computation
- [ ] `PortfolioInterface` usecase (holdings, allocation, P&L, summary) using `decimal`
- [ ] `GET /portfolio`, `/portfolio/summary` (prices read from `asset_prices`, manual until Slice 4)

### Slice 4 — Market-data repositories + in-process scheduler
- [ ] CoinGecko / Yahoo / MetalPrice / ExchangeRate repos (interfaces + `helper/httpclient` impls)
- [ ] Price-fetch usecase + value-snapshot usecase
- [ ] Scheduler loop wired into `main.go` (runs once at boot, then on interval; graceful shutdown)
- [ ] `GET /price`, `/price/history`, `/portfolio/history`, `GET /asset/search`
- [ ] **Milestone:** live P&L, allocation, sparkline data, ticker autocomplete

### Slice 5 — Telegram alerts
- [ ] Telegram repo (`sendMessage` via `helper/httpclient`) + alert CRUD + alert engine in the fetch job
- [ ] `POST/GET/PATCH/DELETE /alert`; re-arm/cooldown logic; activity log on alert writes
- [ ] **Milestone:** threshold breach → Telegram message

### Slice 6 — Single-container packaging & harden
- [ ] Embed React `dist/` via `embed.FS`; Gin serves SPA at `/` (API routes first, SPA catch-all last)
- [ ] Multi-stage Dockerfile (node build → `CGO_ENABLED=0` go build → distroless) + docker-compose
      with a volume for `./data`
- [ ] Error paths, validation coverage, provider-failure handling (no crashes)
- [ ] Swagger complete; `/ping` health check confirmed; README for self-host run

---

## 14. Acceptance Checklist (per CLAUDE.md)

- [ ] Every write/sensitive-read usecase ends with `activityLog.Log(...)` using OWASP actions and
      `context.WithoutCancel(ctx)`, after tx commit.
- [ ] All third-party API calls go through `helper/httpclient`, live in `repository/`, behind
      interfaces, wired in `repository.go`.
- [ ] Money/quantity use `decimal.Decimal` stored as `TEXT`, never float; zero-meaningful fields are
      pointers; FKs explicit; `PRAGMA foreign_keys` ON.
- [ ] Lists follow the pagination pattern with whitelisted `search_by`.
- [ ] Handlers contain no business logic; usecases return final response structs; no re-fetch
      after write.
- [ ] Validation lives in `src/entity/validation/`; no inline message tables.
- [ ] No magic business literals without an inline `// TODO:`; no panic/fatal in business code.
- [ ] Usecase functions ≤ cognitive complexity 15.
- [ ] One commit per layer; each builds standalone; PR docs include an English Test Plan.

---

## 15. Open Questions / Future

- **Multi-user:** schema already supports it (`created_by` on every row). Flip on by scoping
  transaction/alert queries per user and removing the single-seed restriction.
- **Provider rate limits & caching:** CoinGecko (30/min), MetalpriceAPI (100/mo) — may need
  batching, caching the last good price, and backoff in the provider repos.
- **Bonds & precious-metal pricing:** no free live feed for IDN bonds / Antam precious metals;
  default to last user-entered price unless a manual price-update endpoint is added.
- **SQLite concurrency:** single-writer is fine for one owner; revisit (WAL tuning / Postgres) only
  if write contention between the scheduler loop and the API ever shows up.

---

## 16. Deployment — Single Self-Hosted Container

The whole product (React UI + Go API + in-process scheduler + SQLite) ships as **one container**.
SQLite is a file, so no database service is needed — only a volume for persistence.

**Image (multi-stage):**

```dockerfile
# 1. Build the React frontend
FROM node:20-alpine AS web
WORKDIR /web
COPY web/ .
RUN npm ci && npm run build              # -> /web/dist

# 2. Build the Go binary with the embedded dist (pure-Go SQLite, no CGO)
FROM golang:1.25-alpine AS api
WORKDIR /app
COPY . .
COPY --from=web /web/dist ./web/dist     # embedded via embed.FS
RUN CGO_ENABLED=0 go build -o /wealthpulse .

# 3. Minimal final image
FROM gcr.io/distroless/static
COPY --from=api /wealthpulse /wealthpulse
EXPOSE 8080
ENTRYPOINT ["/wealthpulse"]
```

**docker-compose (one service + a volume):**

```yaml
services:
  wealthpulse:
    build: .
    ports: ["8080:8080"]
    env_file: [.env]
    environment:
      DB_PATH: /data/wealthpulse.db
    volumes:
      - ./data:/data        # SQLite file persists here; back up by copying it
    restart: unless-stopped
```

- The pure-Go SQLite driver (`glebarez/sqlite`) means `CGO_ENABLED=0` works, keeping the static
  binary and the `distroless/static` base.
- Backup = copy `wealthpulse.db`. Migration to a new host = move the file.
- Run goose migrations on startup (small init step in `main.go`) or as a one-shot
  `make migrate` against the mounted volume before first boot.
- Because everything is one process, the scheduler loop (§8) and alerts run as long as the container
  is up — no extra services or supervisor needed.
```
