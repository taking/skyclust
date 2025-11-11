/**
 * Virtual Machines Page
 * Virtual Machine 관리 페이지
 * 
 * ResourceListPage 템플릿을 사용한 리팩토링 버전
 */

'use client';

import { Suspense } from 'react';
import { useMemo, useCallback } from 'react';
import * as React from 'react';
import dynamic from 'next/dynamic';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { EVENTS } from '@/lib/constants';
import { useWorkspaceStore } from '@/store/workspace';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Filter } from 'lucide-react';
import { FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useCredentialAutoSelect } from '@/hooks/use-credential-auto-select';
import { useCreateDialog } from '@/hooks/use-create-dialog';
import { useTranslation } from '@/hooks/use-translation';
import { useGenericResource } from '@/hooks/use-generic-resource';
import { useResourceListState } from '@/hooks/use-resource-list-state';
import { VMPageHeader, useVMs, useVMActions } from '@/features/vms';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import { log } from '@/lib/logging';
import type { CreateVMForm, VM } from '@/lib/types';

// Dynamic import for VMTable
const VMTable = dynamic(
  () => import('@/features/vms').then(mod => ({ default: mod.VMTable })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={7} rows={5} showCheckbox={true} />,
  }
);

function VMsPageContent() {
  const { success: showSuccess } = useToast();
  const { handleError } = useErrorHandler();
  const { t } = useTranslation();

  // Get workspace from store
  const { currentWorkspace } = useWorkspaceStore();

  // Get credential context from global store (Header에서 관리)
  const { selectedCredentialId } = useCredentialContext();
  
  // Auto-select credential if not selected
  useCredentialAutoSelect({
    enabled: !!currentWorkspace,
    resourceType: 'compute',
    updateUrl: true,
  });

  // Local state
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useCreateDialog(EVENTS.CREATE_DIALOG.VM);
  
  // 공통 리스트 상태 관리 (deleteDialogState와 showFilters만 사용)
  const {
    showFilters,
    setShowFilters,
    deleteDialogState,
    openDeleteDialog,
    closeDeleteDialog,
  } = useResourceListState({
    storageKey: 'vms-page',
  });

  // VMs hook
  const {
    credentials,
    vms,
    isLoading,
    createVMMutation,
    deleteVMMutation,
    startVMMutation,
    stopVMMutation,
  } = useVMs({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || '',
  });

  // Filter configurations
  const filterConfigs: FilterConfig[] = useMemo(() => [
    {
      id: 'status',
      label: t('filters.status'),
      type: 'multiselect',
      options: [
        { id: 'running', label: t('filters.running'), value: 'running' },
        { id: 'stopped', label: t('filters.stopped'), value: 'stopped' },
        { id: 'starting', label: t('filters.starting'), value: 'starting' },
        { id: 'stopping', label: t('filters.stopping'), value: 'stopping' },
      ],
    },
    {
      id: 'provider',
      label: t('filters.provider'),
      type: 'multiselect',
      options: [
        { id: 'aws', label: 'AWS', value: 'aws' },
        { id: 'gcp', label: 'GCP', value: 'gcp' },
        { id: 'azure', label: 'Azure', value: 'azure' },
      ],
    },
    {
      id: 'region',
      label: t('filters.region'),
      type: 'select',
      options: Array.from(new Set(vms.map(vm => vm.region)))
        .filter(Boolean)
        .map((region, idx) => ({
          id: `region-${idx}`,
          label: region,
          value: region,
        })),
    },
  ], [vms, t]);

  // Generic Resource Hook (검색, 필터, 페이지네이션 통합)
  const {
    searchQuery,
    setSearchQuery,
    isSearching,
    clearSearch,
    filters,
    setFilters,
    clearFilters,
    filteredItems: filteredVMs,
    paginatedItems: paginatedVMs,
    page,
    pageSize,
    setPage,
    setPageSize,
    selectedIds: selectedVMIds,
    setSelectedIds: setSelectedVMIds,
    clearSelection: _clearSelection,
  } = useGenericResource<VM>({
    resourceName: 'vms',
    items: vms,
    isLoading,
    searchKeys: ['name', 'id', 'region', 'instance_type', 'status', 'provider'],
    filterConfigs: selectedCredentialId && vms.length > 0 ? filterConfigs : [],
    filterFn: useCallback((vm: VM, filters: FilterValue): boolean => {
      if (filters.status) {
        const statusArray = Array.isArray(filters.status) ? filters.status : [filters.status];
        if (!statusArray.includes(vm.status)) return false;
      }
      if (filters.provider) {
        const providerArray = Array.isArray(filters.provider) ? filters.provider : [filters.provider];
        if (!providerArray.includes(vm.provider)) return false;
      }
      if (filters.region && vm.region !== filters.region) return false;
      return true;
    }, []),
  });

  // VM actions hook
  const {
    handleDeleteVM,
    handleStartVM,
    handleStopVM,
  } = useVMActions({
    workspaceId: currentWorkspace?.id,
    deleteMutation: deleteVMMutation,
    startMutation: startVMMutation,
    stopMutation: stopVMMutation,
    onSuccess: showSuccess,
    onError: (error: unknown) => handleError(error, { operation: 'vmAction', resource: 'VM' }),
  });

  // VM 액션 핸들러 메모이제이션 (VMTable에 전달)
  const handleVMStart = useCallback((vmId: string) => {
    handleStartVM(vmId, vms);
  }, [handleStartVM, vms]);

  const handleVMStop = useCallback((vmId: string) => {
    handleStopVM(vmId, vms);
  }, [handleStopVM, vms]);

  const handleVMDelete = useCallback((vmId: string) => {
    const vm = vms.find(v => v.id === vmId);
    openDeleteDialog(vmId, undefined, vm?.name);
  }, [vms, openDeleteDialog]);

  const handleConfirmDelete = useCallback(() => {
    if (deleteDialogState.id) {
      handleDeleteVM(deleteDialogState.id, vms);
      closeDeleteDialog();
    }
  }, [deleteDialogState.id, handleDeleteVM, vms, closeDeleteDialog]);

  // Event handlers
  const handleCreateVM = useCallback((data: CreateVMForm) => {
    if (!currentWorkspace) return;
    createVMMutation.mutate(
      { workspaceId: currentWorkspace.id, data },
      {
        onSuccess: () => {
          setIsCreateDialogOpen(false);
          showSuccess(t('vm.creationInitiated'));
        },
        onError: (error: unknown) => {
          handleError(error, { operation: 'createVM', resource: 'VM' });
        },
      }
    );
  }, [currentWorkspace, createVMMutation, setIsCreateDialogOpen, showSuccess, t, handleError]);

  const handlePageSizeChange = useCallback((newSize: number) => {
    setPageSize(newSize);
  }, [setPageSize]);

  // Toggle filters handler (memoized)
  const handleToggleFilters = useCallback(() => {
    setShowFilters(prev => !prev);
  }, []);

  // Empty toggle sort handler (memoized)
  const handleToggleSort = useCallback(() => {
    // No-op for now
  }, []);

  // Determine empty state
  const isEmpty = !selectedCredentialId || filteredVMs.length === 0;

  // Empty state component
  const emptyStateComponent = credentials.length === 0 ? (
    <CredentialRequiredState serviceName={t('vm.title')} />
  ) : !selectedCredentialId ? (
    <CredentialRequiredState
      title={t('credential.selectCredential')}
      description={t('credential.selectCredential')}
      serviceName={t('vm.title')}
    />
  ) : filteredVMs.length === 0 ? (
    <ResourceEmptyState
      resourceName={t('nav.vms')}
      title={t('vm.noVMsFound')}
      description={t('vm.createFirst')}
      onCreateClick={() => setIsCreateDialogOpen(true)}
      withCard={true}
    />
  ) : null;

  return (
    <>
    <ResourceListPage
        title={t('vm.title')}
        resourceName={t('nav.vms')}
        storageKey="vms-page"
        header={
          <VMPageHeader
            workspaceName={currentWorkspace?.name}
            credentials={credentials}
            selectedCredentialId={selectedCredentialId || ''}
            onCredentialChange={() => {}} // Handled by Header
            onCreateVM={handleCreateVM}
            isCreatePending={createVMMutation.isPending}
            isCreateDialogOpen={isCreateDialogOpen}
            onCreateDialogChange={setIsCreateDialogOpen}
          />
        }
      items={filteredVMs}
      isLoading={isLoading}
      isEmpty={isEmpty}
      searchQuery={searchQuery}
      onSearchChange={setSearchQuery}
      onSearchClear={clearSearch}
      isSearching={isSearching}
      searchPlaceholder="Search VMs by name, provider, instance type, region, or status..."
      filterConfigs={selectedCredentialId && vms.length > 0 ? filterConfigs : []}
      filters={filters}
      onFiltersChange={setFilters}
      onFiltersClear={clearFilters}
      showFilters={showFilters}
      onToggleFilters={useCallback(() => setShowFilters(prev => !prev), [])}
      filterCount={Object.keys(filters).length}
      toolbar={
        selectedCredentialId && filteredVMs.length > 0 ? (
          <BulkActionsToolbar
            items={filteredVMs}
            selectedIds={selectedVMIds}
            onSelectionChange={setSelectedVMIds}
            onBulkDelete={(vmIds) => {
              // TODO: Implement bulk delete
              log.debug('Bulk delete VMs', { vmIds });
            }}
            getItemDisplayName={(vm) => vm.name}
          />
        ) : null
      }
      additionalControls={
        selectedCredentialId && vms.length > 0 ? (
          <>
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
          </>
        ) : null
      }
      emptyState={emptyStateComponent}
      content={
        selectedCredentialId && filteredVMs.length > 0 ? (
          <>
            <Card>
              <CardHeader>
                <CardTitle>VMs</CardTitle>
                <CardDescription>
                  {filteredVMs.length} of {vms.length} VM{vms.length !== 1 ? 's' : ''} 
                  {isSearching && ` (${searchQuery})`}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <VMTable
                  vms={paginatedVMs}
                  sortConfig={[]}
                  onToggleSort={handleToggleSort}
                  onStart={handleVMStart}
                  onStop={handleVMStop}
                  onDelete={handleVMDelete}
                  isStarting={startVMMutation.isPending}
                  isStopping={stopVMMutation.isPending}
                  isDeleting={deleteVMMutation.isPending}
                  page={page}
                  pageSize={pageSize}
                  total={filteredVMs.length}
                  onPageChange={setPage}
                  onPageSizeChange={handlePageSizeChange}
                />
              </CardContent>
            </Card>
          </>
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

      {/* Delete VM Confirmation Dialog */}
      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => {
          if (open && deleteDialogState.id) {
            openDeleteDialog(deleteDialogState.id, deleteDialogState.region || undefined, deleteDialogState.name);
          } else {
            closeDeleteDialog();
          }
        }}
        onConfirm={handleConfirmDelete}
        title={t('vm.deleteVM')}
        description="이 VM을 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다."
        isLoading={deleteVMMutation.isPending}
        resourceName={deleteDialogState.name}
        resourceNameLabel="VM 이름"
      />
    </>
  );
}

const MemoizedVMsPageContent = React.memo(VMsPageContent);

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
      <MemoizedVMsPageContent />
    </Suspense>
  );
}

