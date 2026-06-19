package account

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status UserStatus) error
	UpdateKYCStatus(ctx context.Context, id uuid.UUID, status KYCStatus) error
}

type MerchantRepository interface {
	Create(ctx context.Context, merchant *Merchant) (*Merchant, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Merchant, error)
	FindByEmail(ctx context.Context, email string) (*Merchant, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status MerchantStatus) error
}
