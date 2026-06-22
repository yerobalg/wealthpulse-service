-- +goose Up
-- Roles
INSERT INTO roles (name, code, created_at, updated_at) VALUES
    ('Admin', 'admin', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
    ('User',  'user',  EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT);

-- Permissions
INSERT INTO permissions (name, code, created_at, updated_at) VALUES
    ('Mengelola pengguna', 'manageUser', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
    ('Mengelola item',     'manageItem', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT);

-- Admin role gets every permission
INSERT INTO role_permissions (role_id, permission_id, created_at, updated_at)
SELECT r.id, p.id, EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'admin';

-- The single WealthPulse owner is seeded separately from environment variables
-- in 20260616000003_seed_superuser.sql — no hardcoded user is created here.

-- +goose Down
DELETE FROM role_permissions
    WHERE role_id IN (SELECT id FROM roles WHERE code = 'admin');
DELETE FROM permissions WHERE code IN ('manageUser', 'manageItem');
DELETE FROM roles WHERE code IN ('admin', 'user');
