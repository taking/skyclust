/**
 * useMultiProviderVPCsSingleRegion Hook
 * Multi-provider VPC 조회 및 병합 (단일 Region 모드)
 * 
 * 기능:
 * - 여러 provider (AWS, GCP, Azure) 동시 조회
 * - 단일 Region 선택
 * - 결과 병합
 * - 부분 에러 처리
 */

import { useMemo } from 'react';
import { useQueries } from '@tanstack/react-query';
import { networkService } from '@/services/network';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import type { CloudProvider } from '@/lib/types/kubernetes';
import type { VPC } from '@/lib/types';

interface UseMultiProviderVPCsSingleRegionOptions {
  workspaceId: string;
  credentialIds: string[];
  credentials: Array<{ id: string; provider: CloudProvider }>;
  region?: string;
  enabled?: boolean;
}

/**
 * Multi-provider VPC 조회 Hook (단일 Region 모드)
 */
export function useMultiProviderVPCsSingleRegion({
  workspaceId,
  credentialIds,
  credentials,
  region,
  enabled = true,
}: UseMultiProviderVPCsSingleRegionOptions) {
  
  const queries = useMemo(() => {
    return credentialIds.map(credentialId => {
      const credential = credentials.find(c => c.id === credentialId);
      if (!credential) {
        return {
          queryKey: queryKeys.vpcs.list('aws' as CloudProvider, credentialId, region),
          queryFn: async () => [] as VPC[],
          enabled: false,
        };
      }
      
      const provider = credential.provider as CloudProvider;
      
      return {
        queryKey: queryKeys.vpcs.list(provider, credentialId, region),
        queryFn: async () => {
          if (!region) return [];
          return networkService.listVPCs(provider, credentialId, region);
        },
        enabled: enabled && !!workspaceId && !!credentialId && !!region,
      };
    });
  }, [credentialIds, credentials, region, enabled, workspaceId]);
  
  const results = useQueries({
    queries: queries.map(q => ({
      ...q,
      staleTime: CACHE_TIMES.MONITORING,
      gcTime: GC_TIMES.MEDIUM,
      retry: 1,
      retryOnMount: false,
    })),
  });
  
  const result = useMemo(() => {
    const allVPCs: Array<VPC & { provider?: CloudProvider; credential_id?: string }> = [];
    const errors: Array<{ provider: CloudProvider; credentialId: string; region?: string; error: Error }> = [];
    
    results.forEach((result, index) => {
      const credentialId = credentialIds[index];
      const credential = credentials.find(c => c.id === credentialId);
      if (!credential) return;
      
      const provider = credential.provider as CloudProvider;
      
      if (result.error) {
        errors.push({
          provider,
          credentialId,
          region,
          error: result.error as Error,
        });
      } else if (result.data) {
        const vpcsWithProvider = result.data.map(vpc => ({
          ...vpc,
          provider,
          credential_id: credentialId,
        }));
        allVPCs.push(...vpcsWithProvider);
      }
    });
    
    return {
      vpcs: allVPCs,
      isLoading: results.some(r => r.isLoading),
      errors,
      hasError: errors.length > 0,
      totalCount: allVPCs.length,
    };
  }, [results, credentialIds, credentials, region]);
  
  return result;
}

