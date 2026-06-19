import { useState } from 'react';
import { adminDeposit, adminGetBalances } from '../api/client';
import type { DepositResult, UserBalance } from '../api/client';

const CURRENCIES = [
  { value: 'SOL',       label: 'SOL',  icon: '◎' },
  { value: 'ETH',       label: 'ETH',  icon: 'Ξ' },
  { value: 'USDT_TRC20', label: 'USDT', icon: '$' },
];

const PRESETS = [
  { label: '0.1 SOL',  currency: 'SOL',       amount: '0.1' },
  { label: '1 SOL',    currency: 'SOL',       amount: '1' },
  { label: '10 SOL',   currency: 'SOL',       amount: '10' },
  { label: '0.01 ETH', currency: 'ETH',       amount: '0.01' },
  { label: '0.1 ETH',  currency: 'ETH',       amount: '0.1' },
  { label: '100 USDT', currency: 'USDT_TRC20', amount: '100' },
  { label: '500 USDT', currency: 'USDT_TRC20', amount: '500' },
];

export function AdminPage() {
  const [email,    setEmail]    = useState('');
  const [currency, setCurrency] = useState('SOL');
  const [amount,   setAmount]   = useState('');
  const [loading,  setLoading]  = useState(false);
  const [error,    setError]    = useState<string | null>(null);
  const [result,   setResult]   = useState<DepositResult | null>(null);
  const [balances, setBalances] = useState<UserBalance | null>(null);

  const applyPreset = (p: typeof PRESETS[number]) => {
    setCurrency(p.currency);
    setAmount(p.amount);
    setError(null);
  };

  const handleLookup = async () => {
    if (!email.trim()) return;
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      setBalances(await adminGetBalances(email.trim()));
    } catch (e) {
      setError(e instanceof Error ? e.message : '조회 실패');
      setBalances(null);
    } finally {
      setLoading(false);
    }
  };

  const handleDeposit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim() || !amount) return;
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      const res = await adminDeposit(email.trim(), currency, amount);
      setResult(res);
      // 입금 후 잔액 갱신
      setBalances(await adminGetBalances(email.trim()));
    } catch (e) {
      setError(e instanceof Error ? e.message : '입금 실패');
    } finally {
      setLoading(false);
    }
  };

  const currencyLabel = (c: string) =>
    CURRENCIES.find(x => x.value === c)?.label ?? c;

  return (
    <div className="admin-page">
      <div className="admin-wrap">

        {/* 헤더 */}
        <div className="admin-header">
          <div className="admin-logo">
            Crypto<span className="admin-logo-accent">2KRW</span>
            <span className="admin-logo-tag">Admin</span>
          </div>
          <div className="admin-subtitle">테스트 입금 도구 — 시연용</div>
        </div>

        {/* 입금 폼 */}
        <form className="admin-card" onSubmit={handleDeposit}>
          <div className="admin-card-title">잔액 입금</div>

          {/* 이메일 */}
          <div className="field-group">
            <label className="field-label">유저 이메일</label>
            <div className="admin-email-row">
              <input
                type="email"
                className="field-input"
                placeholder="user@crypto2krw.com"
                value={email}
                onChange={e => { setEmail(e.target.value); setBalances(null); setResult(null); }}
                required
              />
              <button
                type="button"
                className="admin-lookup-btn"
                onClick={handleLookup}
                disabled={!email.trim() || loading}
              >
                조회
              </button>
            </div>
          </div>

          {/* 빠른 프리셋 */}
          <div className="field-group">
            <label className="field-label">빠른 선택</label>
            <div className="admin-presets">
              {PRESETS.map(p => (
                <button
                  key={`${p.currency}-${p.amount}`}
                  type="button"
                  className={`admin-preset-btn ${currency === p.currency && amount === p.amount ? 'active' : ''}`}
                  onClick={() => applyPreset(p)}
                >
                  {p.label}
                </button>
              ))}
            </div>
          </div>

          {/* 통화 + 금액 */}
          <div className="admin-amount-row">
            <div className="field-group" style={{ flex: '0 0 auto' }}>
              <label className="field-label">통화</label>
              <div className="admin-currency-group">
                {CURRENCIES.map(c => (
                  <button
                    key={c.value}
                    type="button"
                    className={`admin-currency-btn ${currency === c.value ? 'active' : ''}`}
                    onClick={() => setCurrency(c.value)}
                  >
                    <span className="admin-currency-icon">{c.icon}</span>
                    {c.label}
                  </button>
                ))}
              </div>
            </div>
            <div className="field-group" style={{ flex: 1 }}>
              <label className="field-label">금액</label>
              <input
                type="number"
                className="field-input"
                placeholder="0.0"
                min="0"
                step="any"
                value={amount}
                onChange={e => setAmount(e.target.value)}
                required
              />
            </div>
          </div>

          {error && <div className="admin-error">{error}</div>}

          {/* 성공 결과 */}
          {result && (
            <div className="admin-result">
              <span className="admin-result-icon">✓</span>
              <span>
                <strong>{result.deposited} {currencyLabel(result.currency)}</strong> 입금 완료
                &nbsp;→&nbsp;잔액 <strong>{parseFloat(result.new_balance).toFixed(6)} {currencyLabel(result.currency)}</strong>
              </span>
            </div>
          )}

          <button
            type="submit"
            className="admin-submit-btn"
            disabled={!email.trim() || !amount || loading}
          >
            {loading ? '처리 중...' : `${amount || '0'} ${currencyLabel(currency)} 입금`}
          </button>
        </form>

        {/* 잔액 현황 */}
        {balances && (
          <div className="admin-card">
            <div className="admin-card-title">
              현재 잔액
              <span className="admin-card-subtitle">{balances.email}</span>
            </div>
            {balances.balances.length === 0 ? (
              <div className="admin-empty">잔액 없음</div>
            ) : (
              <div className="admin-balances">
                {balances.balances.map(b => {
                  const c = CURRENCIES.find(x => x.value === b.currency);
                  return (
                    <div key={b.currency} className="admin-balance-row">
                      <span className="admin-balance-icon">{c?.icon ?? '?'}</span>
                      <span className="admin-balance-label">{c?.label ?? b.currency}</span>
                      <span className="admin-balance-value">
                        {parseFloat(b.balance).toLocaleString('ko-KR', { maximumFractionDigits: 6 })}
                      </span>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        )}

        <div className="admin-back">
          <a href="/" className="admin-back-link">← POS 터미널로</a>
        </div>

      </div>
    </div>
  );
}
