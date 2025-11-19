/**
 * VM Actions Hook
 * VM 액션 (시작, 중지, 삭제) 로직 및 실시간 업데이트
 */

import { useCallback } from 'react';
import { useQueryClient, type UseMutationResult } from '@tanstack/react-query';
import type { VM } from '@/lib/types';
import { getLiveRegionMessage } from '@/lib/accessibility';
import { queryKeys } from '@/lib/query';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';

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
  const { success } = useToast();
  const { handleError } = useErrorHandler();

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

    // 실시간 업데이트
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
        // 실시간 업데이트 롤백
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

    // 실시간 업데이트
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
        // 실시간 업데이트 롤백
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

  // Bulk delete VMs
  const handleBulkDelete = useCallback(async (vmIds: string[], vms: VM[]) => {
    if (!workspaceId) return;
    
    const vmsToDelete = vms.filter(vm => vmIds.includes(vm.id));
    const deletePromises = vmsToDelete.map(vm =>
      deleteMutation.mutateAsync(vm.id)
    );

    try {
      await Promise.all(deletePromises);
      queryClient.invalidateQueries({ queryKey: queryKeys.vms.list(workspaceId) });
      success(`Successfully deleted ${vmIds.length} VM(s)`);
    } catch (error) {
      handleError(error, { operation: 'bulkDeleteVMs', resource: 'VM' });
      throw error;
    }
  }, [workspaceId, deleteMutation, queryClient, success, handleError]);

  // Bulk start VMs
  const handleBulkStart = useCallback(async (vmIds: string[], vms: VM[]) => {
    if (!workspaceId) return;
    
    const vmsToStart = vms.filter(vm => vmIds.includes(vm.id));
    const startPromises = vmsToStart.map(vm =>
      startMutation.mutateAsync(vm.id)
    );

    try {
      await Promise.all(startPromises);
      queryClient.invalidateQueries({ queryKey: queryKeys.vms.list(workspaceId) });
      success(`Successfully started ${vmIds.length} VM(s)`);
    } catch (error) {
      handleError(error, { operation: 'bulkStartVMs', resource: 'VM' });
      throw error;
    }
  }, [workspaceId, startMutation, queryClient, success, handleError]);

  // Bulk stop VMs
  const handleBulkStop = useCallback(async (vmIds: string[], vms: VM[]) => {
    if (!workspaceId) return;
    
    const vmsToStop = vms.filter(vm => vmIds.includes(vm.id));
    const stopPromises = vmsToStop.map(vm =>
      stopMutation.mutateAsync(vm.id)
    );

    try {
      await Promise.all(stopPromises);
      queryClient.invalidateQueries({ queryKey: queryKeys.vms.list(workspaceId) });
      success(`Successfully stopped ${vmIds.length} VM(s)`);
    } catch (error) {
      handleError(error, { operation: 'bulkStopVMs', resource: 'VM' });
      throw error;
    }
  }, [workspaceId, stopMutation, queryClient, success, handleError]);

  // Bulk restart VMs (stop then start)
  const handleBulkRestart = useCallback(async (vmIds: string[], vms: VM[]) => {
    if (!workspaceId) return;
    
    const vmsToRestart = vms.filter(vm => vmIds.includes(vm.id));
    
    // First stop all VMs
    const stopPromises = vmsToRestart.map(vm =>
      stopMutation.mutateAsync(vm.id)
    );
    
    try {
      await Promise.all(stopPromises);
      
      // Then start all VMs
      const startPromises = vmsToRestart.map(vm =>
        startMutation.mutateAsync(vm.id)
      );
      
      await Promise.all(startPromises);
      queryClient.invalidateQueries({ queryKey: queryKeys.vms.list(workspaceId) });
      success(`Successfully restarted ${vmIds.length} VM(s)`);
    } catch (error) {
      handleError(error, { operation: 'bulkRestartVMs', resource: 'VM' });
      throw error;
    }
  }, [workspaceId, stopMutation, startMutation, queryClient, success, handleError]);

  return {
    handleDeleteVM,
    handleStartVM,
    handleStopVM,
    handleBulkDelete,
    handleBulkStart,
    handleBulkStop,
    handleBulkRestart,
  };
}

