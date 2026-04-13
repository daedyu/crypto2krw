# Data Models (Database Schema)

## 1. Users & Merchants
- **User:** `id`, `email`, `password_hash`, `created_at`, `status`
- **Merchant:** `id`, `business_name`, `email`, `password_hash`, `krw_balance` (정산 대기금)

## 2. Wallets & Ledger (Wallet Service)
- **User_Wallets:** `id`, `user_id`, `currency` (SOL, USDT, ETH), `address` (온체인 주소)
- **Offchain_Ledger (매우 중요 - 엄격한 트랜잭션 필요):** - `id`, `user_id`, `currency`, `balance`, `locked_balance`, `updated_at`

## 3. Payments (Payment Service)
- **QR_Sessions:** `id`, `merchant_id`, `amount_krw`, `qr_token` (UUID), `status` (PENDING, SUCCESS, EXPIRED), `expires_at`
- **Transactions:** - `id`, `tx_hash` (내부 고유값), `user_id`, `merchant_id`, `type` (PAYMENT, DEPOSIT, WITHDRAW)
  - `amount_krw` (결제 원화 금액)
  - `used_currency` (사용된 코인)
  - `used_amount` (차감된 코인 수량)
  - `applied_rate` (적용 환율)
  - `status`, `created_at`
