package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/crypto2krw/payment/internal/client"
	"github.com/crypto2krw/payment/internal/config"
	"github.com/crypto2krw/payment/internal/handler"
	"github.com/crypto2krw/payment/internal/session"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	sessStore, err := session.NewStore(cfg.RedisURL)
	if err != nil {
		slog.Error("init session store", "error", err)
		os.Exit(1)
	}

	oracleClient := client.NewOracleClient(cfg.OracleURL)
	coreClient := client.NewCoreClient(cfg.CoreInternalURL)

	h := handler.New(sessStore, oracleClient, coreClient, cfg.JWTAccessSecret)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("payment service listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server error", "error", err)
		}
	}()

	<-ctx.Done()
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutCtx)
	slog.Info("payment service stopped")
}
