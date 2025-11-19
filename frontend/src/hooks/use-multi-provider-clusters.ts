/**
 * useMultiProviderClusters Hook
 * Multi-provider 클러스터 조회 및 병합
 * 
 * 기능:
 * - 여러 provider (AWS, GCP, Azure) 동시 조회
 * - 결과 병합
 * - 부분 에러 처리
 * - Provider별 로딩 상태 관리
 */

import { useMemo } from 'react';
import { useQuery, useQueries } from '@tanstack/react-query';
import { kubernetesService } from '@/features/kubernetes';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import type { CloudProvider } from '@/lib/types/kubernetes';
import type { KubernetesCluster } from '@/lib/types';

interface UseMultiProviderClustersOptions {
  workspaceId: string;
  credentialIds: string[]; // Multi-credential IDs
  credentials: Array<{ id: string; provider: CloudProvider }>; // Credential 정보
  region?: string;
  enabled?: boolean;
}

interface ProviderQueryResult {
  provider: CloudProvider;
  credentialId: string;
  data: KubernetesCluster[];
  isLoading: boolean;
  error: Error | null;
}

/**
 * Multi-provider 클러스터 조회 Hook
 * 
 * @example
 * ```tsx
 * const { 
 *   clusters, 
 *   isLoading, 
 *   errors,
 *   providersStatus 
 * } = useMultiProviderClusters({
 *   workspaceId: 'ws-1',
 *   credentialIds: ['cred-aws-1', 'cred-gcp-1'],
 *   credentials: [
 *     { id: 'cred-aws-1', provider: 'aws' },
 *     { id: 'cred-gcp-1', provider: 'gcp' },
 *   ],
 *   region: 'ap-northeast-3',
 * });
 * ```
 */
export function useMultiProviderClusters({
  workspaceId,
  credentialIds,
  credentials,
  region,
  enabled = true,
}: UseMultiProviderClustersOptions) {
  
  // Credential ID를 Provider별로 그룹화
  const credentialsByProvider = useMemo(() => {
    const grouped: Record<CloudProvider, Array<{ id: string; provider: CloudProvider }>> = {
      aws: [],
      gcp: [],
      azure: [],
    };
    
    credentialIds.forEach(credentialId => {
      const credential = credentials.find(c => c.id === credentialId);
      if (credential && credential.provider in grouped) {
        grouped[credential.provider as CloudProvider].push(credential);
      }
    });
    
    return grouped;
  }, [credentialIds, credentials]);
  
  // 각 Provider별 쿼리 생성
  const queries = useQueries({
    queries: Object.entries(credentialsByProvider).flatMap(([provider, creds]) =>
      creds.map(cred => ({
        queryKey: queryKeys.kubernetesClusters.list(
          workspaceId,
          provider as CloudProvider,
          cred.id,
          region
        ),
        queryFn: async () => {
          return kubernetesService.listClusters(
            provider as CloudProvider,
            cred.id,
            region
          );
        },
        enabled: enabled && !!workspaceId && !!cred.id,
        staleTime: CACHE_TIMES.MONITORING,
        gcTime: GC_TIMES.MEDIUM,
        retry: 1, // 부분 실패 허용
        retryOnMount: false,
      }))
    ),
  });
  
  // 결과 병합 및 상태 계산
  const result = useMemo(() => {
    const allClusters: KubernetesCluster[] = [];
    const errors: Array<{ provider: CloudProvider; credentialId: string; error: Error }> = [];
    const providersStatus: Record<string, ProviderQueryResult> = {};
    
    let queryIndex = 0;
    
    Object.entries(credentialsByProvider).forEach(([provider, creds]) => {
      creds.forEach(cred => {
        const query = queries[queryIndex];
        const key = `${provider}-${cred.id}`;
        
        if (query.data) {
          // Provider 정보 추가
          const clustersWithProvider = query.data.map(cluster => ({
            ...cluster,
            provider: provider as CloudProvider,
            credential_id: cred.id,
          }));
          allClusters.push(...clustersWithProvider);
        }
        
        if (query.error) {
          errors.push({
            provider: provider as CloudProvider,
            credentialId: cred.id,
            error: query.error as Error,
          });
        }
        
        providersStatus[key] = {
          provider: provider as CloudProvider,
          credentialId: cred.id,
          data: query.data || [],
          isLoading: query.isLoading,
          error: query.error as Error | null,
        };
        
        queryIndex++;
      });
    });
    
    const isLoading = queries.some(q => q.isLoading);
    const hasError = errors.length > 0;
    
    return {
      clusters: allClusters,
      isLoading,
      errors,
      hasError,
      providersStatus,
      totalCount: allClusters.length,
    };
  }, [queries, credentialsByProvider]);
  
  return result;
}

