# 주간 개발 보고서

> 작성일: 2026년 4월 13일

---

## 1. 계획한 작업

**프론트엔드 제작** — Crypto2KRW 사용자 앱 (iOS) UI 구현

---

## 2. 일정 준수 여부

✅ **계획대로 진행**

| 항목 | 계획 | 실제 |
|------|------|------|
| 화면 설계 및 네비게이션 구성 | 1주차 | 1주차 완료 |
| 주요 화면 구현 (로그인·홈·QR·설정·입금) | 1주차 | 1주차 완료 |
| 디자인 시스템 정립 및 전체 적용 | 1주차 | 1주차 완료 |

---

## 3. 주간 작업 내용

### 3-1. 기존 계획 (프로젝트계획서 기준)

> ※ 아래 내용은 발표자료의 원문을 그대로 옮길 것

- 사용자 앱 프론트엔드를 React Native 기반으로 구현
- 로그인 / 회원가입 / 자산 대시보드 / QR 결제 / 입금 / 설정 화면 구현
- 서버 미연동 상태의 목업(Mock) 데이터 기반 UI 프로토타입

---

### 3-2. 실제 작업 내용

#### ① 기술 스택 및 프로젝트 초기 설정

| 항목 | 선택 |
|------|------|
| 프레임워크 | Expo SDK 54 + React Native 0.81.5 |
| 언어 | TypeScript |
| 라우팅 | Expo Router v6 (파일 기반 라우팅) |
| 스타일링 | `StyleSheet.create` (NativeWind 미사용) |
| 주요 라이브러리 | `react-native-qrcode-svg`, `expo-clipboard`, `@expo/vector-icons`, `react-native-safe-area-context` |

`package.json` main 필드를 `expo-router/entry`로 설정하고, `app.json`에 `"scheme": "crypto2krw"` 및 Expo Router 플러그인을 추가해 파일 기반 라우팅을 활성화했다.

---

#### ② 네비게이션 아키텍처 설계

**최초 설계 (React Navigation 기반)**
```
Stack Navigator
 ├── AuthStack (Login, Register)
 └── HomeStack (Home, QR, History, Settings, Deposit)
```

**1차 변경 — Expo Router 마이그레이션**

React Navigation을 제거하고 Expo Router의 파일 기반 구조로 전환했다. 기존 `src/navigation/` 하위 파일을 전부 삭제하고 `app/` 디렉토리 구조로 재편했다.

```
app/
 ├── index.tsx            ← 인증 상태에 따라 분기 (Redirect)
 ├── _layout.tsx          ← Root Stack (AuthProvider 포함)
 ├── (auth)/
 │   ├── login.tsx
 │   └── register.tsx
 ├── (tabs)/
 │   ├── _layout.tsx      ← Stack (탭 바 없음, 커스텀 바텀 바 사용)
 │   ├── index.tsx        ← 홈 (자산 대시보드 + 내역 토글)
 │   ├── qr.tsx           ← QR 결제 (fullScreenModal)
 │   └── settings.tsx     ← 설정
 └── deposit/
     ├── index.tsx
     └── [currency].tsx
```

**2차 변경 — 탭 바 제거 및 커스텀 바텀 바 도입**

iOS 기본 탭 바 대신, 홈 화면 내부에 커스텀 3버튼 하단 바를 직접 구현했다. `NativeTabs` → `Stack` 전환으로 UITabBarController를 완전히 제거하고, 하단 바는 `position: absolute` 없이 `SafeAreaInsets`를 활용해 안전 영역 위에 고정했다.

```
[+ 잔액추가]  [▦ 결제 (보라색 pill)]  [⚙ 설정]
```

---

#### ③ AI 활용 내역

**사용 도구:** Claude Code (claude-sonnet-4-6)

**주요 프롬프트 및 결과**

| 프롬프트 요약 | 결과 |
|--------------|------|
| "Expo Router로 마이그레이션해줘" | `app/` 구조 전환, `index.tsx` Redirect 패턴 적용 |
| "Liquid Glass 탭 바 적용" | `NativeTabs` 연구 → 결국 커스텀 바텀 바로 교체 |
| "이미지 디자인대로 전체 재설계" | 이미지 기반 디자인 시스템 도출, 전 화면 재작성 |
| "로그인 제외 모든 화면 디자인 재작성" | 5개 화면 동시 리팩터 |

**프로젝트 반영 결과**

AI가 제안한 코드를 그대로 반영하지 않고, 다음 기준으로 직접 검토 후 수정했다:
- `any` 타입 사용 여부 검사
- SafeAreaInsets 처리 누락 여부
- 불필요한 컴포넌트 분리 제거 (1회성 UI는 인라인 유지)

---

#### ④ 구현 화면 목록

| 화면 | 파일 | 주요 요소 |
|------|------|-----------|
| 로그인 | `app/(auth)/login.tsx` | 이메일·비밀번호 입력, 보라색 CTA |
| 회원가입 | `app/(auth)/register.tsx` | 비밀번호 확인 유효성, 눈 토글, 지갑 발급 안내 |
| 홈 (자산) | `app/(tabs)/index.tsx` | 코인별 컬러 카드, 총 자산, 커스텀 바텀 바 |
| 홈 (내역) | 동일 파일, 상태 토글 | 타입별 색상 구분 트랜잭션 리스트 |
| QR 결제 | `src/screens/QRPayScreen.tsx` | 코인 선택 → QR 생성 → 30초 자동 갱신 |
| 설정 | `src/screens/SettingsScreen.tsx` | 주소 아코디언, 복사 버튼, 로그아웃 |
| 입금 선택 | `app/deposit/index.tsx` | 3개 코인 카드 |
| 입금 상세 | `src/screens/DepositDetailScreen.tsx` | 네트워크 탭, QR, 주소 복사, 입금 내역 |

---

#### ⑤ 디자인 시스템

참고 이미지(AiOS 앱 디자인)에서 추출한 원칙:

| 토큰 | 값 |
|------|----|
| 배경 | `#FFFFFF` |
| 서피스 | `#F2F2F7` |
| 텍스트 Primary | `#1C1C1E` |
| 텍스트 Secondary | `#8E8E93` |
| Primary (보라) | `#5A4FCF` |
| USDT | `#26A17B` / `#D4EDE5` |
| SOL | `#9945FF` / `#E4D9FF` |
| ETH | `#627EEA` / `#D4DEF5` |
| 경고 | `#FF9500` / `#FFF8E6` |
| 위험 | `#FF3B30` |
| 카드 radius | 20–24px |
| 타이틀 | 34px, weight 800 |

---

## 4. 이슈

**`NativeTabs` Liquid Glass 미적용 문제**

Expo Router의 탭 바에 iOS 26 Liquid Glass 효과를 적용하려 했으나, 시뮬레이터에서 일반 탭 바로만 표시됨.

---

## 5. 원인

`expo-router`의 `<Tabs>` 컴포넌트는 내부적으로 React Native의 JS 레이어에서 탭 바를 렌더링한다. Liquid Glass는 iOS 26에서 네이티브 `UITabBarController`에만 자동 적용되는 시스템 효과이므로, JS 렌더링 탭 바에는 적용되지 않는다.

`expo-router/unstable-native-tabs`의 `NativeTabs`를 적용해 네이티브 탭 바로 전환했으나, 개발 빌드(Expo Go)에서는 동작을 완전히 확인할 수 없었고, iOS 26 실기기 또는 시뮬레이터 상에서 최종 확인이 필요하다.

---

## 6. 해결책

`NativeTabs` 적용을 시도했으나, 실 기기 검증 전까지 불확실성을 줄이기 위해 **탭 바 자체를 제거하는 방향으로 설계를 전환**했다.

- `(tabs)/_layout.tsx`를 `NativeTabs` → `Stack`으로 교체
- 홈 화면 하단에 커스텀 3버튼 바 (`잔액추가 / 결제 / 설정`) 직접 구현
- Liquid Glass 의존성을 제거함으로써 OS 버전 종속 없이 일관된 UX 확보

---

## 7. 배운 점 및 느낀 점

- **Expo Router의 파일 기반 라우팅**은 초기 설정 비용이 있지만, 화면이 늘어날수록 React Navigation 대비 구조 파악이 직관적이었다.
- **OS 시스템 UI에 의존하는 효과**(Liquid Glass 등)는 개발 초기에 실기기 검증 계획을 미리 세워두지 않으면 방향을 잃기 쉽다. 의존성이 불확실할 때는 커스텀 구현으로 fallback하는 결정이 더 빠른 전진을 가능하게 했다.
- AI 도구는 반복적인 보일러플레이트 작성에서 시간을 크게 줄여줬지만, 타입 안전성이나 SafeArea 처리 같은 세부 사항은 직접 검토가 필수임을 확인했다.
- 디자인을 코드로 옮길 때 레퍼런스 이미지에서 **색상·타이포그래피·간격 토큰을 먼저 추출**한 뒤 일관되게 적용하면, 중간에 스타일이 혼재되는 문제를 막을 수 있었다.

---

## 8. 다음 계획

### 기존 계획 (프로젝트계획서 기준)

> ※ 발표자료의 원문을 그대로 옮길 것

- 백엔드 API 연동 (Auth Service, Wallet Service 등)
- 실시간 환율 연동 (Oracle Service WebSocket)
- QR 결제 플로우 End-to-End 구현

---

### 수정 계획

| 항목 | 기존 계획 | 수정 계획 | 사유 |
|------|-----------|-----------|------|
| API 연동 시작 | 2주차 | 2주차 유지 | 프론트 목업 완성으로 일정 여유 확보 |
| QR 결제 플로우 | 3주차 | 2주차 병행 | UI가 완성되어 연동 작업 즉시 착수 가능 |
| 실기기 디자인 검증 | 미계획 | 2주차 추가 | Liquid Glass 등 OS 의존 UI 최종 확인 필요 |

**2주차 구체 작업 목록**

- [ ] Auth Service REST API 연동 (로그인 · 회원가입)
- [ ] Oracle Service WebSocket 연결 → 실시간 환율 홈 화면 반영
- [ ] Wallet Service API 연동 → 실제 잔액 조회
- [ ] Payment Service QR 토큰 발급 API 연동
- [ ] iOS 실기기 또는 TestFlight 빌드 후 UI 검증

---

## 📸 필요한 이미지 목록

보고서 제출 전 아래 스크린샷을 직접 캡처해서 삽입할 것:

| 번호 | 이미지 | 캡처 방법 |
|------|--------|-----------|
| 1 | 로그인 화면 | 시뮬레이터 캡처 (`⌘ + S`) |
| 2 | 회원가입 화면 | 동일 |
| 3 | 홈 — 자산 뷰 | 동일 |
| 4 | 홈 — 결제 내역 뷰 (헤더 토글 후) | 동일 |
| 5 | QR 결제 — 코인 선택 화면 | 동일 |
| 6 | QR 결제 — QR 표시 화면 | 동일 |
| 7 | 입금 선택 화면 | 동일 |
| 8 | 입금 상세 화면 (네트워크 탭 + QR) | 동일 |
| 9 | 설정 화면 (주소 아코디언 열린 상태) | 동일 |
| 10 | 전체 파일 구조 | VS Code 또는 `tree app/ src/` 터미널 |
| 11 | 디자인 레퍼런스 이미지 | 이미 보유 (AiOS 스크린샷) |
