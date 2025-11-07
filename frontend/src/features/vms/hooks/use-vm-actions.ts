/**
 * VM Actions Hook
 * VM 액션 (시작, 중지, 삭제) 로직 및 optimistic updates
 */

import { useQueryClient, type UseMutationResult } from '@tanstack/react-query';
import type { VM } from '@/lib/types';
import { getLiveRegionMessage } from '@/lib/accessibility';
import { queryKeys } from '@/lib/query-keys';

interface UseVMActionsOptions {
  workspaceId?: string;
  deleteMutation: UseMutationResult<void, unknown, string, unknown>;
  startMutation: UseMutationResult<void, unknown, string, unknown>;
  stopMutation: UseMutationResult<void, unknown, string, unknown>;
  onSuccess?: (message: string) => void;
  onError?: (message: string) => void;
  setLiveMessage?: (message: string) => void;
}

export function useVMActions({
  workspaceId,
  deleteMutation,
  startMutation,
  stopMutation,
  onSuccess,
  onError,
  setLiveMessage,
}: UseVMActionsOptions) {
  const queryClient = useQueryClient();

  const handleDeleteVM = (vmId: string, vms: VM[]) => {
    const vm = vms.find(v => v.id === vmId);
    // 모달은 컴포넌트에서 관리하므로 여기서는 바로 삭제 실행
    deleteMutation.mutate(vmId, {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: queryKeys.vms.list(workspaceId) });
        onSuccess?.('VM deleted successfully');
        if (setLiveMessage) {
          setLiveMessage(getLiveRegionMessage('deleted', vm?.name || 'VM', true));
        }
      },
      onError: (error: unknown) => {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error';
        onError?.(`Failed to delete VM: ${errorMessage}`);
        if (setLiveMessage) {
          setLiveMessage(getLiveRegionMessage('deleted', vm?.name || 'VM', false));
        }
      },
    });
  };

  const handleStartVM = (vmId: string, vms: VM[]) => {
    const vm = vms.find(v => v.id === vmId);
    if (!vm) return;

    // Optimistic update
    queryClient.setQueryData(['vms', workspaceId], (old: VM[] | undefined) => {
      if (!old) return old;
      return old.map(v => v.id === vmId ? { ...v, status: 'starting' as const } : v);
    });

    startMutation.mutate(vmId, {
      onSuccess: () => {
        // Invalidate to get fresh data from server
        queryClient.invalidateQueries({ queryKey: ['vms', workspaceId] });
        onSuccess?.('VM started successfully');
        if (setLiveMessage) {
          setLiveMessage(getLiveRegionMessage('started', vm?.name || 'VM', true));
        }
      },
      onError: (error: unknown) => {
        // Rollback optimistic update
        queryClient.setQueryData(['vms', workspaceId], (old: VM[] | undefined) => {
          if (!old) return old;
          return old.map(v => v.id === vmId ? { ...v, status: vm.status } : v);
        });
        const errorMessage = error instanceof Error ? error.message : 'Unknown error';
        onError?.(`Failed to start VM: ${errorMessage}`);
        if (setLiveMessage) {
          setLiveMessage(getLiveRegionMessage('started', vm?.name || 'VM', false));
        }
      },
    });
  };

  const handleStopVM = (vmId: string, vms: VM[]) => {
    const vm = vms.find(v => v.id === vmId);
    if (!vm) return;

    // Optimistic update
    queryClient.setQueryData(['vms', workspaceId], (old: VM[] | undefined) => {
      if (!old) return old;
      return old.map(v => v.id === vmId ? { ...v, status: 'stopping' as const } : v);
    });

    stopMutation.mutate(vmId, {
      onSuccess: () => {
        // Invalidate to get fresh data from server
        queryClient.invalidateQueries({ queryKey: ['vms', workspaceId] });
        onSuccess?.('VM stopped successfully');
        if (setLiveMessage) {
          setLiveMessage(getLiveRegionMessage('stopped', vm?.name || 'VM', true));
        }
      },
      onError: (error: unknown) => {
        // Rollback optimistic update
        queryClient.setQueryData(['vms', workspaceId], (old: VM[] | undefined) => {
          if (!old) return old;
          return old.map(v => v.id === vmId ? { ...v, status: vm.status } : v);
        });
        const errorMessage = error instanceof Error ? error.message : 'Unknown error';
        onError?.(`Failed to stop VM: ${errorMessage}`);
        if (setLiveMessage) {
          setLiveMessage(getLiveRegionMessage('stopped', vm?.name || 'VM', false));
        }
      },
    });
  };

  return {
    handleDeleteVM,
    handleStartVM,
    handleStopVM,
  };
}

