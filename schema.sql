DROP DATABASE IF EXISTS passwordless_demo CASCADE;
CREATE DATABASE IF NOT EXISTS passwordless_demo;
SET DATABASE = passwordless_demo;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email STRING UNIQUE,
    username STRING UNIQUE
);

CREATE TABLE IF NOT EXISTS verification_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO users (email, username) VALUES
    ('john@passwordless.local', 'john_doe');
