/**
 * usePageRefresh Hook
 * 
 * 페이지 새로고침 시 현재 페이지의 React Query 쿼리를 자동으로 무효화하고 재요청합니다.
 * 새로고침 시 현재 페이지를 유지하면서 데이터만 갱신하는 기능을 제공합니다.
 * 
 * @example
 * ```tsx
 * function MyPage() {
 *   // 현재 페이지의 쿼리 키를 지정하여 새로고침 시 자동 무효화
 *   usePageRefresh([
 *     queryKeys.vpcs.list(provider, credentialId, region),
 *   ]);
 *   
 *   // ... 나머지 컴포넌트 코드
 * }
 * ```
 */

import { useEffect, useRef } from 'react';
import { usePathname } from 'next/navigation';
import { useQueryClient, QueryClient } from '@tanstack/react-query';
import { QueryKey } from '@/lib/query';
import { log } from '@/lib/logging';

export interface UsePageRefreshOptions {
  /**
   * 쿼리 무효화 시 refetch 여부
   * @default true
   */
  refetch?: boolean;

  /**
   * 새로고침 감지 방식
   * - 'mount': 컴포넌트 마운트 시 (기본값)
   * - 'navigation': Next.js 네비게이션 이벤트 시
   * @default 'mount'
   */
  trigger?: 'mount' | 'navigation';

  /**
   * 무효화할 쿼리 키 목록
   * 빈 배열이면 현재 페이지의 모든 쿼리를 무효화
   */
  queryKeys?: QueryKey[];

  /**
   * 무효화할 쿼리 키 패턴 (predicate 함수)
   * queryKeys와 함께 사용 가능
   */
  predicate?: (query: { queryKey: QueryKey }) => boolean;

  /**
   * 무효화 전 딜레이 (ms)
   * @default 0
   */
  delay?: number;
}

/**
 * usePageRefresh Hook
 * 
 * 페이지 새로고침 시 React Query 쿼리를 자동으로 무효화합니다.
 * 
 * @param options - 설정 옵션
 */
export function usePageRefresh(options: UsePageRefreshOptions = {}) {
  const {
    refetch = true,
    trigger = 'mount',
    queryKeys = [],
    predicate,
    delay = 0,
  } = options;

  const queryClient = useQueryClient();
  const pathname = usePathname();
  const hasInvalidatedRef = useRef(false);

  useEffect(() => {
    // 이미 무효화했으면 스킵 (중복 방지)
    if (hasInvalidatedRef.current) {
      return;
    }

    // 마운트 시 트리거인 경우에만 실행
    if (trigger === 'mount') {
      const invalidateQueries = async () => {
        try {
          // 딜레이 적용
          if (delay > 0) {
            await new Promise(resolve => setTimeout(resolve, delay));
          }

          // 쿼리 키가 지정된 경우
          if (queryKeys.length > 0) {
            for (const queryKey of queryKeys) {
              await queryClient.invalidateQueries({
                queryKey,
                exact: false,
              });
            }
          }

          // Predicate가 지정된 경우
          if (predicate) {
            await queryClient.invalidateQueries({
              predicate,
            });
          }

          // 쿼리 키와 predicate가 모두 없는 경우, 현재 페이지 관련 쿼리만 무효화
          if (queryKeys.length === 0 && !predicate) {
            // 경로 기반으로 쿼리 무효화
            await invalidateQueriesByPath(pathname, queryClient);
          }

          // Refetch가 활성화된 경우 재요청
          if (refetch) {
            if (queryKeys.length > 0) {
              for (const queryKey of queryKeys) {
                await queryClient.refetchQueries({
                  queryKey,
                  exact: false,
                });
              }
            } else if (predicate) {
              await queryClient.refetchQueries({
                predicate,
              });
            } else {
              await refetchQueriesByPath(pathname, queryClient);
            }
          }

          hasInvalidatedRef.current = true;

          log.debug('[Page Refresh] Queries invalidated and refetched', {
            pathname,
            queryKeysCount: queryKeys.length,
            hasPredicate: !!predicate,
            refetch,
          });
        } catch (error) {
          log.error('[Page Refresh] Failed to invalidate queries', error, {
            pathname,
            service: 'PageRefresh',
            action: 'invalidateQueries',
          });
        }
      };

      invalidateQueries();
    }
  }, [queryClient, pathname, refetch, trigger, delay, queryKeys, predicate]);

  // 경로 변경 시 리셋 (다른 페이지로 이동 시)
  useEffect(() => {
    hasInvalidatedRef.current = false;
  }, [pathname]);
}

/**
 * 경로 기반으로 쿼리 무효화
 */
async function invalidateQueriesByPath(
  pathname: string,
  queryClient: QueryClient
): Promise<void> {
  const pathPatterns = getQueryKeyPatternsByPath(pathname);
  
  for (const pattern of pathPatterns) {
    await queryClient.invalidateQueries({
      queryKey: pattern,
      exact: false,
    });
  }
}

/**
 * 경로 기반으로 쿼리 재요청
 */
async function refetchQueriesByPath(
  pathname: string,
  queryClient: QueryClient
): Promise<void> {
  const pathPatterns = getQueryKeyPatternsByPath(pathname);
  
  for (const pattern of pathPatterns) {
    await queryClient.refetchQueries({
      queryKey: pattern,
      exact: false,
    });
  }
}

/**
 * 경로에 따른 쿼리 키 패턴 반환
 */
function getQueryKeyPatternsByPath(pathname: string): QueryKey[] {
  const patterns: QueryKey[] = [];

  // 경로별 쿼리 키 패턴 매핑
  if (pathname.startsWith('/kubernetes/clusters')) {
    patterns.push(['kubernetes-clusters']);
    if (pathname.match(/\/kubernetes\/clusters\/[^/]+$/)) {
      // 클러스터 상세 페이지
      patterns.push(['kubernetes-clusters', 'detail']);
    }
  } else if (pathname.startsWith('/kubernetes/node-groups')) {
    patterns.push(['node-groups']);
    if (pathname.match(/\/kubernetes\/node-groups\/[^/]+$/)) {
      // 노드 그룹 상세 페이지
      patterns.push(['node-groups', 'detail']);
    }
  } else if (pathname.startsWith('/kubernetes/nodes') || pathname.includes('/k8s/nodes')) {
    patterns.push(['nodes']);
  } else if (pathname.startsWith('/networks/vpcs')) {
    patterns.push(['vpcs']);
    if (pathname.match(/\/networks\/vpcs\/[^/]+$/)) {
      // VPC 상세 페이지
      patterns.push(['vpcs', 'detail']);
    }
  } else if (pathname.startsWith('/networks/subnets')) {
    patterns.push(['subnets']);
    if (pathname.match(/\/networks\/subnets\/[^/]+$/)) {
      // 서브넷 상세 페이지
      patterns.push(['subnets', 'detail']);
    }
  } else if (pathname.startsWith('/networks/security-groups')) {
    patterns.push(['security-groups']);
    if (pathname.match(/\/networks\/security-groups\/[^/]+$/)) {
      // 보안 그룹 상세 페이지
      patterns.push(['security-groups', 'detail']);
    }
  } else if (pathname.startsWith('/compute/vms')) {
    patterns.push(['vms']);
    if (pathname.match(/\/compute\/vms\/[^/]+$/)) {
      // VM 상세 페이지
      patterns.push(['vms', 'detail']);
    }
  } else if (pathname.startsWith('/compute/images')) {
    patterns.push(['images']);
  } else if (pathname.startsWith('/compute/snapshots')) {
    patterns.push(['snapshots']);
  } else if (pathname.startsWith('/dashboard')) {
    patterns.push(['dashboard']);
  } else if (pathname.startsWith('/workspaces')) {
    patterns.push(['workspaces']);
    if (pathname.match(/\/workspaces\/[^/]+\/settings$/)) {
      patterns.push(['workspaces', 'detail']);
    }
    if (pathname.match(/\/workspaces\/[^/]+\/members$/)) {
      patterns.push(['workspaces', 'detail']);
    }
  } else if (pathname.startsWith('/azure/iam/resource-groups')) {
    patterns.push(['azure-resource-groups']);
  } else if (pathname.startsWith('/notifications')) {
    patterns.push(['notifications']);
  } else if (pathname.startsWith('/exports')) {
    patterns.push(['exports']);
  } else if (pathname.startsWith('/cost-analysis')) {
    patterns.push(['cost-analysis']);
  }

  return patterns;
}

