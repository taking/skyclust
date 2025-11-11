/**
 * React Query Client Configuration
 * 
 * 캐시 전략:
 * - 리소스별 staleTime 설정으로 불필요한 refetch 방지
 * - gcTime (구 cacheTime) 설정으로 메모리 사용 최적화
 * - 에러별 retry 전략 최적화
 */

import { QueryClient } from '@tanstack/react-query';

// 캐시 시간 상수 (밀리초)
export const CACHE_TIMES = {
  // 빠르게 변경되는 데이터 (30초)
  REALTIME: 30 * 1000,
  
  // 실시간 모니터링 데이터 (1분)
  MONITORING: 60 * 1000,
  
  // 일반 리소스 데이터 (5분)
  RESOURCE: 5 * 60 * 1000,
  
  // 자주 변경되지 않는 데이터 (10분)
  STABLE: 10 * 60 * 1000,
  
  // 거의 변경되지 않는 데이터 (30분)
  STATIC: 30 * 60 * 1000,
  
  // 매우 안정적인 데이터 (1시간)
  CACHEABLE: 60 * 60 * 1000,
} as const;

// Garbage Collection 시간 상수 (메모리 정리)
export const GC_TIMES = {
  // 짧은 GC 시간 (5분)
  SHORT: 5 * 60 * 1000,
  
  // 중간 GC 시간 (10분)
  MEDIUM: 10 * 60 * 1000,
  
  // 긴 GC 시간 (30분)
  LONG: 30 * 60 * 1000,
} as const;

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // 기본 staleTime: 5분 (일반적인 리소스 데이터)
      staleTime: CACHE_TIMES.RESOURCE,
      
      // 기본 gcTime: 10분 (메모리 정리 시간)
      gcTime: GC_TIMES.MEDIUM,
      
      // 기본 refetch 설정
      refetchOnWindowFocus: false, // 윈도우 포커스 시 자동 refetch 비활성화
      refetchOnMount: true, // 컴포넌트 마운트 시 refetch 활성화
      refetchOnReconnect: true, // 네트워크 재연결 시 refetch 활성화
      
      // 재시도 전략
      retry: (failureCount, error: unknown) => {
        const errorResponse = error as { response?: { status?: number } };
        const status = errorResponse?.response?.status;
        
        // 인증 에러는 재시도하지 않음
        if (status === 401 || status === 403) {
          return false;
        }
        
        // 4xx 클라이언트 에러는 재시도하지 않음
        if (status && status >= 400 && status < 500) {
          return false;
        }
        
        // 최대 3번 재시도
        return failureCount < 3;
      },
      
      // 재시도 딜레이 (지수 백오프)
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    },
    mutations: {
      // Mutation 재시도 전략
      retry: (failureCount, error: unknown) => {
        const errorResponse = error as { response?: { status?: number } };
        const status = errorResponse?.response?.status;
        
        // 인증 에러나 클라이언트 에러는 재시도하지 않음
        if (status && (status === 401 || status === 403 || (status >= 400 && status < 500))) {
          return false;
        }
        
        // 네트워크 에러만 1번 재시도
        return failureCount < 1;
      },
      
      // Mutation 재시도 딜레이
      retryDelay: 1000,
    },
  },
});

