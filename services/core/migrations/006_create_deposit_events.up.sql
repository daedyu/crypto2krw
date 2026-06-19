CREATE TABLE core.deposit_events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_tx_hash TEXT    NOT NULL,
    network       TEXT    NOT NULL CHECK (network IN ('SOLANA', 'ETHEREUM', 'TRON')),
    to_address    TEXT    NOT NULL,
    currency      TEXT    NOT NULL
                      CHECK (currency IN ('SOL', 'USDT_ERC20', 'USDT_TRC20', 'ETH')),
    amount        NUMERIC(30, 8) NOT NULL CHECK (amount > 0),
    block_number  BIGINT,
    detected_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    credited_at   TIMESTAMPTZ,  -- NULL이면 아직 장부 미적립
    user_id       UUID REFERENCES core.users (id),
    -- 이중 입금 방지 최후 보루: 동일 tx_hash + network는 1건만 허용
    UNIQUE (chain_tx_hash, network)
);

CREATE INDEX idx_deposit_events_user_id    ON core.deposit_events (user_id, detected_at DESC);
CREATE INDEX idx_deposit_events_uncredited ON core.deposit_events (credited_at) WHERE credited_at IS NULL;

-- transactions 테이블에 deposit_event_id FK 추가
ALTER TABLE core.transactions
    ADD CONSTRAINT fk_transactions_deposit_event
        FOREIGN KEY (deposit_event_id) REFERENCES core.deposit_events (id);
