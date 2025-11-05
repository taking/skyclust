/**
 * Security Group Actions Hook
 * Security Group 관련 mutations 및 핸들러 통합 관리
 */

import { useStandardMutation } from '@/hooks/use-standard-mutation';
import { networkService } from '@/services/network';
import { queryKeys } from '@/lib/query-keys';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
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
  const createSecurityGroupMutation = useStandardMutation({
    mutationFn: (data: CreateSecurityGroupForm) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return networkService.createSecurityGroup(selectedProvider as CloudProvider, data);
    },
    invalidateQueries: [queryKeys.securityGroups.all],
    successMessage: 'Security group creation initiated',
    errorContext: { operation: 'createSecurityGroup', resource: 'SecurityGroup' },
    onSuccess,
  });

  // Delete Security Group mutation
  const deleteSecurityGroupMutation = useStandardMutation({
    mutationFn: async ({ securityGroupId, credentialId, region }: { securityGroupId: string; credentialId: string; region: string }) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return networkService.deleteSecurityGroup(selectedProvider as CloudProvider, securityGroupId, credentialId, region);
    },
    invalidateQueries: [queryKeys.securityGroups.all],
    successMessage: 'Security group deletion initiated',
    errorContext: { operation: 'deleteSecurityGroup', resource: 'SecurityGroup' },
  });

  const handleBulkDeleteSecurityGroups = async (securityGroupIds: string[], securityGroups: Array<{ id: string; region: string }>) => {
    if (!selectedCredentialId || !selectedProvider) return;
    
    const securityGroupsToDelete = securityGroups.filter(sg => securityGroupIds.includes(sg.id));
    const deletePromises = securityGroupsToDelete.map(securityGroup =>
      deleteSecurityGroupMutation.mutateAsync({
        securityGroupId: securityGroup.id,
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
    if (confirm(`Are you sure you want to delete this security group? This action cannot be undone.`)) {
      deleteSecurityGroupMutation.mutate({ securityGroupId, credentialId: selectedCredentialId, region });
    }
  };

  return {
    createSecurityGroupMutation,
    deleteSecurityGroupMutation,
    handleBulkDeleteSecurityGroups,
    handleDeleteSecurityGroup,
  };
}

