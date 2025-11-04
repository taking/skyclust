/**
 * Sentry Provider
 * Sentry 초기화를 위한 클라이언트 컴포넌트
 */

'use client';

import { useEffect } from 'react';
import { initSentry } from '@/lib/sentry-client';

export function SentryProvider({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    initSentry();
  }, []);

  return <>{children}</>;
}

