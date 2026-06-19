export type SessionStatus = 'PENDING' | 'COMPLETED' | 'EXPIRED';

export interface QRSession {
  token: string;
  merchant_id: string;
  amount_krw: string;
  status: SessionStatus;
  currency?: string;
  user_id?: string;
  transaction_id?: string;
  expires_at: string;
  created_at: string;
}

export interface CreateQRData {
  token: string;
  amount_krw: string;
  expires_at: string;
  qr_payload: string;
}

export interface AuthData {
  access_token: string;
  refresh_token: string;
  merchant_id?: string;
  business_name?: string;
}

export interface PayResult {
  transaction_id: string;
  merchant_id: string;
  amount_krw: string;
  used_currency: string;
  used_amount: string;
  applied_rate: string;
  remaining_balance: string;
}
