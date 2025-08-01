-- +goose Up
-- +goose StatementBegin


CREATE TABLE users (
    id TEXT PRIMARY KEY,
    shop_domain TEXT UNIQUE NOT NULL,
    access_token TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE shopify_store (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    shop_domain TEXT UNIQUE NOT NULL,
    access_token TEXT NOT NULL,
    scope TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS shopify_store;
DROP TABLE IF EXISTS users;

-- +goose StatementEnd
