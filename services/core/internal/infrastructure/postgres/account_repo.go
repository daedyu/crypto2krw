package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/crypto2krw/core/internal/domain/account"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *account.User) (*account.User, error) {
	const q = `
		INSERT INTO core.users (email, password_hash, kyc_status, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, q,
		u.Email, u.PasswordHash, u.KYCStatus, u.Status,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, account.ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*account.User, error) {
	const q = `
		SELECT id, email, password_hash, kyc_status, status, created_at, updated_at
		FROM core.users WHERE id = $1`

	u := &account.User{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.KYCStatus, &u.Status, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, account.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*account.User, error) {
	const q = `
		SELECT id, email, password_hash, kyc_status, status, created_at, updated_at
		FROM core.users WHERE email = $1`

	u := &account.User{}
	err := r.db.QueryRowContext(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.KYCStatus, &u.Status, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, account.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return u, nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status account.UserStatus) error {
	const q = `UPDATE core.users SET status = $1, updated_at = now() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("update user status: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return account.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) UpdateKYCStatus(ctx context.Context, id uuid.UUID, status account.KYCStatus) error {
	const q = `UPDATE core.users SET kyc_status = $1, updated_at = now() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("update kyc status: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return account.ErrUserNotFound
	}
	return nil
}

// MerchantRepository

type MerchantRepository struct {
	db *sql.DB
}

func NewMerchantRepository(db *sql.DB) *MerchantRepository {
	return &MerchantRepository{db: db}
}

func (r *MerchantRepository) Create(ctx context.Context, m *account.Merchant) (*account.Merchant, error) {
	const q = `
		INSERT INTO core.merchants (business_name, email, password_hash, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, krw_balance, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, q,
		m.BusinessName, m.Email, m.PasswordHash, m.Status,
	).Scan(&m.ID, &m.KRWBalance, &m.CreatedAt, &m.UpdatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, account.ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("insert merchant: %w", err)
	}
	return m, nil
}

func (r *MerchantRepository) FindByID(ctx context.Context, id uuid.UUID) (*account.Merchant, error) {
	const q = `
		SELECT id, business_name, email, password_hash, krw_balance, status, created_at, updated_at
		FROM core.merchants WHERE id = $1`

	m := &account.Merchant{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&m.ID, &m.BusinessName, &m.Email, &m.PasswordHash,
		&m.KRWBalance, &m.Status, &m.CreatedAt, &m.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, account.ErrMerchantNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find merchant by id: %w", err)
	}
	return m, nil
}

func (r *MerchantRepository) FindByEmail(ctx context.Context, email string) (*account.Merchant, error) {
	const q = `
		SELECT id, business_name, email, password_hash, krw_balance, status, created_at, updated_at
		FROM core.merchants WHERE email = $1`

	m := &account.Merchant{}
	err := r.db.QueryRowContext(ctx, q, email).Scan(
		&m.ID, &m.BusinessName, &m.Email, &m.PasswordHash,
		&m.KRWBalance, &m.Status, &m.CreatedAt, &m.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, account.ErrMerchantNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find merchant by email: %w", err)
	}
	return m, nil
}

func (r *MerchantRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status account.MerchantStatus) error {
	const q = `UPDATE core.merchants SET status = $1, updated_at = now() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("update merchant status: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return account.ErrMerchantNotFound
	}
	return nil
}
