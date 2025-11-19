/**
 * useMultiProviderSubnetsSingleRegion Hook
 * Multi-provider Subnet 조회 및 병합 (단일 Region 모드)
 */

import { useMemo } from 'react';
import { useQueries } from '@tanstack/react-query';
import { networkService } from '@/services/network';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import type { CloudProvider } from '@/lib/types/kubernetes';
import type { Subnet } from '@/lib/types';

interface UseMultiProviderSubnetsSingleRegionOptions {
  workspaceId: string;
  credentialIds: string[];
  credentials: Array<{ id: string; provider: CloudProvider }>;
  vpcId: string;
  region?: string;
  enabled?: boolean;
}

export function useMultiProviderSubnetsSingleRegion({
  workspaceId,
  credentialIds,
  credentials,
  vpcId,
  region,
  enabled = true,
}: UseMultiProviderSubnetsSingleRegionOptions) {
  
  const queries = useMemo(() => {
    if (!vpcId || !region) return [];
    
    return credentialIds.map(credentialId => {
      const credential = credentials.find(c => c.id === credentialId);
      if (!credential) {
        return {
          queryKey: queryKeys.subnets.list('aws' as CloudProvider, credentialId, vpcId, region),
          queryFn: async () => [] as Subnet[],
          enabled: false,
        };
      }
      
      const provider = credential.provider as CloudProvider;
      
      return {
        queryKey: queryKeys.subnets.list(provider, credentialId, vpcId, region),
        queryFn: async () => {
          return networkService.listSubnets(provider, credentialId, vpcId, region);
        },
        enabled: enabled && !!workspaceId && !!credentialId && !!vpcId && !!region,
      };
    });
  }, [credentialIds, credentials, vpcId, region, enabled, workspaceId]);
  
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
        const subnetsWithProvider = result.data.map(subnet => ({
          ...subnet,
          provider,
          credential_id: credentialId,
        }));
        allSubnets.push(...subnetsWithProvider);
      }
    });
    
    return {
      subnets: allSubnets,
      isLoading: results.some(r => r.isLoading),
      errors,
      hasError: errors.length > 0,
      totalCount: allSubnets.length,
    };
  }, [results, credentialIds, credentials, region]);
  
  return result;
}

