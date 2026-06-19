CREATE TABLE core.merchants (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_name TEXT NOT NULL,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    krw_balance   NUMERIC(20, 2) NOT NULL DEFAULT 0 CHECK (krw_balance >= 0),
    status        TEXT NOT NULL DEFAULT 'ACTIVE'
                      CHECK (status IN ('ACTIVE', 'SUSPENDED')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_merchants_email ON core.merchants (email);
