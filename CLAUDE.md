# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 프로젝트 개요

Crypto2KRW — 암호화폐(SOL, USDT, ETH)를 원화(KRW)로 실시간 환산하여 오프라인 가맹점에서 QR 결제할 수 있는 모바일 결제 플랫폼. 온체인 트랜잭션 없이 오프체인 장부(Off-chain Ledger)로 1초 이내 결제 승인을 목표로 한다.

## 동적 컨텍스트 로딩

코드 작성 전 **반드시** `docs/` 디렉토리에서 관련 문서를 읽어 요구사항과 설계를 파악할 것:
- `docs/01_PRD_Crypto2KRW.md` — 제품 요구사항
- `docs/02_ARCHITECTURE.md` — MSA 구조, 서비스 간 통신, QR 결제 플로우
- `docs/03_DATA_MODELS.md` — DB 스키마 (Users, Wallets, Ledger, Payments)
- `docs/04_API_CONTRACTS.md` — RESTful API 계약
- `docs/05_TECH_STACK_AND_GUIDELINES.md` — 코딩 가이드라인
- `docs/06_ADMIN_AND_DB.md` — 백오피스 시스템

각 마이크로서비스 내부의 `README.md`도 작업 전 확인할 것.

## MSA 아키텍처

### Node.js (TypeScript/NestJS) 서비스
- **API Gateway** — 요청 라우팅, Rate Limiting, JWT 검증
- **Auth Service** — 회원가입, KYC, 가맹점 계정 관리
- **Admin Service** — 백오피스 대시보드, CS, 수동 정산

### Go 서비스 (고성능/동시성 핵심 도메인)
- **Wallet & Ledger Service** — 온체인 입금 감지, 오프체인 장부 잔액 관리 (가장 엄격한 DB 트랜잭션 요구)
- **Oracle Service** — 바이낸스/업비트 웹소켓 구독, 환율 계산, Redis 캐싱
- **Payment Service** — QR 결제 승인, Wallet에 장부 차감 요청, 가맹점 매출 업데이트

### 프론트엔드
- **User App** — Flutter (Cross-platform)
- **Merchant Web** — React.js 또는 Vue.js

### 인프라
- **DB:** PostgreSQL (논리적 스키마 분리, 추후 서비스별 DB 분리), Redis (환율 캐싱)
- **서비스 간 통신:** gRPC 또는 Kafka 우선
- **API 응답 포맷:** `{ "success": true/false, "data": {...}, "error": { "code": "...", "message": "..." } }`

## QR 결제 핵심 플로우

1. 가맹점 → Payment Service: KRW 금액으로 QR 토큰 생성
2. 유저 앱 → API Gateway → Payment Service: QR 스캔 결제 요청
3. Payment Service → Oracle Service: 현재 환율 조회
4. Payment Service → Wallet Service: 유저 코인 장부 차감
5. 차감 성공 시 결제 완료 → 가맹점에 WebSocket 이벤트 발송

## Git 커밋 규칙

- **커밋 메시지는 반드시 한국어**로 작성
- **금지어:** `claude`, `ai`, `generated` 등 AI 관련 단어 절대 불포함
- Conventional Commits: `<타입>(<스코프>): <제목>`
- 타입: `feat` / `fix` / `refactor` / `chore` / `docs` / `style` / `test`
- 예시: `feat(wallet): 사용자의 USDT 입금 주소 생성 API 추가`

## 코드 작성 규칙

### 공통
- **숫자 처리 (매우 중요):** 원화/암호화폐 계산 시 부동소수점 금지. Node.js는 `decimal.js`, Go는 `shopspring/decimal` 사용
- 조기 반환(Early Return)으로 들여쓰기 최소화
- 서술형 네이밍 (축약어 금지): `calculateTotalPaymentAmount()`
- 함수 50줄 이상 시 분리 검토 (단일 책임 원칙)
- 주석은 '왜(Why)'만 설명 ('어떻게(How)' 금지)
- Ledger 업데이트 시 반드시 DB Transaction(ACID) 적용

### Node.js / TypeScript
- NestJS 사용 시 객체 지향 및 Decorator 패턴 활용
- `any` 타입 엄격히 금지

### Go
- `gofmt` 스타일 엄수
- 동시성: `goroutine` + `channel` 적절히 사용
- 에러는 반드시 `if err != nil`로 명시적 핸들링, panic 최소화
- 트랜잭션 롤백은 `defer` 활용

### Frontend (React Native / React)
- 1 파일 = 1 컴포넌트
- 비즈니스 로직은 Custom Hook(`use...`)으로 분리
- RN `StyleSheet.create`는 파일 최하단 배치
- Props/State는 `interface` 또는 `type`으로 정의, `any` 금지

## 작업 프로세스

1. `docs/` 읽어 요구사항과 설계 방침 파악
2. 모노레포 내부 검색으로 기존 컴포넌트/유틸과 중복 방지
3. 논리적 단위(API 1개, 컴포넌트 1개)로 나누어 구현 후 즉시 커밋
4. 모든 네트워크 I/O와 DB 트랜잭션에 예외 처리 검증
