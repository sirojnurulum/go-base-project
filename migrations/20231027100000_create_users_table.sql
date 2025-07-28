-- +goose Up
-- +goose StatementBegin
-- SQL in this section is executed when the migration is applied.


-- Aktifkan ekstensi pgcrypto untuk gen_random_bytes()
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Membuat fungsi untuk generate UUIDv7.
-- UUIDv7 adalah time-ordered, yang sangat baik untuk performa primary key di database.
CREATE OR REPLACE FUNCTION uuid_generate_v7()
RETURNS UUID AS $$
DECLARE
  unix_ts_ms BYTEA;
  rand_bytes BYTEA;
  version_bits_and_rand BYTEA;
  variant_bits_and_rand BYTEA;
BEGIN
  unix_ts_ms := int8send((FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000))::BIGINT);
  rand_bytes := gen_random_bytes(10);

  version_bits_and_rand := SET_BYTE(
    SUBSTRING(rand_bytes FROM 1 FOR 2), 0,
    (GET_BYTE(SUBSTRING(rand_bytes FROM 1 FOR 1), 0) & 15) | 112 -- Version 7 (0111)
  );

  variant_bits_and_rand := SET_BYTE(
    SUBSTRING(rand_bytes FROM 3 FOR 2), 0,
    (GET_BYTE(SUBSTRING(rand_bytes FROM 3 FOR 1), 0) & 63) | 128 -- Variant (10)
  );

  RETURN ENCODE(SUBSTRING(unix_ts_ms FROM 3 FOR 6) || version_bits_and_rand || variant_bits_and_rand || SUBSTRING(rand_bytes FROM 5 FOR 6), 'hex')::UUID;
END;
$$ LANGUAGE plpgsql VOLATILE;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE,
    password VARCHAR(255), -- Dibuat nullable untuk user yang daftar via Google/SSO
    role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    google_id VARCHAR(255) UNIQUE,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ -- Kolom ini penting untuk fitur GORM soft delete
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_google_id ON users(google_id);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SQL in this section is executed when the migration is rolled back.
DROP FUNCTION IF EXISTS uuid_generate_v7();
DROP TABLE IF EXISTS users;
-- +goose StatementEnd