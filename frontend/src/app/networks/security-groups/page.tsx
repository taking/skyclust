/**
 * Security Groups Page (Refactored)
 * 보안 그룹 관리 페이지 - 리팩토링된 버전
 */

'use client';

import { Suspense } from 'react';
import { useState, useMemo, useCallback, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import dynamic from 'next/dynamic';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { usePagination } from '@/hooks/use-pagination';
import { EVENTS, UI } from '@/lib/constants';
import { Shield, Filter, Plus } from 'lucide-react';
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
  useSecurityGroups,
  useSecurityGroupActions,
  SecurityGroupsPageHeader,
} from '@/features/networks';
import type { CreateSecurityGroupForm, SecurityGroup } from '@/lib/types';

const SecurityGroupTable = dynamic(
  () => import('@/features/networks').then(mod => ({ default: mod.SecurityGroupTable })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={4} rows={5} showCheckbox={true} />,
  }
);

function SecurityGroupsPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { t } = useTranslation();
  const { currentWorkspace } = useWorkspaceStore();

  const {
    securityGroups,
    isLoadingSecurityGroups,
    vpcs,
    selectedVPCId,
    setSelectedVPCId,
    credentials,
    selectedProvider,
    selectedCredentialId,
    selectedRegion,
  } = useSecurityGroups();
  
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
  const [selectedSecurityGroupIds, setSelectedSecurityGroupIds] = useState<string[]>([]);
  const [pageSize, setPageSize] = useState(UI.PAGINATION.DEFAULT_PAGE_SIZE);

  const {
    deleteSecurityGroupMutation,
    handleBulkDeleteSecurityGroups: handleBulkDelete,
    executeDeleteSecurityGroup,
  } = useSecurityGroupActions({
    selectedProvider,
    selectedCredentialId,
  });

  const handleCreateSecurityGroup = () => {
    const params = new URLSearchParams();
    if (selectedVPCId) {
      params.set('vpc_id', selectedVPCId);
    }
    router.push(`/networks/security-groups/create${params.toString() ? `?${params.toString()}` : ''}`);
  };

  const [deleteDialogState, setDeleteDialogState] = useState<{
    open: boolean;
    securityGroupId: string | null;
    region: string | null;
    securityGroupName?: string;
  }>({
    open: false,
    securityGroupId: null,
    region: null,
    securityGroupName: undefined,
  });

  const handleDeleteSecurityGroup = (securityGroupId: string, region: string) => {
    const securityGroup = securityGroups.find(sg => sg.id === securityGroupId);
    setDeleteDialogState({ open: true, securityGroupId, region, securityGroupName: securityGroup?.name || securityGroup?.id });
  };

  const handleConfirmDelete = () => {
    if (deleteDialogState.securityGroupId && deleteDialogState.region) {
      executeDeleteSecurityGroup(deleteDialogState.securityGroupId, deleteDialogState.region);
      setDeleteDialogState({ open: false, securityGroupId: null, region: null, securityGroupName: undefined });
    }
  };

  // Search functionality
  const [searchQuery, setSearchQuery] = useState('');

  // Custom filter function for security group filtering (memoized)
  const filterFn = useCallback((_sg: SecurityGroup, _filters: FilterValue): boolean => {
    // Add any security group specific filters here
    return true;
  }, []);

  // Filtered security groups (memoized for consistency)
  const filteredSecurityGroups = useMemo(() => {
    let result = DataProcessor.search(securityGroups, searchQuery, {
      keys: ['name', 'id', 'description'],
      threshold: 0.3,
    });

    result = DataProcessor.filter(result, filters, filterFn);
    
    return result;
  }, [securityGroups, searchQuery, filters, filterFn]);

  const isSearching = searchQuery.length > 0;

  const clearSearch = useCallback(() => {
    setSearchQuery('');
  }, []);

  // Pagination
  const {
    page,
    paginatedItems: paginatedSecurityGroups,
    setPage,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredSecurityGroups, {
    totalItems: filteredSecurityGroups.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });


  const handleBulkDeleteSecurityGroups = useCallback(async (securityGroupIds: string[]) => {
    try {
      await handleBulkDelete(securityGroupIds, filteredSecurityGroups);
      setSelectedSecurityGroupIds([]);
    } catch {
      // Error already handled in hook
    }
  }, [handleBulkDelete, filteredSecurityGroups]);

  const handlePageSizeChange = useCallback((newSize: number) => {
    setPageSize(newSize);
    setPaginationPageSize(newSize);
  }, [setPaginationPageSize]);

  const handleVPCChange = useCallback((vpcId: string) => {
    setSelectedVPCId(vpcId);
  }, []);

  // Filter configurations
  const filterConfigs: FilterConfig[] = useMemo(() => [], []);

  // Determine empty state
  const isEmpty = !selectedProvider || !selectedCredentialId || !selectedVPCId || !selectedRegion || filteredSecurityGroups.length === 0;

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
        <Shield className="h-12 w-12 text-gray-400 mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">
          {t('network.selectVPCAndRegion')}
        </h3>
        <p className="text-sm text-gray-500 text-center">
          {t('network.selectVPCAndRegionMessage')}
        </p>
      </CardContent>
    </Card>
  ) : filteredSecurityGroups.length === 0 ? (
      <ResourceEmptyState
        resourceName={t('network.securityGroups')}
        icon={Shield}
        onCreateClick={handleCreateSecurityGroup}
        description={t('network.noSecurityGroupsFoundForVPC')}
        withCard={true}
      />
  ) : null;

  // Header component
  const header = (
    <SecurityGroupsPageHeader
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
        title={t('network.securityGroups')}
        resourceName={t('network.securityGroups')}
        storageKey="security-groups-page"
        header={header}
        items={filteredSecurityGroups}
        isLoading={isLoadingSecurityGroups}
        isEmpty={isEmpty}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        onSearchClear={clearSearch}
        isSearching={isSearching}
        searchPlaceholder={t('network.searchSecurityGroupsPlaceholder')}
        filterConfigs={selectedCredentialId && selectedVPCId && securityGroups.length > 0 ? filterConfigs : []}
        filters={filters}
        onFiltersChange={setFilters}
        onFiltersClear={() => setFilters({})}
        showFilters={showFilters}
        onToggleFilters={() => setShowFilters(!showFilters)}
        filterCount={Object.keys(filters).length}
        toolbar={
          selectedProvider && selectedCredentialId && selectedVPCId && selectedRegion && filteredSecurityGroups.length > 0 ? (
            <BulkActionsToolbar
              items={filteredSecurityGroups}
              selectedIds={selectedSecurityGroupIds}
              onSelectionChange={setSelectedSecurityGroupIds}
              onBulkDelete={handleBulkDeleteSecurityGroups}
              getItemDisplayName={(sg) => sg.name}
            />
          ) : null
        }
        additionalControls={
          selectedCredentialId && selectedVPCId ? (
            <>
              <Button
                onClick={handleCreateSecurityGroup}
                className="flex items-center"
                disabled={!selectedVPCId}
              >
                <Plus className="mr-2 h-4 w-4" />
                Create Security Group
              </Button>
              {securityGroups.length > 0 && (
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
          selectedProvider && selectedCredentialId && selectedVPCId && selectedRegion && filteredSecurityGroups.length > 0 ? (
            <>
              <SecurityGroupTable
                securityGroups={securityGroups}
                filteredSecurityGroups={filteredSecurityGroups}
                paginatedSecurityGroups={paginatedSecurityGroups}
                selectedSecurityGroupIds={selectedSecurityGroupIds}
                onSelectionChange={setSelectedSecurityGroupIds}
                onDelete={handleDeleteSecurityGroup}
                searchQuery={searchQuery}
                onSearchChange={setSearchQuery}
                onSearchClear={clearSearch}
                page={page}
                pageSize={pageSize}
                onPageChange={setPage}
                onPageSizeChange={handlePageSizeChange}
                isDeleting={deleteSecurityGroupMutation.isPending}
              />
            </>
          ) : emptyStateComponent
        }
        pageSize={pageSize}
        onPageSizeChange={handlePageSizeChange}
        searchResultsCount={filteredSecurityGroups.length}
        skeletonColumns={4}
        skeletonRows={5}
        skeletonShowCheckbox={true}
      showFilterButton={false}
      showSearchResultsInfo={false}
      />

      {/* Delete Security Group Confirmation Dialog */}
      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => setDeleteDialogState({ ...deleteDialogState, open })}
        onConfirm={handleConfirmDelete}
        title={t('network.deleteSecurityGroup')}
        description="이 보안 그룹을 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다."
        isLoading={deleteSecurityGroupMutation.isPending}
        resourceName={deleteDialogState.securityGroupName}
        resourceNameLabel="보안 그룹 이름"
      />
    </>
  );
}

export default function SecurityGroupsPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <SecurityGroupsPageContent />
    </Suspense>
  );
}

