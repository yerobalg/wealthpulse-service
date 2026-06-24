-- +goose Up
CREATE TABLE asset_types (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at        INTEGER NOT NULL DEFAULT 0,
    updated_at        INTEGER NOT NULL DEFAULT 0,
    deleted_at        DATETIME NULL,
    created_by        INTEGER NULL,
    updated_by        INTEGER NULL,
    deleted_by        INTEGER NULL,
    name              TEXT NOT NULL,
    code              TEXT NOT NULL,            -- crypto, idx, us, precious_metal, bonds, cash
    target_allocation TEXT NOT NULL DEFAULT '0' -- percent as decimal string, e.g. '20'
);
CREATE UNIQUE INDEX idx_asset_types_code ON asset_types (code);
CREATE INDEX idx_asset_types_deleted_at ON asset_types (deleted_at);

CREATE TABLE assets (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at    INTEGER NOT NULL DEFAULT 0,
    updated_at    INTEGER NOT NULL DEFAULT 0,
    deleted_at    DATETIME NULL,
    created_by    INTEGER NULL,
    updated_by    INTEGER NULL,
    deleted_by    INTEGER NULL,
    asset_type_id INTEGER NOT NULL,
    name          TEXT NOT NULL,
    ticker        TEXT NOT NULL,
    unique_id     TEXT NOT NULL,                -- stable unique key per asset
    image_url     TEXT NULL,                    -- asset logo/icon url
    external_id   TEXT NULL,                    -- provider id (e.g. coingecko "bitcoin")
    CONSTRAINT fk_assets_asset_type FOREIGN KEY (asset_type_id) REFERENCES asset_types (id)
);
CREATE UNIQUE INDEX idx_assets_unique_id ON assets (unique_id);
CREATE INDEX idx_assets_ticker ON assets (ticker);
CREATE INDEX idx_assets_asset_type_id ON assets (asset_type_id);
CREATE INDEX idx_assets_deleted_at ON assets (deleted_at);

CREATE TABLE asset_prices (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at INTEGER NOT NULL DEFAULT 0,
    updated_at INTEGER NOT NULL DEFAULT 0,
    deleted_at DATETIME NULL,
    asset_id   INTEGER NOT NULL,
    price_usd  TEXT NULL,                       -- decimal string
    price_idr  TEXT NULL,
    CONSTRAINT fk_asset_prices_asset FOREIGN KEY (asset_id) REFERENCES assets (id)
);
CREATE INDEX idx_asset_prices_asset_created ON asset_prices (asset_id, created_at DESC);

CREATE TABLE transactions (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at          INTEGER NOT NULL DEFAULT 0,
    updated_at          INTEGER NOT NULL DEFAULT 0,
    deleted_at          DATETIME NULL,
    created_by          INTEGER NULL,
    updated_by          INTEGER NULL,
    deleted_by          INTEGER NULL,
    asset_id            INTEGER NOT NULL,
    type                TEXT NOT NULL,          -- 'buy' | 'sell'
    quantity            TEXT NOT NULL,          -- decimal string
    price_usd           TEXT NULL,
    price_idr           TEXT NULL,
    annual_return_bonds TEXT NULL,              -- bonds coupon rate, percent decimal string
    transaction_date    INTEGER NOT NULL,       -- epoch; the user-entered trade date
    notes               TEXT NULL,
    CONSTRAINT fk_transactions_asset FOREIGN KEY (asset_id) REFERENCES assets (id)
);
CREATE INDEX idx_transactions_asset_id ON transactions (asset_id);
CREATE INDEX idx_transactions_date ON transactions (transaction_date DESC);
CREATE INDEX idx_transactions_deleted_at ON transactions (deleted_at);

CREATE TABLE alerts (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at        INTEGER NOT NULL DEFAULT 0,
    updated_at        INTEGER NOT NULL DEFAULT 0,
    deleted_at        DATETIME NULL,
    created_by        INTEGER NULL,
    updated_by        INTEGER NULL,
    deleted_by        INTEGER NULL,
    asset_id          INTEGER NULL,             -- NULL for portfolio-wide pct alerts
    type              TEXT NOT NULL,            -- 'upper' | 'lower' | 'pct'
    threshold         TEXT NOT NULL,            -- decimal string
    is_active         BOOLEAN NOT NULL DEFAULT 1,
    last_triggered_at INTEGER NULL,
    CONSTRAINT fk_alerts_asset FOREIGN KEY (asset_id) REFERENCES assets (id)
);
CREATE INDEX idx_alerts_asset_id ON alerts (asset_id);
CREATE INDEX idx_alerts_is_active ON alerts (is_active);
CREATE INDEX idx_alerts_deleted_at ON alerts (deleted_at);

CREATE TABLE asset_value (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at      INTEGER NOT NULL DEFAULT 0,
    updated_at      INTEGER NOT NULL DEFAULT 0,
    deleted_at      DATETIME NULL,
    asset_type_id   INTEGER NOT NULL,
    total_value_usd TEXT NULL,                  -- decimal string
    total_value_idr TEXT NULL,
    CONSTRAINT fk_asset_value_asset_type FOREIGN KEY (asset_type_id) REFERENCES asset_types (id)
);
CREATE INDEX idx_asset_value_type_created ON asset_value (asset_type_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS asset_value;
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS asset_prices;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS asset_types;
