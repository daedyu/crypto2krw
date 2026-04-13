# Coding Guidelines & Instructions for AI (Claude)

## 1. Core Principles
- **Safety First:** 금융/자산을 다루는 시스템이므로 Ledger(장부) 업데이트 시 반드시 Database Transaction(ACID) 또는 분산 트랜잭션(Saga pattern 등)을 적용할 것.
- **Floating Point Math:** 암호화폐 수량과 원화 계산 시 부동소수점 오차를 방지하기 위해 반드시 `Decimal` 타입이나 정밀도(Precision)를 보장하는 라이브러리(예: `bignumber.js`, `decimal.js`)를 사용할 것.

## 2. Backend (MSA) Rules
- 각 도메인(Auth, Wallet, Payment)은 독립적으로 동작 가능해야 하며, 서비스 간 통신은 gRPC 또는 Message Queue(Kafka)를 우선적으로 고려해야함.
- 모든 API 응답은 일관된 포맷을 유지할 것.
  - `{ "success": true/false, "data": {...}, "error": { "code": "...", "message": "..." } }`

## 3. Mobile (React native) Rules
- QR 결제 시나리오에서 네트워크 지연이 발생할 수 있으므로, 명확한 Loading Indicator와 Timeout 에러 핸들링을 구현할 것.
