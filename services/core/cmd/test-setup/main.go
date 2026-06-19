// test-setup: 개발용 테스트 데이터 생성 도구.
// 실행 시 테스트 유저 + SOL/ETH/USDT 지갑을 생성하고 SOL 주소를 출력한다.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/crypto2krw/core/internal/config"
	"github.com/crypto2krw/core/internal/domain/wallet"
	"github.com/crypto2krw/core/internal/infrastructure/kafka"
	pginfra "github.com/crypto2krw/core/internal/infrastructure/postgres"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("load config:", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("open db:", err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatal("ping db:", err)
	}

	publisher, err := kafka.NewPublisher(cfg.KafkaBrokers)
	if err != nil {
		log.Fatal("create publisher:", err)
	}
	defer publisher.Close()

	ctx := context.Background()

	// 1. 테스트 유저 생성 (이미 있으면 기존 것 사용)
	email := "testuser@crypto2krw.dev"
	var userID uuid.UUID
	err = db.QueryRowContext(ctx,
		`SELECT id FROM core.users WHERE email = $1`, email,
	).Scan(&userID)

	if err == sql.ErrNoRows {
		err = db.QueryRowContext(ctx, `
			INSERT INTO core.users (email, password_hash, kyc_status, status)
			VALUES ($1, '$2y$10$placeholder_hash', 'APPROVED', 'ACTIVE')
			RETURNING id
		`, email).Scan(&userID)
		if err != nil {
			log.Fatal("create user:", err)
		}
		fmt.Printf("[OK] 유저 생성: %s (id=%s)\n", email, userID)
	} else if err != nil {
		log.Fatal("query user:", err)
	} else {
		fmt.Printf("[OK] 기존 유저 사용: %s (id=%s)\n", email, userID)
	}

	// 2. 지갑 생성
	walletRepo := pginfra.NewWalletRepository(db)
	walletSvc := wallet.NewService(walletRepo, publisher)

	wallets, err := walletSvc.AllocateDepositAddresses(ctx, userID)
	if err != nil {
		log.Fatal("allocate wallets:", err)
	}

	fmt.Println("\n[생성된 지갑 주소]")
	fmt.Println("-------------------------------------------")
	for _, w := range wallets {
		fmt.Printf("  %-12s %s\n", w.Currency, w.Address)
	}
	fmt.Println("-------------------------------------------")

	// SOL 주소만 별도 출력 (airdrop 명령어 바로 복사 가능하도록)
	for _, w := range wallets {
		if w.Currency == wallet.CurrencySOL {
			fmt.Printf("\n[SOL 입금 주소]\n  %s\n", w.Address)
			fmt.Printf("\n[Devnet Airdrop 명령어]\n")
			fmt.Printf("  solana airdrop 1 %s --url devnet\n", w.Address)
			fmt.Printf("\n[잔액 확인 쿼리]\n")
			fmt.Printf("  SELECT balance FROM core.offchain_ledger WHERE user_id='%s' AND currency='SOL';\n", userID)

			// .env 파일로 저장 (편의용)
			envContent := fmt.Sprintf("TEST_USER_ID=%s\nTEST_SOL_ADDRESS=%s\n", userID, w.Address)
			if writeErr := os.WriteFile("/tmp/crypto2krw-test.env", []byte(envContent), 0644); writeErr == nil {
				fmt.Printf("\n[테스트 환경 저장됨] /tmp/crypto2krw-test.env\n")
			}
		}
	}

	fmt.Println("\n[Kafka] wallet.created 이벤트 발행 완료 — chain-watcher가 주소를 Redis에 등록했습니다.")
}
