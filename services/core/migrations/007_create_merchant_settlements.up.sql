CREATE TABLE core.merchant_settlements (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id  UUID        NOT NULL REFERENCES core.merchants (id),
    period_start DATE        NOT NULL,
    period_end   DATE        NOT NULL,
    gross_krw    NUMERIC(20, 2) NOT NULL,
    fee_krw      NUMERIC(20, 2) NOT NULL DEFAULT 0 CHECK (fee_krw >= 0),
    net_krw      NUMERIC(20, 2) NOT NULL,
    status       TEXT        NOT NULL DEFAULT 'PENDING'
                     CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED')),
    processed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (period_start <= period_end),
    CHECK (net_krw = gross_krw - fee_krw)
);

CREATE INDEX idx_merchant_settlements_merchant_id ON core.merchant_settlements (merchant_id, period_start DESC);
CREATE INDEX idx_merchant_settlements_status      ON core.merchant_settlements (status) WHERE status IN ('PENDING', 'PROCESSING');
