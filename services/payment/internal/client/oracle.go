package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

type OracleClient struct {
	baseURL string
	http    *http.Client
}

func NewOracleClient(baseURL string) *OracleClient {
	return &OracleClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 3 * time.Second},
	}
}

type rateResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Currency string `json:"currency"`
		RateKRW  string `json:"rate_krw"`
	} `json:"data"`
}

// GetRate는 지정 통화의 KRW 환율을 Oracle 서비스에서 조회한다.
func (c *OracleClient) GetRate(ctx context.Context, currency string) (decimal.Decimal, error) {
	url := fmt.Sprintf("%s/api/v1/rates/%s", c.baseURL, currency)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return decimal.Zero, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return decimal.Zero, fmt.Errorf("oracle request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("oracle returned %d for currency=%s", resp.StatusCode, currency)
	}

	var body rateResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return decimal.Zero, err
	}

	rate, err := decimal.NewFromString(body.Data.RateKRW)
	if err != nil || rate.IsZero() {
		return decimal.Zero, fmt.Errorf("invalid rate from oracle: %s", body.Data.RateKRW)
	}
	return rate, nil
}
