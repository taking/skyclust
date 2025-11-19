/**
 * Resource Empty State Component
 * 리소스가 없을 때 표시되는 통합 빈 상태 컴포넌트
 * 
 * VMs, Kubernetes, Networks 등 모든 리소스의 빈 상태를 일관되게 표시
 */

'use client';

import * as React from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { LucideIcon, Plus, Server } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';

export interface ResourceEmptyStateProps {
  /**
   * 리소스 타입 (예: 'VMs', 'Clusters', 'VPCs')
   */
  resourceName: string;
  
  /**
   * 검색 중인지 여부
   */
  isSearching?: boolean;
  
  /**
   * 검색 쿼리 (검색 중일 때 표시)
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
   * 생성 버튼 클릭 핸들러
   */
  onCreateClick?: () => void;
  
  /**
   * 생성 버튼 텍스트 (기본값: `Create {resourceName}`)
   */
  createButtonText?: string;
  
  /**
   * 아이콘 컴포넌트 (기본값: Server)
   */
  icon?: LucideIcon;
  
  /**
   * 커스텀 제목 (기본값: `No {resourceName} found` 또는 `No {resourceName}`)
   */
  title?: string;
  
  /**
   * 커스텀 설명 (기본값: 검색 중이면 다른 메시지, 아니면 생성 메시지)
   */
  description?: string;
  
  /**
   * 생성 버튼 표시 여부 (기본값: onCreateClick이 있으면 true)
   */
  showCreateButton?: boolean;
  
  /**
   * Card로 감싸기 여부 (기본값: false)
   */
  withCard?: boolean;
  
  /**
   * 권한 문제인지 여부
   */
  isPermissionError?: boolean;
  
  /**
   * 권한 문제 해결 액션 핸들러
   */
  onCheckCredentials?: () => void;
}

/**
 * ResourceEmptyState Component
 * 
 * @example
 * ```tsx
 * <ResourceEmptyState
 *   resourceName="VMs"
 *   isSearching={isSearching}
 *   searchQuery={searchQuery}
 *   onCreateClick={() => setIsCreateDialogOpen(true)}
 * />
 * ```
 */
function ResourceEmptyStateComponent({
  resourceName,
  isSearching = false,
  searchQuery,
  hasFilters = false,
  onClearFilters,
  onClearSearch,
  onCreateClick,
  createButtonText,
  icon: Icon = Server,
  title,
  description,
  showCreateButton = !!onCreateClick,
  withCard = false,
  isPermissionError = false,
  onCheckCredentials,
}: ResourceEmptyStateProps) {
  const { t } = useTranslation();
  
  // 컨텍스트별 제목 결정
  let defaultTitle: string;
  if (isPermissionError) {
    defaultTitle = t('emptyState.permissionError', { resource: resourceName });
  } else if (hasFilters) {
    defaultTitle = t('emptyState.noResourceMatchFilters', { resource: resourceName });
  } else if (isSearching && searchQuery) {
    defaultTitle = t('emptyState.noResourceFound', { resource: resourceName, query: searchQuery });
  } else if (isSearching) {
    defaultTitle = t('emptyState.noResourceFound', { resource: resourceName });
  } else {
    defaultTitle = t('emptyState.noResource', { resource: resourceName });
  }
  
  // 컨텍스트별 설명 결정
  let defaultDescription: string;
  if (isPermissionError) {
    defaultDescription = t('emptyState.checkCredentialsMessage');
  } else if (hasFilters) {
    defaultDescription = t('emptyState.tryAdjustingFilters');
  } else if (isSearching && searchQuery) {
    defaultDescription = t('emptyState.tryAdjustingWithQuery', { query: searchQuery });
  } else if (isSearching) {
    defaultDescription = t('emptyState.tryAdjusting');
  } else {
    defaultDescription = t('emptyState.createFirst', { resource: resourceName });
  }

  const displayTitle = title || defaultTitle;
  const displayDescription = description || defaultDescription;
  
  // 복수형 제거 시도 (영어만 지원, 번역된 문자열은 그대로 사용)
  // 영어 복수형 패턴: "s"로 끝나는 경우 (예: "VMs" -> "VM", "Clusters" -> "Cluster")
  const singularResourceName = resourceName.endsWith('s') && resourceName.length > 1 && /^[A-Za-z]+$/.test(resourceName)
    ? resourceName.slice(0, -1)
    : resourceName;
  
  const defaultCreateButtonText = createButtonText || t('emptyState.createResource', { resource: singularResourceName });

  const content = (
    <div className="text-center py-12">
      <div className="mx-auto h-12 w-12 text-gray-400">
        <Icon className="h-12 w-12" />
      </div>
      <h3 className="mt-2 text-sm font-medium text-gray-900">
        {displayTitle}
      </h3>
      <p className="mt-1 text-sm text-gray-500">
        {displayDescription}
      </p>
      <div className="mt-6 flex flex-col sm:flex-row gap-2 justify-center">
        {/* 권한 문제인 경우 */}
        {isPermissionError && onCheckCredentials && (
          <Button 
            onClick={onCheckCredentials} 
            variant="default"
            aria-label={t('emptyState.checkCredentials') || 'Check credentials'}
          >
            {t('emptyState.checkCredentials')}
          </Button>
        )}
        
        {/* 필터가 적용된 경우 */}
        {hasFilters && onClearFilters && (
          <Button 
            onClick={onClearFilters} 
            variant="outline"
            aria-label={t('emptyState.clearFilters') || 'Clear filters'}
          >
            {t('emptyState.clearFilters')}
          </Button>
        )}
        
        {/* 검색 중인 경우 */}
        {isSearching && searchQuery && onClearSearch && (
          <Button 
            onClick={onClearSearch} 
            variant="outline"
            aria-label={t('emptyState.clearSearch') || 'Clear search'}
          >
            {t('emptyState.clearSearch')}
          </Button>
        )}
        
        {/* 생성 버튼 */}
        {showCreateButton && onCreateClick && !isPermissionError && (
          <Button 
            onClick={onCreateClick}
            aria-label={defaultCreateButtonText}
          >
            <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
            {defaultCreateButtonText}
          </Button>
        )}
      </div>
    </div>
  );

  if (withCard) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          {content}
        </CardContent>
      </Card>
    );
  }

  return content;
}

export const ResourceEmptyState = React.memo(ResourceEmptyStateComponent);

