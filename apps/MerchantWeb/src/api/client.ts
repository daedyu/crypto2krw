import type { AuthData, CreateQRData, PayResult, QRSession } from '../types';

interface ApiResponse<T> {
  success: boolean;
  data: T | null;
  error?: { code?: string; message: string };
}

async function apiFetch<T>(path: string, options: RequestInit): Promise<T> {
  const res = await fetch('/api/v1' + path, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...(options.headers as Record<string, string>) },
  });
  const body: ApiResponse<T> = await res.json();
  if (!body.success || body.data == null) {
    throw new Error(body.error?.message ?? '알 수 없는 오류');
  }
  return body.data;
}

export async function login(email: string, password: string): Promise<AuthData> {
  return apiFetch<AuthData>('/auth/merchant/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
}

export async function registerMerchant(
  email: string,
  password: string,
  businessName: string,
): Promise<AuthData> {
  return apiFetch<AuthData>('/auth/merchant/register', {
    method: 'POST',
    body: JSON.stringify({ email, password, business_name: businessName }),
  });
}

export async function updateBusinessName(
  businessName: string,
  accessToken: string,
): Promise<{ business_name: string }> {
  return apiFetch<{ business_name: string }>('/merchant/me', {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify({ business_name: businessName }),
  });
}

export interface DepositResult {
  user_id: string;
  email: string;
  currency: string;
  deposited: string;
  new_balance: string;
}

export interface UserBalance {
  user_id: string;
  email: string;
  balances: { currency: string; balance: string }[];
}

export async function adminDeposit(
  email: string,
  currency: string,
  amount: string,
): Promise<DepositResult> {
  return apiFetch<DepositResult>('/admin/deposit', {
    method: 'POST',
    body: JSON.stringify({ email, currency, amount }),
  });
}

export async function adminGetBalances(email: string): Promise<UserBalance> {
  return apiFetch<UserBalance>(`/admin/balances?email=${encodeURIComponent(email)}`, {
    method: 'GET',
  });
}

export async function createQR(
  merchantId: string,
  amountKrw: string,
  accessToken: string,
): Promise<CreateQRData> {
  return apiFetch<CreateQRData>('/payment/qr', {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify({ merchant_id: merchantId, amount_krw: amountKrw }),
  });
}

export async function getQRStatus(token: string, accessToken: string): Promise<QRSession> {
  return apiFetch<QRSession>(`/payment/qr/${token}`, {
    method: 'GET',
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

// POS가 유저 QR을 스캔한 뒤 유저의 토큰으로 결제 처리
export async function payWithUserToken(
  sessionToken: string,
  currency: string,
  userAccessToken: string,
): Promise<PayResult> {
  return apiFetch<PayResult>(`/payment/qr/${sessionToken}/pay`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${userAccessToken}` },
    body: JSON.stringify({ currency }),
  });
}
