-- +goose Up
-- Roles
INSERT INTO roles (name, code, created_at, updated_at) VALUES
    ('Admin', 'admin', CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER)),
    ('User',  'user',  CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER));

-- Permissions
INSERT INTO permissions (name, code, created_at, updated_at) VALUES
    ('Mengelola pengguna',   'manageUser',        CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER)),
    ('Mengelola portofolio', 'managePortfolio',   CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER)),
    ('Mengelola transaksi',  'manageTransaction', CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER)),
    ('Mengelola peringatan', 'manageAlert',       CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER));

-- Admin role gets every permission
INSERT INTO role_permissions (role_id, permission_id, created_at, updated_at)
SELECT r.id, p.id, CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER)
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'admin';

-- The single WealthPulse owner is seeded separately from environment variables
-- in 20260616000003_seed_superuser.sql — no hardcoded user is created here.

-- +goose Down
DELETE FROM role_permissions
    WHERE role_id IN (SELECT id FROM roles WHERE code = 'admin');
DELETE FROM permissions WHERE code IN ('manageUser', 'managePortfolio', 'manageTransaction', 'manageAlert');
DELETE FROM roles WHERE code IN ('admin', 'user');
