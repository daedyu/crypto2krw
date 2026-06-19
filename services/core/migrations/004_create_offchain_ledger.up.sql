CREATE TABLE core.offchain_ledger (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES core.users (id),
    currency       TEXT NOT NULL
                       CHECK (currency IN ('SOL', 'USDT_ERC20', 'USDT_TRC20', 'ETH')),
    balance        NUMERIC(30, 8) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    locked_balance NUMERIC(30, 8) NOT NULL DEFAULT 0 CHECK (locked_balance >= 0),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, currency)
);

-- 결제 차감 hot-path: (user_id, currency) 조회
CREATE INDEX idx_offchain_ledger_user_currency ON core.offchain_ledger (user_id, currency);
