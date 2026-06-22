-- +goose Up
-- Default asset types with target allocation (percent as decimal string).
INSERT INTO asset_types (name, code, target_allocation, created_at, updated_at) VALUES
    ('Crypto',          'crypto',         '20', CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER)),
    ('IDX Stocks',      'idx',            '30', CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER)),
    ('US Stocks/ETFs',  'us',             '20', CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER)),
    ('Precious Metals', 'precious_metal', '15', CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER)),
    ('IDN Bonds',       'bonds',          '15', CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER)),
    ('Cash',            'cash',           '0',  CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER));

-- +goose Down
DELETE FROM asset_types
    WHERE code IN ('crypto', 'idx', 'us', 'precious_metal', 'bonds', 'cash');
