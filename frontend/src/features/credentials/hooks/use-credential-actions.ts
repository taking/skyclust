/**
 * Credential Actions Hook
 * Credential 관련 mutations 및 핸들러 통합 관리
 * Use Case 패턴 적용
 */

import { useMemo } from 'react';
import { useStandardMutation } from '@/hooks/use-standard-mutation';
import { credentialRepository } from '@/infrastructure/repositories';
import { CreateCredentialUseCase } from '@/domain/use-cases';
import { queryKeys } from '@/lib/query';
import type { CreateCredentialForm } from '@/lib/types';

export interface UseCredentialActionsOptions {
  workspaceId: string | undefined;
  onSuccess?: () => void;
}

export function useCredentialActions({ workspaceId, onSuccess }: UseCredentialActionsOptions) {
  // Use Case 생성 (메모이제이션으로 재생성 방지)
  const createCredentialUseCase = useMemo(
    () => new CreateCredentialUseCase(credentialRepository),
    []
  );

  // Create credential mutation
  const createCredentialMutation = useStandardMutation({
    mutationFn: (data: CreateCredentialForm & { workspace_id?: string; name?: string }) => {
      if (!workspaceId) throw new Error('Workspace ID is required');
      return createCredentialUseCase.execute({
        workspaceId,
        data,
        name: data.name,
      });
    },
    invalidateQueries: [queryKeys.credentials.list(workspaceId)],
    successMessage: 'Credential created successfully',
    errorContext: { operation: 'createCredential', resource: 'Credential' },
    onSuccess,
  });

  // Create credential from file mutation (for GCP)
  const createCredentialFromFileMutation = useStandardMutation({
    mutationFn: ({ workspaceId, name, provider, file }: { workspaceId: string; name: string; provider: string; file: File }) =>
      credentialRepository.createFromFile(workspaceId, name, provider, file),
    invalidateQueries: [queryKeys.credentials.list(workspaceId)],
    successMessage: 'Credential created successfully from file',
    errorContext: { operation: 'createCredentialFromFile', resource: 'Credential' },
    onSuccess,
  });

  // Update credential mutation
  const updateCredentialMutation = useStandardMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CreateCredentialForm> }) =>
      credentialRepository.update(id, data),
    invalidateQueries: [queryKeys.credentials.list(workspaceId)],
    successMessage: 'Credential updated successfully',
    errorContext: { operation: 'updateCredential', resource: 'Credential' },
    onSuccess,
  });

  // Delete credential mutation
  const deleteCredentialMutation = useStandardMutation({
    mutationFn: (id: string) => credentialRepository.delete(id),
    invalidateQueries: [queryKeys.credentials.list(workspaceId)],
    successMessage: 'Credential deleted successfully',
    errorContext: { operation: 'deleteCredential', resource: 'Credential' },
  });

  return {
    createCredentialMutation,
    createCredentialFromFileMutation,
    updateCredentialMutation,
    deleteCredentialMutation,
  };
}

