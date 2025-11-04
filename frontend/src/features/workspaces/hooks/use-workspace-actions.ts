/**
 * useWorkspaceActions Hook
 * Workspace 관련 mutations (create, update, delete, members)
 */

import { useQueryClient } from '@tanstack/react-query';
import { useStandardMutation } from '@/hooks/use-standard-mutation';
import { workspaceService } from '../services/workspace';
import { queryKeys } from '@/lib/query-keys';
import type { CreateWorkspaceForm, Workspace } from '@/lib/types';

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
    invalidateQueries: [{ queryKey: queryKeys.workspaces.all }],
    successMessage: 'Workspace created successfully',
    errorMessage: 'Failed to create workspace',
    onSuccess,
    onError,
  });

  // Update workspace mutation
  const updateWorkspaceMutation = useStandardMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CreateWorkspaceForm> }) =>
      workspaceService.updateWorkspace(id, data),
    invalidateQueries: [
      { queryKey: queryKeys.workspaces.all },
      { queryKey: (variables: { id: string }) => queryKeys.workspaces.detail(variables.id) },
    ],
    successMessage: 'Workspace updated successfully',
    errorMessage: 'Failed to update workspace',
    onSuccess,
    onError,
  });

  // Delete workspace mutation
  const deleteWorkspaceMutation = useStandardMutation({
    mutationFn: (id: string) => workspaceService.deleteWorkspace(id),
    invalidateQueries: [{ queryKey: queryKeys.workspaces.all }],
    successMessage: 'Workspace deleted successfully',
    errorMessage: 'Failed to delete workspace',
    onSuccess,
    onError,
  });

  // Add member mutation
  const addMemberMutation = useStandardMutation({
    mutationFn: ({ workspaceId, userId, role }: { workspaceId: string; userId: string; role?: string }) =>
      workspaceService.addMember(workspaceId, userId, role),
    invalidateQueries: [
      { queryKey: (variables: { workspaceId: string }) => queryKeys.workspaces.detail(variables.workspaceId) },
    ],
    successMessage: 'Member added successfully',
    errorMessage: 'Failed to add member',
    onSuccess,
    onError,
  });

  // Remove member mutation
  const removeMemberMutation = useStandardMutation({
    mutationFn: ({ workspaceId, userId }: { workspaceId: string; userId: string }) =>
      workspaceService.removeMember(workspaceId, userId),
    invalidateQueries: [
      { queryKey: (variables: { workspaceId: string }) => queryKeys.workspaces.detail(variables.workspaceId) },
    ],
    successMessage: 'Member removed successfully',
    errorMessage: 'Failed to remove member',
    onSuccess,
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

