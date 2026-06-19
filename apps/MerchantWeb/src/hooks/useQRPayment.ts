import { useEffect, useRef, useState } from 'react';
import { getQRStatus } from '../api/client';
import type { QRSession } from '../types';

const POLL_INTERVAL_MS = 2000;

export function useQRPayment(token: string | null, accessToken: string | null) {
  const [session, setSession] = useState<QRSession | null>(null);
  const [error, setError] = useState<string | null>(null);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    if (!token || !accessToken) return;

    setSession(null);
    setError(null);

    const poll = async () => {
      try {
        const data = await getQRStatus(token, accessToken);
        setSession(data);
        if (data.status !== 'PENDING') {
          if (intervalRef.current) clearInterval(intervalRef.current);
        }
      } catch (e) {
        setError(e instanceof Error ? e.message : '상태 조회 실패');
      }
    };

    poll();
    intervalRef.current = setInterval(poll, POLL_INTERVAL_MS);

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [token, accessToken]);

  return { session, error };
}
