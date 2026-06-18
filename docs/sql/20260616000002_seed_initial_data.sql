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

-- Initial admin user (username: admin / password: admin123)
INSERT INTO users (username, password, name, is_male, role_id, has_changed_password, created_at, updated_at)
SELECT 'admin',
       '$2a$08$LtrZK9DGlxsYuUxNf.1cu.CHnG2acwUy9cZgbZHPfjK09WiCmwDca',
       'Administrator',
       TRUE,
       r.id,
       FALSE,
       EXTRACT(EPOCH FROM NOW())::BIGINT,
       EXTRACT(EPOCH FROM NOW())::BIGINT
FROM roles r
WHERE r.code = 'admin';

-- +goose Down
DELETE FROM users WHERE username = 'admin';
DELETE FROM role_permissions
    WHERE role_id IN (SELECT id FROM roles WHERE code = 'admin');
DELETE FROM permissions WHERE code IN ('manageUser', 'manageItem');
DELETE FROM roles WHERE code IN ('admin', 'user');
