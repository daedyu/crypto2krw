ALTER TABLE core.transactions DROP CONSTRAINT IF EXISTS fk_transactions_deposit_event;
DROP TABLE IF EXISTS core.deposit_events;
