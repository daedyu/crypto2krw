package account

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrMerchantNotFound   = errors.New("merchant not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountSuspended   = errors.New("account suspended")
)

type Service struct {
	userRepo     UserRepository
	merchantRepo MerchantRepository
}

func NewService(userRepo UserRepository, merchantRepo MerchantRepository) *Service {
	return &Service{
		userRepo:     userRepo,
		merchantRepo: merchantRepo,
	}
}

func (s *Service) RegisterUser(ctx context.Context, email, passwordHash string) (*User, error) {
	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return nil, fmt.Errorf("check existing user: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	user := &User{
		Email:        email,
		PasswordHash: passwordHash,
		KYCStatus:    KYCStatusUnverified,
		Status:       UserStatusActive,
	}

	return s.userRepo.Create(ctx, user)
}

func (s *Service) ValidateUserCredentials(ctx context.Context, email, passwordHash string) (*User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if user.PasswordHash != passwordHash {
		return nil, ErrInvalidCredentials
	}

	if user.Status == UserStatusSuspended {
		return nil, ErrAccountSuspended
	}

	return user, nil
}

func (s *Service) SuspendUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.UpdateStatus(ctx, userID, UserStatusSuspended); err != nil {
		return fmt.Errorf("suspend user: %w", err)
	}
	return nil
}

func (s *Service) UpdateKYCStatus(ctx context.Context, userID uuid.UUID, status KYCStatus) error {
	if err := s.userRepo.UpdateKYCStatus(ctx, userID, status); err != nil {
		return fmt.Errorf("update kyc status: %w", err)
	}
	return nil
}

func (s *Service) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

func (s *Service) RegisterMerchant(ctx context.Context, businessName, email, passwordHash string) (*Merchant, error) {
	existing, err := s.merchantRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, ErrMerchantNotFound) {
		return nil, fmt.Errorf("check existing merchant: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	merchant := &Merchant{
		BusinessName: businessName,
		Email:        email,
		PasswordHash: passwordHash,
		KRWBalance:   "0",
		Status:       MerchantStatusActive,
	}

	return s.merchantRepo.Create(ctx, merchant)
}

func (s *Service) ValidateMerchantCredentials(ctx context.Context, email, passwordHash string) (*Merchant, error) {
	merchant, err := s.merchantRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrMerchantNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find merchant: %w", err)
	}

	if merchant.PasswordHash != passwordHash {
		return nil, ErrInvalidCredentials
	}

	if merchant.Status == MerchantStatusSuspended {
		return nil, ErrAccountSuspended
	}

	return merchant, nil
}
