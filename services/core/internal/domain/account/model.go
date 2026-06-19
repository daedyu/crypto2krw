package account

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string
type KYCStatus string

const (
	UserStatusActive     UserStatus = "ACTIVE"
	UserStatusSuspended  UserStatus = "SUSPENDED"
	UserStatusPendingKYC UserStatus = "PENDING_KYC"

	KYCStatusUnverified KYCStatus = "UNVERIFIED"
	KYCStatusSubmitted  KYCStatus = "SUBMITTED"
	KYCStatusApproved   KYCStatus = "APPROVED"
	KYCStatusRejected   KYCStatus = "REJECTED"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	KYCStatus    KYCStatus
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type MerchantStatus string

const (
	MerchantStatusActive    MerchantStatus = "ACTIVE"
	MerchantStatusSuspended MerchantStatus = "SUSPENDED"
)

type Merchant struct {
	ID           uuid.UUID
	BusinessName string
	Email        string
	PasswordHash string
	KRWBalance   string // Decimal string
	Status       MerchantStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
