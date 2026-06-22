-- +goose Up
-- Seeds the single WealthPulse owner from environment variables using goose ENVSUB.
-- Requires SUPERUSER_USERNAME, SUPERUSER_PASSWORD_HASH, and SUPERUSER_NAME to be
-- present in the environment when `goose up` runs (the Makefile `migrate` target
-- exports them). Generate the bcrypt hash with `make gen-password password=...`.
--
-- NOTE: a bcrypt hash contains '$'. When supplying SUPERUSER_PASSWORD_HASH via .env,
-- double every '$' (e.g. $$2a$$08$$...) so `make` does not treat it as a variable
-- reference; make expands it back to a single '$' before exporting to goose.
-- +goose ENVSUB ON
INSERT INTO users (
    username,
    password,
    name,
    is_male,
    role_id,
    has_changed_password,
    is_inactive,
    created_at,
    updated_at
)
SELECT
    '${SUPERUSER_USERNAME}',
    '${SUPERUSER_PASSWORD_HASH}',
    '${SUPERUSER_NAME}',
    1,
    r.id,
    1,
    0,
    CAST(strftime('%s','now') AS INTEGER),
    CAST(strftime('%s','now') AS INTEGER)
FROM roles r
WHERE r.code = 'admin';
-- +goose ENVSUB OFF

-- +goose Down
-- +goose ENVSUB ON
DELETE FROM users WHERE username = '${SUPERUSER_USERNAME}';
-- +goose ENVSUB OFF
