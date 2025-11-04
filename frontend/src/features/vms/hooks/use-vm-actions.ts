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
  deleteMutation: UseMutationResult<void, Error, string, unknown>;
  startMutation: UseMutationResult<void, Error, string, unknown>;
  stopMutation: UseMutationResult<void, Error, string, unknown>;
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
    if (confirm('Are you sure you want to delete this VM?')) {
      deleteMutation.mutate(vmId, {
        onSuccess: () => {
          queryClient.invalidateQueries({ queryKey: queryKeys.vms.list(workspaceId) });
          onSuccess?.('VM deleted successfully');
          if (setLiveMessage) {
            setLiveMessage(getLiveRegionMessage('deleted', vm?.name || 'VM', true));
          }
        },
        onError: (error: Error) => {
          onError?.(`Failed to delete VM: ${error.message}`);
          if (setLiveMessage) {
            setLiveMessage(getLiveRegionMessage('deleted', vm?.name || 'VM', false));
          }
        },
      });
    }
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
      onError: (error: Error) => {
        // Rollback optimistic update
        queryClient.setQueryData(['vms', workspaceId], (old: VM[] | undefined) => {
          if (!old) return old;
          return old.map(v => v.id === vmId ? { ...v, status: vm.status } : v);
        });
        onError?.(`Failed to start VM: ${error.message}`);
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
      onError: (error: Error) => {
        // Rollback optimistic update
        queryClient.setQueryData(['vms', workspaceId], (old: VM[] | undefined) => {
          if (!old) return old;
          return old.map(v => v.id === vmId ? { ...v, status: vm.status } : v);
        });
        onError?.(`Failed to stop VM: ${error.message}`);
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

