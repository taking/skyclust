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

import * as React from 'react';
import { useQuery } from '@tanstack/react-query';
import { credentialService } from '@/services/credential';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
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
  // 1. 함수 오버로딩 지원: 옵션 객체 또는 개별 파라미터 처리
  const options: UseCredentialsOptions = typeof optionsOrWorkspaceId === 'object' 
    ? optionsOrWorkspaceId 
    : { workspaceId: optionsOrWorkspaceId, selectedCredentialId };

  // 2. 옵션에서 값 추출
  const { workspaceId, selectedCredentialId: selectedId, enabled } = options;
  
  // 3. Query 활성화 여부 결정: enabled가 명시되지 않으면 workspaceId 존재 여부로 결정
  const shouldEnable = enabled !== undefined ? enabled : !!workspaceId;

  // 4. React Query를 사용하여 자격 증명 목록 가져오기
  const { data: credentialsData = [], isLoading, error } = useQuery({
    queryKey: queryKeys.credentials.list(workspaceId),
    queryFn: () => workspaceId ? credentialService.getCredentials(workspaceId) : Promise.resolve([]),
    enabled: shouldEnable,
    staleTime: CACHE_TIMES.STABLE, // 10분 - 자격 증명은 자주 변경되지 않음
    gcTime: GC_TIMES.LONG, // 30분 - GC 시간
  });

  // 5. 데이터가 항상 배열인지 보장 (타입 안전성)
  const credentials = React.useMemo(() => {
    return Array.isArray(credentialsData) ? credentialsData : [];
  }, [credentialsData]);

  // 6. 선택된 자격 증명 찾기
  const selectedCredential = useMemo(() => {
    if (!selectedId || credentials.length === 0) return undefined;
    return credentials.find(c => c.id === selectedId);
  }, [credentials, selectedId]);

  // 7. 선택된 자격 증명의 프로바이더 추출
  const selectedProvider = useMemo(() => {
    return selectedCredential?.provider as CloudProvider | undefined;
  }, [selectedCredential]);

  // 8. 프로바이더별로 자격 증명 그룹화
  const credentialsByProvider = useMemo(() => {
    const grouped: Record<string, Credential[]> = {};
    credentials.forEach(credential => {
      const provider = credential.provider;
      // 프로바이더별 그룹이 없으면 생성
      if (!grouped[provider]) {
        grouped[provider] = [];
      }
      // 해당 프로바이더 그룹에 추가
      grouped[provider].push(credential);
    });
    return grouped;
  }, [credentials]);

  // 9. 모든 정보 반환
  return {
    credentials,
    isLoading,
    error: error as Error | null,
    selectedCredential,
    selectedProvider,
    credentialsByProvider,
  };
}

