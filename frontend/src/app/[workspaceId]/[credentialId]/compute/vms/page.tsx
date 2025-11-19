/**
 * Virtual Machines Page
 * Virtual Machine 관리 페이지
 * 
 * 새로운 라우팅 구조: /{workspaceId}/{credentialId}/compute/vms
 */

'use client';

import { Suspense } from 'react';
import { useState, useMemo, useCallback, useEffect } from 'react';
import * as React from 'react';
import dynamic from 'next/dynamic';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { EVENTS } from '@/lib/constants';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Filter } from 'lucide-react';
import { FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { useTranslation } from '@/hooks/use-translation';
import { useGenericResource } from '@/hooks/use-generic-resource';
import { useResourceListState } from '@/hooks/use-resource-list-state';
import { VMPageHeader, useVMs, useVMActions } from '@/features/vms';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import { log } from '@/lib/logging';
import { useSSEStatus } from '@/hooks/use-sse-status';
import { useSSESubscription } from '@/hooks/use-sse-subscription';
import { useSSEErrorRecovery } from '@/hooks/use-sse-error-recovery';
import { useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '@/lib/query';
import { usePageRefresh } from '@/hooks/use-page-refresh';
import { useRequiredResourceContext } from '@/hooks/use-resource-context';
import { buildResourceDetailPath } from '@/lib/routing/helpers';
import type { CreateVMForm, VM } from '@/lib/types';

const VMTable = dynamic(
  () => import('@/features/vms').then(mod => ({ default: mod.VMTable })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={7} rows={5} showCheckbox={true} />,
  }
);

function VMsPageContent() {
  const { handleError } = useErrorHandler();
  const { t } = useTranslation();

  // Path Parameter에서 컨텍스트 추출
  const { workspaceId, credentialId, region } = useRequiredResourceContext();

  const queryClient = useQueryClient();

  // 페이지 새로고침 시 쿼리 무효화 및 재요청
  usePageRefresh({
    queryKeys: [
      queryKeys.vms.list(workspaceId),
    ],
    refetch: true,
    trigger: 'mount',
  });

  // SSE 상태 확인
  const { status: sseStatus } = useSSEStatus();

  // SSE 이벤트 구독 (VM 실시간 업데이트)
  useSSESubscription({
    eventTypes: [
      'vm-created',
      'vm-updated',
      'vm-deleted',
      'vm-list',
    ],
    enabled: sseStatus.isConnected && !!workspaceId,
  });

  // SSE 에러 복구
  useSSEErrorRecovery({
    autoReconnect: true,
    showNotifications: true,
  });

  // 공통 리스트 상태 관리
  const {
    filters,
    setFilters,
    showFilters,
    setShowFilters,
    selectedIds: selectedVMIds,
    setSelectedIds: setSelectedVMIds,
    pageSize,
    setPageSize,
    deleteDialogState,
    openDeleteDialog,
    closeDeleteDialog,
  } = useResourceListState({
    storageKey: 'vms-page',
  });

  // VMs hook
  const {
    vms,
    isLoading,
    createVMMutation,
    deleteVMMutation,
    startVMMutation,
    stopVMMutation,
  } = useVMs({
    workspaceId: workspaceId || '',
  });

  // VM actions hook
  const {
    handleBulkDelete,
    handleBulkStart,
    handleBulkStop,
    handleBulkRestart,
  } = useVMActions({
    workspaceId: workspaceId || '',
    deleteMutation: deleteVMMutation,
    startMutation: startVMMutation,
    stopMutation: stopVMMutation,
  });

  // Filter configurations
  const filterConfigs: FilterConfig[] = useMemo(() => [
    {
      id: 'status',
      label: t('filters.status'),
      type: 'select',
      options: [
        { id: 'running', value: 'running', label: t('filters.running') },
        { id: 'stopped', value: 'stopped', label: t('filters.stopped') },
        { id: 'pending', value: 'pending', label: t('status.pending') },
      ],
    },
  ], [t]);

  // Custom filter function
  const filterFn = useCallback((vm: VM, filters: FilterValue): boolean => {
    if (filters.status && vm.status !== filters.status) return false;
    return true;
  }, []);

  // Apply filters
  const filteredVMs = useMemo(() => {
    return vms.filter(vm => filterFn(vm, filters));
  }, [vms, filters, filterFn]);

  // Pagination state
  const [page, setPage] = useState(1);

  // Pagination calculation
  const totalPages = Math.max(1, Math.ceil(filteredVMs.length / pageSize));
  const currentPage = Math.min(page, totalPages);
  const paginatedVMs = useMemo(() => {
    const startIndex = (currentPage - 1) * pageSize;
    const endIndex = startIndex + pageSize;
    return filteredVMs.slice(startIndex, endIndex);
  }, [filteredVMs, currentPage, pageSize]);

  const handlePageSizeChange = useCallback((newSize: number) => {
    setPageSize(newSize);
    // Adjust page if current page would be out of bounds
    const newTotalPages = Math.max(1, Math.ceil(filteredVMs.length / newSize));
    if (currentPage > newTotalPages) {
      setPage(newTotalPages);
    }
  }, [setPageSize, filteredVMs.length, currentPage]);

  const handleDeleteVM = useCallback((vmId: string) => {
    openDeleteDialog(vmId, undefined, vmId);
  }, [openDeleteDialog]);

  const handleConfirmDelete = useCallback(() => {
    if (!deleteDialogState.id) return;
    
    deleteVMMutation.mutate(deleteDialogState.id, {
      onSuccess: () => {
        closeDeleteDialog();
      },
      onError: (error: unknown) => {
        handleError(error, { operation: 'deleteVM', resource: 'VM' });
      },
    });
  }, [deleteDialogState.id, deleteVMMutation, closeDeleteDialog, handleError]);

  const handleVMClick = useCallback((vmId: string) => {
    if (!workspaceId || !credentialId) return;
    const path = buildResourceDetailPath(
      workspaceId,
      credentialId,
      'compute',
      'vms',
      vmId,
      { region: region || undefined }
    );
    // router.push will be handled by the component that uses this
  }, [workspaceId, credentialId, region]);

  // Empty handlers (memoized to prevent unnecessary re-renders)
  const handleToggleSort = useCallback(() => {
    // Sort functionality not implemented yet
  }, []);

  const handleStart = useCallback(() => {
    // Start functionality not implemented yet
  }, []);

  const handleStop = useCallback(() => {
    // Stop functionality not implemented yet
  }, []);

  // Determine empty state
  const isEmpty = !credentialId || filteredVMs.length === 0;

  // Empty state component
  const emptyStateComponent = !credentialId ? (
    <CredentialRequiredState serviceName={t('compute.title')} />
  ) : filteredVMs.length === 0 ? (
    <ResourceEmptyState
      resourceName={t('compute.vms')}
      title={t('compute.noVMsFound')}
      description={t('compute.createFirst')}
      isSearching={false}
      searchQuery=""
      withCard={true}
    />
  ) : null;

  return (
    <>
      <ResourceListPage
        title={t('compute.vms')}
        resourceName={t('compute.vms')}
        storageKey="vms-page"
        header={null}
        items={filteredVMs}
        isLoading={isLoading}
        isEmpty={isEmpty}
        searchQuery=""
        onSearchChange={() => {}}
        onSearchClear={() => {}}
        isSearching={false}
        searchPlaceholder={t('compute.searchVMs') || 'Search VMs...'}
        filterConfigs={credentialId && vms.length > 0 ? filterConfigs : []}
        filters={filters}
        onFiltersChange={setFilters}
        onFiltersClear={() => setFilters({})}
        showFilters={showFilters}
        onToggleFilters={() => setShowFilters(!showFilters)}
        filterCount={Object.keys(filters).length}
        toolbar={
          credentialId && filteredVMs.length > 0 ? (
            <BulkActionsToolbar
              items={filteredVMs}
              selectedIds={selectedVMIds}
              onSelectionChange={setSelectedVMIds}
              onBulkDelete={(ids) => handleBulkDelete(ids, filteredVMs)}
              getItemDisplayName={(vm) => vm.name || vm.id}
            />
          ) : null
        }
        additionalControls={
          credentialId && vms.length > 0 ? (
            <Button
              variant="outline"
              onClick={() => setShowFilters(!showFilters)}
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
          ) : null
        }
        emptyState={emptyStateComponent}
        content={
          credentialId && filteredVMs.length > 0 ? (
            <Card>
              <CardHeader>
                <CardTitle>Virtual Machines</CardTitle>
                <CardDescription>
                  {filteredVMs.length} of {vms.length} VM{vms.length !== 1 ? 's' : ''}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <VMTable
                  vms={paginatedVMs}
                  sortConfig={[]}
                  onToggleSort={handleToggleSort}
                  onStart={handleStart}
                  onStop={handleStop}
                  onDelete={handleDeleteVM}
                  page={currentPage}
                  pageSize={pageSize}
                  total={filteredVMs.length}
                  onPageChange={setPage}
                  onPageSizeChange={handlePageSizeChange}
                  isDeleting={deleteVMMutation.isPending}
                />
              </CardContent>
            </Card>
          ) : emptyStateComponent
        }
        pageSize={pageSize}
        onPageSizeChange={handlePageSizeChange}
        searchResultsCount={filteredVMs.length}
        skeletonColumns={7}
        skeletonRows={5}
        skeletonShowCheckbox={true}
        showFilterButton={false}
        showSearchResultsInfo={false}
      />

      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => {
          if (open && deleteDialogState.id) {
            openDeleteDialog(deleteDialogState.id, undefined, deleteDialogState.name);
          } else {
            closeDeleteDialog();
          }
        }}
        onConfirm={handleConfirmDelete}
        title={t('compute.deleteVM')}
        description={deleteDialogState.id ? t('compute.confirmDeleteVM', { vmId: deleteDialogState.id }) : ''}
        isLoading={deleteVMMutation.isPending}
        resourceName={deleteDialogState.name || deleteDialogState.id || undefined}
        resourceNameLabel="VM ID"
      />
    </>
  );
}

export default function VMsPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <VMsPageContent />
    </Suspense>
  );
}

