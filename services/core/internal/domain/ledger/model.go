package ledger

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OffchainLedger struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Currency      string
	Balance       decimal.Decimal
	LockedBalance decimal.Decimal
	UpdatedAt     time.Time
}

// AvailableBalance는 실제 출금 가능 잔액을 반환한다.
func (l *OffchainLedger) AvailableBalance() decimal.Decimal {
	return l.Balance.Sub(l.LockedBalance)
}

type DebitResult struct {
	TransactionID    uuid.UUID
	RemainingBalance decimal.Decimal
}

type CreditResult struct {
	TransactionID uuid.UUID
	NewBalance    decimal.Decimal
}
