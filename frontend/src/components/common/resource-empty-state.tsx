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
  onCreateClick,
  createButtonText,
  icon: Icon = Server,
  title,
  description,
  showCreateButton = !!onCreateClick,
  withCard = false,
}: ResourceEmptyStateProps) {
  const { t } = useTranslation();
  
  const defaultTitle = isSearching 
    ? t('emptyState.noResourceFound', { resource: resourceName })
    : t('emptyState.noResource', { resource: resourceName });
  
  const defaultDescription = isSearching
    ? searchQuery 
      ? t('emptyState.tryAdjustingWithQuery', { query: searchQuery })
      : t('emptyState.tryAdjusting')
    : t('emptyState.createFirst', { resource: resourceName.toLowerCase() });

  const displayTitle = title || defaultTitle;
  const displayDescription = description || defaultDescription;
  
  const defaultCreateButtonText = createButtonText || t('emptyState.createResource', { resource: resourceName.slice(0, -1) });

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
      {showCreateButton && onCreateClick && (
        <div className="mt-6">
          <Button onClick={onCreateClick}>
            <Plus className="mr-2 h-4 w-4" />
            {defaultCreateButtonText}
          </Button>
        </div>
      )}
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

