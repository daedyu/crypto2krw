package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/crypto2krw/core/internal/domain/ledger"
	"github.com/crypto2krw/core/internal/domain/wallet"
	"github.com/shopspring/decimal"
)

// DepositDetectedData는 chain-watcher가 발행하는 입금 감지 이벤트 데이터이다.
type DepositDetectedData struct {
	ChainTxHash string `json:"chain_tx_hash"`
	Network     string `json:"network"`
	ToAddress   string `json:"to_address"`
	Currency    string `json:"currency"`
	Amount      string `json:"amount"` // decimal string, float 금지
	BlockNumber *int64 `json:"block_number,omitempty"`
}

type cloudEventEnvelope struct {
	Data json.RawMessage `json:"data"`
}

// DepositConsumer는 deposit.detected 이벤트를 소비해 장부에 입금을 반영한다.
type DepositConsumer struct {
	consumer     *kafka.Consumer
	walletSvc    *wallet.Service
	ledgerSvc    *ledger.Service
}

func NewDepositConsumer(brokers string, walletSvc *wallet.Service, ledgerSvc *ledger.Service) (*DepositConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"group.id":           "core-deposit-processor",
		"auto.offset.reset":  "earliest",
		// 자동 커밋 비활성화: 처리 성공 후 수동 커밋으로 at-least-once 보장
		"enable.auto.commit": false,
	})
	if err != nil {
		return nil, fmt.Errorf("create kafka consumer: %w", err)
	}

	if err := c.Subscribe(TopicDepositDetected, nil); err != nil {
		c.Close()
		return nil, fmt.Errorf("subscribe to %s: %w", TopicDepositDetected, err)
	}

	return &DepositConsumer{
		consumer:  c,
		walletSvc: walletSvc,
		ledgerSvc: ledgerSvc,
	}, nil
}

// Run은 컨텍스트가 취소될 때까지 메시지를 소비한다.
func (c *DepositConsumer) Run(ctx context.Context) {
	slog.Info("deposit consumer started", "topic", TopicDepositDetected)
	for {
		select {
		case <-ctx.Done():
			slog.Info("deposit consumer stopping")
			c.consumer.Close()
			return
		default:
			msg, err := c.consumer.ReadMessage(500) // 500ms 폴링 타임아웃
			if err != nil {
				var kafkaErr kafka.Error
				if errors.As(err, &kafkaErr) && kafkaErr.Code() == kafka.ErrTimedOut {
					continue // 폴링 타임아웃은 정상; 다음 루프
				}
				slog.Error("kafka read error", "error", err)
				continue
			}

			if processErr := c.processMessage(ctx, msg); processErr != nil {
				slog.Error("deposit processing failed",
					"offset", msg.TopicPartition.Offset,
					"error", processErr,
				)
				// 처리 실패 시 커밋하지 않아 재처리 보장
				continue
			}

			// 처리 성공 후 오프셋 커밋
			if _, err := c.consumer.CommitMessage(msg); err != nil {
				slog.Error("commit offset failed", "error", err)
			}
		}
	}
}

func (c *DepositConsumer) processMessage(ctx context.Context, msg *kafka.Message) error {
	var envelope cloudEventEnvelope
	if err := json.Unmarshal(msg.Value, &envelope); err != nil {
		// 파싱 불가 메시지는 재처리해도 의미 없으므로 스킵 (dead-letter 처리는 추후 구현)
		slog.Warn("malformed deposit.detected message, skipping", "error", err)
		return nil
	}

	var data DepositDetectedData
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		slog.Warn("malformed deposit data, skipping", "error", err)
		return nil
	}

	amount, err := decimal.NewFromString(data.Amount)
	if err != nil || amount.IsZero() || amount.IsNegative() {
		slog.Warn("invalid deposit amount, skipping", "amount", data.Amount)
		return nil
	}

	currency := wallet.Currency(data.Currency)

	// to_address → user_id 역조회
	w, err := c.walletSvc.ResolveAddressToUser(ctx, data.ToAddress, currency)
	if err != nil {
		if errors.Is(err, wallet.ErrAddressNotFound) {
			// 플랫폼 외 주소로 전송된 tx — 스킵
			slog.Debug("unknown deposit address, skipping", "address", data.ToAddress)
			return nil
		}
		return fmt.Errorf("resolve address %s: %w", data.ToAddress, err)
	}

	_, err = c.ledgerSvc.CreditDeposit(
		ctx,
		w.UserID,
		data.Currency,
		amount,
		data.ChainTxHash,
		data.Network,
		data.ToAddress,
		data.BlockNumber,
	)
	if err != nil {
		if errors.Is(err, ledger.ErrDepositAlreadyCredited) {
			slog.Debug("deposit already credited, skipping", "tx", data.ChainTxHash)
			return nil // 멱등 처리
		}
		return fmt.Errorf("credit deposit %s: %w", data.ChainTxHash, err)
	}

	slog.Info("deposit credited",
		"user_id", w.UserID,
		"currency", data.Currency,
		"amount", amount.String(),
		"tx", data.ChainTxHash,
	)
	return nil
}
