# Slice 0 — SQLite + Portfolio Foundation

## Summary

Converts the boilerplate from PostgreSQL to SQLite (single-file, single-container target),
removes the sample `item` resource, and lays the groundwork for the WealthPulse portfolio
domain: the database schema, the single env-seeded superuser, portfolio permissions, and a
shared outbound HTTP client for the upcoming third-party market-data integrations.

See [docs/technical/wealthpulse-project-plan.md](../technical/wealthpulse-project-plan.md) for the
full plan; this PR is **Slice 0 (Foundation)**.

## Changes

- **Database driver → SQLite.** `glebarez/sqlite` (pure-Go, builds with `CGO_ENABLED=0`). DSN enables
  `foreign_keys`, `WAL`, and `busy_timeout`. `db.Credential` is now a single `Path`; the app reads
  `DB_PATH`.
- **Constraint-violation detection.** `helper/db` now recognizes SQLite UNIQUE/FK errors. SQLite FK
  errors carry no constraint name, so the name argument is ignored for FK checks (documented).
- **Migrations → SQLite.** Auth tables rewritten in SQLite syntax. New portfolio tables
  (`asset_types`, `assets`, `asset_prices`, `transactions`, `alerts`, `asset_value`) with money/qty
  stored as `TEXT` (for `shopspring/decimal`). Seeds: six default asset types and portfolio
  permissions (`managePortfolio`, `manageTransaction`, `manageAlert`).
- **Single superuser at startup.** `usecase.User.EnsureSuperuser` idempotently inserts the
  `admin`-role owner from `SUPERUSER_USERNAME` / `SUPERUSER_PASSWORD_HASH` / `SUPERUSER_NAME`, storing
  the pre-bcrypt-hashed password verbatim (read via `os.Getenv`, so the hash's `$` needs no
  escaping). No SQL seed, no goose `ENVSUB`.
- **Removed sample `item` resource** across entity/repository/usecase/handler/validation, its routes,
  wiring, and the `manageItem` permission/migration.
- **`helper/httpclient`.** Single outbound HTTP gateway (timeout, JSON, status-error mapping) that
  returns the decoded body as `(map[string]any, error)`. All future provider repos call through it.
- **Dependencies.** Added `glebarez/sqlite` and `shopspring/decimal`; dropped pgx/postgres via
  `go mod tidy`.

## Test Plan

- [ ] `go build ./...` succeeds.
- [ ] `go vet ./...` is clean.
- [ ] `make migrate` (or `goose -dir docs/sql sqlite3 "$DB_PATH" up`) applies all migrations against a
      fresh SQLite file with no errors.
- [ ] `permissions` contains `manageUser`, `managePortfolio`, `manageTransaction`, `manageAlert`
      (no `manageItem`); the admin role is granted all of them.
- [ ] `asset_types` contains the six categories with target allocations
      (crypto 20, idx 30, us 20, precious_metal 15, bonds 15, cash 0).
- [ ] `goose ... down` reverses each migration cleanly.
- [ ] On boot with `SUPERUSER_*` set (hash single-quoted in `.env`), exactly one `admin`-role user is
      created with the env hash stored verbatim, and `POST /user/login` succeeds.
- [ ] Restarting the app does not create a duplicate user (EnsureSuperuser is idempotent).
- [ ] `GET /ping` returns success; removed `/item` routes return 404.
- [ ] With `foreign_keys` on, inserting a row that references a missing FK fails and is detected by
      `IsForeignKeyViolation`.

## Checklist

- [x] One commit per layer; each builds standalone.
- [x] Migrations use goose `Up`/`Down` directives (SQLite dialect).
- [x] No money/quantity stored as float (decimal-as-`TEXT`).
- [x] No panic/fatal added in business code.
- [x] Migration chain verified end-to-end against a temp SQLite DB.
