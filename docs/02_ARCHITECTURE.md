# System Architecture & Tech Stack

## 1. Tech Stack
- **Mobile (User):** Flutter (Cross-platform)
- **Web (Merchant):** React.js 또는 Vue.js
- **Backend Services:** NestJS (Node.js), Go 또는 Spring Boot (Kotlin)(user 정보같은 채널계 서비스들만 spring 사용)
- **Database:** PostgreSQL (Relational DB for Ledger/Users), Redis (Caching/Rates)
- **Infrastructure:** AWS (EKS/ECS)(추후적용), Docker, API Gateway

## 2. Microservices Architecture (MSA)
1. **API Gateway:** 클라이언트 요청 라우팅, 통합 인증 처리.
2. **Auth Service:** 사용자 가입, 로그인, JWT 토큰 발행, 가맹점 계정 관리.
3. **Wallet Service:** 사용자별 암호화폐 지갑 주소 발급, 외부 온체인 입금(Deposit) 감지 데몬, 오프체인 장부(Ledger) 잔액 관리.
4. **Oracle Service:** 외부 거래소 API(Binance, Upbit) 실시간 구독, 내부 기준 KRW 환율 계산 및 Redis 캐싱.
5. **Payment Service:** QR 코드 생성, QR 스캔 처리, 결제 승인 로직(환율 확인 -> 지갑 차감 -> 가맹점 매출 추가)의 트랜잭션 관리.

## 2. Microservices Architecture & Language Allocation

### 🟢 Node.js (TypeScript) 영역 (빠른 개발 & 유연성)
1. **API Gateway:** 클라이언트 요청 라우팅, Rate Limiting, 통합 인증(JWT 검증).
2. **Auth Service:** 사용자 회원가입, KYC 상태 관리, 가맹점 계정 생성 및 권한 관리.
3. **Admin (Backoffice) Service:** 내부 운영진을 위한 통합 DB 관리, 유저 CS 처리, 수동 정산 및 대시보드 API 제공.

### 🔵 Go (Golang) 영역 (고성능 & 동시성 & 안전성)
4. **Wallet & Ledger Service:** 온체인 입금 감지(블록체인 노드 통신), 오프체인 장부(Ledger) 잔액 증감. **(DB 트랜잭션 성능과 타입 안정성이 가장 중요한 핵심 서비스)**
5. **Oracle Service:** 바이낸스/업비트 등 외부 API 웹소켓 다중 구독, 내부 기준 환율 계산 후 Redis에 밀리초 단위로 업데이트.
6. **Payment Service:** QR 결제 승인 로직 처리. Wallet Service에 장부 차감을 요청하고 가맹점 매출 테이블을 업데이트하는 매우 빠른 트랜잭션 제어.

## 3. Database Strategy
- **DB Migration:** 서비스별 분리 원칙(Database per Service)을 지향하되, 초기 MVP 단계에서는 하나의 PostgreSQL을 논리적 스키마(Schema)로 분리하여 사용.
- **Migration Tool:** 데이터베이스 마이그레이션(DDL 변경)은 가급적 하나로 통일. (예: Go의 `golang-migrate` 또는 Node의 `Prisma Migrate` 중 하나를 주력으로 선택하여 스키마 버전 관리)

## 3. QR Payment Flow (User Scans Merchant QR)
1. 가맹점이 [Payment Service]를 통해 10,000 KRW 짜리 결제 QR(Token 포함) 생성.
2. 유저가 앱에서 QR 스캔 후 결제 요청 (UserApp -> API Gateway -> Payment Service).
3. [Payment Service]는 [Oracle Service]에서 현재 환율 조회.
4. [Payment Service]는 [Wallet Service]에 유저의 1순위 코인 차감 요청 (예: 10,000원에 해당하는 USDT 차감).
5. [Wallet Service]의 장부(DB)에서 USDT 차감 성공 시, [Payment Service]는 결제 완료 처리.
6. [Payment Service]는 가맹점 앱에 결제 성공 웹소켓 이벤트 발송.
