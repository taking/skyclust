/**
 * Security Groups Hook
 * Security Group 데이터 fetching 및 상태 관리
 */

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { networkService } from '@/services/network';
import { queryKeys } from '@/lib/query-keys';
import { CACHE_TIMES, GC_TIMES } from '@/lib/query-client';
import { useCredentials } from '@/hooks/use-credentials';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useWorkspaceStore } from '@/store/workspace';

export function useSecurityGroups() {
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedCredentialId, selectedRegion } = useCredentialContext();
  const [selectedVPCId, setSelectedVPCId] = useState<string>('');

  const { credentials, selectedCredential, selectedProvider } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
  });

  const watchedCredentialId = selectedCredentialId || '';
  const watchedRegion = selectedRegion || '';

  // Fetch VPCs for selection
  const { data: vpcs = [] } = useQuery({
    queryKey: queryKeys.vpcs.list(selectedProvider, watchedCredentialId, watchedRegion),
    queryFn: async () => {
      if (!selectedProvider || !watchedCredentialId) return [];
      return networkService.listVPCs(selectedProvider, watchedCredentialId, watchedRegion);
    },
    enabled: !!selectedProvider && !!watchedCredentialId && !!currentWorkspace,
  });

  // Fetch Security Groups
  const { data: securityGroups = [], isLoading: isLoadingSecurityGroups } = useQuery({
    queryKey: queryKeys.securityGroups.list(selectedProvider, watchedCredentialId, selectedVPCId, watchedRegion),
    queryFn: async () => {
      if (!selectedProvider || !watchedCredentialId || !currentWorkspace || !selectedVPCId || !watchedRegion) {
        return [];
      }
      // 파라미터 순서: provider, credentialId, vpcId, region
      return networkService.listSecurityGroups(selectedProvider, watchedCredentialId, selectedVPCId, watchedRegion);
    },
    enabled: !!selectedProvider && !!watchedCredentialId && !!currentWorkspace && !!selectedVPCId && !!watchedRegion,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: 30000,
  });

  return {
    securityGroups,
    isLoadingSecurityGroups,
    vpcs,
    selectedVPCId,
    setSelectedVPCId,
    credentials,
    selectedCredential,
    selectedProvider,
    selectedCredentialId: watchedCredentialId,
    selectedRegion: watchedRegion,
  };
}

