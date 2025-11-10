/**
 * VPCs Page (Refactored)
 * Virtual Private Cloud 관리 페이지 - 리팩토링된 버전
 */

'use client';

import { Suspense } from 'react';
import { useState, useMemo, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import dynamic from 'next/dynamic';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { usePagination } from '@/hooks/use-pagination';
import { EVENTS, UI } from '@/lib/constants';
import { Network, Filter, Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { useCredentialAutoSelect } from '@/hooks/use-credential-auto-select';
import { useWorkspaceStore } from '@/store/workspace';
import { useTranslation } from '@/hooks/use-translation';
import { DataProcessor } from '@/lib/data-processor';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import {
  useVPCs,
  useVPCActions,
  VPCsPageHeader,
} from '@/features/networks';
import type { CreateVPCForm, VPC } from '@/lib/types';

const VPCTable = dynamic(
  () => import('@/features/networks').then(mod => ({ default: mod.VPCTable })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={5} rows={5} showCheckbox={true} />,
  }
);

function VPCsPageContent() {
  const router = useRouter();
  const { t } = useTranslation();
  const { currentWorkspace } = useWorkspaceStore();

  const {
    vpcs,
    isLoadingVPCs,
    credentials,
    selectedProvider,
    selectedCredentialId,
    selectedRegion,
  } = useVPCs();
  
  // Auto-select credential if not selected
  useCredentialAutoSelect({
    enabled: !!currentWorkspace,
    resourceType: 'network',
    updateUrl: true,
  });

  const [filters, setFilters] = useState<FilterValue>({});
  const [showFilters, setShowFilters] = useState(false);
  const [selectedVPCIds, setSelectedVPCIds] = useState<string[]>([]);
  const [pageSize, setPageSize] = useState<number>(UI.PAGINATION.DEFAULT_PAGE_SIZE);

  const {
    deleteVPCMutation,
    handleBulkDeleteVPCs: handleBulkDelete,
    executeDeleteVPC,
  } = useVPCActions({
    selectedProvider,
    selectedCredentialId,
    selectedRegion,
  });

  const handleCreateVPC = () => {
    router.push('/networks/vpcs/create');
  };

  const [deleteDialogState, setDeleteDialogState] = useState<{
    open: boolean;
    vpcId: string | null;
    region: string | null;
    vpcName?: string;
  }>({
    open: false,
    vpcId: null,
    region: null,
    vpcName: undefined,
  });

  const handleDeleteVPC = (vpcId: string, region?: string) => {
    if (!region) return;
    const vpc = vpcs.find(v => v.id === vpcId);
    setDeleteDialogState({ open: true, vpcId, region, vpcName: vpc?.name || vpc?.id });
  };

  const handleConfirmDelete = () => {
    if (deleteDialogState.vpcId && deleteDialogState.region) {
      executeDeleteVPC(deleteDialogState.vpcId, deleteDialogState.region);
      setDeleteDialogState({ open: false, vpcId: null, region: null, vpcName: undefined });
    }
  };

  const [searchQuery, setSearchQuery] = useState('');

  // Custom filter function for VPC-specific filtering (memoized)
  const filterFn = useCallback((vpc: VPC, filters: FilterValue): boolean => {
    if (filters.state && vpc.state !== filters.state) return false;
    if (filters.is_default !== undefined) {
      const isDefault = filters.is_default === 'true';
      if (vpc.is_default !== isDefault) return false;
    }
    return true;
  }, []);

  // Apply search and filter using DataProcessor (memoized)
  const filteredVPCs = useMemo(() => {
    let result = DataProcessor.search(vpcs, searchQuery, {
      keys: ['name', 'id', 'state'],
      threshold: 0.3,
    });

    result = DataProcessor.filter(result, filters, filterFn);
    
    return result as VPC[];
  }, [vpcs, searchQuery, filters, filterFn]);

  const isSearching = searchQuery.length > 0;

  const clearSearch = useCallback(() => {
    setSearchQuery('');
  }, []);

  // Pagination
  const {
    page,
    paginatedItems: paginatedVPCs,
    setPage,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredVPCs, {
    totalItems: filteredVPCs.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });


  const handleBulkDeleteVPCs = useCallback(async (vpcIds: string[]) => {
    try {
      await handleBulkDelete(vpcIds, filteredVPCs);
      setSelectedVPCIds([]);
    } catch {
      // Error already handled in hook
    }
  }, [handleBulkDelete, filteredVPCs, setSelectedVPCIds]);

  const handlePageSizeChange = useCallback((newSize: number) => {
    setPageSize(newSize);
    setPaginationPageSize(newSize);
  }, [setPaginationPageSize]);

  // Filter configurations
  const filterConfigs: FilterConfig[] = useMemo(() => [
    {
      id: 'state',
      label: t('filters.status'),
      type: 'select',
      options: [
        { id: 'available', value: 'available', label: t('filters.available') },
        { id: 'pending', value: 'pending', label: t('status.pending') },
        { id: 'deleting', value: 'deleting', label: t('status.deleting') },
      ],
    },
    {
      id: 'is_default',
      label: t('common.type'),
      type: 'select',
      options: [
        { id: 'true', value: 'true', label: t('network.defaultVPC') },
        { id: 'false', value: 'false', label: t('network.customVPC') },
      ],
    },
  ], [t]);

  // Determine empty state
  const isEmpty = !selectedProvider || !selectedCredentialId || filteredVPCs.length === 0;

  // Empty state component
  const emptyStateComponent = credentials.length === 0 ? (
    <CredentialRequiredState serviceName={t('network.title')} />
  ) : !selectedProvider || !selectedCredentialId ? (
    <CredentialRequiredState
      title={t('credential.selectCredential')}
      description={t('credential.selectCredential')}
      serviceName={t('network.title')}
    />
  ) : filteredVPCs.length === 0 ? (
      <ResourceEmptyState
        resourceName={t('network.vpcs')}
        icon={Network}
        onCreateClick={handleCreateVPC}
        isSearching={isSearching}
      searchQuery={searchQuery}
      withCard={true}
    />
  ) : null;

  return (
    <>
    <ResourceListPage
        title={t('network.vpcs')}
        resourceName={t('network.vpcs')}
        storageKey="vpcs-page"
        header={<VPCsPageHeader />}
        items={filteredVPCs}
        isLoading={isLoadingVPCs}
        isEmpty={isEmpty}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        onSearchClear={clearSearch}
        isSearching={isSearching}
        searchPlaceholder={t('network.searchVPCsPlaceholder')}
        filterConfigs={selectedCredentialId && vpcs.length > 0 ? filterConfigs : []}
        filters={filters}
        onFiltersChange={setFilters}
        onFiltersClear={() => setFilters({})}
        showFilters={showFilters}
        onToggleFilters={() => setShowFilters(!showFilters)}
        filterCount={Object.keys(filters).length}
        toolbar={
          selectedProvider && selectedCredentialId && filteredVPCs.length > 0 ? (
            <BulkActionsToolbar
              items={filteredVPCs}
              selectedIds={selectedVPCIds}
              onSelectionChange={setSelectedVPCIds}
              onBulkDelete={handleBulkDeleteVPCs}
              getItemDisplayName={(vpc) => vpc.name}
            />
          ) : null
        }
        additionalControls={
          selectedCredentialId ? (
            <>
              <Button
                onClick={handleCreateVPC}
                className="flex items-center"
              >
                <Plus className="mr-2 h-4 w-4" />
                Create VPC
              </Button>
              {vpcs.length > 0 && (
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
              )}
            </>
          ) : null
        }
        emptyState={emptyStateComponent}
        content={
          selectedProvider && selectedCredentialId && filteredVPCs.length > 0 ? (
            <>
              <VPCTable
                vpcs={vpcs}
                filteredVPCs={filteredVPCs}
                paginatedVPCs={paginatedVPCs}
                selectedVPCIds={selectedVPCIds}
                onSelectionChange={setSelectedVPCIds}
                onDelete={handleDeleteVPC}
                selectedRegion={selectedRegion}
                page={page}
                pageSize={pageSize}
                onPageChange={setPage}
                onPageSizeChange={handlePageSizeChange}
                isDeleting={deleteVPCMutation.isPending}
              />
            </>
          ) : emptyStateComponent
        }
        pageSize={pageSize}
        onPageSizeChange={handlePageSizeChange}
        searchResultsCount={filteredVPCs.length}
        skeletonColumns={5}
        skeletonRows={5}
        skeletonShowCheckbox={true}
      showFilterButton={false}
      showSearchResultsInfo={false}
      />

      {/* Delete VPC Confirmation Dialog */}
      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => setDeleteDialogState({ ...deleteDialogState, open })}
        onConfirm={handleConfirmDelete}
        title={t('network.deleteVPC')}
        description="이 VPC를 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다."
        isLoading={deleteVPCMutation.isPending}
        resourceName={deleteDialogState.vpcName}
        resourceNameLabel="VPC 이름"
      />
    </>
  );
}

export default function VPCsPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <VPCsPageContent />
    </Suspense>
  );
}

