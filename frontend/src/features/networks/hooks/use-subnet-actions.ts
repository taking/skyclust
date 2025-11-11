/**
 * Subnet Actions Hook
 * Subnet 관련 mutations 및 핸들러 통합 관리
 * useResourceMutations 통합
 */

import { useCallback } from 'react';
import { useResourceMutation } from '@/hooks/use-resource-mutation';
import { networkService } from '@/services/network';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { ErrorHandler } from '@/lib/error-handling';
import type { CreateSubnetForm, CloudProvider } from '@/lib/types';

export interface UseSubnetActionsOptions {
  selectedProvider: string | undefined;
  selectedCredentialId: string;
  onSuccess?: () => void;
}

export function useSubnetActions({
  selectedProvider,
  selectedCredentialId,
  onSuccess,
}: UseSubnetActionsOptions) {
  const { success } = useToast();
  const { handleError } = useErrorHandler();

  // Create mutation
  const createSubnetMutation = useResourceMutation({
    resourceType: 'subnets',
    operation: 'create',
    mutationFn: (data: CreateSubnetForm) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return networkService.createSubnet(selectedProvider as CloudProvider, data);
    },
    successMessage: 'Subnet creation initiated',
    errorContext: { operation: 'createSubnet', resource: 'Subnet' },
    onSuccess,
    onError: (error) => {
      const errorMessage = ErrorHandler.getNetworkErrorMessage(
        error,
        'create',
        'Subnet',
        selectedProvider
      );
      handleError(error, {
        operation: 'createSubnet',
        resource: 'Subnet',
        customMessage: errorMessage,
      });
    },
  });

  // Delete mutation (내부적으로 id 사용, 외부에서는 subnetId로 래핑)
  const deleteSubnetMutationBase = useResourceMutation({
    resourceType: 'subnets',
    operation: 'delete',
    mutationFn: ({ id, credentialId, region }: { id: string; credentialId: string; region: string }) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return networkService.deleteSubnet(selectedProvider as CloudProvider, id, credentialId, region);
    },
    successMessage: 'Subnet deletion initiated',
    errorContext: { operation: 'deleteSubnet', resource: 'Subnet' },
    onError: (error) => {
      const errorMessage = ErrorHandler.getNetworkErrorMessage(
        error,
        'delete',
        'Subnet',
        selectedProvider
      );
      handleError(error, {
        operation: 'deleteSubnet',
        resource: 'Subnet',
        customMessage: errorMessage,
      });
    },
  });

  // Subnet 전용 delete mutation (subnetId 사용)
  const deleteSubnetMutation = {
    ...deleteSubnetMutationBase,
    mutate: (params: { subnetId: string; credentialId: string; region: string }) => {
      deleteSubnetMutationBase.mutate({
        id: params.subnetId,
        credentialId: params.credentialId,
        region: params.region,
      });
    },
    mutateAsync: (params: { subnetId: string; credentialId: string; region: string }) => {
      return deleteSubnetMutationBase.mutateAsync({
        id: params.subnetId,
        credentialId: params.credentialId,
        region: params.region,
      });
    },
  };

  // Subnet 전용 executeDelete (subnetId 사용)
  const executeDeleteSubnet = useCallback((subnetId: string, region: string) => {
    if (!selectedCredentialId) return;
    deleteSubnetMutationBase.mutate({
      id: subnetId,
      credentialId: selectedCredentialId,
      region,
    });
  }, [selectedCredentialId, deleteSubnetMutationBase]);

  // Subnet 전용 bulk delete (subnetId 사용)
  const handleBulkDeleteSubnets = useCallback(async (subnetIds: string[], subnets: Array<{ id: string; region: string }>) => {
    if (!selectedCredentialId || !selectedProvider) return;
    
    const itemsToDelete = subnets.filter((item) => subnetIds.includes(item.id));
    const deletePromises = itemsToDelete.map((item) =>
      deleteSubnetMutationBase.mutateAsync({
        id: item.id,
        credentialId: selectedCredentialId,
        region: item.region,
      })
    );

    try {
      await Promise.all(deletePromises);
      success(`Successfully initiated deletion of ${subnetIds.length} Subnet(s)`);
    } catch (error) {
      handleError(error, {
        operation: 'bulkDeleteSubnets',
        resource: 'Subnet',
      });
      throw error;
    }
  }, [selectedCredentialId, selectedProvider, deleteSubnetMutationBase, success, handleError]);

  const handleDeleteSubnet = (subnetId: string, region: string) => {
    if (!selectedCredentialId) return;
    // 모달은 컴포넌트에서 관리하므로 여기서는 바로 삭제 실행
    executeDeleteSubnet(subnetId, region);
  };

  return {
    createSubnetMutation,
    deleteSubnetMutation,
    handleBulkDeleteSubnets,
    handleDeleteSubnet,
    executeDeleteSubnet,
  };
}

