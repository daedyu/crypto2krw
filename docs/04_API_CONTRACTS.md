# API Contracts (RESTful)

## 1. Wallet Service
- `GET /api/v1/wallets/balance` : 유저의 코인별 잔액 및 KRW 환산 총액 조회
- `GET /api/v1/wallets/deposit-address` : 유저의 입금용 온체인 주소 조회
- `PUT /api/v1/wallets/priority` : 결제 코인 우선순위 변경

## 2. Oracle Service
- `GET /api/v1/rates` : 현재 서비스 내부 적용 환율 리스트 (SOL/KRW, USDT/KRW 등)

## 3. Payment Service
- `POST /api/v1/payments/qr/create` : (Merchant) 결제용 QR 토큰 생성
  - Body: `{ amount_krw: 10000 }`
  - Response: `{ qr_token: "uuid...", expires_in: 300 }`
- `POST /api/v1/payments/qr/pay` : (User) 스캔한 QR 토큰으로 결제 실행
  - Body: `{ qr_token: "uuid..." }`
  - Response: `{ status: "SUCCESS", tx_id: "...", amount_krw: 10000, deducted_coin: "USDT", deducted_amount: 7.2 }`
