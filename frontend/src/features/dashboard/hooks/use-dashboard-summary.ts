/**
 * useDashboardSummary Hook
 * 대시보드 요약 정보를 조회하는 React Query 훅
 * 
 * 폴링 대신 SSE 이벤트를 통해 실시간으로 업데이트됩니다.
 * useSSEEvents 훅에서 리소스 변경 이벤트를 수신하면 자동으로 쿼리가 무효화됩니다.
 */

import { useQuery } from '@tanstack/react-query';
import { dashboardService } from '../services/dashboard';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import type { DashboardSummary } from '@/lib/types/dashboard';

interface UseDashboardSummaryOptions {
  workspaceId: string;
  credentialId?: string;
  region?: string;
  enabled?: boolean;
}

/**
 * 대시보드 요약 정보 조회 훅
 * 
 * 폴링이 제거되었으며, SSE 이벤트를 통해 실시간 업데이트됩니다.
 * VM, Kubernetes, Network 리소스 변경 시 자동으로 쿼리가 무효화되어 최신 데이터를 가져옵니다.
 * 
 * @param options - 조회 옵션
 * @returns 대시보드 요약 정보 및 로딩 상태
 */
export function useDashboardSummary({
  workspaceId,
  credentialId,
  region,
  enabled = true,
}: UseDashboardSummaryOptions) {
  return useQuery<DashboardSummary>({
    queryKey: queryKeys.dashboard.summary(workspaceId, credentialId, region),
    queryFn: () => dashboardService.getDashboardSummary(workspaceId, credentialId, region),
    enabled: enabled && !!workspaceId,
    staleTime: CACHE_TIMES.REALTIME, // 30초 - 대시보드 데이터는 자주 변경될 수 있음
    gcTime: GC_TIMES.SHORT, // 5분
    // refetchInterval 제거: SSE 이벤트로 자동 업데이트
    retry: 3,
    retryDelay: 1000,
  });
}

