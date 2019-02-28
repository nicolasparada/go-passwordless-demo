-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DROP DATABASE IF EXISTS passwordless_demo CASCADE;
CREATE DATABASE IF NOT EXISTS passwordless_demo;
SET DATABASE = passwordless_demo;

CREATE TABLE IF NOT EXISTS users (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR NOT NULL UNIQUE,
    username VARCHAR(18) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS verification_codes (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO users (email, username) VALUES
    ('john@passwordless.local', 'john_doe');
