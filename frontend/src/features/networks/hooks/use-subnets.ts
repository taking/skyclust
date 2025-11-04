/**
 * Subnets Hook
 * Subnet 데이터 fetching 및 상태 관리
 */

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { networkService } from '@/services/network';
import { queryKeys } from '@/lib/query-keys';
import { CACHE_TIMES, GC_TIMES } from '@/lib/query-client';
import { useCredentials } from '@/hooks/use-credentials';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useWorkspaceStore } from '@/store/workspace';

export function useSubnets() {
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

  // Fetch Subnets
  const { data: subnets = [], isLoading: isLoadingSubnets } = useQuery({
    queryKey: queryKeys.subnets.list(selectedProvider, watchedCredentialId, selectedVPCId, watchedRegion),
    queryFn: async () => {
      if (!selectedProvider || !watchedCredentialId || !selectedVPCId || !watchedRegion) return [];
      return networkService.listSubnets(selectedProvider, watchedCredentialId, selectedVPCId, watchedRegion);
    },
    enabled: !!selectedProvider && !!watchedCredentialId && !!selectedVPCId && !!watchedRegion && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: 30000,
  });

  return {
    subnets,
    isLoadingSubnets,
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

