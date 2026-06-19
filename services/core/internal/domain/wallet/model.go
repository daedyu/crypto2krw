package wallet

import (
	"time"

	"github.com/google/uuid"
)

type Currency string

const (
	CurrencySOL       Currency = "SOL"
	CurrencyUSDTERC20 Currency = "USDT_ERC20"
	CurrencyUSDTTRC20 Currency = "USDT_TRC20"
	CurrencyETH       Currency = "ETH"
)

var AllCurrencies = []Currency{
	CurrencySOL,
	CurrencyUSDTERC20,
	CurrencyUSDTTRC20,
	CurrencyETH,
}

type UserWallet struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	Currency        Currency
	Address         string
	PrivateKeyHex   string // 개발 전용; 프로덕션은 HSM
	PaymentPriority int
	CreatedAt       time.Time
}
