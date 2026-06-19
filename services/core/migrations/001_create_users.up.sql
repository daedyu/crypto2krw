CREATE SCHEMA IF NOT EXISTS core;

CREATE TABLE core.users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    kyc_status    TEXT NOT NULL DEFAULT 'UNVERIFIED'
                      CHECK (kyc_status IN ('UNVERIFIED', 'SUBMITTED', 'APPROVED', 'REJECTED')),
    status        TEXT NOT NULL DEFAULT 'ACTIVE'
                      CHECK (status IN ('ACTIVE', 'SUSPENDED', 'PENDING_KYC')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON core.users (email);
