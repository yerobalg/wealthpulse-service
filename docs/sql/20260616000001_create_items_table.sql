-- +goose Up
CREATE TABLE items (
    id          BIGSERIAL PRIMARY KEY,
    created_at  BIGINT NOT NULL DEFAULT 0,
    updated_at  BIGINT NOT NULL DEFAULT 0,
    deleted_at  TIMESTAMPTZ NULL,
    created_by  BIGINT NULL,
    updated_by  BIGINT NULL,
    deleted_by  BIGINT NULL,
    name        VARCHAR(255) NOT NULL,
    description TEXT NULL,
    price       BIGINT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX idx_items_deleted_at ON items (deleted_at);
CREATE INDEX idx_items_name ON items (name);

-- +goose Down
DROP TABLE IF EXISTS items;
