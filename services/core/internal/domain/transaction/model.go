package transaction

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Type string
type Status string

const (
	TypePayment    Type = "PAYMENT"
	TypeDeposit    Type = "DEPOSIT"
	TypeWithdrawal Type = "WITHDRAWAL"

	StatusPending   Status = "PENDING"
	StatusCompleted Status = "COMPLETED"
	StatusFailed    Status = "FAILED"
	StatusReversed  Status = "REVERSED"
)

type Transaction struct {
	ID             uuid.UUID
	InternalRef    string // 멱등성 키 (qr_session_id, "deposit:{network}:{txHash}" 등)
	UserID         uuid.UUID
	MerchantID     *uuid.UUID
	Type           Type
	AmountKRW      *decimal.Decimal
	UsedCurrency   string
	UsedAmount     decimal.Decimal
	AppliedRate    *decimal.Decimal
	Status         Status
	DepositEventID *uuid.UUID
	CreatedAt      time.Time
}
