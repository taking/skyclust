/**
 * Resource Groups Page
 * Azure Resource Groups 관리 페이지
 */

'use client';

import { Suspense } from 'react';
import { useState, useMemo, useCallback, useEffect } from 'react';
import * as React from 'react';
import { useRouter } from 'next/navigation';
import dynamic from 'next/dynamic';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { usePagination } from '@/hooks/use-pagination';
import { Filter, Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { useEmptyState } from '@/hooks/use-empty-state';
import { useCredentialAutoSelect } from '@/hooks/use-credential-auto-select';
import { useWorkspaceStore } from '@/store/workspace';
import { useTranslation } from '@/hooks/use-translation';
import { DataProcessor } from '@/lib/data';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import { useResourceListState } from '@/hooks/use-resource-list-state';
import { sseService } from '@/services/sse';
import { log } from '@/lib/logging';
import { useSSEStatus } from '@/hooks/use-sse-status';
import {
  useResourceGroupActions,
  useResourceGroups,
  ResourceGroupsPageHeader,
  ResourceGroupTable,
} from '@/features/resource-groups';
import type { ResourceGroupInfo } from '@/services/resource-group';

const ResourceGroupTableDynamic = dynamic(
  () => Promise.resolve({ default: ResourceGroupTable }),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={6} rows={5} showCheckbox={true} />,
  }
);

function ResourceGroupsPageContent() {
  const router = useRouter();
  const { t } = useTranslation();
  const { currentWorkspace } = useWorkspaceStore();

  const {
    resourceGroups,
    isLoadingResourceGroups,
    credentials,
    selectedProvider,
    selectedCredentialId,
  } = useResourceGroups({
    limit: undefined, // 목록 페이지용: pagination 사용
    enabled: !!currentWorkspace,
  });
  
  // Auto-select credential if not selected
  useCredentialAutoSelect({
    enabled: !!currentWorkspace,
    provider: 'azure',
    updateUrl: true,
  });

  // SSE 상태 확인
  const { status: sseStatus } = useSSEStatus();

  // SSE 이벤트 구독 (Azure Resource Groups 실시간 업데이트)
  useEffect(() => {
    // SSE 연결 완료 확인 (clientId는 subscribeToEvent 내부에서 대기 처리)
    if (!sseStatus.isConnected) {
      log.debug('[Resource Groups Page] SSE not connected, skipping subscription', {
        isConnected: sseStatus.isConnected,
        readyState: sseStatus.readyState,
      });
      return;
    }

    const filters = {
      credential_ids: selectedCredentialId ? [selectedCredentialId] : undefined,
    };

    const subscribeToResourceGroupEvents = async () => {
      try {
        await sseService.subscribeToEvent('azure-resource-group-created', filters);
        await sseService.subscribeToEvent('azure-resource-group-updated', filters);
        await sseService.subscribeToEvent('azure-resource-group-deleted', filters);
        await sseService.subscribeToEvent('azure-resource-group-list', filters);
        
        log.debug('[Resource Groups Page] Subscribed to Azure Resource Group events', { 
          filters,
          clientId: sseService.getClientId(),
        });
      } catch (error) {
        log.error('[Resource Groups Page] Failed to subscribe to Azure Resource Group events', error, {
          service: 'SSE',
          action: 'subscribeResourceGroupEvents',
        });
      }
    };

    subscribeToResourceGroupEvents();

    // Cleanup: 페이지를 떠날 때 또는 필터가 변경될 때 구독 해제
    return () => {
      const unsubscribe = async () => {
        try {
          await sseService.unsubscribeFromEvent('azure-resource-group-created', filters);
          await sseService.unsubscribeFromEvent('azure-resource-group-updated', filters);
          await sseService.unsubscribeFromEvent('azure-resource-group-deleted', filters);
          await sseService.unsubscribeFromEvent('azure-resource-group-list', filters);
          
          log.debug('[Resource Groups Page] Unsubscribed from Azure Resource Group events', { filters });
        } catch (error) {
          log.warn('[Resource Groups Page] Failed to unsubscribe from Azure Resource Group events', error, {
            service: 'SSE',
            action: 'unsubscribeResourceGroupEvents',
          });
        }
      };
      unsubscribe();
    };
  }, [selectedCredentialId, sseStatus.isConnected]);

  // 공통 리스트 상태 관리
  const {
    filters,
    setFilters,
    showFilters,
    setShowFilters,
    selectedIds: selectedResourceGroupNames,
    setSelectedIds: setSelectedResourceGroupNames,
    pageSize,
    setPageSize,
    deleteDialogState,
    openDeleteDialog,
    closeDeleteDialog,
  } = useResourceListState({
    storageKey: 'resource-groups-page',
  });

  const {
    deleteResourceGroupMutation,
    handleBulkDeleteResourceGroups: handleBulkDelete,
    executeDeleteResourceGroup,
  } = useResourceGroupActions({
    selectedCredentialId,
  });

  const handleCreateResourceGroup = useCallback(() => {
    router.push('/azure/iam/resource-groups/create');
  }, [router]);

  const handleDeleteResourceGroup = useCallback((name: string) => {
    const rg = resourceGroups.find(r => r.name === name);
    openDeleteDialog(name, '', rg?.name || name);
  }, [resourceGroups, openDeleteDialog]);

  const handleConfirmDelete = useCallback(() => {
    if (deleteDialogState.id) {
      executeDeleteResourceGroup(deleteDialogState.id);
      closeDeleteDialog();
    }
  }, [deleteDialogState.id, executeDeleteResourceGroup, closeDeleteDialog]);

  const [searchQuery, setSearchQuery] = useState('');

  // Custom filter function for Resource Group-specific filtering
  const filterFn = useCallback((rg: ResourceGroupInfo, filters: FilterValue): boolean => {
    if (filters.location && rg.location !== filters.location) return false;
    if (filters.provisioning_state && rg.provisioning_state !== filters.provisioning_state) return false;
    return true;
  }, []);

  // Apply search and filter using DataProcessor
  const filteredResourceGroups = useMemo(() => {
    let result = DataProcessor.search(resourceGroups, searchQuery, {
      keys: ['name', 'location', 'provisioning_state'],
      threshold: 0.3,
    });

    result = DataProcessor.filter(result, filters, filterFn);
    
    return result as ResourceGroupInfo[];
  }, [resourceGroups, searchQuery, filters, filterFn]);

  const isSearching = searchQuery.length > 0;

  const clearSearch = useCallback(() => {
    setSearchQuery('');
  }, []);

  const handleToggleFilters = useCallback(() => {
    setShowFilters(prev => !prev);
  }, [setShowFilters]);

  // Pagination
  const {
    page,
    paginatedItems: paginatedResourceGroups,
    setPage,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredResourceGroups, {
    totalItems: filteredResourceGroups.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  const handleBulkDeleteResourceGroups = useCallback(async (names: string[]) => {
    try {
      await handleBulkDelete(names, filteredResourceGroups);
      setSelectedResourceGroupNames([]);
    } catch {
      // Error already handled in hook
    }
  }, [handleBulkDelete, filteredResourceGroups, setSelectedResourceGroupNames]);

  const handlePageSizeChange = useCallback((newSize: number) => {
    setPageSize(newSize);
    setPaginationPageSize(newSize);
  }, [setPageSize, setPaginationPageSize]);

  // Filter configurations
  const filterConfigs: FilterConfig[] = useMemo(() => {
    if (!selectedCredentialId || resourceGroups.length === 0) return [];

    const locations = Array.from(new Set(resourceGroups.map(rg => rg.location))).sort();
    const states = Array.from(new Set(resourceGroups.map(rg => rg.provisioning_state))).sort();

    return [
      {
        id: 'location',
        label: t('filters.location') || 'Location',
        type: 'select',
        options: locations.map((loc, idx) => ({ 
          id: `location-${idx}`, 
          value: loc, 
          label: loc 
        })),
      },
      {
        id: 'provisioning_state',
        label: 'Provisioning State',
        type: 'select',
        options: states.map((state, idx) => ({ 
          id: `state-${idx}`, 
          value: state, 
          label: state 
        })),
      },
    ];
  }, [selectedCredentialId, resourceGroups, t]);

  // Empty state
  const { isEmpty, emptyStateComponent } = useEmptyState({
    credentials,
    selectedProvider,
    selectedCredentialId,
    filteredItems: filteredResourceGroups,
    resourceName: t('nav.resourceGroups'),
    serviceName: 'Azure IAM',
    onCreateClick: handleCreateResourceGroup,
    icon: Plus,
    isSearching,
    searchQuery,
  });

  return (
    <>
      <ResourceListPage
        title={t('nav.resourceGroups')}
        resourceName={t('nav.resourceGroups')}
        storageKey="resource-groups-page"
        header={<ResourceGroupsPageHeader />}
        items={filteredResourceGroups}
        isLoading={isLoadingResourceGroups}
        isEmpty={isEmpty}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        onSearchClear={clearSearch}
        isSearching={isSearching}
        searchPlaceholder="Search resource groups by name, location..."
        filterConfigs={selectedCredentialId && resourceGroups.length > 0 ? filterConfigs : []}
        filters={filters}
        onFiltersChange={setFilters}
        onFiltersClear={() => setFilters({})}
        showFilters={showFilters}
        onToggleFilters={handleToggleFilters}
        filterCount={Object.keys(filters).length}
        toolbar={
          selectedProvider === 'azure' && selectedCredentialId && filteredResourceGroups.length > 0 ? (
            <BulkActionsToolbar
              items={filteredResourceGroups}
              selectedIds={selectedResourceGroupNames}
              onSelectionChange={setSelectedResourceGroupNames}
              onBulkDelete={handleBulkDeleteResourceGroups}
              getItemDisplayName={(rg) => rg.name}
            />
          ) : null
        }
        additionalControls={
          selectedCredentialId && selectedProvider === 'azure' ? (
            <>
              <Button
                onClick={handleCreateResourceGroup}
                className="flex items-center"
              >
                <Plus className="mr-2 h-4 w-4" />
                Create Resource Group
              </Button>
              {resourceGroups.length > 0 && (
                <Button
                  variant="outline"
                  onClick={handleToggleFilters}
                  className="flex items-center"
                >
                  <Filter className="mr-2 h-4 w-4" />
                  Filters
                  {Object.keys(filters).length > 0 && (
                    <span className="ml-2 px-2 py-1 bg-gray-100 rounded text-sm">
                      {Object.keys(filters).length}
                    </span>
                  )}
                </Button>
              )}
            </>
          ) : null
        }
        emptyState={emptyStateComponent}
        content={
          selectedProvider === 'azure' && selectedCredentialId && filteredResourceGroups.length > 0 ? (
            <ResourceGroupTableDynamic
              resourceGroups={resourceGroups}
              filteredResourceGroups={filteredResourceGroups}
              paginatedResourceGroups={paginatedResourceGroups}
              selectedResourceGroupNames={selectedResourceGroupNames}
              onSelectionChange={setSelectedResourceGroupNames}
              onDelete={handleDeleteResourceGroup}
              page={page}
              pageSize={pageSize}
              onPageChange={setPage}
              onPageSizeChange={handlePageSizeChange}
              isDeleting={deleteResourceGroupMutation.isPending}
            />
          ) : emptyStateComponent
        }
        pageSize={pageSize}
        onPageSizeChange={handlePageSizeChange}
        searchResultsCount={filteredResourceGroups.length}
        skeletonColumns={6}
        skeletonRows={5}
        skeletonShowCheckbox={true}
        showFilterButton={false}
        showSearchResultsInfo={false}
      />

      {/* Delete Resource Group Confirmation Dialog */}
      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => {
          if (open && deleteDialogState.id) {
            openDeleteDialog(deleteDialogState.id, '', deleteDialogState.name);
          } else {
            closeDeleteDialog();
          }
        }}
        onConfirm={handleConfirmDelete}
        title="Delete Resource Group"
        description="이 Resource Group을 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다."
        isLoading={deleteResourceGroupMutation.isPending}
        resourceName={deleteDialogState.name}
        resourceNameLabel="Resource Group 이름"
      />
    </>
  );
}

const MemoizedResourceGroupsPageContent = React.memo(ResourceGroupsPageContent);

export default function ResourceGroupsPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <MemoizedResourceGroupsPageContent />
    </Suspense>
  );
}

