import { useState } from 'react';
import type { AuthData } from './types';
import { LoginPage } from './pages/LoginPage';
import { POSPage } from './pages/POSPage';
import { AdminPage } from './pages/AdminPage';

const TOKEN_KEY       = 'c2k_access_token';
const MERCHANT_ID_KEY = 'c2k_merchant_id';
const BIZ_NAME_KEY    = 'c2k_business_name';

function POSApp() {
  const [token,      setToken]      = useState(() => localStorage.getItem(TOKEN_KEY) ?? '');
  const [merchantId, setMerchantId] = useState(() => localStorage.getItem(MERCHANT_ID_KEY) ?? '');
  const [bizName,    setBizName]    = useState(() => localStorage.getItem(BIZ_NAME_KEY) ?? '');

  const handleLogin = (auth: AuthData) => {
    localStorage.setItem(TOKEN_KEY, auth.access_token);
    setToken(auth.access_token);
    if (auth.merchant_id) {
      localStorage.setItem(MERCHANT_ID_KEY, auth.merchant_id);
      setMerchantId(auth.merchant_id);
    }
    if (auth.business_name) {
      localStorage.setItem(BIZ_NAME_KEY, auth.business_name);
      setBizName(auth.business_name);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(MERCHANT_ID_KEY);
    localStorage.removeItem(BIZ_NAME_KEY);
    setToken('');
    setMerchantId('');
    setBizName('');
  };

  const handleBizNameChange = (name: string) => {
    localStorage.setItem(BIZ_NAME_KEY, name);
    setBizName(name);
  };

  if (!token) return <LoginPage onLogin={handleLogin} />;

  return (
    <POSPage
      posAccessToken={token}
      merchantId={merchantId}
      businessName={bizName}
      onBizNameChange={handleBizNameChange}
      onLogout={handleLogout}
    />
  );
}

export function App() {
  if (window.location.pathname.startsWith('/admin')) return <AdminPage />;
  return <POSApp />;
}
