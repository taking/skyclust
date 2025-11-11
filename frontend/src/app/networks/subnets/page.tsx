/**
 * Subnets Page (Refactored)
 * 서브넷 관리 페이지 - 리팩토링된 버전
 */

'use client';

import { Suspense } from 'react';
import { useState, useMemo, useCallback, useEffect } from 'react';
import * as React from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import dynamic from 'next/dynamic';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { usePagination } from '@/hooks/use-pagination';
import { EVENTS, UI } from '@/lib/constants';
import { Layers, Filter, Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { useCredentialAutoSelect } from '@/hooks/use-credential-auto-select';
import { useWorkspaceStore } from '@/store/workspace';
import { useTranslation } from '@/hooks/use-translation';
import { DataProcessor } from '@/lib/data';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import {
  useSubnetActions,
  SubnetsPageHeader,
} from '@/features/networks';
import { useNetworkResources } from '@/features/networks/hooks/use-network-resources';
import type { CreateSubnetForm, Subnet } from '@/lib/types';

const SubnetTable = dynamic(
  () => import('@/features/networks').then(mod => ({ default: mod.SubnetTable })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={6} rows={5} showCheckbox={true} />,
  }
);

function SubnetsPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { t } = useTranslation();
  const { currentWorkspace } = useWorkspaceStore();

  const {
    subnets = [],
    isLoadingSubnets = false,
    vpcs,
    selectedVPCId = '',
    setSelectedVPCId = () => {},
    credentials,
    selectedProvider,
    selectedCredentialId,
    selectedRegion,
  } = useNetworkResources({ resourceType: 'subnets', requireVPC: true });
  
  // Auto-select credential if not selected
  useCredentialAutoSelect({
    enabled: !!currentWorkspace,
    resourceType: 'network',
    updateUrl: true,
  });

  // URL 파라미터에서 VPC ID를 가져와서 자동 선택
  useEffect(() => {
    const vpcIdFromUrl = searchParams?.get('vpc_id');
    if (vpcIdFromUrl && vpcIdFromUrl !== selectedVPCId) {
      // VPC가 존재하는지 확인
      const vpcExists = vpcs.some(vpc => vpc.id === vpcIdFromUrl);
      if (vpcExists) {
        setSelectedVPCId(vpcIdFromUrl);
        // URL에서 vpc_id 파라미터 제거 (한 번만 적용)
        const newParams = new URLSearchParams(searchParams?.toString());
        newParams.delete('vpc_id');
        const newUrl = newParams.toString() 
          ? `${window.location.pathname}?${newParams.toString()}`
          : window.location.pathname;
        router.replace(newUrl);
      }
    }
  }, [searchParams, selectedVPCId, vpcs, setSelectedVPCId, router]);

  const [filters, setFilters] = useState<FilterValue>({});
  const [showFilters, setShowFilters] = useState(false);
  const [selectedSubnetIds, setSelectedSubnetIds] = useState<string[]>([]);
  const [pageSize, setPageSize] = useState<number>(UI.PAGINATION.DEFAULT_PAGE_SIZE);

  const {
    deleteSubnetMutation,
    handleBulkDeleteSubnets: handleBulkDelete,
    executeDeleteSubnet,
  } = useSubnetActions({
    selectedProvider,
    selectedCredentialId,
  });

  const handleCreateSubnet = useCallback(() => {
    const params = new URLSearchParams();
    if (selectedVPCId) {
      params.set('vpc_id', selectedVPCId);
    }
    router.push(`/networks/subnets/create${params.toString() ? `?${params.toString()}` : ''}`);
  }, [router, selectedVPCId]);

  const [deleteDialogState, setDeleteDialogState] = useState<{
    open: boolean;
    subnetId: string | null;
    region: string | null;
    subnetName?: string;
  }>({
    open: false,
    subnetId: null,
    region: null,
    subnetName: undefined,
  });

  const handleDeleteSubnet = useCallback((subnetId: string, region: string) => {
    const subnet = subnets.find(s => s.id === subnetId);
    setDeleteDialogState({ open: true, subnetId, region, subnetName: subnet?.name || subnet?.id });
  }, [subnets]);

  const handleConfirmDelete = useCallback(() => {
    if (deleteDialogState.subnetId && deleteDialogState.region) {
      executeDeleteSubnet(deleteDialogState.subnetId, deleteDialogState.region);
      setDeleteDialogState({ open: false, subnetId: null, region: null, subnetName: undefined });
    }
  }, [deleteDialogState.subnetId, deleteDialogState.region, executeDeleteSubnet]);

  // Search functionality
  const [searchQuery, setSearchQuery] = useState('');

  // Custom filter function for subnet filtering (memoized)
  const filterFn = useCallback((subnet: Subnet, filters: FilterValue): boolean => {
    if (filters.state && subnet.state !== filters.state) return false;
    return true;
  }, []);

  // Filtered subnets (memoized for consistency)
  const filteredSubnets = useMemo(() => {
    let result = DataProcessor.search(subnets, searchQuery, {
      keys: ['name', 'id', 'cidr_block', 'state'],
      threshold: 0.3,
    });

    result = DataProcessor.filter(result, filters, filterFn);
    
    return result;
  }, [subnets, searchQuery, filters, filterFn]);

  const isSearching = searchQuery.length > 0;

  const clearSearch = useCallback(() => {
    setSearchQuery('');
  }, []);

  // Pagination
  const {
    page,
    paginatedItems: paginatedSubnets,
    setPage,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredSubnets, {
    totalItems: filteredSubnets.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });


  const handleBulkDeleteSubnets = useCallback(async (subnetIds: string[]) => {
    try {
      await handleBulkDelete(subnetIds, filteredSubnets);
      setSelectedSubnetIds([]);
    } catch {
      // Error already handled in hook
    }
  }, [handleBulkDelete, filteredSubnets]);

  const handlePageSizeChange = useCallback((newSize: number) => {
    setPageSize(newSize);
    setPaginationPageSize(newSize);
  }, [setPaginationPageSize]);

  const handleVPCChange = useCallback((vpcId: string) => {
    setSelectedVPCId(vpcId);
  }, []);

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
  ], [t]);

  // Determine empty state
  const isEmpty = !selectedProvider || !selectedCredentialId || !selectedVPCId || !selectedRegion || filteredSubnets.length === 0;

  // Empty state component
  const emptyStateComponent = credentials.length === 0 ? (
    <CredentialRequiredState serviceName={t('network.title')} />
  ) : !selectedProvider || !selectedCredentialId ? (
    <CredentialRequiredState
      title={t('credential.selectCredential')}
      description={t('credential.selectCredential')}
      serviceName={t('network.title')}
    />
  ) : !selectedVPCId || !selectedRegion ? (
    <Card>
      <CardContent className="flex flex-col items-center justify-center py-12">
        <Layers className="h-12 w-12 text-gray-400 mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">
          {t('network.selectVPCAndRegion')}
        </h3>
        <p className="text-sm text-gray-500 text-center">
          {t('network.selectVPCAndRegionMessage')}
        </p>
      </CardContent>
    </Card>
  ) : filteredSubnets.length === 0 ? (
    <ResourceEmptyState
      resourceName={t('network.subnets')}
      icon={Layers}
      onCreateClick={() => setIsCreateDialogOpen(true)}
      description={t('network.noSubnetsFoundForVPC')}
      withCard={true}
    />
  ) : null;

  // Header component
  const header = (
    <SubnetsPageHeader
      selectedProvider={selectedProvider}
      selectedCredentialId={selectedCredentialId}
      selectedVPCId={selectedVPCId}
      vpcs={vpcs}
      onVPCChange={handleVPCChange}
    />
  );

  return (
    <>
    <ResourceListPage
        title={t('network.subnets')}
        resourceName={t('network.subnets')}
        storageKey="subnets-page"
        header={header}
        items={filteredSubnets}
        isLoading={isLoadingSubnets}
        isEmpty={isEmpty}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        onSearchClear={clearSearch}
        isSearching={isSearching}
        searchPlaceholder={t('network.searchSubnetsPlaceholder')}
        filterConfigs={selectedCredentialId && selectedVPCId && subnets.length > 0 ? filterConfigs : []}
        filters={filters}
        onFiltersChange={setFilters}
        onFiltersClear={() => setFilters({})}
        showFilters={showFilters}
        onToggleFilters={() => setShowFilters(!showFilters)}
        filterCount={Object.keys(filters).length}
        toolbar={
          selectedProvider && selectedCredentialId && selectedVPCId && selectedRegion && filteredSubnets.length > 0 ? (
            <BulkActionsToolbar
              items={filteredSubnets}
              selectedIds={selectedSubnetIds}
              onSelectionChange={setSelectedSubnetIds}
              onBulkDelete={handleBulkDeleteSubnets}
              getItemDisplayName={(subnet) => subnet.name}
            />
          ) : null
        }
        additionalControls={
          selectedCredentialId && selectedVPCId ? (
            <>
              <Button
                onClick={handleCreateSubnet}
                className="flex items-center"
                disabled={!selectedVPCId}
              >
                <Plus className="mr-2 h-4 w-4" />
                Create Subnet
              </Button>
              {subnets.length > 0 && (
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
          selectedProvider && selectedCredentialId && selectedVPCId && selectedRegion && filteredSubnets.length > 0 ? (
            <>
              <SubnetTable
                subnets={subnets}
                filteredSubnets={filteredSubnets}
                paginatedSubnets={paginatedSubnets}
                selectedSubnetIds={selectedSubnetIds}
                onSelectionChange={setSelectedSubnetIds}
                onDelete={handleDeleteSubnet}
                searchQuery={searchQuery}
                onSearchChange={setSearchQuery}
                onSearchClear={clearSearch}
                page={page}
                pageSize={pageSize}
                onPageChange={setPage}
                onPageSizeChange={handlePageSizeChange}
                isDeleting={deleteSubnetMutation.isPending}
              />
            </>
          ) : emptyStateComponent
        }
        pageSize={pageSize}
        onPageSizeChange={handlePageSizeChange}
        searchResultsCount={filteredSubnets.length}
        skeletonColumns={6}
        skeletonRows={5}
        skeletonShowCheckbox={true}
      showFilterButton={false}
      showSearchResultsInfo={false}
      />

      {/* Delete Subnet Confirmation Dialog */}
      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => setDeleteDialogState({ ...deleteDialogState, open })}
        onConfirm={handleConfirmDelete}
        title={t('network.deleteSubnet')}
        description="이 서브넷을 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다."
        isLoading={deleteSubnetMutation.isPending}
        resourceName={deleteDialogState.subnetName}
        resourceNameLabel="서브넷 이름"
      />
    </>
  );
}

const MemoizedSubnetsPageContent = React.memo(SubnetsPageContent);

export default function SubnetsPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <MemoizedSubnetsPageContent />
    </Suspense>
  );
}

