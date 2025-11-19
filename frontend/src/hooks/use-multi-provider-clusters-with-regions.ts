/**
 * useMultiProviderClustersWithRegions Hook
 * Multi-provider 클러스터 조회 및 병합 (Provider별 Region 지원)
 * 
 * 기능:
 * - 여러 provider (AWS, GCP, Azure) 동시 조회
 * - Provider별 Region 선택 지원
 * - 결과 병합
 * - 부분 에러 처리
 * - Provider별 로딩 상태 관리
 */

import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { kubernetesService } from '@/features/kubernetes';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import type { CloudProvider } from '@/lib/types/kubernetes';
import type { KubernetesCluster } from '@/lib/types';
import type { ProviderRegionSelection } from './use-provider-region-filter';

interface UseMultiProviderClustersWithRegionsOptions {
  workspaceId: string;
  credentialIds: string[]; // Multi-credential IDs
  credentials: Array<{ id: string; provider: CloudProvider }>; // Credential 정보
  selectedRegions?: ProviderRegionSelection; // Provider별 Region 선택
  enabled?: boolean;
}

interface ProviderQueryResult {
  provider: CloudProvider;
  credentialId: string;
  region: string;
  data: KubernetesCluster[];
  isLoading: boolean;
  error: Error | null;
}

/**
 * Multi-provider 클러스터 조회 Hook (Provider별 Region 지원)
 * 
 * @example
 * ```tsx
 * const { 
 *   clusters, 
 *   isLoading, 
 *   errors,
 *   providersStatus 
 * } = useMultiProviderClustersWithRegions({
 *   workspaceId: 'ws-1',
 *   credentialIds: ['cred-aws-1', 'cred-gcp-1'],
 *   credentials: [
 *     { id: 'cred-aws-1', provider: 'aws' },
 *     { id: 'cred-gcp-1', provider: 'gcp' },
 *   ],
 *   selectedRegions: {
 *     aws: ['us-east-1', 'ap-northeast-3'],
 *     gcp: ['asia-northeast3'],
 *   },
 * });
 * ```
 */
export function useMultiProviderClustersWithRegions({
  workspaceId,
  credentialIds,
  credentials,
  selectedRegions,
  enabled = true,
}: UseMultiProviderClustersWithRegionsOptions) {
  
  // 배치 쿼리 생성: 모든 credential과 region 조합
  const batchQueries = useMemo(() => {
    const queries: Array<{
      credential_id: string;
      region: string;
      resource_group?: string;
    }> = [];
    
    credentialIds.forEach(credentialId => {
      const credential = credentials.find(c => c.id === credentialId);
      if (!credential) return;
      
      const providerRegions = selectedRegions?.[credential.provider as CloudProvider] || [];
      providerRegions.forEach(region => {
        queries.push({
          credential_id: credentialId,
          region,
        });
      });
    });
    
    return queries;
  }, [credentialIds, credentials, selectedRegions]);
  
  // 배치 API 호출 (첫 번째 provider 사용 - 모든 provider에서 동일한 엔드포인트 사용)
  const primaryProvider = useMemo(() => {
    if (credentialIds.length === 0) return 'aws' as CloudProvider;
    const firstCredential = credentials.find(c => c.id === credentialIds[0]);
    return (firstCredential?.provider as CloudProvider) || 'aws';
  }, [credentialIds, credentials]);
  
  const { data: batchResult, isLoading, error } = useQuery({
    queryKey: queryKeys.kubernetesClusters.batch(workspaceId, batchQueries),
    queryFn: async () => {
      if (batchQueries.length === 0) {
        return {
          results: [],
          total: 0,
        };
      }
      return kubernetesService.batchListClusters(primaryProvider, batchQueries);
    },
    enabled: enabled && !!workspaceId && batchQueries.length > 0,
    staleTime: CACHE_TIMES.MONITORING,
    gcTime: GC_TIMES.MEDIUM,
    retry: 1,
    retryOnMount: false,
  });
  
  // 결과 병합 및 상태 계산
  const result = useMemo(() => {
    if (!batchResult) {
      return {
        clusters: [],
        isLoading,
        errors: [],
        hasError: !!error,
        providersStatus: {},
        totalCount: 0,
      };
    }
    
    const allClusters: KubernetesCluster[] = [];
    const errors: Array<{ provider: CloudProvider; credentialId: string; region: string; error: Error }> = [];
    const providersStatus: Record<string, ProviderQueryResult> = {};
    
    batchResult.results.forEach(result => {
      const key = `${result.provider}-${result.credential_id}-${result.region}`;
      
      if (result.error) {
        errors.push({
          provider: result.provider as CloudProvider,
          credentialId: result.credential_id,
          region: result.region,
          error: new Error(result.error.message),
        });
      } else if (result.clusters && Array.isArray(result.clusters)) {
        const clustersWithProvider = result.clusters.map(cluster => ({
          ...cluster,
          provider: result.provider as CloudProvider,
          credential_id: result.credential_id,
        }));
        allClusters.push(...clustersWithProvider);
      }
      
      providersStatus[key] = {
        provider: result.provider as CloudProvider,
        credentialId: result.credential_id,
        region: result.region,
        data: result.clusters || [],
        isLoading: false,
        error: result.error ? new Error(result.error.message) : null,
      };
    });
    
    return {
      clusters: allClusters,
      isLoading,
      errors,
      hasError: errors.length > 0 || !!error,
      providersStatus,
      totalCount: allClusters.length,
    };
  }, [batchResult, isLoading, error]);
  
  return result;
}

