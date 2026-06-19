package wallet

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, wallet *UserWallet) (*UserWallet, error)
	FindByUserAndCurrency(ctx context.Context, userID uuid.UUID, currency Currency) (*UserWallet, error)
	FindAllByUser(ctx context.Context, userID uuid.UUID) ([]*UserWallet, error)
	FindByAddress(ctx context.Context, address string, currency Currency) (*UserWallet, error)
	UpdatePaymentPriorities(ctx context.Context, userID uuid.UUID, priorities []Currency) error
}
