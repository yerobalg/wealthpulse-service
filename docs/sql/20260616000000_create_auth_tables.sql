-- +goose Up
CREATE TABLE roles (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at  INTEGER NOT NULL DEFAULT 0,
    updated_at  INTEGER NOT NULL DEFAULT 0,
    deleted_at  DATETIME NULL,
    created_by  INTEGER NULL,
    updated_by  INTEGER NULL,
    deleted_by  INTEGER NULL,
    name        TEXT NOT NULL,
    code        TEXT NOT NULL
);
CREATE UNIQUE INDEX idx_roles_code ON roles (code);
CREATE INDEX idx_roles_deleted_at ON roles (deleted_at);

CREATE TABLE permissions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at  INTEGER NOT NULL DEFAULT 0,
    updated_at  INTEGER NOT NULL DEFAULT 0,
    deleted_at  DATETIME NULL,
    created_by  INTEGER NULL,
    updated_by  INTEGER NULL,
    deleted_by  INTEGER NULL,
    name        TEXT NOT NULL,
    code        TEXT NOT NULL
);
CREATE UNIQUE INDEX idx_permissions_code ON permissions (code);
CREATE INDEX idx_permissions_deleted_at ON permissions (deleted_at);

CREATE TABLE role_permissions (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at    INTEGER NOT NULL DEFAULT 0,
    updated_at    INTEGER NOT NULL DEFAULT 0,
    deleted_at    DATETIME NULL,
    created_by    INTEGER NULL,
    updated_by    INTEGER NULL,
    deleted_by    INTEGER NULL,
    role_id       INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
    CONSTRAINT fk_role_permissions_permission FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE
);
CREATE INDEX idx_role_permissions_role_id ON role_permissions (role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions (permission_id);
CREATE INDEX idx_role_permissions_deleted_at ON role_permissions (deleted_at);

CREATE TABLE users (
    id                   INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at           INTEGER NOT NULL DEFAULT 0,
    updated_at           INTEGER NOT NULL DEFAULT 0,
    deleted_at           DATETIME NULL,
    created_by           INTEGER NULL,
    updated_by           INTEGER NULL,
    deleted_by           INTEGER NULL,
    username             TEXT NOT NULL,
    password             TEXT NOT NULL,
    name                 TEXT NOT NULL,
    position             TEXT NULL,
    is_male              BOOLEAN NOT NULL,
    role_id              INTEGER NOT NULL,
    has_changed_password BOOLEAN NOT NULL DEFAULT 0,
    is_inactive          BOOLEAN NOT NULL DEFAULT 0,
    CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES roles (id)
);
CREATE UNIQUE INDEX idx_users_username ON users (username);
CREATE INDEX idx_users_deleted_at ON users (deleted_at);
CREATE INDEX idx_users_role_id ON users (role_id);

-- +goose Down
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
