/**
 * VMs Hook
 * VM 데이터 fetching 및 mutations 관리
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { vmService } from '../services/vm';
import { credentialService } from '@/services/credential';
import type { VM, CreateVMForm } from '@/lib/types';

interface UseVMsOptions {
  workspaceId?: string;
  selectedCredentialId?: string;
}

export function useVMs({
  workspaceId,
  selectedCredentialId,
}: UseVMsOptions) {
  const queryClient = useQueryClient();

  // Fetch credentials (자주 변경되지 않으므로 긴 staleTime)
  const { data: credentials = [] } = useQuery({
    queryKey: ['credentials', workspaceId],
    queryFn: () => workspaceId ? credentialService.getCredentials(workspaceId) : Promise.resolve([]),
    enabled: !!workspaceId,
    staleTime: 10 * 60 * 1000, // 10분 - 자격 증명은 자주 변경되지 않음
    gcTime: 30 * 60 * 1000, // 30분 - GC 시간
  });

  // Fetch VMs (실시간 상태 변화를 반영하기 위해 짧은 staleTime과 polling)
  const { data: vms = [], isLoading } = useQuery({
    queryKey: ['vms', workspaceId, selectedCredentialId],
    queryFn: () => workspaceId ? vmService.getVMs(workspaceId) : Promise.resolve([]),
    enabled: !!workspaceId,
    staleTime: 30 * 1000, // 30초 - VM 상태는 빠르게 변경될 수 있음
    gcTime: 5 * 60 * 1000, // 5분 - GC 시간
    refetchInterval: 30000, // Poll every 30 seconds
    refetchIntervalInBackground: false, // 백그라운드 polling 비활성화
  });

  // Create VM mutation
  const createVMMutation = useMutation({
    mutationFn: ({ workspaceId: wsId, data }: { workspaceId: string; data: CreateVMForm }) => {
      return vmService.createVM({
        ...data,
        workspace_id: wsId,
      } as CreateVMForm & { workspace_id: string });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vms'] });
    },
  });

  // Delete VM mutation
  const deleteVMMutation = useMutation({
    mutationFn: vmService.deleteVM,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vms'] });
    },
  });

  // Start VM mutation
  const startVMMutation = useMutation({
    mutationFn: vmService.startVM,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vms'] });
    },
  });

  // Stop VM mutation
  const stopVMMutation = useMutation({
    mutationFn: vmService.stopVM,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vms'] });
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

