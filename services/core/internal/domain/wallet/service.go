package wallet

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mr-tron/base58"
)

var (
	ErrWalletNotFound   = errors.New("wallet not found")
	ErrAddressNotFound  = errors.New("address not found")
)

// EventPublisher는 지갑 생성 이벤트를 발행하는 인터페이스이다.
type EventPublisher interface {
	PublishWalletCreated(ctx context.Context, w *UserWallet) error
}

type Service struct {
	repo      Repository
	publisher EventPublisher
}

func NewService(repo Repository, publisher EventPublisher) *Service {
	return &Service{repo: repo, publisher: publisher}
}

// AllocateDepositAddresses creates wallet entries for all supported currencies.
// SOL: 실제 ed25519 키페어로 Solana 주소 생성. 기타 체인은 추후 SDK 연동 시 교체.
func (s *Service) AllocateDepositAddresses(ctx context.Context, userID uuid.UUID) ([]*UserWallet, error) {
	wallets := make([]*UserWallet, 0, len(AllCurrencies))

	for i, currency := range AllCurrencies {
		existing, err := s.repo.FindByUserAndCurrency(ctx, userID, currency)
		if err != nil && !errors.Is(err, ErrWalletNotFound) {
			return nil, fmt.Errorf("check existing wallet for %s: %w", currency, err)
		}
		if existing != nil {
			wallets = append(wallets, existing)
			continue
		}

		address, privateKeyHex, err := generateAddress(currency)
		if err != nil {
			return nil, fmt.Errorf("generate address for %s: %w", currency, err)
		}

		w := &UserWallet{
			UserID:          userID,
			Currency:        currency,
			Address:         address,
			PrivateKeyHex:   privateKeyHex,
			PaymentPriority: i + 1,
		}

		created, err := s.repo.Create(ctx, w)
		if err != nil {
			return nil, fmt.Errorf("create wallet for %s: %w", currency, err)
		}

		// chain-watcher가 감시 주소를 등록할 수 있도록 이벤트 발행.
		// 실패해도 지갑 생성 자체는 롤백하지 않는다 (Kafka 재시도로 복구).
		if err := s.publisher.PublishWalletCreated(ctx, created); err != nil {
			return nil, fmt.Errorf("publish wallet.created for %s: %w", currency, err)
		}

		wallets = append(wallets, created)
	}

	return wallets, nil
}

func (s *Service) GetUserWallets(ctx context.Context, userID uuid.UUID) ([]*UserWallet, error) {
	return s.repo.FindAllByUser(ctx, userID)
}

func (s *Service) SetPaymentPriority(ctx context.Context, userID uuid.UUID, currencies []Currency) error {
	if err := s.repo.UpdatePaymentPriorities(ctx, userID, currencies); err != nil {
		return fmt.Errorf("update payment priorities: %w", err)
	}
	return nil
}

func (s *Service) ResolveAddressToUser(ctx context.Context, address string, currency Currency) (*UserWallet, error) {
	return s.repo.FindByAddress(ctx, address, currency)
}

// generateAddress는 체인별 실제 또는 임시 주소와 개인키를 생성한다.
// SOL: ed25519 키페어에서 Solana 주소(base58 공개키) 생성.
// 기타 체인은 추후 각 SDK로 교체 예정.
func generateAddress(currency Currency) (address, privateKeyHex string, err error) {
	switch currency {
	case CurrencySOL:
		return generateSolanaAddress()
	default:
		addr, genErr := generatePlaceholderAddress(currency)
		return addr, "", genErr
	}
}

func generateSolanaAddress() (address, privateKeyHex string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("generate ed25519 key: %w", err)
	}
	// Solana 주소 = ed25519 공개키 32바이트의 base58 인코딩
	address = base58.Encode(pub)
	// 개인키 64바이트(seed+pub) 전체를 hex로 저장
	privateKeyHex = hex.EncodeToString(priv)
	return address, privateKeyHex, nil
}

func generatePlaceholderAddress(currency Currency) (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	switch currency {
	case CurrencyETH, CurrencyUSDTERC20:
		// Ethereum 주소: 0x + 40자리 hex (20바이트, hex.EncodeToString으로 제로 패딩 보장)
		return "0x" + hex.EncodeToString(b), nil
	case CurrencyUSDTTRC20:
		return generateTronAddress(b)
	default:
		return fmt.Sprintf("%x", b), nil
	}
}

// generateTronAddress는 20바이트 랜덤 주소 바이트로 유효한 Tron base58check 주소를 생성한다.
// Tron 주소 형식: Base58Check(0x41 || address_bytes)
// 체크섬: SHA256(SHA256(21바이트))의 앞 4바이트
func generateTronAddress(addrBytes []byte) (string, error) {
	// 21바이트: [0x41(mainnet prefix)] + [20 address bytes]
	payload := make([]byte, 21)
	payload[0] = 0x41
	copy(payload[1:], addrBytes)

	// Double SHA256 checksum
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	checksum := second[:4]

	// 25바이트: payload + checksum
	full := append(payload, checksum...)
	return base58.Encode(full), nil
}
