package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/crypto2krw/oracle/internal/config"
	"github.com/crypto2krw/oracle/internal/feed"
	"github.com/crypto2krw/oracle/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	rateStore, err := store.NewRateStore(cfg.RedisURL)
	if err != nil {
		slog.Error("init rate store", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go feed.RunUpbit(ctx, rateStore)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/rates", ratesHandler(rateStore))
	mux.HandleFunc("GET /api/v1/rates/{currency}", rateHandler(rateStore))
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("oracle service listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server error", "error", err)
		}
	}()

	<-ctx.Done()
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutCtx)
	slog.Info("oracle service stopped")
}

func ratesHandler(s *store.RateStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rates, err := s.GetAll(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiResponse(false, nil, "rates unavailable"))
			return
		}
		writeJSON(w, http.StatusOK, apiResponse(true, rates, ""))
	}
}

func rateHandler(s *store.RateStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currency := r.PathValue("currency")
		rate, err := s.GetRate(r.Context(), currency)
		if err != nil {
			writeJSON(w, http.StatusNotFound, apiResponse(false, nil, err.Error()))
			return
		}
		writeJSON(w, http.StatusOK, apiResponse(true, map[string]string{"currency": currency, "rate_krw": rate}, ""))
	}
}

func apiResponse(success bool, data any, errMsg string) map[string]any {
	resp := map[string]any{"success": success, "data": data}
	if errMsg != "" {
		resp["error"] = map[string]string{"message": errMsg}
	}
	return resp
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
