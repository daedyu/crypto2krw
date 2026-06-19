package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/crypto2krw/core/internal/domain/ledger"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type LedgerRepository struct {
	db *sql.DB
}

func NewLedgerRepository(db *sql.DB) *LedgerRepository {
	return &LedgerRepository{db: db}
}

func (r *LedgerRepository) FindForUpdate(ctx context.Context, tx *sql.Tx, userID uuid.UUID, currency string) (*ledger.OffchainLedger, error) {
	const q = `
		SELECT id, user_id, currency, balance, locked_balance, updated_at
		FROM core.offchain_ledger
		WHERE user_id = $1 AND currency = $2
		FOR UPDATE`

	l := &ledger.OffchainLedger{}
	err := tx.QueryRowContext(ctx, q, userID, currency).Scan(
		&l.ID, &l.UserID, &l.Currency, &l.Balance, &l.LockedBalance, &l.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("ledger row not found for user %s currency %s", userID, currency)
	}
	if err != nil {
		return nil, fmt.Errorf("find ledger for update: %w", err)
	}
	return l, nil
}

func (r *LedgerRepository) FindByUser(ctx context.Context, userID uuid.UUID) ([]*ledger.OffchainLedger, error) {
	const q = `
		SELECT id, user_id, currency, balance, locked_balance, updated_at
		FROM core.offchain_ledger
		WHERE user_id = $1
		ORDER BY currency`

	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("query ledger by user: %w", err)
	}
	defer rows.Close()

	var result []*ledger.OffchainLedger
	for rows.Next() {
		l := &ledger.OffchainLedger{}
		if err := rows.Scan(&l.ID, &l.UserID, &l.Currency, &l.Balance, &l.LockedBalance, &l.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan ledger row: %w", err)
		}
		result = append(result, l)
	}
	return result, rows.Err()
}

func (r *LedgerRepository) Credit(ctx context.Context, tx *sql.Tx, userID uuid.UUID, currency string, amount decimal.Decimal) error {
	const q = `
		UPDATE core.offchain_ledger
		SET balance = balance + $1, updated_at = now()
		WHERE user_id = $2 AND currency = $3`

	result, err := tx.ExecContext(ctx, q, amount, userID, currency)
	if err != nil {
		return fmt.Errorf("credit ledger: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("ledger row not found for credit")
	}
	return nil
}

func (r *LedgerRepository) Debit(ctx context.Context, tx *sql.Tx, userID uuid.UUID, currency string, amount decimal.Decimal) error {
	const q = `
		UPDATE core.offchain_ledger
		SET balance = balance - $1, updated_at = now()
		WHERE user_id = $2 AND currency = $3 AND (balance - locked_balance) >= $1`

	result, err := tx.ExecContext(ctx, q, amount, userID, currency)
	if err != nil {
		return fmt.Errorf("debit ledger: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		// 업데이트가 0행이면 잔액 부족 (FOR UPDATE 이후 재확인이므로 안전)
		return ledger.ErrInsufficientFunds
	}
	return nil
}

func (r *LedgerRepository) EnsureExists(ctx context.Context, tx *sql.Tx, userID uuid.UUID, currency string) error {
	const q = `
		INSERT INTO core.offchain_ledger (user_id, currency, balance, locked_balance)
		VALUES ($1, $2, 0, 0)
		ON CONFLICT (user_id, currency) DO NOTHING`

	_, err := tx.ExecContext(ctx, q, userID, currency)
	if err != nil {
		return fmt.Errorf("ensure ledger exists: %w", err)
	}
	return nil
}

// DepositEventRepository

type DepositEventRepository struct {
	db *sql.DB
}

func NewDepositEventRepository(db *sql.DB) *DepositEventRepository {
	return &DepositEventRepository{db: db}
}

func (r *DepositEventRepository) FindByTxHash(ctx context.Context, txHash, network string) (*ledger.DepositEvent, error) {
	const q = `
		SELECT id, chain_tx_hash, network, to_address, currency, amount,
		       block_number, detected_at, credited_at, user_id
		FROM core.deposit_events
		WHERE chain_tx_hash = $1 AND network = $2`

	e := &ledger.DepositEvent{}
	err := r.db.QueryRowContext(ctx, q, txHash, network).Scan(
		&e.ID, &e.ChainTxHash, &e.Network, &e.ToAddress, &e.Currency, &e.Amount,
		&e.BlockNumber, &e.DetectedAt, &e.CreditedAt, &e.UserID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ledger.ErrDepositNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find deposit event: %w", err)
	}
	return e, nil
}

func (r *DepositEventRepository) Create(ctx context.Context, tx *sql.Tx, e *ledger.DepositEvent) (*ledger.DepositEvent, error) {
	const q = `
		INSERT INTO core.deposit_events (chain_tx_hash, network, to_address, currency, amount, block_number)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (chain_tx_hash, network) DO NOTHING
		RETURNING id, detected_at`

	err := tx.QueryRowContext(ctx, q,
		e.ChainTxHash, e.Network, e.ToAddress, e.Currency, e.Amount, e.BlockNumber,
	).Scan(&e.ID, &e.DetectedAt)

	if errors.Is(err, sql.ErrNoRows) {
		// ON CONFLICT DO NOTHING: 이미 존재하는 입금 이벤트
		return nil, ledger.ErrDepositAlreadyCredited
	}
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ledger.ErrDepositAlreadyCredited
		}
		return nil, fmt.Errorf("insert deposit event: %w", err)
	}
	return e, nil
}

func (r *DepositEventRepository) MarkCredited(ctx context.Context, tx *sql.Tx, eventID uuid.UUID, userID uuid.UUID) error {
	now := time.Now()
	const q = `
		UPDATE core.deposit_events
		SET credited_at = $1, user_id = $2
		WHERE id = $3`

	_, err := tx.ExecContext(ctx, q, now, userID, eventID)
	if err != nil {
		return fmt.Errorf("mark deposit credited: %w", err)
	}
	return nil
}
