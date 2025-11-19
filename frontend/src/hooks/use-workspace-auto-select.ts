/**
 * useWorkspaceAutoSelect Hook
 * Workspace 자동 선택 로직
 * 
 * 기능:
 * - localStorage에서 마지막 선택된 workspace 복원
 * - Workspace 목록 로드 시 자동 선택
 * - URL과 동기화
 */

import { useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { useWorkspaceStore } from '@/store/workspace';
import { useQuery } from '@tanstack/react-query';
import { workspaceService } from '@/features/workspaces';
import { queryKeys } from '@/lib/query';
import { API } from '@/lib/constants';
import { buildWorkspaceManagementPath } from '@/lib/routing/helpers';

interface UseWorkspaceAutoSelectOptions {
  /**
   * 자동 선택 활성화 여부
   */
  enabled?: boolean;
  
  /**
   * 자동 선택 후 리다이렉트할 경로
   */
  redirectTo?: 'dashboard' | 'workspaces' | null;
  
  /**
   * 리다이렉트 시 URL 업데이트 여부
   */
  updateUrl?: boolean;
}

/**
 * Workspace 자동 선택 Hook
 * 
 * @example
 * ```tsx
 * // Dashboard 페이지에서 사용
 * useWorkspaceAutoSelect({
 *   enabled: true,
 *   redirectTo: 'dashboard',
 * });
 * ```
 */
export function useWorkspaceAutoSelect(options: UseWorkspaceAutoSelectOptions = {}) {
  const {
    enabled = true,
    redirectTo = 'dashboard',
    updateUrl = true,
  } = options;
  
  const router = useRouter();
  const { 
    currentWorkspace, 
    setCurrentWorkspace, 
    setWorkspaces,
    autoSelectWorkspace,
  } = useWorkspaceStore();
  
  // Workspace 목록 조회
  const { data: workspaces = [], isLoading } = useQuery({
    queryKey: queryKeys.workspaces.list(),
    queryFn: () => workspaceService.getWorkspaces(),
    retry: API.REQUEST.MAX_RETRIES,
    retryDelay: API.REQUEST.RETRY_DELAY,
    enabled,
  });
  
  // Workspace 목록 업데이트
  useEffect(() => {
    if (!isLoading && workspaces.length > 0) {
      setWorkspaces(workspaces);
    }
  }, [workspaces, isLoading, setWorkspaces]);
  
  // 자동 선택 로직
  useEffect(() => {
    if (!enabled || isLoading || workspaces.length === 0) {
      return;
    }
    
    // 1. 이미 선택된 workspace가 있고 유효하면 유지
    if (currentWorkspace) {
      const isValid = workspaces.some(w => w.id === currentWorkspace.id);
      if (isValid) {
        return; // 유효한 workspace가 이미 선택됨
      }
    }
    
    // 2. 자동 선택 실행
    const autoSelected = autoSelectWorkspace(workspaces);
    if (autoSelected) {
      setCurrentWorkspace(autoSelected);
      
      // 3. 리다이렉트
      if (redirectTo && updateUrl) {
        const redirectPath = buildWorkspaceManagementPath(autoSelected.id, redirectTo);
        router.replace(redirectPath);
      }
    }
  }, [
    enabled,
    isLoading,
    workspaces,
    currentWorkspace,
    autoSelectWorkspace,
    setCurrentWorkspace,
    redirectTo,
    updateUrl,
    router,
  ]);
  
  return {
    isLoading,
    workspaces,
    currentWorkspace,
    autoSelected: currentWorkspace !== null,
  };
}

