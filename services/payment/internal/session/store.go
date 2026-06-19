package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

const (
	sessionTTL = 10 * time.Minute
	keyPrefix  = "payment:qr:"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusCompleted Status = "COMPLETED"
	StatusExpired   Status = "EXPIRED"
)

type QRSession struct {
	Token      string          `json:"token"`
	MerchantID string          `json:"merchant_id"`
	AmountKRW  decimal.Decimal `json:"amount_krw"`
	Status     Status          `json:"status"`
	Currency   string          `json:"currency,omitempty"`
	UserID     string          `json:"user_id,omitempty"`
	TxID       string          `json:"transaction_id,omitempty"`
	ExpiresAt  time.Time       `json:"expires_at"`
	CreatedAt  time.Time       `json:"created_at"`
}

type Store struct {
	client *redis.Client
}

func NewStore(redisURL string) (*Store, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	return &Store{client: redis.NewClient(opt)}, nil
}

func (s *Store) Create(ctx context.Context, sess *QRSession) error {
	sess.ExpiresAt = time.Now().Add(sessionTTL)
	sess.CreatedAt = time.Now()
	sess.Status = StatusPending

	raw, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, keyPrefix+sess.Token, raw, sessionTTL).Err()
}

func (s *Store) Get(ctx context.Context, token string) (*QRSession, error) {
	raw, err := s.client.Get(ctx, keyPrefix+token).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("qr session not found or expired")
	}
	if err != nil {
		return nil, err
	}
	var sess QRSession
	if err := json.Unmarshal(raw, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *Store) Complete(ctx context.Context, token, userID, currency, txID string) error {
	sess, err := s.Get(ctx, token)
	if err != nil {
		return err
	}
	if sess.Status != StatusPending {
		return fmt.Errorf("session already %s", sess.Status)
	}

	sess.Status = StatusCompleted
	sess.UserID = userID
	sess.Currency = currency
	sess.TxID = txID

	raw, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	// 완료 후 1분만 유지 (영수증 조회용)
	return s.client.Set(ctx, keyPrefix+token, raw, time.Minute).Err()
}
