package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/crypto2krw/payment/internal/client"
	"github.com/crypto2krw/payment/internal/session"
)

type Handler struct {
	sessions      *session.Store
	oracleClient  *client.OracleClient
	coreClient    *client.CoreClient
	jwtSecret     []byte
}

func New(
	sessions *session.Store,
	oracle *client.OracleClient,
	core *client.CoreClient,
	jwtSecret string,
) *Handler {
	return &Handler{
		sessions:     sessions,
		oracleClient: oracle,
		coreClient:   core,
		jwtSecret:    []byte(jwtSecret),
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// QR 생성: 가맹점이 호출 (merchant_id를 body에 포함, 단순화된 인증)
	mux.HandleFunc("POST /api/v1/payment/qr", h.createQR)
	// QR 결제: 유저가 JWT와 함께 호출
	mux.HandleFunc("POST /api/v1/payment/qr/{token}/pay", h.payQR)
	// QR 상태 조회
	mux.HandleFunc("GET /api/v1/payment/qr/{token}", h.getQR)
}

// createQR은 가맹점이 KRW 금액으로 결제 QR 세션을 생성한다.
func (h *Handler) createQR(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MerchantID string `json:"merchant_id"`
		AmountKRW  string `json:"amount_krw"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid request body"))
		return
	}

	merchantID := req.MerchantID
	if _, err := uuid.Parse(merchantID); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid merchant_id"))
		return
	}

	amountKRW, err := decimal.NewFromString(req.AmountKRW)
	if err != nil || amountKRW.IsZero() || amountKRW.IsNegative() {
		writeJSON(w, http.StatusBadRequest, errResp("invalid amount_krw"))
		return
	}

	token := uuid.New().String()
	sess := &session.QRSession{
		Token:      token,
		MerchantID: merchantID,
		AmountKRW:  amountKRW,
	}

	if err := h.sessions.Create(r.Context(), sess); err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("failed to create session"))
		return
	}

	writeJSON(w, http.StatusCreated, successResp(map[string]any{
		"token":      token,
		"amount_krw": amountKRW.String(),
		"expires_at": sess.ExpiresAt.Format(time.RFC3339),
		"qr_payload": fmt.Sprintf("crypto2krw://pay?token=%s", token),
	}))
}

// payQR은 유저가 QR을 스캔하고 결제를 처리한다.
func (h *Handler) payQR(w http.ResponseWriter, r *http.Request) {
	// JWT에서 user_id 추출
	userID, err := h.extractUserID(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errResp("unauthorized"))
		return
	}

	token := r.PathValue("token")
	if token == "" {
		writeJSON(w, http.StatusBadRequest, errResp("missing token"))
		return
	}

	var req struct {
		Currency string `json:"currency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid request body"))
		return
	}

	currency := strings.ToUpper(req.Currency)
	if currency == "" {
		currency = "SOL"
	}

	// 앱은 "USDT"로 보내지만 Oracle은 "USDT", 원장은 "USDT_TRC20"으로 저장됨.
	// 두 가지 정규화 변수를 분리해서 사용한다.
	var oracleCurrency, ledgerCurrency string
	switch currency {
	case "SOL", "ETH":
		oracleCurrency = currency
		ledgerCurrency = currency
	case "USDT", "USDT_TRC20", "USDT_ERC20":
		oracleCurrency = "USDT"
		ledgerCurrency = "USDT_TRC20"
	default:
		writeJSON(w, http.StatusBadRequest, errResp("unsupported currency: "+currency))
		return
	}

	// QR 세션 조회
	sess, err := h.sessions.Get(r.Context(), token)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errResp("qr session not found or expired"))
		return
	}
	if sess.Status != session.StatusPending {
		writeJSON(w, http.StatusConflict, errResp("qr session already "+string(sess.Status)))
		return
	}

	// 환율 조회 (Oracle은 "USDT" 키로 저장)
	rate, err := h.oracleClient.GetRate(r.Context(), oracleCurrency)
	if err != nil {
		slog.Error("oracle get rate failed", "currency", currency, "error", err)
		writeJSON(w, http.StatusServiceUnavailable, errResp("exchange rate unavailable"))
		return
	}

	// 코인 금액 계산: amount_krw / rate (소수점 8자리 반올림)
	coinAmount := sess.AmountKRW.Div(rate).Round(8)
	if coinAmount.IsZero() {
		writeJSON(w, http.StatusInternalServerError, errResp("computed zero coin amount"))
		return
	}

	// Core DebitForPayment 호출 (원장 통화명 사용)
	debitResult, err := h.coreClient.Debit(r.Context(), client.DebitRequest{
		UserID:      userID,
		MerchantID:  sess.MerchantID,
		Currency:    ledgerCurrency,
		Amount:      coinAmount.String(),
		AmountKRW:   sess.AmountKRW.String(),
		AppliedRate: rate.String(),
		PaymentRef:  token, // QR 토큰 = 멱등성 키
	})
	if err != nil {
		slog.Error("debit failed", "error", err, "token", token)
		switch {
		case strings.Contains(err.Error(), "insufficient funds"):
			writeJSON(w, http.StatusPaymentRequired, errResp("잔액이 부족합니다"))
		case strings.Contains(err.Error(), "account suspended"):
			writeJSON(w, http.StatusForbidden, errResp("계정이 정지되었습니다"))
		default:
			writeJSON(w, http.StatusInternalServerError, errResp("payment processing failed"))
		}
		return
	}

	// 세션 완료 처리
	if err := h.sessions.Complete(r.Context(), token, userID, ledgerCurrency, debitResult.TransactionID); err != nil {
		slog.Warn("complete session failed (payment succeeded)", "error", err, "token", token)
	}

	writeJSON(w, http.StatusOK, successResp(map[string]any{
		"transaction_id":    debitResult.TransactionID,
		"merchant_id":       sess.MerchantID,
		"amount_krw":        sess.AmountKRW.String(),
		"used_currency":     ledgerCurrency,
		"used_amount":       coinAmount.String(),
		"applied_rate":      rate.String(),
		"remaining_balance": debitResult.RemainingBalance.String(),
	}))
}

// getQR은 QR 세션의 현재 상태를 반환한다.
func (h *Handler) getQR(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	sess, err := h.sessions.Get(r.Context(), token)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errResp("qr session not found"))
		return
	}
	writeJSON(w, http.StatusOK, successResp(sess))
}

// extractUserID는 Authorization Bearer JWT에서 user_id(sub)를 추출한다.
func (h *Handler) extractUserID(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("missing bearer token")
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return h.jwtSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", errors.New("missing sub claim")
	}
	return sub, nil
}

func successResp(data any) map[string]any {
	return map[string]any{"success": true, "data": data}
}

func errResp(msg string) map[string]any {
	return map[string]any{
		"success": false,
		"data":    nil,
		"error":   map[string]string{"message": msg},
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
