/**
 * VMs Hook
 * VM 데이터 fetching 및 mutations 관리
 */

import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useStandardMutation } from '@/hooks/use-standard-mutation';
import { vmService } from '../services/vm';
import type { CreateVMForm } from '@/lib/types';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import { useTranslation } from '@/hooks/use-translation';
import { useCredentials } from '@/hooks/use-credentials';

interface UseVMsOptions {
  workspaceId?: string;
  selectedCredentialId?: string;
}

export function useVMs({
  workspaceId,
  selectedCredentialId,
}: UseVMsOptions) {
  const { t } = useTranslation();

  // Fetch credentials using unified hook
  const { credentials, selectedProvider } = useCredentials({
    workspaceId,
    selectedCredentialId,
  });

  // Fetch VMs (SSE 이벤트로 실시간 업데이트)
  const { data: allVms = [], isLoading } = useQuery({
    queryKey: queryKeys.vms.list(workspaceId),
    queryFn: () => workspaceId ? vmService.getVMs(workspaceId) : Promise.resolve([]),
    enabled: !!workspaceId,
    staleTime: CACHE_TIMES.REALTIME, // 30초 - VM 상태는 빠르게 변경될 수 있음
    gcTime: GC_TIMES.SHORT, // 5분 - GC 시간
    // refetchInterval 제거: SSE 이벤트로 자동 업데이트
  });

  // Filter VMs by selected credential's provider
  const vms = useMemo(() => {
    if (!selectedCredentialId || !selectedProvider) {
      return allVms;
    }
    return allVms.filter(vm => vm.provider === selectedProvider);
  }, [allVms, selectedCredentialId, selectedProvider]);

  // Create VM mutation
  const createVMMutation = useStandardMutation({
    mutationFn: ({ workspaceId: wsId, data }: { workspaceId: string; data: CreateVMForm }) => {
      return vmService.createVM({
        ...data,
        workspace_id: wsId,
      } as CreateVMForm & { workspace_id: string });
    },
    invalidateQueries: [queryKeys.vms.all],
    successMessage: t('vm.creationInitiated'),
    errorContext: { operation: 'createVM', resource: 'VM' },
  });

  // Delete VM mutation
  const deleteVMMutation = useStandardMutation({
    mutationFn: (id: string) => vmService.deleteVM(id),
    invalidateQueries: [queryKeys.vms.all],
    successMessage: t('messages.deletionInitiated', { resource: 'VM' }),
    errorContext: { operation: 'deleteVM', resource: 'VM' },
  });

  // Start VM mutation
  const startVMMutation = useStandardMutation({
    mutationFn: (id: string) => vmService.startVM(id),
    invalidateQueries: [queryKeys.vms.all],
    successMessage: t('messages.operationSuccess'),
    errorContext: { operation: 'startVM', resource: 'VM' },
  });

  // Stop VM mutation
  const stopVMMutation = useStandardMutation({
    mutationFn: (id: string) => vmService.stopVM(id),
    invalidateQueries: [queryKeys.vms.all],
    successMessage: t('messages.operationSuccess'),
    errorContext: { operation: 'stopVM', resource: 'VM' },
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

