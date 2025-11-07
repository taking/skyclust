/**
 * Sentry Provider
 * 
 * Note: Sentry는 sentry.client.config.ts에서 자동으로 초기화됩니다.
 * 이 컴포넌트는 Sentry가 정상적으로 작동하는지 확인하기 위한 것입니다.
 */

'use client';

import { useEffect } from 'react';
import { logger } from '@/lib/logger';

export function SentryProvider({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    // 클라이언트에서만 확인
    if (typeof window !== 'undefined') {
      const dsn = process.env.NEXT_PUBLIC_SENTRY_DSN;
      if (dsn) {
        logger.info('[Sentry] Provider: Sentry is configured and ready');
      } else {
        logger.warn('[Sentry] Provider: DSN not configured. Check your .env.local file.');
      }
    }
  }, []);

  return <>{children}</>;
}

