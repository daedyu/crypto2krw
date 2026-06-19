package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/crypto2krw/core/internal/domain/wallet"
	"github.com/google/uuid"
)

const TopicUserRegistered = "crypto2krw.core.user.registered"

type userRegisteredData struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// UserConsumer는 auth 서비스가 발행한 user.registered 이벤트를 수신해
// 해당 유저의 SOL/ETH/USDT 지갑 주소를 자동으로 생성한다.
type UserConsumer struct {
	consumer  *kafka.Consumer
	walletSvc *wallet.Service
}

func NewUserConsumer(brokers string, walletSvc *wallet.Service) (*UserConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"group.id":           "core-user-provisioner",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	})
	if err != nil {
		return nil, fmt.Errorf("create user consumer: %w", err)
	}

	if err := c.Subscribe(TopicUserRegistered, nil); err != nil {
		c.Close()
		return nil, fmt.Errorf("subscribe to %s: %w", TopicUserRegistered, err)
	}

	return &UserConsumer{consumer: c, walletSvc: walletSvc}, nil
}

func (c *UserConsumer) Run(ctx context.Context) {
	slog.Info("user consumer started", "topic", TopicUserRegistered)
	for {
		select {
		case <-ctx.Done():
			slog.Info("user consumer stopping")
			c.consumer.Close()
			return
		default:
			msg, err := c.consumer.ReadMessage(500)
			if err != nil {
				var kafkaErr kafka.Error
				if errors.As(err, &kafkaErr) && kafkaErr.Code() == kafka.ErrTimedOut {
					continue
				}
				slog.Error("user consumer read error", "error", err)
				continue
			}

			if err := c.processMessage(ctx, msg); err != nil {
				slog.Error("user provisioning failed", "error", err)
				continue
			}

			if _, err := c.consumer.CommitMessage(msg); err != nil {
				slog.Error("commit offset failed", "error", err)
			}
		}
	}
}

func (c *UserConsumer) processMessage(ctx context.Context, msg *kafka.Message) error {
	var envelope cloudEventEnvelope
	if err := json.Unmarshal(msg.Value, &envelope); err != nil {
		slog.Warn("malformed user.registered message, skipping", "error", err)
		return nil
	}

	var data userRegisteredData
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		slog.Warn("malformed user data, skipping", "error", err)
		return nil
	}

	userID, err := uuid.Parse(data.UserID)
	if err != nil {
		slog.Warn("invalid user_id in user.registered, skipping", "user_id", data.UserID)
		return nil
	}

	wallets, err := c.walletSvc.AllocateDepositAddresses(ctx, userID)
	if err != nil {
		return fmt.Errorf("allocate wallets for user %s: %w", userID, err)
	}

	slog.Info("wallets provisioned for new user",
		"user_id", userID,
		"wallet_count", len(wallets),
	)
	return nil
}
