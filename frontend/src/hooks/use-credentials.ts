/**
 * Credentials Hook
 * 
 * Credential 데이터 fetching을 위한 공통 hook
 * 모든 페이지에서 동일한 패턴으로 credentials를 가져올 수 있도록 통합
 * 
 * @example
 * ```tsx
 * const { currentWorkspace } = useWorkspaceStore();
 * const { credentials, isLoading, selectedCredential, selectedProvider } = useCredentials(currentWorkspace?.id);
 * ```
 */

import { useQuery } from '@tanstack/react-query';
import { credentialService } from '@/services/credential';
import { queryKeys } from '@/lib/query-keys';
import { CACHE_TIMES, GC_TIMES } from '@/lib/query-client';
import type { Credential, CloudProvider } from '@/lib/types';
import { useMemo } from 'react';

interface UseCredentialsOptions {
  workspaceId?: string;
  selectedCredentialId?: string;
  enabled?: boolean;
}

interface UseCredentialsReturn {
  credentials: Credential[];
  isLoading: boolean;
  error: Error | null;
  selectedCredential: Credential | undefined;
  selectedProvider: CloudProvider | undefined;
  credentialsByProvider: Record<string, Credential[]>;
}

/**
 * Credentials를 가져오는 hook
 * 
 * @param options - Hook 옵션
 * @param options.workspaceId - Workspace ID (필수)
 * @param options.selectedCredentialId - 선택된 credential ID (선택)
 * @param options.enabled - Query 활성화 여부 (기본값: !!workspaceId)
 * @returns Credentials 데이터 및 관련 헬퍼
 */
export function useCredentials(options?: UseCredentialsOptions): UseCredentialsReturn;
export function useCredentials(workspaceId?: string, selectedCredentialId?: string): UseCredentialsReturn;
export function useCredentials(
  optionsOrWorkspaceId?: UseCredentialsOptions | string,
  selectedCredentialId?: string
): UseCredentialsReturn {
  // 옵션 객체 또는 개별 파라미터 처리
  const options: UseCredentialsOptions = typeof optionsOrWorkspaceId === 'object' 
    ? optionsOrWorkspaceId 
    : { workspaceId: optionsOrWorkspaceId, selectedCredentialId };

  const { workspaceId, selectedCredentialId: selectedId, enabled } = options;
  const shouldEnable = enabled !== undefined ? enabled : !!workspaceId;

  // Fetch credentials
  const { data: credentialsData = [], isLoading, error } = useQuery({
    queryKey: queryKeys.credentials.list(workspaceId),
    queryFn: () => workspaceId ? credentialService.getCredentials(workspaceId) : Promise.resolve([]),
    enabled: shouldEnable,
    staleTime: CACHE_TIMES.STABLE, // 10분 - 자격 증명은 자주 변경되지 않음
    gcTime: GC_TIMES.LONG, // 30분 - GC 시간
  });

  // Ensure credentials is always an array
  const credentials = Array.isArray(credentialsData) ? credentialsData : [];

  // Find selected credential
  const selectedCredential = useMemo(() => {
    if (!selectedId || credentials.length === 0) return undefined;
    return credentials.find(c => c.id === selectedId);
  }, [credentials, selectedId]);

  // Get selected provider
  const selectedProvider = useMemo(() => {
    return selectedCredential?.provider as CloudProvider | undefined;
  }, [selectedCredential]);

  // Group credentials by provider
  const credentialsByProvider = useMemo(() => {
    const grouped: Record<string, Credential[]> = {};
    credentials.forEach(credential => {
      const provider = credential.provider;
      if (!grouped[provider]) {
        grouped[provider] = [];
      }
      grouped[provider].push(credential);
    });
    return grouped;
  }, [credentials]);

  return {
    credentials,
    isLoading,
    error: error as Error | null,
    selectedCredential,
    selectedProvider,
    credentialsByProvider,
  };
}

