/**
 * useWorkspaceActions Hook
 * Workspace 관련 mutations (create, update, delete, members)
 */

import { useQueryClient } from '@tanstack/react-query';
import { useStandardMutation } from '@/hooks/use-standard-mutation';
import { workspaceService } from '../services/workspace';
import { queryKeys } from '@/lib/query';
import type { CreateWorkspaceForm } from '@/lib/types';

interface UseWorkspaceActionsOptions {
  onSuccess?: () => void;
  onError?: (error: unknown) => void;
}

export function useWorkspaceActions(options: UseWorkspaceActionsOptions = {}) {
  const { onSuccess, onError } = options;
  const queryClient = useQueryClient();

  // Create workspace mutation
  const createWorkspaceMutation = useStandardMutation({
    mutationFn: (data: CreateWorkspaceForm) => workspaceService.createWorkspace(data),
    invalidateQueries: [queryKeys.workspaces.all],
    successMessage: 'Workspace created successfully',
    onSuccess,
    onError,
  });

  // Update workspace mutation
  const updateWorkspaceMutation = useStandardMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CreateWorkspaceForm> }) =>
      workspaceService.updateWorkspace(id, data),
    invalidateQueries: [queryKeys.workspaces.all],
    successMessage: 'Workspace updated successfully',
    onSuccess: (data, variables) => {
      // 특정 workspace detail도 무효화
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.detail(variables.id) });
      onSuccess?.();
    },
    onError,
  });

  // Delete workspace mutation
  const deleteWorkspaceMutation = useStandardMutation({
    mutationFn: (id: string) => workspaceService.deleteWorkspace(id),
    invalidateQueries: [queryKeys.workspaces.all],
    successMessage: 'Workspace deleted successfully',
    onSuccess,
    onError,
  });

  // Add member mutation
  const addMemberMutation = useStandardMutation({
    mutationFn: ({ workspaceId, email, role }: { workspaceId: string; email: string; role?: string }) =>
      workspaceService.addMember(workspaceId, email, role),
    invalidateQueries: [queryKeys.workspaces.all],
    successMessage: 'Member added successfully',
    onSuccess: (data, variables) => {
      // 특정 workspace detail도 무효화
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.detail(variables.workspaceId) });
      onSuccess?.();
    },
    onError,
  });

  // Remove member mutation
  const removeMemberMutation = useStandardMutation({
    mutationFn: ({ workspaceId, userId }: { workspaceId: string; userId: string }) =>
      workspaceService.removeMember(workspaceId, userId),
    invalidateQueries: [queryKeys.workspaces.all],
    successMessage: 'Member removed successfully',
    onSuccess: (data, variables) => {
      // 특정 workspace detail도 무효화
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.detail(variables.workspaceId) });
      onSuccess?.();
    },
    onError,
  });

  return {
    createWorkspaceMutation,
    updateWorkspaceMutation,
    deleteWorkspaceMutation,
    addMemberMutation,
    removeMemberMutation,
  };
}

