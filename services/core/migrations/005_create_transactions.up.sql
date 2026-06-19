CREATE TABLE core.transactions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    internal_ref     TEXT NOT NULL UNIQUE,  -- 멱등성 키 (qr_session_id 등)
    user_id          UUID NOT NULL REFERENCES core.users (id),
    merchant_id      UUID REFERENCES core.merchants (id),
    type             TEXT NOT NULL CHECK (type IN ('PAYMENT', 'DEPOSIT', 'WITHDRAWAL')),
    amount_krw       NUMERIC(20, 2),
    used_currency    TEXT NOT NULL
                         CHECK (used_currency IN ('SOL', 'USDT_ERC20', 'USDT_TRC20', 'ETH')),
    used_amount      NUMERIC(30, 8) NOT NULL CHECK (used_amount > 0),
    applied_rate     NUMERIC(20, 8),  -- KRW per 1 unit, 입금 시 NULL
    status           TEXT NOT NULL DEFAULT 'COMPLETED'
                         CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'REVERSED')),
    deposit_event_id UUID,  -- FK는 006 마이그레이션 후 추가
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_transactions_user_id     ON core.transactions (user_id, created_at DESC);
CREATE INDEX idx_transactions_merchant_id ON core.transactions (merchant_id, created_at DESC);
CREATE INDEX idx_transactions_type        ON core.transactions (type, created_at DESC);
