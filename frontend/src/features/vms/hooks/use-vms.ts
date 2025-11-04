/**
 * VMs Hook
 * VM 데이터 fetching 및 mutations 관리
 */

import { useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { vmService } from '../services/vm';
import type { VM, CreateVMForm } from '@/lib/types';
import { queryKeys } from '@/lib/query-keys';
import { CACHE_TIMES, GC_TIMES } from '@/lib/query-client';
import { useCredentials } from '@/hooks/use-credentials';

interface UseVMsOptions {
  workspaceId?: string;
  selectedCredentialId?: string;
}

export function useVMs({
  workspaceId,
  selectedCredentialId,
}: UseVMsOptions) {
  const queryClient = useQueryClient();

  // Fetch credentials using unified hook
  const { credentials, selectedCredential, selectedProvider } = useCredentials({
    workspaceId,
    selectedCredentialId,
  });

  // Fetch VMs (실시간 상태 변화를 반영하기 위해 짧은 staleTime과 polling)
  const { data: allVms = [], isLoading } = useQuery({
    queryKey: queryKeys.vms.list(workspaceId),
    queryFn: () => workspaceId ? vmService.getVMs(workspaceId) : Promise.resolve([]),
    enabled: !!workspaceId,
    staleTime: CACHE_TIMES.REALTIME, // 30초 - VM 상태는 빠르게 변경될 수 있음
    gcTime: GC_TIMES.SHORT, // 5분 - GC 시간
    refetchInterval: 30000, // Poll every 30 seconds
    refetchIntervalInBackground: false, // 백그라운드 polling 비활성화
  });

  // Filter VMs by selected credential's provider
  const vms = useMemo(() => {
    if (!selectedCredentialId || !selectedProvider) {
      return allVms;
    }
    return allVms.filter(vm => vm.provider === selectedProvider);
  }, [allVms, selectedCredentialId, selectedProvider]);

  // Create VM mutation
  const createVMMutation = useMutation({
    mutationFn: ({ workspaceId: wsId, data }: { workspaceId: string; data: CreateVMForm }) => {
      return vmService.createVM({
        ...data,
        workspace_id: wsId,
      } as CreateVMForm & { workspace_id: string });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.vms.all });
    },
  });

  // Delete VM mutation
  const deleteVMMutation = useMutation({
    mutationFn: (id: string) => vmService.deleteVM(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.vms.all });
    },
  });

  // Start VM mutation
  const startVMMutation = useMutation({
    mutationFn: (id: string) => vmService.startVM(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.vms.all });
    },
  });

  // Stop VM mutation
  const stopVMMutation = useMutation({
    mutationFn: (id: string) => vmService.stopVM(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.vms.all });
    },
  });

  return {
    credentials,
    vms,
    isLoading,
    createVMMutation,
    deleteVMMutation,
    startVMMutation,
    stopVMMutation,
  };
}

