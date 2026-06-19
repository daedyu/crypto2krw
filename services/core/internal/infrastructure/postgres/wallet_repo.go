package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/crypto2krw/core/internal/domain/wallet"
	"github.com/google/uuid"
)

type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) Create(ctx context.Context, w *wallet.UserWallet) (*wallet.UserWallet, error) {
	const q = `
		INSERT INTO core.user_wallets (user_id, currency, address, private_key_hex, payment_priority)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	var privateKeyHex *string
	if w.PrivateKeyHex != "" {
		privateKeyHex = &w.PrivateKeyHex
	}

	err := r.db.QueryRowContext(ctx, q, w.UserID, w.Currency, w.Address, privateKeyHex, w.PaymentPriority).
		Scan(&w.ID, &w.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}
	return w, nil
}

func (r *WalletRepository) FindByUserAndCurrency(ctx context.Context, userID uuid.UUID, currency wallet.Currency) (*wallet.UserWallet, error) {
	const q = `
		SELECT id, user_id, currency, address, payment_priority, created_at
		FROM core.user_wallets WHERE user_id = $1 AND currency = $2`

	w := &wallet.UserWallet{}
	err := r.db.QueryRowContext(ctx, q, userID, currency).Scan(
		&w.ID, &w.UserID, &w.Currency, &w.Address, &w.PaymentPriority, &w.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, wallet.ErrWalletNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find wallet: %w", err)
	}
	return w, nil
}

func (r *WalletRepository) FindAllByUser(ctx context.Context, userID uuid.UUID) ([]*wallet.UserWallet, error) {
	const q = `
		SELECT id, user_id, currency, address, payment_priority, created_at
		FROM core.user_wallets WHERE user_id = $1
		ORDER BY payment_priority ASC`

	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("find all wallets: %w", err)
	}
	defer rows.Close()

	var result []*wallet.UserWallet
	for rows.Next() {
		w := &wallet.UserWallet{}
		if err := rows.Scan(&w.ID, &w.UserID, &w.Currency, &w.Address, &w.PaymentPriority, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan wallet row: %w", err)
		}
		result = append(result, w)
	}
	return result, rows.Err()
}

func (r *WalletRepository) FindByAddress(ctx context.Context, address string, currency wallet.Currency) (*wallet.UserWallet, error) {
	const q = `
		SELECT id, user_id, currency, address, payment_priority, created_at
		FROM core.user_wallets WHERE address = $1 AND currency = $2`

	w := &wallet.UserWallet{}
	err := r.db.QueryRowContext(ctx, q, address, currency).Scan(
		&w.ID, &w.UserID, &w.Currency, &w.Address, &w.PaymentPriority, &w.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, wallet.ErrAddressNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find wallet by address: %w", err)
	}
	return w, nil
}

func (r *WalletRepository) UpdatePaymentPriorities(ctx context.Context, userID uuid.UUID, currencies []wallet.Currency) error {
	for i, currency := range currencies {
		const q = `
			UPDATE core.user_wallets SET payment_priority = $1
			WHERE user_id = $2 AND currency = $3`
		if _, err := r.db.ExecContext(ctx, q, i+1, userID, currency); err != nil {
			return fmt.Errorf("update priority for %s: %w", currency, err)
		}
	}
	return nil
}
