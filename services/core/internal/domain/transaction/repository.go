package transaction

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("transaction not found")

type Repository interface {
	Create(ctx context.Context, tx *sql.Tx, t *Transaction) (*Transaction, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	FindByInternalRef(ctx context.Context, internalRef string) (*Transaction, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Transaction, error)
	FindByMerchantID(ctx context.Context, merchantID uuid.UUID, limit, offset int) ([]*Transaction, error)
}
