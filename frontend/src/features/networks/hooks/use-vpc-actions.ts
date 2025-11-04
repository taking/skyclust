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

  // Create VPC mutation
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
    if (confirm(`Are you sure you want to delete this VPC? This action cannot be undone.`)) {
      deleteVPCMutation.mutate({ vpcId, credentialId: selectedCredentialId, region });
    }
  };

  return {
    createVPCMutation,
    deleteVPCMutation,
    handleBulkDeleteVPCs,
    handleDeleteVPC,
  };
}

