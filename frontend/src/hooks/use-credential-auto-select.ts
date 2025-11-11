/**
 * useCredentialAutoSelect 훅
 * 
 * 자격 증명 자동 선택 로직을 제공하는 커스텀 훅입니다.
 * 
 * @example
 * ```tsx
 * function MyPage() {
 *   useCredentialAutoSelect({
 *     enabled: true,
 *     resourceType: 'kubernetes',
 *     updateUrl: true,
 *   });
 *   
 *   return <div>...</div>;
 * }
 * ```
 * 
 * 자동 선택 우선순위:
 * 1. 이미 선택된 자격 증명이 있으면 유지
 * 2. 자격 증명이 1개면 자동 선택
 * 3. 프로바이더별 기본 자격 증명
 * 4. 최근 사용한 자격 증명
 * 5. 첫 번째 자격 증명
 * 
 * 자격 증명 변경 시:
 * - 리전 초기화
 * - 프로바이더가 리전을 지원하면 기본 리전 자동 설정 (GCP: asia-northeast3, AWS: ap-northeast-3)
 */

'use client';

import { useEffect, useMemo } from 'react';
import { useCredentialContext } from './use-credential-context';
import { useCredentials } from './use-credentials';
import { useWorkspaceStore } from '@/store/workspace';
import { getRecommendedCredential, trackCredentialUsage } from '@/lib/credential';
import { getDefaultRegionForProvider, supportsRegionSelection } from '@/lib/regions';
import type { CloudProvider } from '@/lib/types/kubernetes';

interface UseCredentialAutoSelectOptions {
  /** 자동 선택 활성화 여부 */
  enabled?: boolean;
  
  /** 필터링할 프로바이더 */
  provider?: CloudProvider;
  
  /** 리소스 타입 (사용 추적용) */
  resourceType?: 'kubernetes' | 'network' | 'compute';
  
  /** 자동 선택 시 URL 업데이트 여부 */
  updateUrl?: boolean;
}
export function useCredentialAutoSelect(options: UseCredentialAutoSelectOptions = {}) {
  const {
    enabled = true,
    provider,
    resourceType,
    updateUrl = true,
  } = options;
  
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedCredentialId, selectedRegion: _selectedRegion, setSelectedCredential, setSelectedRegion } = useCredentialContext();
  const { credentials, selectedCredential, isLoading } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
    enabled: enabled && !!currentWorkspace,
  });
  
  // 추천 자격 증명 계산
  const recommendedCredentialId = useMemo(() => {
    // 1. 자동 선택이 비활성화되었거나 워크스페이스/자격 증명이 없으면 null 반환
    if (!enabled || !currentWorkspace || credentials.length === 0) {
      return null;
    }
    
    // 2. 이미 선택된 자격 증명이 있고 유효하면 유지
    if (selectedCredentialId && selectedCredential) {
      return selectedCredentialId;
    }
    
    // 3. 우선순위에 따라 추천 자격 증명 계산
    // - 자격 증명이 1개면 자동 선택
    // - 프로바이더별 기본 자격 증명
    // - 최근 사용한 자격 증명
    // - 첫 번째 자격 증명
    return getRecommendedCredential(
      currentWorkspace.id,
      credentials,
      provider,
      resourceType
    );
  }, [
    enabled,
    currentWorkspace,
    credentials,
    selectedCredentialId,
    selectedCredential,
    provider,
    resourceType,
  ]);
  
  // 자동 선택 실행
  useEffect(() => {
    // 1. 자동 선택이 비활성화되었거나 로딩 중이면 스킵
    if (!enabled || isLoading) {
      return;
    }
    
    // 2. 자격 증명 목록이 비어있으면 스킵 (무한 루프 방지)
    if (credentials.length === 0) {
      return;
    }
    
    // 3. 추천 자격 증명이 없으면 스킵
    if (!recommendedCredentialId) {
      return;
    }
    
    // 4. 이미 추천 자격 증명이 선택되어 있으면 스킵
    if (selectedCredentialId === recommendedCredentialId) {
      return;
    }
    
    // 5. 추천 자격 증명 선택
    setSelectedCredential(recommendedCredentialId);
    
    // 6. 자격 증명 변경 시 리전 처리
    const newCredential = credentials.find(c => c.id === recommendedCredentialId);
    if (!newCredential) {
      // 6-1. 자격 증명을 찾을 수 없으면 리전 제거
      setSelectedRegion(null);
      if (updateUrl && typeof window !== 'undefined') {
        const params = new URLSearchParams(window.location.search);
        params.set('credentialId', recommendedCredentialId);
        params.delete('region');
        window.history.replaceState(
          {},
          '',
          `${window.location.pathname}?${params.toString()}`
        );
      }
      return;
    }
    
    // 6-2. 프로바이더가 리전을 지원하는지 확인
    if (supportsRegionSelection(newCredential.provider as CloudProvider)) {
      // 리전을 지원하면 기본 리전으로 설정
      const defaultRegion = getDefaultRegionForProvider(newCredential.provider);
      if (defaultRegion) {
        setSelectedRegion(defaultRegion);
      } else {
        // 기본 리전이 없으면 null
        setSelectedRegion(null);
      }
    } else {
      // 리전을 지원하지 않으면 리전 제거
      setSelectedRegion(null);
    }
    
    // 7. URL 업데이트 (옵션이 활성화된 경우)
    if (updateUrl && typeof window !== 'undefined') {
      const params = new URLSearchParams(window.location.search);
      params.set('credentialId', recommendedCredentialId);
      
      // 7-1. 리전 정보도 URL에 반영
      if (supportsRegionSelection(newCredential.provider as CloudProvider)) {
        const defaultRegion = getDefaultRegionForProvider(newCredential.provider);
        if (defaultRegion) {
          params.set('region', defaultRegion);
        } else {
          params.delete('region');
        }
      } else {
        params.delete('region');
      }
      
      // 7-2. 브라우저 히스토리 업데이트 (페이지 리로드 없이)
      window.history.replaceState(
        {},
        '',
        `${window.location.pathname}?${params.toString()}`
      );
    }
  }, [
    enabled,
    isLoading,
    credentials,
    recommendedCredentialId,
    selectedCredentialId,
    setSelectedCredential,
    setSelectedRegion,
    updateUrl,
  ]);
  
  // 자격 증명 사용 추적 (localStorage에 저장하여 최근 사용 기록 관리)
  useEffect(() => {
    // 1. 추적 조건 확인: 활성화 여부, 워크스페이스, 선택된 자격 증명, 리소스 타입
    if (
      !enabled ||
      !currentWorkspace ||
      !selectedCredential ||
      !resourceType
    ) {
      return;
    }
    
    // 2. 자격 증명 사용 기록 저장 (최근 사용 자격 증명 추천에 사용)
    trackCredentialUsage(
      currentWorkspace.id,
      selectedCredential.id,
      selectedCredential.provider as CloudProvider,
      resourceType
    );
  }, [
    enabled,
    currentWorkspace,
    selectedCredential,
    resourceType,
  ]);
  
  return {
    recommendedCredentialId,
    selectedCredentialId,
    selectedCredential,
    credentials,
    isLoading,
  };
}

