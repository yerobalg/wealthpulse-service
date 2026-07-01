# Slice 2.0 — Remove Activity Log & Revoked Token

## Summary

Removes the activity-logging (audit) and revoked-token features end to end, along with the endpoints
that depended on them (`Logout`, `ChangePassword`). WealthPulse is a single-owner app with stateless
JWT auth; server-side token revocation and OWASP audit logging add wiring the product does not use.

## Changes

- **Deleted feature files:** `entity/activity_log.go`, `entity/revoked_token.go`,
  `repository/activity_log.go`, `repository/revoked_token.go`, `usecase/activity_log.go`,
  `usecase/revoked_token.go`, and `helper/logger/audit_log.go` (its only consumers were the two
  removed features). `helper/logger/log.go` is unaffected.
- **Removed audit logging** from the `User` usecase — `Login`, `CreateUser`, and `UpdateUser` no
  longer call `activityLog.Log(...)`.
- **Removed `Logout`** (usecase method, handler, and `POST /user/logout` route). With stateless JWTs
  and no revocation table there is nothing to invalidate server-side; clients drop the token.
- **Removed `ChangePassword`** (usecase method, handler, `PATCH /user/password` route, and the
  `entity.ChangePasswordRequest` struct).
- **Middleware `checkToken`** no longer checks the revoked-token store or asynchronously revokes
  expired tokens; it just decodes the JWT and sets the auth context. Removed the now-unused
  `revokeExpiredToken` helper and the `types` / `entity` imports.
- **Wiring:** dropped `ActivityLog` / `RevokedToken` from `repository.Repository` and `usecase.Usecase`,
  and `RevokedTokenRepo` / `ActivityLog` / `TxManager` from `UserInitParam` (the user usecase no longer
  needs a transaction). `main.go` is unchanged.
- **Migration:** dropped the `revoked_tokens` and `activity_logs` `CREATE TABLE` blocks (and their
  `DROP` lines) from `docs/sql/20260616000000_create_auth_tables.sql`, so a fresh DB no longer creates
  the orphaned tables.
- **Docs:** removed the **Activity Logging** section from `CLAUDE.md` and its incidental mentions in
  the Pagination and PR-Files sections.

### Behavioral impact

- **Logout is gone** — a token stays valid until it expires; logout is now purely client-side.
- **Password change is gone** — there is no in-app way to change a password; the superuser password is
  set via `SUPERUSER_PASSWORD_HASH` at startup.
- **No audit trail** is written for auth or user CRUD.

### Not included (follow-up)

- Swagger docs under `docs/api` still describe the removed routes until regenerated (`swag init`).
- The migration is edited in place (correct for a not-yet-applied schema). If it was already applied
  anywhere, add a separate `DROP TABLE` migration instead.

## Test Plan

- [ ] `go build ./...` and `go vet ./...` are clean.
- [ ] `POST /user/login` still succeeds and returns a token.
- [ ] A valid token still authorizes protected routes (`GET /user/profile`, `GET /user`).
- [ ] An expired/invalid token is rejected with 401 (no revoke side-effect attempted).
- [ ] `POST /user/logout` returns 404 (route removed).
- [ ] `PATCH /user/password` returns 404 (route removed).
- [ ] `POST /user` and `PUT /user/:id` still create/update users successfully (no audit-log call).
- [ ] Grep confirms no remaining references to `ActivityLog`, `RevokedToken`, `Logout`,
      `ChangePassword`, or `AuditLog` in `src/`, `helper/`, or `main.go`.
- [ ] `goose ... up` on a fresh SQLite file creates neither `revoked_tokens` nor `activity_logs`;
      `goose ... down` reverses cleanly.

## Checklist

- [x] Build and vet pass after removal.
- [x] No dangling references in code or `CLAUDE.md`.
- [x] Behavioral impact (logout / password change / audit trail) documented above.
- [x] Orphaned tables removed from the migration; stale swagger flagged as follow-up.
