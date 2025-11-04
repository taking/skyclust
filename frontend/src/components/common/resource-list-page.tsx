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
  // Page configuration
  title: string;
  description?: string;
  resourceName: string; // For search placeholder, empty state messages
  storageKey?: string; // For localStorage persistence (optional)

  // Header component
  header: ReactNode;

  // Data and loading
  items: TItem[];
  isLoading?: boolean;
  isEmpty?: boolean;

  // Search
  searchQuery: string;
  onSearchChange: (query: string) => void;
  onSearchClear: () => void;
  isSearching?: boolean;
  searchPlaceholder?: string;

  // Filtering
  filterConfigs?: FilterConfig[];
  filters?: FilterValue;
  onFiltersChange?: (filters: FilterValue) => void;
  onFiltersClear?: () => void;
  showFilters?: boolean;
  onToggleFilters?: () => void;
  filterCount?: number;

  // Sorting (optional - for pages that support it)
  sortIndicator?: ReactNode;

  // Additional toolbar (e.g., BulkActionsToolbar, BulkOperationProgress)
  toolbar?: ReactNode;

  // Additional controls (e.g., TagFilter, FilterPresetsManager)
  additionalControls?: ReactNode;

  // Content
  emptyState?: ReactNode;
  content: ReactNode; // Table or list component

  // Pagination
  pageSize?: number;
  onPageSizeChange?: (size: number) => void;
  showPagination?: boolean;

  // Optional features
  showFilterButton?: boolean;
  showSearchResultsInfo?: boolean;
  searchResultsCount?: number;
  keyboardShortcuts?: ReactNode;
  liveRegion?: ReactNode;
  
  // Custom skeleton columns (for loading state)
  skeletonColumns?: number;
  skeletonRows?: number;
  skeletonShowCheckbox?: boolean;
}

export function ResourceListPage<TItem = unknown>({
  title,
  description,
  resourceName,
  header,
  items,
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
  pageSize = 20,
  onPageSizeChange,
  showPagination = true,
  showFilterButton = true,
  showSearchResultsInfo = true,
  searchResultsCount,
  keyboardShortcuts,
  liveRegion,
  skeletonColumns = 7,
  skeletonRows = 5,
  skeletonShowCheckbox = false,
}: ResourceListPageProps<TItem>) {
  const { currentWorkspace } = useWorkspaceStore();
  const { isLoading: authLoading } = useRequireAuth();
  useSSEMonitoring();

  // Loading state
  if (authLoading || isLoading) {
    return (
      <WorkspaceRequired>
        <Layout>
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-3xl font-bold text-gray-900">{title}</h1>
                {description && (
                  <p className="text-gray-600 mt-1">{description}</p>
                )}
              </div>
            </div>
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
          {/* Header */}
          {header}

          {/* Toolbar (e.g., Bulk Actions) */}
          {toolbar}

          {/* Search and Filter Controls */}
          {(searchQuery || filterConfigs.length > 0 || additionalControls) && (
            <Card>
              <CardContent className="pt-6">
                <div className="space-y-4">
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
                    {additionalControls && (
                      <div className="flex items-center gap-2">
                        {additionalControls}
                      </div>
                    )}
                  </div>

                  {/* Search Results Info and Sort */}
                  {(isSearching || searchResultsCount !== undefined || sortIndicator) && (
                    <div className="flex items-center justify-between">
                      <div className="flex-1">
                        {showSearchResultsInfo && isSearching && searchResultsCount !== undefined && (
                          <div className="text-sm text-gray-600">
                            Found {searchResultsCount} {resourceName}{searchResultsCount !== 1 ? 's' : ''}
                            {searchQuery && ` matching "${searchQuery}"`}
                          </div>
                        )}
                      </div>
                      {sortIndicator && (
                        <div className="flex items-center">
                          {sortIndicator}
                        </div>
                      )}
                    </div>
                  )}

                  {/* Filter Panel */}
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

          {/* Content: Empty State or Table/List */}
          {isEmpty ? (
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
            content
          )}
        </div>
      </Layout>
      {keyboardShortcuts}
      {liveRegion}
    </WorkspaceRequired>
  );
}

