/**
 * Devtools Provider Component
 * 개발 도구 통합 Provider
 * 
 * React Query Devtools를 개발 환경에서만 활성화합니다.
 * Zustand Devtools는 각 스토어에서 직접 설정됩니다.
 */

'use client';

import { ReactNode, useEffect, useState } from 'react';
import { log } from '@/lib/logging';

/**
 * 개발 환경 확인
 */
const isDevelopment = process.env.NODE_ENV === 'development';

export interface DevtoolsProviderProps {
  children: ReactNode;
  /**
   * React Query Devtools 초기 오픈 여부
   */
  initialIsOpen?: boolean;
}

/**
 * DevtoolsProvider Component
 * 
 * 개발 환경에서만 개발 도구를 활성화합니다.
 * 
 * @example
 * ```tsx
 * <DevtoolsProvider>
 *   <App />
 * </DevtoolsProvider>
 * ```
 */
export function DevtoolsProvider({ 
  children,
  initialIsOpen = false,
}: DevtoolsProviderProps) {
  const [ReactQueryDevtools, setReactQueryDevtools] = useState<React.ComponentType<{ initialIsOpen?: boolean }> | null>(null);
  const [isMounted, setIsMounted] = useState(false);

  useEffect(() => {
    setIsMounted(true);
    
    // 개발 환경에서만 동적 import
    // 모듈이 없어도 앱이 계속 작동하도록 try-catch로 안전하게 처리
    if (isDevelopment && typeof window !== 'undefined') {
      // 동적 import를 비동기로 처리하여 빌드 타임에 모듈을 필수로 요구하지 않음
      Promise.resolve()
        .then(() => import('@tanstack/react-query-devtools'))
        .then((mod) => {
          if (mod && mod.ReactQueryDevtools) {
            setReactQueryDevtools(() => mod.ReactQueryDevtools);
          }
        })
        .catch((error) => {
          // 모듈이 없거나 로드 실패 시 조용히 무시
          log.warn('React Query Devtools is not available', error instanceof Error ? error : new Error(String(error)));
        });
    }
  }, []);

  // 개발 환경이 아니면 children만 반환
  if (!isDevelopment || !isMounted) {
    return <>{children}</>;
  }

  return (
    <>
      {children}
      {ReactQueryDevtools && (
        <ReactQueryDevtools initialIsOpen={initialIsOpen} />
      )}
    </>
  );
}

