package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	keyPrefix = "oracle:rate:"
	rateTTL   = 60 * time.Second
)

var Currencies = []string{"SOL", "ETH", "USDT"}

type RateStore struct {
	client *redis.Client
}

func NewRateStore(redisURL string) (*RateStore, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	return &RateStore{client: redis.NewClient(opt)}, nil
}

func (s *RateStore) SetRate(ctx context.Context, currency, rateKRW string) error {
	return s.client.Set(ctx, keyPrefix+currency, rateKRW, rateTTL).Err()
}

func (s *RateStore) GetRate(ctx context.Context, currency string) (string, error) {
	val, err := s.client.Get(ctx, keyPrefix+currency).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("no cached rate for %s", currency)
	}
	return val, err
}

func (s *RateStore) GetAll(ctx context.Context) (map[string]string, error) {
	result := make(map[string]string, len(Currencies))
	for _, c := range Currencies {
		val, err := s.client.Get(ctx, keyPrefix+c).Result()
		if err == redis.Nil {
			result[c] = "0"
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("get rate for %s: %w", c, err)
		}
		result[c] = val
	}
	return result, nil
}
