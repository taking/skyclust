/**
 * Subnet Actions Hook
 * Subnet 관련 mutations 및 핸들러 통합 관리
 */

import { useStandardMutation } from '@/hooks/use-standard-mutation';
import { networkService } from '@/services/network';
import { queryKeys } from '@/lib/query-keys';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import type { CreateSubnetForm } from '@/lib/types';

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

  // Create Subnet mutation
  const createSubnetMutation = useStandardMutation({
    mutationFn: (data: CreateSubnetForm) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return networkService.createSubnet(selectedProvider, data);
    },
    invalidateQueries: [queryKeys.subnets.all],
    successMessage: 'Subnet creation initiated',
    errorContext: { operation: 'createSubnet', resource: 'Subnet' },
    onSuccess,
  });

  // Delete Subnet mutation
  const deleteSubnetMutation = useStandardMutation({
    mutationFn: async ({ subnetId, credentialId, region }: { subnetId: string; credentialId: string; region: string }) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return networkService.deleteSubnet(selectedProvider, subnetId, credentialId, region);
    },
    invalidateQueries: [queryKeys.subnets.all],
    successMessage: 'Subnet deletion initiated',
    errorContext: { operation: 'deleteSubnet', resource: 'Subnet' },
  });

  const handleBulkDeleteSubnets = async (subnetIds: string[], subnets: Array<{ id: string; region: string }>) => {
    if (!selectedCredentialId || !selectedProvider) return;
    
    const subnetsToDelete = subnets.filter(s => subnetIds.includes(s.id));
    const deletePromises = subnetsToDelete.map(subnet =>
      deleteSubnetMutation.mutateAsync({
        subnetId: subnet.id,
        credentialId: selectedCredentialId,
        region: subnet.region,
      })
    );

    try {
      await Promise.all(deletePromises);
      success(`Successfully initiated deletion of ${subnetIds.length} subnet(s)`);
    } catch (error) {
      handleError(error, { operation: 'bulkDeleteSubnets', resource: 'Subnet' });
      throw error;
    }
  };

  const handleDeleteSubnet = (subnetId: string, region: string) => {
    if (!selectedCredentialId) return;
    if (confirm(`Are you sure you want to delete this subnet? This action cannot be undone.`)) {
      deleteSubnetMutation.mutate({ subnetId, credentialId: selectedCredentialId, region });
    }
  };

  return {
    createSubnetMutation,
    deleteSubnetMutation,
    handleBulkDeleteSubnets,
    handleDeleteSubnet,
  };
}

