/**
 * useMultiProviderVPCs Hook
 * Multi-provider VPC 조회 및 병합
 * 
 * 기능:
 * - 여러 provider (AWS, GCP, Azure) 동시 조회
 * - Provider별 Region 선택 지원
 * - 결과 병합
 * - 부분 에러 처리
 * - Provider별 로딩 상태 관리
 */

import { useMemo } from 'react';
import { useQueries } from '@tanstack/react-query';
import { networkService } from '@/services/network';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import type { CloudProvider } from '@/lib/types/kubernetes';
import type { VPC } from '@/lib/types';
import type { ProviderRegionSelection } from './use-provider-region-filter';

interface UseMultiProviderVPCsOptions {
  workspaceId: string;
  credentialIds: string[];
  credentials: Array<{ id: string; provider: CloudProvider }>;
  selectedRegions?: ProviderRegionSelection;
  enabled?: boolean;
}

interface ProviderQueryResult {
  provider: CloudProvider;
  credentialId: string;
  region: string;
  data: VPC[];
  isLoading: boolean;
  error: Error | null;
}

/**
 * Multi-provider VPC 조회 Hook (Provider별 Region 지원)
 * 
 * @example
 * ```tsx
 * const { 
 *   vpcs, 
 *   isLoading, 
 *   errors,
 *   providersStatus 
 * } = useMultiProviderVPCs({
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
export function useMultiProviderVPCs({
  workspaceId,
  credentialIds,
  credentials,
  selectedRegions,
  enabled = true,
}: UseMultiProviderVPCsOptions) {
  
  const queries = useMemo(() => {
    const queryList: Array<{
      queryKey: unknown[];
      queryFn: () => Promise<VPC[]>;
      enabled: boolean;
    }> = [];
    
    credentialIds.forEach(credentialId => {
      const credential = credentials.find(c => c.id === credentialId);
      if (!credential) return;
      
      const provider = credential.provider as CloudProvider;
      const providerRegions = selectedRegions?.[provider] || [];
      
      if (providerRegions.length > 0) {
        providerRegions.forEach(region => {
          queryList.push({
            queryKey: queryKeys.vpcs.list(provider, credentialId, region),
            queryFn: async () => {
              return networkService.listVPCs(provider, credentialId, region);
            },
            enabled: enabled && !!workspaceId && !!credentialId && !!region,
          });
        });
      }
    });
    
    return queryList;
  }, [credentialIds, credentials, selectedRegions, enabled, workspaceId]);
  
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
    const errors: Array<{ provider: CloudProvider; credentialId: string; region: string; error: Error }> = [];
    const providersStatus: Record<string, ProviderQueryResult> = {};
    
    let queryIndex = 0;
    credentialIds.forEach(credentialId => {
      const credential = credentials.find(c => c.id === credentialId);
      if (!credential) return;
      
      const provider = credential.provider as CloudProvider;
      const providerRegions = selectedRegions?.[provider] || [];
      
      providerRegions.forEach(region => {
        const result = results[queryIndex];
        queryIndex++;
        
        if (!result) return;
        
        const key = `${provider}-${credentialId}-${region}`;
        
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
        
        providersStatus[key] = {
          provider,
          credentialId,
          region,
          data: result.data || [],
          isLoading: result.isLoading,
          error: result.error as Error | null,
        };
      });
    });
    
    return {
      vpcs: allVPCs,
      isLoading: results.some(r => r.isLoading),
      errors,
      hasError: errors.length > 0,
      providersStatus,
      totalCount: allVPCs.length,
    };
  }, [results, credentialIds, credentials, selectedRegions]);
  
  return result;
}

