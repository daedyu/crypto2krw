CREATE TABLE core.user_wallets (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES core.users (id),
    currency         TEXT NOT NULL
                         CHECK (currency IN ('SOL', 'USDT_ERC20', 'USDT_TRC20', 'ETH')),
    address          TEXT NOT NULL,
    payment_priority INT  NOT NULL DEFAULT 99,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, currency),
    UNIQUE (address, currency)
);

CREATE INDEX idx_user_wallets_user_id  ON core.user_wallets (user_id);
CREATE INDEX idx_user_wallets_address  ON core.user_wallets (address, currency);
