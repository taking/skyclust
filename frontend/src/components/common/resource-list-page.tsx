/**
 * Resource List Page Template
 * 리소스 목록 페이지 공통 템플릿 컴포넌트
 * 
 * VMs, Kubernetes, Networks 등 리소스 목록 페이지의 공통 구조를 제공합니다.
 * 
 * 사용 예시:
 * ```tsx
 * <ResourceListPage
 *   title="Virtual Machines"
 *   resourceName="VMs"
 *   storageKey="vms-page"
 *   header={<VMPageHeader {...props} />}
 *   items={filteredVMs}
 *   isLoading={isLoading}
 *   isEmpty={filteredVMs.length === 0}
 *   searchQuery={searchQuery}
 *   onSearchChange={setSearchQuery}
 *   onSearchClear={clearSearch}
 *   isSearching={isSearching}
 *   filterConfigs={filterConfigs}
 *   filters={filters}
 *   onFiltersChange={handleFiltersChange}
 *   onFiltersClear={clearFilters}
 *   showFilters={showFilters}
 *   onToggleFilters={() => setShowFilters(!showFilters)}
 *   filterCount={Object.keys(filters).length}
 *   sortIndicator={<MultiSortIndicator {...sortProps} />}
 *   additionalControls={<FilterPresetsManager {...presetProps} />}
 *   emptyState={<VMEmptyState {...emptyStateProps} />}
 *   content={<VMTable {...tableProps} />}
 *   pageSize={pageSize}
 *   onPageSizeChange={handlePageSizeChange}
 *   keyboardShortcuts={<GlobalKeyboardShortcuts />}
 *   liveRegion={<LiveRegion message={liveMessage} />}
 * />
 * ```
 */

'use client';

import * as React from 'react';
import { ReactNode } from 'react';
import { Layout } from '@/components/layout/layout';
import { Card, CardContent } from '@/components/ui/card';
import { SearchBar } from '@/components/ui/search-bar';
import { FilterPanel, FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { useWorkspaceStore } from '@/store/workspace';
import { WorkspaceRequired } from '@/components/common/workspace-required';
import { useRequireAuth } from '@/hooks/use-auth';
import { useSSEMonitoring } from '@/hooks/use-sse-monitoring';

export interface ResourceListPageProps<TItem = unknown> {
  // 페이지 설정
  title: string;
  description?: string;
  resourceName: string; // 검색 플레이스홀더, 빈 상태 메시지용
  storageKey?: string; // localStorage 영속성용 (선택사항)

  // 헤더 컴포넌트
  header: ReactNode;

  // 데이터 및 로딩
  items: TItem[];
  isLoading?: boolean;
  isEmpty?: boolean;

  // 검색
  searchQuery: string;
  onSearchChange: (query: string) => void;
  onSearchClear: () => void;
  isSearching?: boolean;
  searchPlaceholder?: string;

  // 필터링
  filterConfigs?: FilterConfig[];
  filters?: FilterValue;
  onFiltersChange?: (filters: FilterValue) => void;
  onFiltersClear?: () => void;
  showFilters?: boolean;
  onToggleFilters?: () => void;
  filterCount?: number;

  // 정렬 (선택사항 - 지원하는 페이지용)
  sortIndicator?: ReactNode;

  // 추가 툴바 (예: BulkActionsToolbar, BulkOperationProgress)
  toolbar?: ReactNode;

  // 추가 컨트롤 (예: TagFilter, FilterPresetsManager)
  additionalControls?: ReactNode;

  // 콘텐츠
  emptyState?: ReactNode;
  content: ReactNode; // 테이블 또는 리스트 컴포넌트

  // 페이지네이션
  pageSize?: number;
  onPageSizeChange?: (size: number) => void;
  showPagination?: boolean;

  // 선택적 기능
  showFilterButton?: boolean;
  showSearchResultsInfo?: boolean;
  searchResultsCount?: number;
  keyboardShortcuts?: ReactNode;
  liveRegion?: ReactNode;
  
  // 커스텀 스켈레톤 컬럼 (로딩 상태용)
  skeletonColumns?: number;
  skeletonRows?: number;
  skeletonShowCheckbox?: boolean;
}

function ResourceListPageComponent<TItem = unknown>({
  title,
  description,
  resourceName,
  header,
  isLoading = false,
  isEmpty = false,
  searchQuery,
  onSearchChange,
  onSearchClear,
  isSearching = false,
  searchPlaceholder,
  filterConfigs = [],
  filters = {},
  onFiltersChange,
  onFiltersClear,
  showFilters = false,
  onToggleFilters,
  filterCount = 0,
  sortIndicator,
  toolbar,
  additionalControls,
  emptyState,
  content,
  showFilterButton = true,
  showSearchResultsInfo = true,
  searchResultsCount,
  keyboardShortcuts,
  liveRegion,
  skeletonColumns = 7,
  skeletonRows = 5,
  skeletonShowCheckbox = false,
}: ResourceListPageProps<TItem>) {
  // 1. 인증 상태 확인 및 SSE 모니터링 활성화
  const { isLoading: authLoading } = useRequireAuth();
  useSSEMonitoring();

  // 2. 로딩 상태: 인증 로딩 중이거나 데이터 로딩 중이면 스켈레톤 표시
  if (authLoading || isLoading) {
    return (
      <WorkspaceRequired>
        <Layout>
          <div className="space-y-6">
            {/* 2-1. 페이지 헤더 (제목 및 설명) */}
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-3xl font-bold text-gray-900">{title}</h1>
                {description && (
                  <p className="text-gray-600 mt-1">{description}</p>
                )}
              </div>
            </div>
            {/* 2-2. 테이블 스켈레톤 로딩 UI */}
            <Card>
              <CardContent className="pt-6">
                <TableSkeleton 
                  columns={skeletonColumns} 
                  rows={skeletonRows} 
                  showCheckbox={skeletonShowCheckbox}
                />
              </CardContent>
            </Card>
          </div>
        </Layout>
      </WorkspaceRequired>
    );
  }

  return (
    <WorkspaceRequired>
      <Layout>
        <div className="space-y-6">
          {/* 3. 페이지 헤더 컴포넌트 (제목, 설명, 액션 버튼 등) */}
          {header}

          {/* 4. 툴바 (예: Bulk Actions Toolbar) */}
          {toolbar}

          {/* 5. 검색 및 필터 컨트롤: 검색 쿼리, 필터 설정, 추가 컨트롤이 있으면 표시 */}
          {(searchQuery || filterConfigs.length > 0 || additionalControls) && (
            <Card>
              <CardContent className="pt-6">
                <div className="space-y-4">
                  {/* 5-1. 검색 바 및 추가 컨트롤 */}
                  <div className="flex flex-col sm:flex-row gap-4">
                    <div className="flex-1">
                      <SearchBar
                        placeholder={searchPlaceholder || `Search ${resourceName}...`}
                        value={searchQuery}
                        onChange={onSearchChange}
                        onClear={onSearchClear}
                        showFilter={showFilterButton}
                        onFilterClick={onToggleFilters}
                        filterCount={filterCount}
                      />
                    </div>
                    {/* 5-2. 추가 컨트롤 (예: TagFilter, FilterPresetsManager) */}
                    {additionalControls && (
                      <div className="flex items-center gap-2">
                        {additionalControls}
                      </div>
                    )}
                  </div>

                  {/* 5-3. 검색 결과 정보 및 정렬 표시기 */}
                  {(isSearching || searchResultsCount !== undefined || sortIndicator) && (
                    <div className="flex items-center justify-between">
                      <div className="flex-1">
                        {/* 검색 결과 개수 표시 */}
                        {showSearchResultsInfo && isSearching && searchResultsCount !== undefined && (
                          <div className="text-sm text-gray-600">
                            Found {searchResultsCount} {resourceName}{searchResultsCount !== 1 ? 's' : ''}
                            {searchQuery && ` matching "${searchQuery}"`}
                          </div>
                        )}
                      </div>
                      {/* 정렬 표시기 (MultiSortIndicator 등) */}
                      {sortIndicator && (
                        <div className="flex items-center">
                          {sortIndicator}
                        </div>
                      )}
                    </div>
                  )}

                  {/* 5-4. 필터 패널: showFilters가 true이고 필터 설정이 있으면 표시 */}
                  {showFilters && filterConfigs.length > 0 && (
                    <div className="mt-4">
                      <FilterPanel
                        filters={filterConfigs}
                        values={filters}
                        onChange={onFiltersChange || (() => {})}
                        onClear={onFiltersClear || (() => {})}
                        onApply={() => {}}
                        title={`Filter ${resourceName}`}
                        description={`Filter ${resourceName.toLowerCase()} by various criteria`}
                      />
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          )}

          {/* 6. 콘텐츠: 빈 상태 또는 테이블/리스트 */}
          {isEmpty ? (
            // 6-1. 빈 상태: emptyState가 제공되면 사용, 없으면 기본 메시지 표시
            emptyState || (
              <Card>
                <CardContent className="pt-6">
                  <div className="text-center py-12">
                    <p className="text-gray-500">
                      {isSearching 
                        ? `No ${resourceName.toLowerCase()} found matching your search.`
                        : `No ${resourceName.toLowerCase()} found.`
                      }
                    </p>
                  </div>
                </CardContent>
              </Card>
            )
          ) : (
            // 6-2. 데이터가 있으면 테이블/리스트 컴포넌트 표시
            content
          )}
        </div>
      </Layout>
      {/* 7. 키보드 단축키 및 Live Region (접근성) */}
      {keyboardShortcuts}
      {liveRegion}
    </WorkspaceRequired>
  );
}

// React.memo로 최적화: props가 변경되지 않으면 리렌더링 방지
export const ResourceListPage = React.memo(ResourceListPageComponent) as typeof ResourceListPageComponent;

