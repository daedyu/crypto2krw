# Internal Admin & DB Management (Backoffice)

## 1. Overview
Crypto2KRW 서비스의 원활한 운영, 회계 감사, CS 대응, 수동 정산을 위해 내부 운영진이 접근하는 백오피스(Admin) 시스템을 정의한다. 이 서비스는 Node.js 기반으로 구축되며, 민감한 내부 DB 테이블을 안전하게 제어한다.

## 2. Core Features (Admin Dashboard)
### 2.1 User & Wallet Management
- 유저 목록 조회 및 상세 정보 확인 (KYC 상태, 발급된 가상 지갑 주소 등).
- 유저별 오프체인 장부(Ledger) 잔액 및 거래(Transaction) 내역 조회.
- **[위험 관리]** 이상 거래 탐지 시 특정 유저의 지갑 동결(Lock) 기능.

### 2.2 Merchant & Settlement (정산 관리)
- 가맹점별 누적 매출 조회.
- 가맹점 원화(KRW) 정산 처리 (초기에는 수동 이체 후 상태 값 변경, 추후 펌뱅킹 연동).

### 2.3 System & Ledger Audit (장부 무결성 검증)
- 전체 유저의 오프체인 장부 총합과, 회사 법인 소유의 온체인 지갑 총합 간의 차이(Balance Check) 모니터링 대시보드.
- 오라클(Oracle) 서버가 적용한 시간대별 환율 스프레드 내역 조회.

## 3. Security Guidelines for Admin
- Admin API는 외부망에 노출되지 않도록 VPN 내부망 처리 또는 IP Whitelist를 적용.
- Admin 시스템의 모든 데이터 CUD(생성, 수정, 삭제) 액션은 `Admin_Audit_Logs` 테이블에 기록되어야 함 (누가, 언제, 누구의 데이터를, 어떻게 변경했는지 기록).
