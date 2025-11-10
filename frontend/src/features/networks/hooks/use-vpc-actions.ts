/**
 * VPC Actions Hook
 * VPC 관련 mutations 및 핸들러 통합 관리
 * Use Case 패턴 적용
 */

import { useMemo } from 'react';
import { useStandardMutation } from '@/hooks/use-standard-mutation';
import { queryKeys } from '@/lib/query-keys';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { vpcRepository } from '@/infrastructure/repositories';
import { CreateVPCUseCase, BulkDeleteVPCsUseCase } from '@/domain/use-cases';
import type { CreateVPCForm, CloudProvider } from '@/lib/types';
import { getVPCCreationErrorMessage, getVPCDeletionErrorMessage } from '@/lib/network-error-messages';

export interface UseVPCActionsOptions {
  selectedProvider: string | undefined;
  selectedCredentialId: string;
  selectedRegion: string;
  onSuccess?: () => void;
}

export function useVPCActions({
  selectedProvider,
  selectedCredentialId,
  selectedRegion,
  onSuccess,
}: UseVPCActionsOptions) {
  const { success } = useToast();
  const { handleError } = useErrorHandler();

  // Use Cases 생성 (메모이제이션으로 재생성 방지)
  const createVPCUseCase = useMemo(
    () => new CreateVPCUseCase(vpcRepository),
    []
  );

  const bulkDeleteVPCsUseCase = useMemo(
    () => new BulkDeleteVPCsUseCase(vpcRepository),
    []
  );

  // Create VPC mutation with enhanced error handling
  const createVPCMutation = useStandardMutation({
    mutationFn: (data: CreateVPCForm) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return createVPCUseCase.execute({
        provider: selectedProvider as CloudProvider,
        data,
      });
    },
    invalidateQueries: [queryKeys.vpcs.all],
    successMessage: 'VPC creation initiated',
    errorContext: { operation: 'createVPC', resource: 'VPC' },
    onSuccess,
    onError: (error) => {
      // Enhanced error handling for VPC creation
      const errorMessage = getVPCCreationErrorMessage(error, selectedProvider);
      handleError(error, { 
        operation: 'createVPC', 
        resource: 'VPC',
        customMessage: errorMessage,
      });
    },
  });

  // Delete VPC mutation (단일 삭제는 기존 방식 유지 - Use Case는 bulk에만 적용)
  const deleteVPCMutation = useStandardMutation({
    mutationFn: async ({ vpcId, credentialId, region }: { vpcId: string; credentialId: string; region: string }) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return vpcRepository.delete(selectedProvider as CloudProvider, vpcId, credentialId, region);
    },
    invalidateQueries: [queryKeys.vpcs.all],
    successMessage: 'VPC deletion initiated',
    errorContext: { operation: 'deleteVPC', resource: 'VPC' },
    onError: (error) => {
      // Enhanced error handling for VPC deletion
      const errorMessage = getVPCDeletionErrorMessage(error, selectedProvider);
      handleError(error, { 
        operation: 'deleteVPC', 
        resource: 'VPC',
        customMessage: errorMessage,
      });
    },
  });

  const handleBulkDeleteVPCs = async (vpcIds: string[], vpcs: Array<{ id: string; region?: string }>) => {
    if (!selectedCredentialId || !selectedProvider) return;

    try {
      await bulkDeleteVPCsUseCase.execute({
        provider: selectedProvider as CloudProvider,
        vpcIds,
        vpcs,
        credentialId: selectedCredentialId,
        defaultRegion: selectedRegion || '',
      });
      success(`Successfully initiated deletion of ${vpcIds.length} VPC(s)`);
    } catch (error) {
      handleError(error, { operation: 'bulkDeleteVPCs', resource: 'VPC' });
      throw error;
    }
  };

  const handleDeleteVPC = (vpcId: string, region?: string) => {
    if (!selectedCredentialId || !region) return;
    // 모달은 컴포넌트에서 관리하므로 여기서는 바로 삭제 실행
    // 컴포넌트에서 모달 확인 후 이 함수를 호출하도록 수정 필요
    deleteVPCMutation.mutate({ vpcId, credentialId: selectedCredentialId, region });
  };

  // 모달 없이 직접 삭제 실행하는 함수 (컴포넌트에서 모달 확인 후 호출)
  const executeDeleteVPC = (vpcId: string, region: string) => {
    if (!selectedCredentialId) return;
    deleteVPCMutation.mutate({ vpcId, credentialId: selectedCredentialId, region });
  };

  return {
    createVPCMutation,
    deleteVPCMutation,
    handleBulkDeleteVPCs,
    handleDeleteVPC,
    executeDeleteVPC,
  };
}

