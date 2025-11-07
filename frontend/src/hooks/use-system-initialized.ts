/**
 * useSystemInitialized Hook
 * 
 * 시스템 초기화 상태를 확인하는 React Query 훅입니다.
 * 
 * @example
 * ```tsx
 * function SetupPage() {
 *   const { data: initStatus, isLoading } = useSystemInitialized();
 *   
 *   if (isLoading) return <div>Loading...</div>;
 *   if (initStatus?.initialized) return <div>Already initialized</div>;
 *   return <SetupForm />;
 * }
 * ```
 */

import { useQuery } from '@tanstack/react-query';
import { BaseService } from '@/lib/service-base';

/**
 * 시스템 초기화 상태 응답 타입
 */
export interface SystemInitializationStatus {
  initialized: boolean;
  user_count: number;
}

/**
 * SystemService 클래스
 * 시스템 관련 API를 제공합니다.
 */
class SystemService extends BaseService {
  /**
   * 시스템 초기화 상태 조회
   * @returns 초기화 상태 및 사용자 수
   */
  async getInitializationStatus(): Promise<SystemInitializationStatus> {
    return this.get<SystemInitializationStatus>('/system/initialized');
  }
}

const systemService = new SystemService();

/**
 * 시스템 초기화 상태 조회 훅
 * @param options - React Query 옵션
 * @returns 초기화 상태 및 로딩 상태
 */
export function useSystemInitialized(options?: {
  enabled?: boolean;
  staleTime?: number;
}) {
  const { enabled = true, staleTime = 5 * 60 * 1000 } = options || {};

  return useQuery<SystemInitializationStatus>({
    queryKey: ['system', 'initialized'],
    queryFn: () => systemService.getInitializationStatus(),
    enabled,
    staleTime, // 5분간 캐시
    gcTime: 10 * 60 * 1000, // 10분간 가비지 컬렉션 방지
    retry: 3,
    retryDelay: 1000,
  });
}

