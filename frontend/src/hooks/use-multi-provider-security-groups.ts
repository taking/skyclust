/**
 * useMultiProviderSecurityGroups Hook
 * Multi-provider Security Group 조회 및 병합
 * 
 * 기능:
 * - 여러 provider (AWS, GCP, Azure) 동시 조회
 * - Provider별 Region 선택 지원
 * - VPC별 Security Group 조회
 * - 결과 병합
 * - 부분 에러 처리
 */

import { useMemo } from 'react';
import { useQueries } from '@tanstack/react-query';
import { networkService } from '@/services/network';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import type { CloudProvider } from '@/lib/types/kubernetes';
import type { SecurityGroup } from '@/lib/types';
import type { ProviderRegionSelection } from './use-provider-region-filter';

interface UseMultiProviderSecurityGroupsOptions {
  workspaceId: string;
  credentialIds: string[];
  credentials: Array<{ id: string; provider: CloudProvider }>;
  vpcId: string;
  selectedRegions?: ProviderRegionSelection;
  enabled?: boolean;
}

/**
 * Multi-provider Security Group 조회 Hook (Provider별 Region 지원)
 */
export function useMultiProviderSecurityGroups({
  workspaceId,
  credentialIds,
  credentials,
  vpcId,
  selectedRegions,
  enabled = true,
}: UseMultiProviderSecurityGroupsOptions) {
  
  const queries = useMemo(() => {
    const queryList: Array<{
      queryKey: unknown[];
      queryFn: () => Promise<SecurityGroup[]>;
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
            queryKey: queryKeys.securityGroups.list(provider, credentialId, vpcId, region),
            queryFn: async () => {
              return networkService.listSecurityGroups(provider, credentialId, vpcId, region);
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
    const allSecurityGroups: Array<SecurityGroup & { provider?: CloudProvider; credential_id?: string }> = [];
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
          const securityGroupsWithProvider = result.data.map(sg => ({
            ...sg,
            provider,
            credential_id: credentialId,
          }));
          allSecurityGroups.push(...securityGroupsWithProvider);
        }
      });
    });
    
    return {
      securityGroups: allSecurityGroups,
      isLoading: results.some(r => r.isLoading),
      errors,
      hasError: errors.length > 0,
      totalCount: allSecurityGroups.length,
    };
  }, [results, credentialIds, credentials, selectedRegions]);
  
  return result;
}

