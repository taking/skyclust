/**
 * Resource Groups Hook
 * Azure Resource Groups 관련 React Query 훅
 */

import { useQuery } from '@tanstack/react-query';
import { resourceGroupService } from '@/services/resource-group';
import type { ResourceGroupInfo } from '@/services/resource-group';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';

export interface UseResourceGroupsOptions {
  credentialId?: string;
  enabled?: boolean;
  limit?: number; // 조회할 최대 개수 (Sidebar용: 100, 목록 페이지용: 페이지 크기)
}

/**
 * Azure Resource Groups 목록 조회 훅
 * 
 * @param options - 옵션 객체 또는 credentialId (하위 호환성)
 * @param enabled - 쿼리 활성화 여부 (옵션 객체 사용 시 무시됨)
 * 
 * @example
 * ```tsx
 * // Sidebar용: 모든 Resource Group 조회
 * const { data: resourceGroups, isLoading } = useResourceGroups({
 *   credentialId: 'credential-id',
 *   limit: 100
 * });
 * 
 * // 목록 페이지용: Pagination 사용
 * const { data: resourceGroups, isLoading } = useResourceGroups({
 *   credentialId: 'credential-id',
 *   limit: 20
 * });
 * 
 * // 하위 호환성: 기존 방식도 지원
 * const { data: resourceGroups, isLoading } = useResourceGroups('credential-id');
 * ```
 */
export function useResourceGroups(
  optionsOrCredentialId?: UseResourceGroupsOptions | string,
  enabled?: boolean
) {
  // 하위 호환성: 문자열로 전달된 경우 옵션 객체로 변환
  const options: UseResourceGroupsOptions = typeof optionsOrCredentialId === 'string'
    ? { credentialId: optionsOrCredentialId, enabled }
    : optionsOrCredentialId || {};

  const { credentialId, limit, enabled: enabledOption } = options;
  const isEnabled = enabledOption !== undefined ? enabledOption : !!credentialId;

  return useQuery<ResourceGroupInfo[]>({
    queryKey: queryKeys.azureResourceGroups.list(credentialId, limit),
    queryFn: () => {
      if (!credentialId) {
        return Promise.resolve([]);
      }
      return resourceGroupService.listResourceGroups(credentialId, limit);
    },
    enabled: isEnabled && !!credentialId,
    staleTime: CACHE_TIMES.STABLE, // 10분 - Resource Groups는 자주 변경되지 않음
    gcTime: GC_TIMES.LONG, // 30분 - GC 시간
  });
}

