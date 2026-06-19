package ledger

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Repository는 ledger 변경 시 반드시 *sql.Tx를 받아 ACID를 보장한다.
// 서비스 계층이 트랜잭션 경계를 소유하고 레포지터리에 전달한다.
type Repository interface {
	// FindForUpdate는 결제 차감 hot-path에서 FOR UPDATE 락과 함께 조회한다.
	FindForUpdate(ctx context.Context, tx *sql.Tx, userID uuid.UUID, currency string) (*OffchainLedger, error)

	// FindByUser는 락 없이 잔액 조회 (읽기 전용)
	FindByUser(ctx context.Context, userID uuid.UUID) ([]*OffchainLedger, error)

	// Credit은 잔액을 증가시킨다. 반드시 트랜잭션 내에서 호출해야 한다.
	Credit(ctx context.Context, tx *sql.Tx, userID uuid.UUID, currency string, amount decimal.Decimal) error

	// Debit은 잔액을 감소시킨다. 반드시 트랜잭션 내에서 호출해야 한다.
	Debit(ctx context.Context, tx *sql.Tx, userID uuid.UUID, currency string, amount decimal.Decimal) error

	// EnsureExists는 (userID, currency) 행이 없으면 0으로 초기화한다.
	EnsureExists(ctx context.Context, tx *sql.Tx, userID uuid.UUID, currency string) error
}
