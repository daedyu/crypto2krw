import { useState } from 'react';
import { login, registerMerchant } from '../api/client';
import type { AuthData } from '../types';

interface Props {
  onLogin: (auth: AuthData) => void;
}

type Mode = 'login' | 'register';

export function LoginPage({ onLogin }: Props) {
  const [mode, setMode]           = useState<Mode>('login');
  const [email, setEmail]         = useState('');
  const [password, setPassword]   = useState('');
  const [bizName, setBizName]     = useState('');
  const [loading, setLoading]     = useState(false);
  const [error, setError]         = useState<string | null>(null);

  const switchMode = (next: Mode) => {
    setMode(next);
    setError(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      if (mode === 'login') {
        onLogin(await login(email, password));
      } else {
        onLogin(await registerMerchant(email, password, bizName));
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '오류가 발생했습니다.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-page">
      <div className="login-form-wrap">

        <div className="login-brand">
          <div className="login-logo">
            Crypto<span className="login-logo-accent">2KRW</span>
          </div>
          <div className="login-tagline">가맹점 POS</div>
        </div>

        {/* 탭 */}
        <div className="login-tabs">
          <button
            type="button"
            className={`login-tab ${mode === 'login' ? 'active' : ''}`}
            onClick={() => switchMode('login')}
          >
            로그인
          </button>
          <button
            type="button"
            className={`login-tab ${mode === 'register' ? 'active' : ''}`}
            onClick={() => switchMode('register')}
          >
            가맹점 등록
          </button>
        </div>

        <form onSubmit={handleSubmit} className="login-fields">
          {mode === 'register' && (
            <div className="field-group">
              <label className="field-label">업체명</label>
              <input
                type="text"
                className="field-input"
                placeholder="스타벅스 강남점"
                value={bizName}
                onChange={e => setBizName(e.target.value)}
                required
              />
            </div>
          )}

          <div className="field-group">
            <label className="field-label">이메일</label>
            <input
              type="email"
              className="field-input"
              placeholder="email@example.com"
              value={email}
              onChange={e => setEmail(e.target.value)}
              autoComplete="email"
              required
            />
          </div>

          <div className="field-group">
            <label className="field-label">비밀번호</label>
            <input
              type="password"
              className="field-input"
              placeholder="••••••••"
              value={password}
              onChange={e => setPassword(e.target.value)}
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
              required
            />
          </div>

          {error && <div className="login-error">{error}</div>}

          <button type="submit" className="login-submit" disabled={loading}>
            {loading ? '처리 중...' : mode === 'login' ? '로그인' : '가맹점 등록'}
          </button>
        </form>

      </div>
    </div>
  );
}
