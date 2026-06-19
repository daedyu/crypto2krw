package ledger

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/crypto2krw/core/internal/domain/transaction"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrAccountSuspended   = errors.New("account suspended")
	ErrAlreadyProcessed   = errors.New("already processed")
	ErrDepositAlreadyCredited = errors.New("deposit already credited")
)

type DepositEventRepository interface {
	FindByTxHash(ctx context.Context, txHash, network string) (*DepositEvent, error)
	Create(ctx context.Context, tx *sql.Tx, event *DepositEvent) (*DepositEvent, error)
	MarkCredited(ctx context.Context, tx *sql.Tx, eventID uuid.UUID, userID uuid.UUID) error
}

type DepositEvent struct {
	ID           uuid.UUID
	ChainTxHash  string
	Network      string
	ToAddress    string
	Currency     string
	Amount       decimal.Decimal
	BlockNumber  *int64
	DetectedAt   time.Time
	CreditedAt   *time.Time
	UserID       *uuid.UUID
}

type DB interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type EventPublisher interface {
	PublishLedgerDebited(ctx context.Context, event LedgerDebitedEvent) error
	PublishLedgerCredited(ctx context.Context, event LedgerCreditedEvent) error
}

type LedgerDebitedEvent struct {
	TransactionID  uuid.UUID
	UserID         uuid.UUID
	MerchantID     uuid.UUID
	PaymentRef     string
	Currency       string
	DeductedAmount decimal.Decimal
	AmountKRW      decimal.Decimal
	AppliedRate    decimal.Decimal
}

type LedgerCreditedEvent struct {
	TransactionID uuid.UUID
	UserID        uuid.UUID
	DepositEventID uuid.UUID
	Currency       string
	CreditedAmount decimal.Decimal
}

type Service struct {
	db              DB
	ledgerRepo      Repository
	txRepo          transaction.Repository
	depositRepo     DepositEventRepository
	publisher       EventPublisher
}

func NewService(
	db DB,
	ledgerRepo Repository,
	txRepo transaction.Repository,
	depositRepo DepositEventRepository,
	publisher EventPublisher,
) *Service {
	return &Service{
		db:          db,
		ledgerRepo:  ledgerRepo,
		txRepo:      txRepo,
		depositRepo: depositRepo,
		publisher:   publisher,
	}
}

// DebitForPayment는 결제 차감의 핵심 메서드.
// qr_session_id를 internal_ref로 사용하여 멱등성을 보장한다.
// FOR UPDATE 락으로 동시 요청 시 잔액 음수 방지.
func (s *Service) DebitForPayment(
	ctx context.Context,
	userID uuid.UUID,
	merchantID uuid.UUID,
	currency string,
	amount decimal.Decimal,
	amountKRW decimal.Decimal,
	appliedRate decimal.Decimal,
	paymentRef string, // qr_session_id — 멱등성 키
) (*DebitResult, error) {
	// 멱등성 선행 검사: 이미 처리된 결제인지 확인
	existing, err := s.txRepo.FindByInternalRef(ctx, paymentRef)
	if err != nil && !errors.Is(err, transaction.ErrNotFound) {
		return nil, fmt.Errorf("check idempotency: %w", err)
	}
	if existing != nil {
		return &DebitResult{
			TransactionID:    existing.ID,
			RemainingBalance: decimal.Zero,
		}, ErrAlreadyProcessed
	}

	dbTx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = dbTx.Rollback()
		}
	}()

	// FOR UPDATE: 동일 유저의 동시 결제 요청이 순차 처리되도록 보장
	ledger, err := s.ledgerRepo.FindForUpdate(ctx, dbTx, userID, currency)
	if err != nil {
		return nil, fmt.Errorf("lock ledger row: %w", err)
	}

	if ledger.AvailableBalance().LessThan(amount) {
		return nil, ErrInsufficientFunds
	}

	if err = s.ledgerRepo.Debit(ctx, dbTx, userID, currency, amount); err != nil {
		return nil, fmt.Errorf("debit ledger: %w", err)
	}

	newTx := &transaction.Transaction{
		InternalRef:  paymentRef,
		UserID:       userID,
		MerchantID:   &merchantID,
		Type:         transaction.TypePayment,
		AmountKRW:    &amountKRW,
		UsedCurrency: currency,
		UsedAmount:   amount,
		AppliedRate:  &appliedRate,
		Status:       transaction.StatusCompleted,
	}

	createdTx, err := s.txRepo.Create(ctx, dbTx, newTx)
	if err != nil {
		return nil, fmt.Errorf("create transaction record: %w", err)
	}

	if err = dbTx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	// 커밋 성공 후 이벤트 발행 (at-least-once, 채널계가 멱등 처리)
	_ = s.publisher.PublishLedgerDebited(ctx, LedgerDebitedEvent{
		TransactionID:  createdTx.ID,
		UserID:         userID,
		MerchantID:     merchantID,
		PaymentRef:     paymentRef,
		Currency:       currency,
		DeductedAmount: amount,
		AmountKRW:      amountKRW,
		AppliedRate:    appliedRate,
	})

	remainingBalance := ledger.Balance.Sub(amount)
	return &DebitResult{
		TransactionID:    createdTx.ID,
		RemainingBalance: remainingBalance,
	}, nil
}

// CreditDeposit는 chain-watcher로부터 받은 입금 이벤트를 장부에 반영한다.
// UNIQUE(chain_tx_hash, network) 제약으로 이중 적립을 방지한다.
func (s *Service) CreditDeposit(
	ctx context.Context,
	userID uuid.UUID,
	currency string,
	amount decimal.Decimal,
	chainTxHash string,
	network string,
	toAddress string,
	blockNumber *int64,
) (*CreditResult, error) {
	// 이미 처리된 입금인지 확인
	existing, err := s.depositRepo.FindByTxHash(ctx, chainTxHash, network)
	if err != nil && !errors.Is(err, ErrDepositNotFound) {
		return nil, fmt.Errorf("check deposit idempotency: %w", err)
	}
	if existing != nil && existing.CreditedAt != nil {
		return nil, ErrDepositAlreadyCredited
	}

	dbTx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = dbTx.Rollback()
		}
	}()

	depositEvent := &DepositEvent{
		ChainTxHash: chainTxHash,
		Network:     network,
		ToAddress:   toAddress,
		Currency:    currency,
		Amount:      amount,
		BlockNumber: blockNumber,
	}

	// ON CONFLICT(chain_tx_hash, network)는 DB에서 이중 INSERT를 거부
	createdDeposit, err := s.depositRepo.Create(ctx, dbTx, depositEvent)
	if err != nil {
		return nil, fmt.Errorf("create deposit event: %w", err)
	}

	if err = s.ledgerRepo.EnsureExists(ctx, dbTx, userID, currency); err != nil {
		return nil, fmt.Errorf("ensure ledger row exists: %w", err)
	}

	if err = s.ledgerRepo.Credit(ctx, dbTx, userID, currency, amount); err != nil {
		return nil, fmt.Errorf("credit ledger: %w", err)
	}

	internalRef := fmt.Sprintf("deposit:%s:%s", network, chainTxHash)
	newTx := &transaction.Transaction{
		InternalRef:     internalRef,
		UserID:          userID,
		Type:            transaction.TypeDeposit,
		UsedCurrency:    currency,
		UsedAmount:      amount,
		Status:          transaction.StatusCompleted,
		DepositEventID:  &createdDeposit.ID,
	}

	createdTx, err := s.txRepo.Create(ctx, dbTx, newTx)
	if err != nil {
		return nil, fmt.Errorf("create transaction record: %w", err)
	}

	if err = s.depositRepo.MarkCredited(ctx, dbTx, createdDeposit.ID, userID); err != nil {
		return nil, fmt.Errorf("mark deposit credited: %w", err)
	}

	if err = dbTx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	_ = s.publisher.PublishLedgerCredited(ctx, LedgerCreditedEvent{
		TransactionID:  createdTx.ID,
		UserID:         userID,
		DepositEventID: createdDeposit.ID,
		Currency:       currency,
		CreditedAmount: amount,
	})

	ledgers, _ := s.ledgerRepo.FindByUser(ctx, userID)
	newBalance := amount
	for _, l := range ledgers {
		if l.Currency == currency {
			newBalance = l.Balance
			break
		}
	}

	return &CreditResult{
		TransactionID: createdTx.ID,
		NewBalance:    newBalance,
	}, nil
}

// GetUserBalances는 유저의 전체 통화 잔액을 반환한다.
func (s *Service) GetUserBalances(ctx context.Context, userID uuid.UUID) ([]*OffchainLedger, error) {
	return s.ledgerRepo.FindByUser(ctx, userID)
}

var ErrDepositNotFound = errors.New("deposit event not found")
