package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/crypto2krw/core/internal/domain/ledger"
	"github.com/crypto2krw/core/internal/domain/wallet"
	"github.com/google/uuid"
)

const (
	TopicLedgerDebited      = "crypto2krw.core.ledger.debited"
	TopicLedgerCredited     = "crypto2krw.core.ledger.credited"
	TopicTransactionCreated = "crypto2krw.core.transaction.created"
	TopicUserSuspended      = "crypto2krw.core.user.suspended"
	TopicDepositDetected    = "crypto2krw.core.deposit.detected"
	TopicWalletCreated      = "crypto2krw.core.wallet.created"
)

// CloudEvent는 CloudEvents v1.0 스펙 봉투이다.
type CloudEvent struct {
	SpecVersion     string `json:"specversion"`
	ID              string `json:"id"`
	Source          string `json:"source"`
	Type            string `json:"type"`
	Time            string `json:"time"`
	DataContentType string `json:"datacontenttype"`
	Data            any    `json:"data"`
}

type Publisher struct {
	producer *kafka.Producer
}

func NewPublisher(brokers string) (*Publisher, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"acks":               "all",   // 리더 + 모든 ISR 확인 후 ack
		"retries":            5,
		"enable.idempotence": true,    // 프로듀서 멱등성
	})
	if err != nil {
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}

	// 비동기 delivery report 고루틴
	go func() {
		for range p.Events() {
		}
	}()

	return &Publisher{producer: p}, nil
}

func (p *Publisher) Close() {
	p.producer.Flush(5000)
	p.producer.Close()
}

func (p *Publisher) PublishLedgerDebited(ctx context.Context, event ledger.LedgerDebitedEvent) error {
	return p.publish(ctx, TopicLedgerDebited, event.UserID.String(), &CloudEvent{
		SpecVersion:     "1.0",
		ID:              uuid.New().String(),
		Source:          "core/ledger-service",
		Type:            "crypto2krw.core.ledger.debited",
		Time:            time.Now().UTC().Format(time.RFC3339),
		DataContentType: "application/json",
		Data: map[string]any{
			"transaction_id":  event.TransactionID.String(),
			"user_id":         event.UserID.String(),
			"merchant_id":     event.MerchantID.String(),
			"payment_ref":     event.PaymentRef,
			"currency":        event.Currency,
			"deducted_amount": event.DeductedAmount.String(),
			"amount_krw":      event.AmountKRW.String(),
			"applied_rate":    event.AppliedRate.String(),
		},
	})
}

func (p *Publisher) PublishLedgerCredited(ctx context.Context, event ledger.LedgerCreditedEvent) error {
	return p.publish(ctx, TopicLedgerCredited, event.UserID.String(), &CloudEvent{
		SpecVersion:     "1.0",
		ID:              uuid.New().String(),
		Source:          "core/ledger-service",
		Type:            "crypto2krw.core.ledger.credited",
		Time:            time.Now().UTC().Format(time.RFC3339),
		DataContentType: "application/json",
		Data: map[string]any{
			"transaction_id":   event.TransactionID.String(),
			"user_id":          event.UserID.String(),
			"deposit_event_id": event.DepositEventID.String(),
			"currency":         event.Currency,
			"credited_amount":  event.CreditedAmount.String(),
		},
	})
}

func (p *Publisher) PublishUserSuspended(ctx context.Context, userID uuid.UUID, reason string) error {
	return p.publish(ctx, TopicUserSuspended, userID.String(), &CloudEvent{
		SpecVersion:     "1.0",
		ID:              uuid.New().String(),
		Source:          "core/account-service",
		Type:            "crypto2krw.core.user.suspended",
		Time:            time.Now().UTC().Format(time.RFC3339),
		DataContentType: "application/json",
		Data: map[string]any{
			"user_id": userID.String(),
			"reason":  reason,
		},
	})
}

func (p *Publisher) PublishWalletCreated(ctx context.Context, w *wallet.UserWallet) error {
	return p.publish(ctx, TopicWalletCreated, w.UserID.String(), &CloudEvent{
		SpecVersion:     "1.0",
		ID:              uuid.New().String(),
		Source:          "core/wallet-service",
		Type:            "crypto2krw.core.wallet.created",
		Time:            time.Now().UTC().Format(time.RFC3339),
		DataContentType: "application/json",
		Data: map[string]any{
			"user_id":  w.UserID.String(),
			"currency": string(w.Currency),
			"address":  w.Address,
			"network":  networkForCurrency(w.Currency),
		},
	})
}

func networkForCurrency(c wallet.Currency) string {
	switch c {
	case wallet.CurrencySOL:
		return "SOLANA"
	case wallet.CurrencyETH, wallet.CurrencyUSDTERC20:
		return "ETHEREUM"
	case wallet.CurrencyUSDTTRC20:
		return "TRON"
	default:
		return string(c)
	}
}

func (p *Publisher) publish(ctx context.Context, topic, key string, event *CloudEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          payload,
	}, nil)
}
