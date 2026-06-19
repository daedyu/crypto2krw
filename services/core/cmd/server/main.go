package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/crypto2krw/core/internal/config"
	"github.com/crypto2krw/core/internal/domain/account"
	"github.com/crypto2krw/core/internal/domain/ledger"
	"github.com/crypto2krw/core/internal/domain/wallet"
	grpcserver "github.com/crypto2krw/core/internal/grpc"
	"github.com/crypto2krw/core/internal/infrastructure/kafka"
	pginfra "github.com/crypto2krw/core/internal/infrastructure/postgres"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	_ "github.com/lib/pq"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		slog.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		slog.Error("ping postgres", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to postgres")

	publisher, err := kafka.NewPublisher(cfg.KafkaBrokers)
	if err != nil {
		slog.Error("create kafka publisher", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()

	userRepo := pginfra.NewUserRepository(db)
	merchantRepo := pginfra.NewMerchantRepository(db)
	walletRepo := pginfra.NewWalletRepository(db)
	ledgerRepo := pginfra.NewLedgerRepository(db)
	txRepo := pginfra.NewTransactionRepository(db)
	depositRepo := pginfra.NewDepositEventRepository(db)

	_ = account.NewService(userRepo, merchantRepo)
	walletSvc := wallet.NewService(walletRepo, publisher)
	ledgerSvc := ledger.NewService(db, ledgerRepo, txRepo, depositRepo, publisher)

	depositConsumer, err := kafka.NewDepositConsumer(cfg.KafkaBrokers, walletSvc, ledgerSvc)
	if err != nil {
		slog.Error("create deposit consumer", "error", err)
		os.Exit(1)
	}

	userConsumer, err := kafka.NewUserConsumer(cfg.KafkaBrokers, walletSvc)
	if err != nil {
		slog.Error("create user consumer", "error", err)
		os.Exit(1)
	}

	grpcSrv := grpcserver.NewServer(cfg.GRPCPort)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go depositConsumer.Run(ctx)
	go userConsumer.Run(ctx)

	go func() {
		slog.Info("starting grpc server", "port", cfg.GRPCPort)
		if err := grpcSrv.ListenAndServe(); err != nil {
			slog.Error("grpc server error", "error", err)
		}
	}()

	// 내부 서비스 간 HTTP API (Payment Service → Core)
	internalPort := getEnv("INTERNAL_HTTP_PORT", "8080")
	internalMux := http.NewServeMux()
	internalMux.HandleFunc("POST /internal/debit", debitHandler(ledgerSvc))
	internalMux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	internalSrv := &http.Server{
		Addr:         fmt.Sprintf(":%s", internalPort),
		Handler:      internalMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("starting internal http server", "port", internalPort)
		if err := internalSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("internal http server error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down gracefully")

	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = internalSrv.Shutdown(shutCtx)
	grpcSrv.GracefulStop()
}

type debitRequest struct {
	UserID     string `json:"user_id"`
	MerchantID string `json:"merchant_id"`
	Currency   string `json:"currency"`
	Amount     string `json:"amount"`
	AmountKRW  string `json:"amount_krw"`
	AppliedRate string `json:"applied_rate"`
	PaymentRef string `json:"payment_ref"`
}

type debitResponse struct {
	TransactionID    string `json:"transaction_id"`
	RemainingBalance string `json:"remaining_balance"`
}

func debitHandler(ledgerSvc *ledger.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req debitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errResponse("invalid request body"))
			return
		}

		userID, err := uuid.Parse(req.UserID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errResponse("invalid user_id"))
			return
		}
		merchantID, err := uuid.Parse(req.MerchantID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errResponse("invalid merchant_id"))
			return
		}
		amount, err := decimal.NewFromString(req.Amount)
		if err != nil || amount.IsZero() || amount.IsNegative() {
			writeJSON(w, http.StatusBadRequest, errResponse("invalid amount"))
			return
		}
		amountKRW, err := decimal.NewFromString(req.AmountKRW)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errResponse("invalid amount_krw"))
			return
		}
		appliedRate, err := decimal.NewFromString(req.AppliedRate)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errResponse("invalid applied_rate"))
			return
		}

		result, err := ledgerSvc.DebitForPayment(
			r.Context(),
			userID, merchantID,
			req.Currency, amount, amountKRW, appliedRate,
			req.PaymentRef,
		)
		if err != nil {
			switch {
			case errors.Is(err, ledger.ErrInsufficientFunds):
				writeJSON(w, http.StatusPaymentRequired, errResponse("insufficient funds"))
			case errors.Is(err, ledger.ErrAccountSuspended):
				writeJSON(w, http.StatusForbidden, errResponse("account suspended"))
			case errors.Is(err, ledger.ErrAlreadyProcessed):
				// 멱등: 이미 처리된 결제는 200 반환
				writeJSON(w, http.StatusOK, map[string]any{
					"success": true,
					"data":    debitResponse{TransactionID: result.TransactionID.String(), RemainingBalance: result.RemainingBalance.String()},
				})
			default:
				slog.Error("debit failed", "error", err, "payment_ref", req.PaymentRef)
				writeJSON(w, http.StatusInternalServerError, errResponse("internal error"))
			}
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data": debitResponse{
				TransactionID:    result.TransactionID.String(),
				RemainingBalance: result.RemainingBalance.String(),
			},
		})
	}
}

func errResponse(msg string) map[string]any {
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

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
