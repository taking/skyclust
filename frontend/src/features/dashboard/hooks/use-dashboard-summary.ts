/**
 * useDashboardSummary Hook
 * 대시보드 요약 정보를 조회하는 React Query 훅
 */

import { useQuery } from '@tanstack/react-query';
import { dashboardService } from '../services/dashboard';
import { queryKeys } from '@/lib/query-keys';
import type { DashboardSummary } from '@/lib/types/dashboard';

interface UseDashboardSummaryOptions {
  workspaceId: string;
  credentialId?: string;
  region?: string;
  enabled?: boolean;
}

/**
 * 대시보드 요약 정보 조회 훅
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
    staleTime: 30 * 1000, // 30초 - 대시보드 데이터는 자주 변경될 수 있음
    gcTime: 5 * 60 * 1000, // 5분
    refetchInterval: 30 * 1000, // 30초마다 자동 갱신
    retry: 3,
    retryDelay: 1000,
  });
}

