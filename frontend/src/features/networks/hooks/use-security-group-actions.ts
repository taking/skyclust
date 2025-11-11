/**
 * Security Group Actions Hook
 * Security Group 관련 mutations 및 핸들러 통합 관리
 */

import { useResourceMutation } from '@/hooks/use-resource-mutation';
import { networkService } from '@/services/network';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { ErrorHandler } from '@/lib/error-handling';
import type { CreateSecurityGroupForm, CloudProvider } from '@/lib/types';

export interface UseSecurityGroupActionsOptions {
  selectedProvider: string | undefined;
  selectedCredentialId: string;
  onSuccess?: () => void;
}

export function useSecurityGroupActions({
  selectedProvider,
  selectedCredentialId,
  onSuccess,
}: UseSecurityGroupActionsOptions) {
  const { success } = useToast();
  const { handleError } = useErrorHandler();

  // Create Security Group mutation
  const createSecurityGroupMutation = useResourceMutation({
    resourceType: 'securityGroups',
    operation: 'create',
    mutationFn: (data: CreateSecurityGroupForm) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return networkService.createSecurityGroup(selectedProvider as CloudProvider, data);
    },
    successMessage: 'Security group creation initiated',
    errorContext: { operation: 'createSecurityGroup', resource: 'SecurityGroup' },
    onSuccess,
    onError: (error) => {
      const errorMessage = ErrorHandler.getNetworkErrorMessage(
        error,
        'create',
        'SecurityGroup',
        selectedProvider
      );
      handleError(error, {
        operation: 'createSecurityGroup',
        resource: 'SecurityGroup',
        customMessage: errorMessage,
      });
    },
  });

  // Delete Security Group mutation
  const deleteSecurityGroupMutation = useResourceMutation({
    resourceType: 'securityGroups',
    operation: 'delete',
    mutationFn: async ({ id, credentialId, region }: { id: string; credentialId: string; region: string }) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return networkService.deleteSecurityGroup(selectedProvider as CloudProvider, id, credentialId, region);
    },
    successMessage: 'Security group deletion initiated',
    errorContext: { operation: 'deleteSecurityGroup', resource: 'SecurityGroup' },
    onError: (error) => {
      const errorMessage = ErrorHandler.getNetworkErrorMessage(
        error,
        'delete',
        'SecurityGroup',
        selectedProvider
      );
      handleError(error, {
        operation: 'deleteSecurityGroup',
        resource: 'SecurityGroup',
        customMessage: errorMessage,
      });
    },
  });

  // Security Group 전용 delete mutation (securityGroupId 사용)
  const deleteSecurityGroupMutationWrapped = {
    ...deleteSecurityGroupMutation,
    mutate: (params: { securityGroupId: string; credentialId: string; region: string }) => {
      deleteSecurityGroupMutation.mutate({
        id: params.securityGroupId,
        credentialId: params.credentialId,
        region: params.region,
      });
    },
    mutateAsync: (params: { securityGroupId: string; credentialId: string; region: string }) => {
      return deleteSecurityGroupMutation.mutateAsync({
        id: params.securityGroupId,
        credentialId: params.credentialId,
        region: params.region,
      });
    },
  };

  const handleBulkDeleteSecurityGroups = async (securityGroupIds: string[], securityGroups: Array<{ id: string; region: string }>) => {
    if (!selectedCredentialId || !selectedProvider) return;
    
    const securityGroupsToDelete = securityGroups.filter(sg => securityGroupIds.includes(sg.id));
    const deletePromises = securityGroupsToDelete.map(securityGroup =>
      deleteSecurityGroupMutation.mutateAsync({
        id: securityGroup.id,
        credentialId: selectedCredentialId,
        region: securityGroup.region,
      })
    );

    try {
      await Promise.all(deletePromises);
      success(`Successfully initiated deletion of ${securityGroupIds.length} security group(s)`);
    } catch (error) {
      handleError(error, { operation: 'bulkDeleteSecurityGroups', resource: 'SecurityGroup' });
      throw error;
    }
  };

  const handleDeleteSecurityGroup = (securityGroupId: string, region: string) => {
    if (!selectedCredentialId) return;
    // 모달은 컴포넌트에서 관리하므로 여기서는 바로 삭제 실행
    deleteSecurityGroupMutationWrapped.mutate({ securityGroupId, credentialId: selectedCredentialId, region });
  };

  // 모달 없이 직접 삭제 실행하는 함수 (컴포넌트에서 모달 확인 후 호출)
  const executeDeleteSecurityGroup = (securityGroupId: string, region: string) => {
    if (!selectedCredentialId) return;
    deleteSecurityGroupMutationWrapped.mutate({ securityGroupId, credentialId: selectedCredentialId, region });
  };

  return {
    createSecurityGroupMutation,
    deleteSecurityGroupMutation: deleteSecurityGroupMutationWrapped,
    handleBulkDeleteSecurityGroups,
    handleDeleteSecurityGroup,
    executeDeleteSecurityGroup,
  };
}

