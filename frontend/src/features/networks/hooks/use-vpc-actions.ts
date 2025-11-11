/**
 * VPC Actions Hook
 * VPC 관련 mutations 및 핸들러 통합 관리
 * Use Case 패턴 적용 + useResourceMutations 통합
 */

import { useMemo, useCallback } from 'react';
import { useResourceMutation } from '@/hooks/use-resource-mutation';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { ErrorHandler } from '@/lib/error-handling';
import { vpcRepository } from '@/infrastructure/repositories';
import { CreateVPCUseCase, BulkDeleteVPCsUseCase } from '@/domain/use-cases';
import type { CreateVPCForm, CloudProvider } from '@/lib/types';

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

  // Create mutation
  const createVPCMutation = useResourceMutation({
    resourceType: 'vpcs',
    operation: 'create',
    mutationFn: (data: CreateVPCForm) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return createVPCUseCase.execute({
        provider: selectedProvider as CloudProvider,
        data,
      });
    },
    successMessage: 'VPC creation initiated',
    errorContext: { operation: 'createVPC', resource: 'VPC' },
    onSuccess,
    onError: (error) => {
      const errorMessage = ErrorHandler.getNetworkErrorMessage(
        error,
        'create',
        'VPC',
        selectedProvider
      );
      handleError(error, {
        operation: 'createVPC',
        resource: 'VPC',
        customMessage: errorMessage,
      });
    },
  });

  // Delete mutation (내부적으로 id 사용, 외부에서는 vpcId로 래핑)
  const deleteVPCMutationBase = useResourceMutation({
    resourceType: 'vpcs',
    operation: 'delete',
    mutationFn: ({ id, credentialId, region }: { id: string; credentialId: string; region: string }) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return vpcRepository.delete(selectedProvider as CloudProvider, id, credentialId, region);
    },
    successMessage: 'VPC deletion initiated',
    errorContext: { operation: 'deleteVPC', resource: 'VPC' },
    onError: (error) => {
      const errorMessage = ErrorHandler.getNetworkErrorMessage(
        error,
        'delete',
        'VPC',
        selectedProvider
      );
      handleError(error, {
        operation: 'deleteVPC',
        resource: 'VPC',
        customMessage: errorMessage,
      });
    },
  });

  // VPC 전용 delete mutation (vpcId 사용)
  const deleteVPCMutation = {
    ...deleteVPCMutationBase,
    mutate: (params: { vpcId: string; credentialId: string; region: string }) => {
      deleteVPCMutationBase.mutate({
        id: params.vpcId,
        credentialId: params.credentialId,
        region: params.region,
      });
    },
    mutateAsync: (params: { vpcId: string; credentialId: string; region: string }) => {
      return deleteVPCMutationBase.mutateAsync({
        id: params.vpcId,
        credentialId: params.credentialId,
        region: params.region,
      });
    },
  };

  // VPC 전용 executeDelete (vpcId 사용)
  const executeDeleteVPC = useCallback((vpcId: string, region: string) => {
    if (!selectedCredentialId) return;
    deleteVPCMutationBase.mutate({
      id: vpcId,
      credentialId: selectedCredentialId,
      region,
    });
  }, [selectedCredentialId, deleteVPCMutationBase]);

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
    executeDeleteVPC(vpcId, region);
  };

  return {
    createVPCMutation,
    deleteVPCMutation,
    handleBulkDeleteVPCs,
    handleDeleteVPC,
    executeDeleteVPC,
  };
}

