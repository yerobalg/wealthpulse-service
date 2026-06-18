-- +goose Up
CREATE TABLE roles (
    id          BIGSERIAL PRIMARY KEY,
    created_at  BIGINT NOT NULL DEFAULT 0,
    updated_at  BIGINT NOT NULL DEFAULT 0,
    deleted_at  TIMESTAMPTZ NULL,
    created_by  BIGINT NULL,
    updated_by  BIGINT NULL,
    deleted_by  BIGINT NULL,
    name        VARCHAR(255) NOT NULL,
    code        VARCHAR(255) NOT NULL
);
CREATE UNIQUE INDEX idx_roles_code ON roles (code);
CREATE INDEX idx_roles_deleted_at ON roles (deleted_at);

CREATE TABLE permissions (
    id          BIGSERIAL PRIMARY KEY,
    created_at  BIGINT NOT NULL DEFAULT 0,
    updated_at  BIGINT NOT NULL DEFAULT 0,
    deleted_at  TIMESTAMPTZ NULL,
    created_by  BIGINT NULL,
    updated_by  BIGINT NULL,
    deleted_by  BIGINT NULL,
    name        VARCHAR(255) NOT NULL,
    code        VARCHAR(255) NOT NULL
);
CREATE UNIQUE INDEX idx_permissions_code ON permissions (code);
CREATE INDEX idx_permissions_deleted_at ON permissions (deleted_at);

CREATE TABLE role_permissions (
    id            BIGSERIAL PRIMARY KEY,
    created_at    BIGINT NOT NULL DEFAULT 0,
    updated_at    BIGINT NOT NULL DEFAULT 0,
    deleted_at    TIMESTAMPTZ NULL,
    created_by    BIGINT NULL,
    updated_by    BIGINT NULL,
    deleted_by    BIGINT NULL,
    role_id       BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
    CONSTRAINT fk_role_permissions_permission FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE
);
CREATE INDEX idx_role_permissions_role_id ON role_permissions (role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions (permission_id);
CREATE INDEX idx_role_permissions_deleted_at ON role_permissions (deleted_at);

CREATE TABLE users (
    id                   BIGSERIAL PRIMARY KEY,
    created_at           BIGINT NOT NULL DEFAULT 0,
    updated_at           BIGINT NOT NULL DEFAULT 0,
    deleted_at           TIMESTAMPTZ NULL,
    created_by           BIGINT NULL,
    updated_by           BIGINT NULL,
    deleted_by           BIGINT NULL,
    username             VARCHAR(255) NOT NULL,
    password             TEXT NOT NULL,
    name                 VARCHAR(255) NOT NULL,
    position             VARCHAR(255) NULL,
    is_male              BOOLEAN NOT NULL,
    role_id              BIGINT NOT NULL,
    has_changed_password BOOLEAN NOT NULL DEFAULT FALSE,
    is_inactive          BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES roles (id)
);
CREATE UNIQUE INDEX idx_users_username ON users (username);
CREATE INDEX idx_users_deleted_at ON users (deleted_at);
CREATE INDEX idx_users_role_id ON users (role_id);

CREATE TABLE revoked_tokens (
    id          BIGSERIAL PRIMARY KEY,
    created_at  BIGINT NOT NULL DEFAULT 0,
    updated_at  BIGINT NOT NULL DEFAULT 0,
    deleted_at  TIMESTAMPTZ NULL,
    created_by  BIGINT NULL,
    updated_by  BIGINT NULL,
    deleted_by  BIGINT NULL,
    user_id     BIGINT NOT NULL,
    token       TEXT NOT NULL,
    expired_at  BIGINT NULL,
    reason      VARCHAR(255) NOT NULL,
    CONSTRAINT fk_revoked_tokens_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX idx_revoked_tokens_token ON revoked_tokens (token);
CREATE INDEX idx_revoked_tokens_user_id ON revoked_tokens (user_id);
CREATE INDEX idx_revoked_tokens_deleted_at ON revoked_tokens (deleted_at);

CREATE TABLE activity_logs (
    id                BIGSERIAL PRIMARY KEY,
    created_at        BIGINT NOT NULL DEFAULT 0,
    updated_at        BIGINT NOT NULL DEFAULT 0,
    deleted_at        TIMESTAMPTZ NULL,
    created_by        BIGINT NULL,
    updated_by        BIGINT NULL,
    deleted_by        BIGINT NULL,
    user_id           BIGINT NOT NULL,
    user_token        TEXT NOT NULL,
    metadata          TEXT NOT NULL,
    activity_event    TEXT NOT NULL,
    activity_name     TEXT NOT NULL,
    additional_fields TEXT NULL,
    CONSTRAINT fk_activity_logs_user FOREIGN KEY (user_id) REFERENCES users (id)
);
CREATE INDEX idx_activity_logs_user_id ON activity_logs (user_id);
CREATE INDEX idx_activity_logs_deleted_at ON activity_logs (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS activity_logs;
DROP TABLE IF EXISTS revoked_tokens;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
