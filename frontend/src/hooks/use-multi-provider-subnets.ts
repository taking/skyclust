/**
 * useMultiProviderSubnets Hook
 * Multi-provider Subnet 조회 및 병합
 * 
 * 기능:
 * - 여러 provider (AWS, GCP, Azure) 동시 조회
 * - Provider별 Region 선택 지원
 * - VPC별 Subnet 조회
 * - 결과 병합
 * - 부분 에러 처리
 */

import { useMemo } from 'react';
import { useQueries } from '@tanstack/react-query';
import { networkService } from '@/services/network';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import type { CloudProvider } from '@/lib/types/kubernetes';
import type { Subnet } from '@/lib/types';
import type { ProviderRegionSelection } from './use-provider-region-filter';

interface UseMultiProviderSubnetsOptions {
  workspaceId: string;
  credentialIds: string[];
  credentials: Array<{ id: string; provider: CloudProvider }>;
  vpcId: string;
  selectedRegions?: ProviderRegionSelection;
  enabled?: boolean;
}

/**
 * Multi-provider Subnet 조회 Hook (Provider별 Region 지원)
 */
export function useMultiProviderSubnets({
  workspaceId,
  credentialIds,
  credentials,
  vpcId,
  selectedRegions,
  enabled = true,
}: UseMultiProviderSubnetsOptions) {
  
  const queries = useMemo(() => {
    const queryList: Array<{
      queryKey: unknown[];
      queryFn: () => Promise<Subnet[]>;
      enabled: boolean;
    }> = [];
    
    if (!vpcId) return queryList;
    
    credentialIds.forEach(credentialId => {
      const credential = credentials.find(c => c.id === credentialId);
      if (!credential) return;
      
      const provider = credential.provider as CloudProvider;
      const providerRegions = selectedRegions?.[provider] || [];
      
      if (providerRegions.length > 0) {
        providerRegions.forEach(region => {
          queryList.push({
            queryKey: queryKeys.subnets.list(provider, credentialId, vpcId, region),
            queryFn: async () => {
              return networkService.listSubnets(provider, credentialId, vpcId, region);
            },
            enabled: enabled && !!workspaceId && !!credentialId && !!vpcId && !!region,
          });
        });
      }
    });
    
    return queryList;
  }, [credentialIds, credentials, vpcId, selectedRegions, enabled, workspaceId]);
  
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
    const allSubnets: Array<Subnet & { provider?: CloudProvider; credential_id?: string }> = [];
    const errors: Array<{ provider: CloudProvider; credentialId: string; region: string; error: Error }> = [];
    
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
        
        if (result.error) {
          errors.push({
            provider,
            credentialId,
            region,
            error: result.error as Error,
          });
        } else if (result.data) {
          const subnetsWithProvider = result.data.map(subnet => ({
            ...subnet,
            provider,
            credential_id: credentialId,
          }));
          allSubnets.push(...subnetsWithProvider);
        }
      });
    });
    
    return {
      subnets: allSubnets,
      isLoading: results.some(r => r.isLoading),
      errors,
      hasError: errors.length > 0,
      totalCount: allSubnets.length,
    };
  }, [results, credentialIds, credentials, selectedRegions]);
  
  return result;
}

