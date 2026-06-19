package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/crypto2krw/oracle/internal/store"
)

// Upbit 마켓 코드 → 내부 통화 코드 매핑
var upbitMarkets = map[string]string{
	"KRW-SOL":  "SOL",
	"KRW-ETH":  "ETH",
	"KRW-USDT": "USDT",
}

type upbitTickerMsg struct {
	Type       string  `json:"type"`
	Code       string  `json:"code"`
	TradePrice float64 `json:"trade_price"`
}

type upbitRestTicker struct {
	Market     string  `json:"market"`
	TradePrice float64 `json:"trade_price"`
}

// RunUpbit는 Upbit WebSocket을 우선 시도하고, 연속 3회 실패 시 REST 폴링으로 전환한다.
func RunUpbit(ctx context.Context, s *store.RateStore) {
	wsFails := 0
	for {
		if wsFails >= 3 {
			slog.Warn("upbit ws unreachable, switching to REST polling")
			runRestPolling(ctx, s)
			return
		}

		if err := connectAndConsume(ctx, s); err != nil {
			wsFails++
			slog.Error("upbit ws disconnected", "error", err, "fail_count", wsFails)
		} else {
			wsFails = 0
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

// runRestPolling은 Upbit REST API를 30초마다 폴링해 Redis에 환율을 저장한다.
func runRestPolling(ctx context.Context, s *store.RateStore) {
	codes := make([]string, 0, len(upbitMarkets))
	for code := range upbitMarkets {
		codes = append(codes, code)
	}
	marketsParam := strings.Join(codes, ",")
	url := "https://api.upbit.com/v1/ticker?markets=" + marketsParam

	client := &http.Client{Timeout: 10 * time.Second}

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(30 * time.Second):
		}

		if err := fetchAndStoreRates(ctx, client, url, s); err != nil {
			slog.Error("upbit rest poll failed", "error", err)
		}
	}
}

func fetchAndStoreRates(ctx context.Context, client *http.Client, url string, s *store.RateStore) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upbit rest get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	var tickers []upbitRestTicker
	if err := json.Unmarshal(body, &tickers); err != nil {
		return fmt.Errorf("parse tickers: %w", err)
	}

	for _, t := range tickers {
		currency, ok := upbitMarkets[t.Market]
		if !ok {
			continue
		}
		rateStr := fmt.Sprintf("%.2f", t.TradePrice)
		if err := s.SetRate(ctx, currency, rateStr); err != nil {
			slog.Error("set rate failed", "currency", currency, "error", err)
			continue
		}
		slog.Info("rate updated via REST", "currency", currency, "krw", rateStr)
	}
	return nil
}

func connectAndConsume(ctx context.Context, s *store.RateStore) error {
	header := http.Header{}
	header.Set("User-Agent", "Mozilla/5.0")

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, "wss://api.upbit.com/websocket/v1", header)
	if err != nil {
		return fmt.Errorf("dial upbit ws: %w", err)
	}
	defer conn.Close()

	codes := make([]string, 0, len(upbitMarkets))
	for code := range upbitMarkets {
		codes = append(codes, code)
	}

	subscribeMsg := []any{
		map[string]string{"ticket": uuid.New().String()},
		map[string]any{"type": "ticker", "codes": codes},
	}
	if err := conn.WriteJSON(subscribeMsg); err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	slog.Info("upbit ws connected", "markets", codes)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read message: %w", err)
		}

		var msg upbitTickerMsg
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		if msg.Type != "ticker" {
			continue
		}

		currency, ok := upbitMarkets[msg.Code]
		if !ok {
			continue
		}

		rateStr := fmt.Sprintf("%.2f", msg.TradePrice)
		if err := s.SetRate(ctx, currency, rateStr); err != nil {
			slog.Error("set rate failed", "currency", currency, "error", err)
			continue
		}
		slog.Debug("rate updated", "currency", currency, "krw", rateStr)
	}
}
