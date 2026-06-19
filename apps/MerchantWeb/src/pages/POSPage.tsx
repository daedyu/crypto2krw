import { useCallback, useEffect, useRef, useState } from 'react';
import jsQR from 'jsqr';
import { createQR, payWithUserToken, updateBusinessName } from '../api/client';
import type { PayResult } from '../types';

// ── 상수 ──────────────────────────────────────────────────────────────────
const MAX_AMOUNT = 9_999_999;
const PRESETS    = [1000, 5000, 10000, 50000];

type Stage =
  | 'scanning'   // 카메라로 유저 QR 스캔
  | 'amount'     // 금액 입력 (유저 QR 인식 완료)
  | 'paying'     // 결제 처리 중
  | 'success'    // 결제 완료
  | 'error';     // 오류

interface ScannedUser {
  coin: string;        // SOL | ETH | USDT
  accessToken: string; // 유저 JWT
}

interface Props {
  posAccessToken: string;
  merchantId: string;
  businessName: string;
  onBizNameChange: (name: string) => void;
  onLogout: () => void;
}

// ── 금액 포맷 ──────────────────────────────────────────────────────────────
function fmtKRW(v: string | number) {
  return `₩${Number(v).toLocaleString('ko-KR')}`;
}

// ── 카메라 스캔 훅 ─────────────────────────────────────────────────────────
function useCameraScanner(onDetect: (data: string) => void) {
  const videoRef  = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const rafRef    = useRef<number>(0);
  const activeRef = useRef(false);
  const [camError, setCamError] = useState<string | null>(null);

  const stop = useCallback(() => {
    activeRef.current = false;
    cancelAnimationFrame(rafRef.current);
    streamRef.current?.getTracks().forEach(t => t.stop());
    streamRef.current = null;
  }, []);

  const start = useCallback(async () => {
    setCamError(null);
    activeRef.current = true;
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: { ideal: 'environment' }, width: { ideal: 1280 } },
      });
      streamRef.current = stream;
      if (videoRef.current) {
        videoRef.current.srcObject = stream;
        await videoRef.current.play();
      }

      const scan = () => {
        if (!activeRef.current) return;
        const v = videoRef.current;
        const c = canvasRef.current;
        if (!v || !c || v.readyState < v.HAVE_ENOUGH_DATA) {
          rafRef.current = requestAnimationFrame(scan);
          return;
        }
        c.width  = v.videoWidth;
        c.height = v.videoHeight;
        const ctx = c.getContext('2d', { willReadFrequently: true });
        if (!ctx) return;
        ctx.drawImage(v, 0, 0);
        const img  = ctx.getImageData(0, 0, c.width, c.height);
        const code = jsQR(img.data, img.width, img.height, { inversionAttempts: 'dontInvert' });
        if (code?.data) {
          activeRef.current = false;
          onDetect(code.data);
          return;
        }
        rafRef.current = requestAnimationFrame(scan);
      };
      rafRef.current = requestAnimationFrame(scan);

    } catch {
      setCamError('카메라 권한이 필요합니다.\n브라우저 설정 → 카메라 → 허용');
    }
  }, [onDetect]);

  return { videoRef, canvasRef, camError, start, stop };
}

// ── 설정 패널 ──────────────────────────────────────────────────────────────
function SettingsPanel({
  businessName, accessToken, onChange, onClose,
}: {
  businessName: string;
  accessToken: string;
  onChange: (name: string) => void;
  onClose: () => void;
}) {
  const [val, setVal]       = useState(businessName);
  const [saving, setSaving] = useState(false);
  const [err, setErr]       = useState<string | null>(null);

  const save = async () => {
    const trimmed = val.trim();
    if (!trimmed) return;
    setSaving(true);
    setErr(null);
    try {
      await updateBusinessName(trimmed, accessToken);
      onChange(trimmed);
      onClose();
    } catch (e) {
      setErr(e instanceof Error ? e.message : '저장 실패');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="settings-backdrop" onClick={onClose}>
      <div className="settings-sheet" onClick={e => e.stopPropagation()}>
        <div className="settings-title">가맹점 설정</div>
        <div className="field-group">
          <label className="field-label">업체명</label>
          <input
            className="field-input"
            value={val}
            onChange={e => setVal(e.target.value)}
            placeholder="업체명을 입력하세요"
            autoFocus
          />
        </div>
        {err && <div className="settings-error">{err}</div>}
        <div className="settings-actions">
          <button className="settings-save-btn" onClick={save} disabled={saving}>
            {saving ? '저장 중...' : '저장'}
          </button>
          <button className="settings-cancel-btn" onClick={onClose}>취소</button>
        </div>
      </div>
    </div>
  );
}

// ── 메인 POS 컴포넌트 ──────────────────────────────────────────────────────
export function POSPage({ posAccessToken, merchantId, businessName, onBizNameChange, onLogout }: Props) {
  const [stage,    setStage]  = useState<Stage>('scanning');
  const [scanned,  setScanned] = useState<ScannedUser | null>(null);
  const [digits,   setDigits]  = useState('');
  const [result,   setResult]  = useState<PayResult | null>(null);
  const [error,    setError]   = useState<string | null>(null);
  const [paying,   setPaying]  = useState(false);
  const [settings, setSettings] = useState(false);
  const autoReset = useRef<ReturnType<typeof setTimeout> | null>(null);

  // QR 인식 콜백
  // 구형식: crypto2krw://pay?coin=SOL&ts=...  (at 없음 → POS 토큰으로 fallback)
  // 신형식: crypto2krw://pay?coin=SOL&at=JWT
  const handleQRData = useCallback((data: string) => {
    const trimmed = data.trim();
    console.log('[POS] QR raw:', trimmed);

    const coinMatch = trimmed.match(/[?&]coin=([A-Za-z0-9_]+)/i);
    const coin = coinMatch?.[1]?.toUpperCase() ?? '';

    if (!['SOL', 'ETH', 'USDT'].includes(coin)) {
      setError(`코인 정보를 찾을 수 없습니다.\n\n스캔된 QR:\n${trimmed.slice(0, 150)}`);
      setStage('error');
      return;
    }

    // 유저 토큰이 있으면 사용, 없으면 POS 로그인 토큰으로 결제
    const atMatch = trimmed.match(/[?&]at=([^&\s]+)/);
    const accessToken = atMatch?.[1]?.trim() || posAccessToken;

    console.log('[POS] coin:', coin, '| 토큰출처:', atMatch ? '유저QR' : 'POS로그인');

    setScanned({ coin, accessToken });
    setDigits('');
    setStage('amount');
  }, [posAccessToken]);

  const { videoRef, canvasRef, camError, start, stop } = useCameraScanner(handleQRData);

  // scanning 상태 진입 시 카메라 시작
  useEffect(() => {
    if (stage === 'scanning') {
      start();
      return stop;
    }
  }, [stage, start, stop]);

  // 언마운트 시 타이머 정리
  useEffect(() => () => { if (autoReset.current) clearTimeout(autoReset.current); }, []);

  // 키보드 지원 (amount 단계)
  useEffect(() => {
    if (stage !== 'amount') return;
    const handler = (e: KeyboardEvent) => {
      if (e.key >= '0' && e.key <= '9') appendDigit(e.key);
      else if (e.key === 'Backspace') delDigit();
      else if (e.key === 'Enter') handlePay();
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  });

  const appendDigit = (d: string) =>
    setDigits(prev => { const n = prev + d; return Number(n) > MAX_AMOUNT ? prev : n; });

  const delDigit = () => setDigits(prev => prev.slice(0, -1));

  const addPreset = (v: number) =>
    setDigits(prev => {
      const n = (Number(prev || '0') + v).toString();
      return Number(n) > MAX_AMOUNT ? prev : n;
    });

  const reset = useCallback(() => {
    if (autoReset.current) clearTimeout(autoReset.current);
    setStage('scanning');
    setScanned(null);
    setDigits('');
    setResult(null);
    setError(null);
    setPaying(false);
  }, []);

  const handlePay = async () => {
    if (!scanned || !digits || Number(digits) < 100) return;
    setPaying(true);
    setError(null);
    try {
      // 1. POS 토큰으로 QR 세션 생성
      const session = await createQR(merchantId, digits, posAccessToken);
      // 2. 유저 토큰으로 즉시 결제
      const pay = await payWithUserToken(session.token, scanned.coin, scanned.accessToken);
      setResult(pay);
      setStage('success');
      autoReset.current = setTimeout(reset, 5000);
    } catch (e) {
      setError(e instanceof Error ? e.message : '결제 처리 실패');
      setStage('error');
    } finally {
      setPaying(false);
    }
  };

  // ── 렌더 ──────────────────────────────────────────────────────────────────
  const COIN_ICON: Record<string, string> = { SOL: '◎', ETH: 'Ξ', USDT: '$' };

  return (
    <div className="pos-layout">

      {/* 헤더 */}
      <header className="pos-header">
        <div className="pos-header-left">
          <div className="pos-logo">
            Crypto<span className="pos-logo-accent">2KRW</span>
          </div>
          {businessName && (
            <div className="pos-biz-name">{businessName}</div>
          )}
        </div>
        <div className="pos-header-actions">
          <button className="pos-header-btn" onClick={() => setSettings(true)}>설정</button>
          <button className="pos-header-btn logout" onClick={onLogout}>로그아웃</button>
        </div>
      </header>

      {/* ── 스캔 화면 ── */}
      {stage === 'scanning' && (
        <div className="scan-stage">
          {camError ? (
            <div className="scan-error">
              <div className="scan-error-icon">📷</div>
              <div className="scan-error-title">카메라 오류</div>
              <div className="scan-error-desc">{camError}</div>
              <button className="scan-retry-btn" onClick={start}>다시 시도</button>
            </div>
          ) : (
            <>
              <video
                ref={videoRef}
                muted
                playsInline
                className="scan-video"
              />
              <div className="scan-overlay">
                <div className="scan-corners">
                  <div className="scan-corner scan-corner--tl" />
                  <div className="scan-corner scan-corner--tr" />
                  <div className="scan-corner scan-corner--bl" />
                  <div className="scan-corner scan-corner--br" />
                </div>
              </div>
              <div className="scan-hint">
                <div className="scan-hint-pill">고객의 앱 QR 코드를 화면에 맞춰주세요</div>
              </div>
            </>
          )}
          <canvas ref={canvasRef} style={{ display: 'none' }} />
        </div>
      )}

      {/* ── 금액 입력 화면 ── */}
      {stage === 'amount' && scanned && (
        <div className="amount-stage">
          <div className="amount-top">
            {/* 코인 배지 */}
            <div className="coin-badge">
              <span className="coin-badge-icon">{COIN_ICON[scanned.coin]}</span>
              <span className="coin-badge-label">{scanned.coin}</span>
            </div>

            {/* 금액 — 화면의 주인공 */}
            <div className="amount-display">
              <span className="amount-unit">₩</span>
              <span className={`amount-value ${digits ? 'has-value' : 'zero'}`}>
                {digits ? Number(digits).toLocaleString('ko-KR') : '0'}
              </span>
            </div>

            {error && <div className="amount-error">{error}</div>}
          </div>

          <div className="amount-bottom">
            {/* 프리셋 */}
            <div className="preset-row">
              {PRESETS.map(v => (
                <button key={v} className="preset-btn" onClick={() => addPreset(v)}>
                  +{v >= 10000 ? `${v / 10000}만` : `${v / 1000}천`}
                </button>
              ))}
            </div>

            {/* 키패드 */}
            <div className="numpad">
              {['1','2','3','4','5','6','7','8','9'].map(d => (
                <button key={d} className="key-btn" onClick={() => appendDigit(d)}>{d}</button>
              ))}
              <button className="key-btn del" onClick={delDigit}>⌫</button>
              <button className="key-btn" onClick={() => appendDigit('0')}>0</button>
              <button
                className={`key-btn pay${paying ? ' loading' : ''}`}
                disabled={!digits || Number(digits) < 100 || paying}
                onClick={handlePay}
              >
                {paying ? <span className="spinner" /> : '결제'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── 결제 처리 중 ── */}
      {stage === 'paying' && (
        <div className="paying-stage">
          <div className="spinner spinner--large" />
          <div className="paying-label">결제 처리 중</div>
        </div>
      )}

      {/* ── 성공 화면 ── */}
      {stage === 'success' && result && (
        <div className="success-stage">
          <div className="success-check">✓</div>
          <div className="success-amount">{fmtKRW(result.amount_krw)}</div>
          <div className="success-label">결제 완료</div>

          <div className="success-rows">
            <div className="success-row">
              <span className="success-row-key">코인</span>
              <span className="success-row-val">{result.used_amount} {result.used_currency}</span>
            </div>
            <div className="success-row">
              <span className="success-row-key">환율</span>
              <span className="success-row-val">{fmtKRW(result.applied_rate)}</span>
            </div>
            <div className="success-row">
              <span className="success-row-key">잔여</span>
              <span className="success-row-val">{result.remaining_balance} {result.used_currency}</span>
            </div>
            {result.transaction_id && (
              <div className="success-row">
                <span className="success-row-key">TX</span>
                <span className="success-row-val mono">
                  {result.transaction_id.slice(0, 18)}…
                </span>
              </div>
            )}
          </div>

          <button className="success-next-btn" onClick={reset}>다음 결제</button>
          <div className="success-auto-hint">5초 후 자동 초기화</div>
        </div>
      )}

      {/* ── 오류 화면 ── */}
      {stage === 'error' && (
        <div className="error-stage">
          <div className="error-icon">⚠</div>
          <div className="error-message">{error}</div>
          <button className="error-retry-btn" onClick={reset}>다시 스캔</button>
        </div>
      )}

      {settings && (
        <SettingsPanel
          businessName={businessName}
          accessToken={posAccessToken}
          onChange={onBizNameChange}
          onClose={() => setSettings(false)}
        />
      )}

    </div>
  );
}
