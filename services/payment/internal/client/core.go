package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

type CoreClient struct {
	baseURL string
	http    *http.Client
}

func NewCoreClient(baseURL string) *CoreClient {
	return &CoreClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type DebitRequest struct {
	UserID      string `json:"user_id"`
	MerchantID  string `json:"merchant_id"`
	Currency    string `json:"currency"`
	Amount      string `json:"amount"`
	AmountKRW   string `json:"amount_krw"`
	AppliedRate string `json:"applied_rate"`
	PaymentRef  string `json:"payment_ref"`
}

type DebitResult struct {
	TransactionID    string
	RemainingBalance decimal.Decimal
}

type debitResp struct {
	Success bool `json:"success"`
	Data    struct {
		TransactionID    string `json:"transaction_id"`
		RemainingBalance string `json:"remaining_balance"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Debit는 Core 내부 API를 호출해 사용자 장부에서 결제 금액을 차감한다.
func (c *CoreClient) Debit(ctx context.Context, req DebitRequest) (*DebitResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/internal/debit", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("core debit request failed: %w", err)
	}
	defer resp.Body.Close()

	var result debitResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		msg := "debit failed"
		if result.Error != nil {
			msg = result.Error.Message
		}
		return nil, fmt.Errorf("%s (http %d)", msg, resp.StatusCode)
	}

	remaining, _ := decimal.NewFromString(result.Data.RemainingBalance)
	return &DebitResult{
		TransactionID:    result.Data.TransactionID,
		RemainingBalance: remaining,
	}, nil
}
