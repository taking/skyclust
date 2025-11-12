/**
 * Resource Group Actions Hook
 * Azure Resource Group 관련 mutations 및 핸들러 통합 관리
 */

import { useCallback } from 'react';
import { useResourceMutation } from '@/hooks/use-resource-mutation';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { useQueryClient } from '@tanstack/react-query';
import { resourceGroupService } from '@/services/resource-group';
import { queryKeys } from '@/lib/query';
import type { ResourceGroupInfo } from '@/services/resource-group';

export interface UseResourceGroupActionsOptions {
  selectedCredentialId: string;
  onSuccess?: () => void;
}

export interface CreateResourceGroupForm {
  credential_id: string;
  name: string;
  location: string;
  tags?: Record<string, string>;
}

export function useResourceGroupActions({
  selectedCredentialId,
  onSuccess,
}: UseResourceGroupActionsOptions) {
  const { success } = useToast();
  const { handleError } = useErrorHandler();
  const queryClient = useQueryClient();

  // Create mutation
  const createResourceGroupMutation = useResourceMutation({
    resourceType: 'resource-groups',
    operation: 'create',
    mutationFn: (data: CreateResourceGroupForm) => {
      return resourceGroupService.createResourceGroup(data);
    },
    successMessage: 'Resource Group creation initiated',
    errorContext: { operation: 'createResourceGroup', resource: 'Resource Group' },
    onSuccess: () => {
      queryClient.invalidateQueries({ 
        queryKey: queryKeys.azureResourceGroups.all 
      });
      onSuccess?.();
    },
    onError: (error) => {
      handleError(error, {
        operation: 'createResourceGroup',
        resource: 'Resource Group',
      });
    },
  });

  // Delete mutation
  const deleteResourceGroupMutation = useResourceMutation({
    resourceType: 'resource-groups',
    operation: 'delete',
    mutationFn: ({ name, credentialId }: { name: string; credentialId: string }) => {
      return resourceGroupService.deleteResourceGroup(name, credentialId);
    },
    successMessage: 'Resource Group deletion initiated',
    errorContext: { operation: 'deleteResourceGroup', resource: 'Resource Group' },
    onSuccess: () => {
      queryClient.invalidateQueries({ 
        queryKey: queryKeys.azureResourceGroups.all 
      });
      onSuccess?.();
    },
    onError: (error) => {
      handleError(error, {
        operation: 'deleteResourceGroup',
        resource: 'Resource Group',
      });
    },
  });

  // Execute delete helper
  const executeDeleteResourceGroup = useCallback((name: string) => {
    if (!selectedCredentialId) return;
    deleteResourceGroupMutation.mutate({
      name,
      credentialId: selectedCredentialId,
    });
  }, [selectedCredentialId, deleteResourceGroupMutation]);

  // Bulk delete handler
  const handleBulkDeleteResourceGroups = useCallback(async (
    names: string[],
    resourceGroups: ResourceGroupInfo[]
  ) => {
    if (!selectedCredentialId) return;

    const itemsToDelete = resourceGroups.filter((rg) => names.includes(rg.name));
    const deletePromises = itemsToDelete.map((rg) =>
      deleteResourceGroupMutation.mutateAsync({
        name: rg.name,
        credentialId: selectedCredentialId,
      })
    );

    try {
      await Promise.all(deletePromises);
      success(`Successfully initiated deletion of ${names.length} Resource Group(s)`);
    } catch (error) {
      handleError(error, {
        operation: 'bulkDeleteResourceGroups',
        resource: 'Resource Group',
      });
      throw error;
    }
  }, [selectedCredentialId, deleteResourceGroupMutation, success, handleError]);

  return {
    createResourceGroupMutation,
    deleteResourceGroupMutation,
    executeDeleteResourceGroup,
    handleBulkDeleteResourceGroups,
  };
}

