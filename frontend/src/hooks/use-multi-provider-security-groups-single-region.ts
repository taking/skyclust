/**
 * useMultiProviderSecurityGroupsSingleRegion Hook
 * Multi-provider Security Group 조회 및 병합 (단일 Region 모드)
 */

import { useMemo } from 'react';
import { useQueries } from '@tanstack/react-query';
import { networkService } from '@/services/network';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import type { CloudProvider } from '@/lib/types/kubernetes';
import type { SecurityGroup } from '@/lib/types';

interface UseMultiProviderSecurityGroupsSingleRegionOptions {
  workspaceId: string;
  credentialIds: string[];
  credentials: Array<{ id: string; provider: CloudProvider }>;
  vpcId: string;
  region?: string;
  enabled?: boolean;
}

export function useMultiProviderSecurityGroupsSingleRegion({
  workspaceId,
  credentialIds,
  credentials,
  vpcId,
  region,
  enabled = true,
}: UseMultiProviderSecurityGroupsSingleRegionOptions) {
  
  const queries = useMemo(() => {
    if (!vpcId || !region) return [];
    
    return credentialIds.map(credentialId => {
      const credential = credentials.find(c => c.id === credentialId);
      if (!credential) {
        return {
          queryKey: queryKeys.securityGroups.list('aws' as CloudProvider, credentialId, vpcId, region),
          queryFn: async () => [] as SecurityGroup[],
          enabled: false,
        };
      }
      
      const provider = credential.provider as CloudProvider;
      
      return {
        queryKey: queryKeys.securityGroups.list(provider, credentialId, vpcId, region),
        queryFn: async () => {
          return networkService.listSecurityGroups(provider, credentialId, vpcId, region);
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
    const allSecurityGroups: Array<SecurityGroup & { provider?: CloudProvider; credential_id?: string }> = [];
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
        const securityGroupsWithProvider = result.data.map(sg => ({
          ...sg,
          provider,
          credential_id: credentialId,
        }));
        allSecurityGroups.push(...securityGroupsWithProvider);
      }
    });
    
    return {
      securityGroups: allSecurityGroups,
      isLoading: results.some(r => r.isLoading),
      errors,
      hasError: errors.length > 0,
      totalCount: allSecurityGroups.length,
    };
  }, [results, credentialIds, credentials, region]);
  
  return result;
}

