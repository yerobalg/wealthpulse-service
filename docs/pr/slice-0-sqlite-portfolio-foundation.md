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
  stored as `TEXT` (for `shopspring/decimal`). Seeds: six default asset types, portfolio permissions
  (`managePortfolio`, `manageTransaction`, `manageAlert`), and the single superuser via goose
  `ENVSUB` from `SUPERUSER_*` env.
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
- [ ] After migrating, `users` contains exactly one row whose `username`/`name`/`password` match the
      `SUPERUSER_*` env values (bcrypt `$` preserved), `role_id` = admin, `has_changed_password` = 1.
- [ ] `permissions` contains `manageUser`, `managePortfolio`, `manageTransaction`, `manageAlert`
      (no `manageItem`); the admin role is granted all of them.
- [ ] `asset_types` contains the six categories with target allocations
      (crypto 20, idx 30, us 20, precious_metal 15, bonds 15, cash 0).
- [ ] `goose ... down` reverses each migration cleanly.
- [ ] App boots with `DB_PATH` set and `POST /user/login` succeeds for the seeded superuser.
- [ ] `GET /ping` returns success; removed `/item` routes return 404.
- [ ] With `foreign_keys` on, inserting a row that references a missing FK fails and is detected by
      `IsForeignKeyViolation`.

## Checklist

- [x] One commit per layer; each builds standalone.
- [x] Migrations use goose `Up`/`Down` directives (SQLite dialect).
- [x] No money/quantity stored as float (decimal-as-`TEXT`).
- [x] No panic/fatal added in business code.
- [x] Migration chain verified end-to-end against a temp SQLite DB.
