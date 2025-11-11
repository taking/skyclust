/**
 * useEmptyState Hook
 * 모든 리스트 페이지의 empty state 로직을 통합한 훅
 * 
 * 중복된 empty state 패턴을 통합하여 일관된 인터페이스를 제공합니다.
 * 
 * @example
 * ```tsx
 * const {
 *   isEmpty,
 *   emptyStateComponent,
 * } = useEmptyState({
 *   credentials,
 *   selectedProvider,
 *   selectedCredentialId,
 *   filteredItems,
 *   resourceName: t('network.vpcs'),
 *   serviceName: t('network.title'),
 *   onCreateClick: handleCreateVPC,
 *   icon: Network,
 * });
 * ```
 */

import { useMemo } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { LucideIcon } from 'lucide-react';
import { useTranslation } from './use-translation';

export interface UseEmptyStateOptions {
  /**
   * 자격 증명 목록
   */
  credentials: Array<unknown>;
  
  /**
   * 선택된 프로바이더
   */
  selectedProvider?: string;
  
  /**
   * 선택된 자격 증명 ID
   */
  selectedCredentialId?: string;
  
  /**
   * 필터링된 아이템 목록
   */
  filteredItems: Array<unknown>;
  
  /**
   * 리소스 이름 (예: 'VPCs', 'Clusters')
   */
  resourceName: string;
  
  /**
   * 서비스 이름 (예: 'Network', 'Kubernetes')
   */
  serviceName: string;
  
  /**
   * 생성 버튼 클릭 핸들러
   */
  onCreateClick?: () => void;
  
  /**
   * 아이콘 (선택적)
   */
  icon?: LucideIcon;
  
  /**
   * 설명 (선택적)
   */
  description?: string;
  
  /**
   * 검색 중인지 여부
   */
  isSearching?: boolean;
  
  /**
   * 검색 쿼리
   */
  searchQuery?: string;
  
  /**
   * 필터가 적용되었는지 여부
   */
  hasFilters?: boolean;
  
  /**
   * 필터 초기화 핸들러
   */
  onClearFilters?: () => void;
  
  /**
   * 검색 초기화 핸들러
   */
  onClearSearch?: () => void;
  
  /**
   * 추가 조건 (예: selectedVPCId, selectedClusterName)
   * 조건이 false이면 커스텀 empty state를 표시
   */
  additionalConditions?: Array<{
    /**
     * 조건 값
     */
    value: boolean | string | undefined;
    
    /**
     * 조건이 false일 때 표시할 제목
     */
    title: string;
    
    /**
     * 조건이 false일 때 표시할 설명
     */
    description: string;
    
    /**
     * 조건이 false일 때 표시할 아이콘
     */
    icon?: LucideIcon;
  }>;
}

export interface UseEmptyStateReturn {
  /**
   * Empty state 여부
   */
  isEmpty: boolean;
  
  /**
   * Empty state 컴포넌트
   */
  emptyStateComponent: React.ReactNode | null;
}

/**
 * Empty State 통합 훅
 */
export function useEmptyState(options: UseEmptyStateOptions): UseEmptyStateReturn {
  const {
    credentials,
    selectedProvider,
    selectedCredentialId,
    filteredItems,
    resourceName,
    serviceName,
    onCreateClick,
    icon,
    description,
    isSearching,
    searchQuery,
    hasFilters,
    onClearFilters,
    onClearSearch,
    additionalConditions = [],
  } = options;
  
  const { t } = useTranslation();

  // isEmpty 계산
  const isEmpty = useMemo(() => {
    // 기본 조건: provider, credentialId, filteredItems
    const basicEmpty = !selectedProvider || !selectedCredentialId || filteredItems.length === 0;
    
    // 추가 조건 확인
    const additionalEmpty = additionalConditions.some(condition => !condition.value);
    
    return basicEmpty || additionalEmpty;
  }, [selectedProvider, selectedCredentialId, filteredItems.length, additionalConditions]);

  // emptyStateComponent 생성
  const emptyStateComponent = useMemo(() => {
    // 1. 자격 증명이 없는 경우
    if (credentials.length === 0) {
      return <CredentialRequiredState serviceName={serviceName} />;
    }
    
    // 2. 프로바이더 또는 자격 증명이 선택되지 않은 경우
    if (!selectedProvider || !selectedCredentialId) {
      return (
        <CredentialRequiredState
          title={t('credential.selectCredential')}
          description={t('credential.selectCredential')}
          serviceName={serviceName}
        />
      );
    }
    
    // 3. 추가 조건이 만족되지 않은 경우
    const unmetCondition = additionalConditions.find(condition => !condition.value);
    if (unmetCondition) {
      const ConditionIcon = unmetCondition.icon;
      return (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            {ConditionIcon && <ConditionIcon className="h-12 w-12 text-gray-400 mb-4" />}
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              {unmetCondition.title}
            </h3>
            <p className="text-sm text-gray-500 text-center">
              {unmetCondition.description}
            </p>
          </CardContent>
        </Card>
      );
    }
    
    // 4. 필터링된 아이템이 없는 경우
    if (filteredItems.length === 0) {
      return (
        <ResourceEmptyState
          resourceName={resourceName}
          icon={icon}
          onCreateClick={onCreateClick}
          description={description}
          isSearching={isSearching}
          searchQuery={searchQuery}
          hasFilters={hasFilters}
          onClearFilters={onClearFilters}
          onClearSearch={onClearSearch}
          withCard={true}
        />
      );
    }
    
    return null;
  }, [
    credentials.length,
    selectedProvider,
    selectedCredentialId,
    additionalConditions,
    filteredItems.length,
    resourceName,
    icon,
    onCreateClick,
    description,
    isSearching,
    searchQuery,
    hasFilters,
    onClearFilters,
    onClearSearch,
    serviceName,
    t,
  ]);

  return {
    isEmpty,
    emptyStateComponent,
  };
}

