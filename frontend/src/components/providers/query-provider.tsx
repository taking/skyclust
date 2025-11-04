/**
 * Query Provider Component
 * React Query Provider로 앱을 래핑
 * 
 * 오프라인 큐 자동 처리 포함
 * React Query Devtools 통합
 */

'use client';

import { QueryClientProvider } from '@tanstack/react-query';
import { queryClient } from '@/lib/query-client';
import { ReactNode } from 'react';
import { useOfflineQueue } from '@/hooks/use-offline-queue';
import { DevtoolsProvider } from './devtools-provider';

interface QueryProviderProps {
  children: ReactNode;
  /**
   * React Query Devtools 초기 오픈 여부
   */
  devtoolsInitialIsOpen?: boolean;
}

function QueryProviderContent({ children }: QueryProviderProps) {
  // 오프라인 큐 자동 처리
  useOfflineQueue();
  
  return <>{children}</>;
}

export function QueryProvider({ 
  children,
  devtoolsInitialIsOpen = false,
}: QueryProviderProps) {
  return (
    <QueryClientProvider client={queryClient}>
      <DevtoolsProvider initialIsOpen={devtoolsInitialIsOpen}>
        <QueryProviderContent>
          {children}
        </QueryProviderContent>
      </DevtoolsProvider>
    </QueryClientProvider>
  );
}
