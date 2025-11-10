/**
 * Subnet Actions Hook
 * Subnet 관련 mutations 및 핸들러 통합 관리
 */

import { useStandardMutation } from '@/hooks/use-standard-mutation';
import { networkService } from '@/services/network';
import { queryKeys } from '@/lib/query-keys';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import type { CreateSubnetForm, CloudProvider } from '@/lib/types';
import { getSubnetCreationErrorMessage } from '@/lib/network-error-messages';

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

  // Create Subnet mutation with enhanced error handling
  const createSubnetMutation = useStandardMutation({
    mutationFn: (data: CreateSubnetForm) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return networkService.createSubnet(selectedProvider as CloudProvider, data);
    },
    invalidateQueries: [queryKeys.subnets.all],
    successMessage: 'Subnet creation initiated',
    errorContext: { operation: 'createSubnet', resource: 'Subnet' },
    onSuccess,
    onError: (error) => {
      // Enhanced error handling for Subnet creation
      const errorMessage = getSubnetCreationErrorMessage(error, selectedProvider);
      handleError(error, { 
        operation: 'createSubnet', 
        resource: 'Subnet',
        customMessage: errorMessage,
      });
    },
  });

  // Delete Subnet mutation
  const deleteSubnetMutation = useStandardMutation({
    mutationFn: async ({ subnetId, credentialId, region }: { subnetId: string; credentialId: string; region: string }) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return networkService.deleteSubnet(selectedProvider as CloudProvider, subnetId, credentialId, region);
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
    // 모달은 컴포넌트에서 관리하므로 여기서는 바로 삭제 실행
    deleteSubnetMutation.mutate({ subnetId, credentialId: selectedCredentialId, region });
  };

  // 모달 없이 직접 삭제 실행하는 함수 (컴포넌트에서 모달 확인 후 호출)
  const executeDeleteSubnet = (subnetId: string, region: string) => {
    if (!selectedCredentialId) return;
    deleteSubnetMutation.mutate({ subnetId, credentialId: selectedCredentialId, region });
  };

  return {
    createSubnetMutation,
    deleteSubnetMutation,
    handleBulkDeleteSubnets,
    handleDeleteSubnet,
    executeDeleteSubnet,
  };
}

