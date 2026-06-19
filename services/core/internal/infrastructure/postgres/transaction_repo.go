package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/crypto2krw/core/internal/domain/transaction"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *sql.Tx, t *transaction.Transaction) (*transaction.Transaction, error) {
	const q = `
		INSERT INTO core.transactions
			(internal_ref, user_id, merchant_id, type, amount_krw,
			 used_currency, used_amount, applied_rate, status, deposit_event_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`

	err := tx.QueryRowContext(ctx, q,
		t.InternalRef, t.UserID, t.MerchantID, t.Type,
		t.AmountKRW, t.UsedCurrency, t.UsedAmount,
		t.AppliedRate, t.Status, t.DepositEventID,
	).Scan(&t.ID, &t.CreatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			// internal_ref UNIQUE 제약 위반 = 이미 처리된 요청
			return nil, transaction.ErrNotFound // 상위에서 ErrAlreadyProcessed로 처리
		}
		return nil, fmt.Errorf("insert transaction: %w", err)
	}
	return t, nil
}

func (r *TransactionRepository) FindByID(ctx context.Context, id uuid.UUID) (*transaction.Transaction, error) {
	const q = `
		SELECT id, internal_ref, user_id, merchant_id, type, amount_krw,
		       used_currency, used_amount, applied_rate, status, deposit_event_id, created_at
		FROM core.transactions WHERE id = $1`

	t := &transaction.Transaction{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&t.ID, &t.InternalRef, &t.UserID, &t.MerchantID, &t.Type,
		&t.AmountKRW, &t.UsedCurrency, &t.UsedAmount,
		&t.AppliedRate, &t.Status, &t.DepositEventID, &t.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, transaction.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find transaction by id: %w", err)
	}
	return t, nil
}

func (r *TransactionRepository) FindByInternalRef(ctx context.Context, internalRef string) (*transaction.Transaction, error) {
	const q = `
		SELECT id, internal_ref, user_id, merchant_id, type, amount_krw,
		       used_currency, used_amount, applied_rate, status, deposit_event_id, created_at
		FROM core.transactions WHERE internal_ref = $1`

	t := &transaction.Transaction{}
	err := r.db.QueryRowContext(ctx, q, internalRef).Scan(
		&t.ID, &t.InternalRef, &t.UserID, &t.MerchantID, &t.Type,
		&t.AmountKRW, &t.UsedCurrency, &t.UsedAmount,
		&t.AppliedRate, &t.Status, &t.DepositEventID, &t.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, transaction.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find transaction by ref: %w", err)
	}
	return t, nil
}

func (r *TransactionRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*transaction.Transaction, error) {
	const q = `
		SELECT id, internal_ref, user_id, merchant_id, type, amount_krw,
		       used_currency, used_amount, applied_rate, status, deposit_event_id, created_at
		FROM core.transactions WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	return r.queryTransactions(ctx, r.db, q, userID, limit, offset)
}

func (r *TransactionRepository) FindByMerchantID(ctx context.Context, merchantID uuid.UUID, limit, offset int) ([]*transaction.Transaction, error) {
	const q = `
		SELECT id, internal_ref, user_id, merchant_id, type, amount_krw,
		       used_currency, used_amount, applied_rate, status, deposit_event_id, created_at
		FROM core.transactions WHERE merchant_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	return r.queryTransactions(ctx, r.db, q, merchantID, limit, offset)
}

func (r *TransactionRepository) queryTransactions(ctx context.Context, db *sql.DB, q string, args ...any) ([]*transaction.Transaction, error) {
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	var result []*transaction.Transaction
	for rows.Next() {
		t := &transaction.Transaction{}
		if err := rows.Scan(
			&t.ID, &t.InternalRef, &t.UserID, &t.MerchantID, &t.Type,
			&t.AmountKRW, &t.UsedCurrency, &t.UsedAmount,
			&t.AppliedRate, &t.Status, &t.DepositEventID, &t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		result = append(result, t)
	}
	return result, rows.Err()
}
