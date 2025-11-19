/**
 * Subnets Page
 * Subnet 관리 페이지
 * 
 * 새로운 라우팅 구조: /{workspaceId}/{credentialId}/networks/subnets
 */

'use client';

import { Suspense } from 'react';
import { useState, useMemo, useCallback, useEffect, useRef } from 'react';
import * as React from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import dynamic from 'next/dynamic';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { usePagination } from '@/hooks/use-pagination';
import { UI } from '@/lib/constants';
import { Layers, Filter, Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { useTranslation } from '@/hooks/use-translation';
import { DataProcessor } from '@/lib/data';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import { log } from '@/lib/logging';
import { useResourceListState } from '@/hooks/use-resource-list-state';
import { useSSEStatus } from '@/hooks/use-sse-status';
import { useSSESubscription } from '@/hooks/use-sse-subscription';
import { useSSEErrorRecovery } from '@/hooks/use-sse-error-recovery';
import { useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '@/lib/query';
import { usePageRefresh } from '@/hooks/use-page-refresh';
import { useRequiredResourceContext } from '@/hooks/use-resource-context';
import { buildResourceCreatePath, updatePathFilters } from '@/lib/routing/helpers';
import {
  useSubnetActions,
  SubnetsPageHeader,
  SubnetsPageHeaderSection,
} from '@/features/networks';
import { useNetworkResources } from '@/features/networks/hooks/use-network-resources';
import { useCredentialRegionSync } from '@/hooks/use-credential-region-sync';
import { useCredentialContextStore } from '@/store/credential-context';
import { useWorkspaceStore } from '@/store/workspace';
import { useProviderRegionFilter, type ProviderRegionSelection } from '@/hooks/use-provider-region-filter';
import { useRegionFilter } from '@/hooks/use-region-filter';
import { useCredentials } from '@/hooks/use-credentials';
import { UnifiedFilterPanel } from '@/features/kubernetes';
import { useMultiProviderVPCs } from '@/hooks/use-multi-provider-vpcs';
import { useMultiProviderVPCsSingleRegion } from '@/hooks/use-multi-provider-vpcs-single-region';
import { useMultiProviderSubnets } from '@/hooks/use-multi-provider-subnets';
import { useMultiProviderSubnetsSingleRegion } from '@/hooks/use-multi-provider-subnets-single-region';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import type { CloudProvider, Subnet, VPC } from '@/lib/types';

const SubnetTable = dynamic(
  () => import('@/features/networks').then(mod => ({ default: mod.SubnetTable })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={6} rows={5} showCheckbox={true} />,
  }
);

function SubnetsPageContent() {
  useCredentialRegionSync();
  const router = useRouter();
  const searchParams = useSearchParams();
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { success } = useToast();
  const { handleError } = useErrorHandler();

  const { workspaceId, credentialId: pathCredentialId, region: pathRegion } = useRequiredResourceContext();
  const { currentWorkspace } = useWorkspaceStore();

  const {
    selectedCredentialIds,
    selectedRegion,
    setSelectedCredentials,
    setSelectedRegion,
  } = useCredentialContextStore();

  const { credentials, isLoading: isLoadingCredentials } = useCredentials({
    workspaceId: workspaceId || '',
    enabled: !!workspaceId,
  });

  useEffect(() => {
    if (!isLoadingCredentials && credentials.length > 0 && selectedCredentialIds.length === 0 && pathCredentialId) {
      setSelectedCredentials([pathCredentialId]);
    }
  }, [isLoadingCredentials, credentials, selectedCredentialIds.length, pathCredentialId, setSelectedCredentials]);

  useEffect(() => {
    if (pathRegion && !selectedRegion) {
      setSelectedRegion(pathRegion);
    }
  }, [pathRegion, selectedRegion, setSelectedRegion]);

  const selectedProviders = useMemo(() => {
    const providers = new Set<CloudProvider>();
    selectedCredentialIds.forEach((credentialId: string) => {
      const credential = credentials.find((c: { id: string; provider: string }) => c.id === credentialId);
      if (credential) {
        providers.add(credential.provider as CloudProvider);
      }
    });
    return Array.from(providers);
  }, [selectedCredentialIds, credentials]);

  const useProviderRegionFilterMode = selectedProviders.length > 1;
  
  const { providerSelectedRegions: providerSelectedRegionsFromStore, setProviderSelectedRegions } = useCredentialContextStore();
  
  const prevProvidersRef = React.useRef<string>('');
  React.useEffect(() => {
    const currentProvidersKey = selectedProviders.sort().join(',');
    if (prevProvidersRef.current && prevProvidersRef.current !== currentProvidersKey && useProviderRegionFilterMode) {
      setProviderSelectedRegions({
        aws: [],
        gcp: [],
        azure: [],
      });
    }
    prevProvidersRef.current = currentProvidersKey;
  }, [selectedProviders, useProviderRegionFilterMode, setProviderSelectedRegions]);
  
  const {
    selectedRegions: providerSelectedRegions,
    hasSelectedRegions,
  } = useProviderRegionFilter({
    providers: selectedProviders,
    initialSelectedRegions: providerSelectedRegionsFromStore,
    onRegionSelectionChange: (regions) => {
      setProviderSelectedRegions(regions);
      if (useProviderRegionFilterMode) {
        setSelectedRegion(null);
      }
    },
  });

  const {
    availableRegions,
    setSelectedRegion: handleRegionChange,
  } = useRegionFilter({
    providers: selectedProviders,
    selectedRegion: selectedRegion || undefined,
    onRegionChange: setSelectedRegion,
  });

  const selectedCredential = credentials.find((c: { id: string; provider: string }) => c.id === selectedCredentialIds[0]);
  const selectedProvider = selectedCredential?.provider as CloudProvider | undefined;
  const isMultiProviderMode = selectedCredentialIds.length > 1 || selectedProviders.length > 1;

  const {
    vpcs: mergedVPCsWithRegions,
    isLoading: isLoadingVPCsWithRegions,
  } = useMultiProviderVPCs({
    workspaceId: workspaceId || '',
    credentialIds: selectedCredentialIds,
    credentials: credentials.map((c: { id: string; provider: string }) => ({ id: c.id, provider: c.provider as CloudProvider })),
    selectedRegions: useProviderRegionFilterMode ? providerSelectedRegions : undefined,
    enabled: selectedCredentialIds.length > 0 && (useProviderRegionFilterMode ? hasSelectedRegions : false),
  });

  const {
    vpcs: mergedVPCsSingleRegion,
    isLoading: isLoadingVPCsSingleRegion,
  } = useMultiProviderVPCsSingleRegion({
    workspaceId: workspaceId || '',
    credentialIds: selectedCredentialIds,
    credentials: credentials.map((c: { id: string; provider: string }) => ({ id: c.id, provider: c.provider as CloudProvider })),
    region: selectedRegion || undefined,
    enabled: selectedCredentialIds.length > 0 && !useProviderRegionFilterMode,
  });

  const mergedVPCs = useProviderRegionFilterMode ? mergedVPCsWithRegions : mergedVPCsSingleRegion;
  const isLoadingVPCs = useProviderRegionFilterMode ? isLoadingVPCsWithRegions : isLoadingVPCsSingleRegion;

  const [selectedVPCId, setSelectedVPCId] = useState<string>('');

  useEffect(() => {
    const vpcIdFromUrl = searchParams?.get('vpc_id');
    if (vpcIdFromUrl && vpcIdFromUrl !== selectedVPCId) {
      const vpcExists = mergedVPCs.some((vpc: VPC & { provider?: CloudProvider }) => vpc.id === vpcIdFromUrl);
      if (vpcExists) {
        setSelectedVPCId(vpcIdFromUrl);
        const newParams = new URLSearchParams(searchParams?.toString());
        newParams.delete('vpc_id');
        const newUrl = newParams.toString() 
          ? `${window.location.pathname}?${newParams.toString()}`
          : window.location.pathname;
        router.replace(newUrl);
      }
    }
  }, [searchParams, selectedVPCId, mergedVPCs, router]);

  const {
    subnets: mergedSubnetsWithRegions,
    isLoading: isLoadingSubnetsWithRegions,
    errors: subnetErrorsWithRegions,
    hasError: hasSubnetErrorWithRegions,
  } = useMultiProviderSubnets({
    workspaceId: workspaceId || '',
    credentialIds: selectedCredentialIds,
    credentials: credentials.map((c: { id: string; provider: string }) => ({ id: c.id, provider: c.provider as CloudProvider })),
    vpcId: selectedVPCId,
    selectedRegions: useProviderRegionFilterMode ? providerSelectedRegions : undefined,
    enabled: selectedCredentialIds.length > 0 && !!selectedVPCId && (useProviderRegionFilterMode ? hasSelectedRegions : false),
  });

  const {
    subnets: mergedSubnetsSingleRegion,
    isLoading: isLoadingSubnetsSingleRegion,
    errors: subnetErrorsSingleRegion,
    hasError: hasSubnetErrorSingleRegion,
  } = useMultiProviderSubnetsSingleRegion({
    workspaceId: workspaceId || '',
    credentialIds: selectedCredentialIds,
    credentials: credentials.map((c: { id: string; provider: string }) => ({ id: c.id, provider: c.provider as CloudProvider })),
    vpcId: selectedVPCId,
    region: selectedRegion || undefined,
    enabled: selectedCredentialIds.length > 0 && !!selectedVPCId && !useProviderRegionFilterMode,
  });

  const mergedSubnets = useProviderRegionFilterMode ? mergedSubnetsWithRegions : mergedSubnetsSingleRegion;
  const isLoadingSubnets = useProviderRegionFilterMode ? isLoadingSubnetsWithRegions : isLoadingSubnetsSingleRegion;
  const subnetErrors = useProviderRegionFilterMode ? subnetErrorsWithRegions : subnetErrorsSingleRegion;
  const hasSubnetError = useProviderRegionFilterMode ? hasSubnetErrorWithRegions : hasSubnetErrorSingleRegion;

  usePageRefresh({
    queryKeys: selectedCredentialIds.map((credId: string) => {
      const cred = credentials.find((c: { id: string; provider: string }) => c.id === credId);
      return queryKeys.subnets.list(
        cred?.provider as CloudProvider,
        credId,
        selectedVPCId,
        selectedRegion || undefined
      );
    }),
    refetch: true,
    trigger: 'mount',
  });

  const { status: sseStatus } = useSSEStatus();

  const [isRefreshing, setIsRefreshing] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const lastRefreshTimeRef = useRef<number>(0);
  const REFRESH_DEBOUNCE_MS = 2000;

  const subscriptionFilters = React.useMemo(() => {
    if (useProviderRegionFilterMode && hasSelectedRegions) {
      const allSelectedRegions = Object.values(providerSelectedRegions).flat();
      return {
        credential_ids: selectedCredentialIds.length > 0 ? selectedCredentialIds : undefined,
        regions: allSelectedRegions.length > 0 ? allSelectedRegions : undefined,
      };
    } else {
      return {
        credential_ids: selectedCredentialIds.length > 0 ? selectedCredentialIds : undefined,
        regions: selectedRegion ? [selectedRegion] : undefined,
      };
    }
  }, [useProviderRegionFilterMode, hasSelectedRegions, providerSelectedRegions, selectedCredentialIds, selectedRegion]);

  useSSESubscription({
    eventTypes: [
      'network-subnet-created',
      'network-subnet-updated',
      'network-subnet-deleted',
      'network-subnet-list',
    ],
    filters: subscriptionFilters,
    enabled: sseStatus.isConnected && selectedCredentialIds.length > 0 && !!selectedVPCId,
  });

  // SSE 에러 복구
  useSSEErrorRecovery({
    autoReconnect: true,
    showNotifications: true,
  });

  const handleRefresh = useCallback(async () => {
    const now = Date.now();
    const timeSinceLastRefresh = now - lastRefreshTimeRef.current;

    if (timeSinceLastRefresh < REFRESH_DEBOUNCE_MS) {
      return;
    }

    lastRefreshTimeRef.current = now;
    setIsRefreshing(true);

    try {
      await Promise.all(
        selectedCredentialIds.map(async (credId: string) => {
          const cred = credentials.find((c: { id: string; provider: string }) => c.id === credId);
          if (!cred) return;
          
          await queryClient.invalidateQueries({
            queryKey: queryKeys.subnets.list(
              cred.provider as CloudProvider,
              credId,
              selectedVPCId,
              selectedRegion || undefined
            ),
          });
        })
      );

      await queryClient.refetchQueries({
        predicate: (query) => {
          const key = query.queryKey;
          if (!Array.isArray(key)) return false;
          return key.includes('subnets');
        },
      });

      setLastUpdated(new Date());
      success(t('network.subnetsRefreshed') || 'Subnet 목록을 새로고침했습니다');
    } catch (error) {
      handleError(error, { operation: 'refreshSubnets', resource: 'Subnet' });
    } finally {
      setIsRefreshing(false);
    }
  }, [
    queryClient,
    selectedCredentialIds,
    credentials,
    selectedVPCId,
    selectedRegion,
    success,
    t,
    handleError,
  ]);

  useEffect(() => {
    if (mergedSubnets && mergedSubnets.length >= 0 && !isLoadingSubnets) {
      if (!lastUpdated) {
        setLastUpdated(new Date());
      }
    }
  }, [mergedSubnets, isLoadingSubnets, lastUpdated]);

  // 공통 리스트 상태 관리
  const {
    filters,
    setFilters,
    showFilters,
    setShowFilters,
    selectedIds: selectedSubnetIds,
    setSelectedIds: setSelectedSubnetIds,
    pageSize,
    setPageSize,
    deleteDialogState,
    openDeleteDialog,
    closeDeleteDialog,
  } = useResourceListState({
    storageKey: 'subnets-page',
  });

  const {
    deleteSubnetMutation,
    handleBulkDeleteSubnets: handleBulkDelete,
    executeDeleteSubnet,
  } = useSubnetActions({
    selectedProvider,
    selectedCredentialId: selectedCredentialIds[0] || '',
  });

  const handleCreateSubnet = useCallback(() => {
    if (!workspaceId || selectedCredentialIds.length === 0) return;
    const firstCredentialId = selectedCredentialIds[0];
    const filters: Record<string, string | undefined> = {};
    if (selectedVPCId) filters.vpc_id = selectedVPCId;
    if (selectedRegion) filters.region = selectedRegion;
    
    const path = buildResourceCreatePath(
      workspaceId,
      firstCredentialId,
      'networks',
      'subnets',
      filters
    );
    router.push(path);
  }, [workspaceId, selectedCredentialIds, selectedVPCId, selectedRegion, router]);

  const handleDeleteSubnet = useCallback((subnetId: string, subnetRegion: string) => {
    const subnet = mergedSubnets.find((s: Subnet & { provider?: CloudProvider }) => s.id === subnetId);
    openDeleteDialog(subnetId, subnetRegion, subnet?.name || subnet?.id);
  }, [mergedSubnets, openDeleteDialog]);

  const handleConfirmDelete = useCallback(() => {
    if (deleteDialogState.id && deleteDialogState.region) {
      executeDeleteSubnet(deleteDialogState.id, deleteDialogState.region);
      closeDeleteDialog();
    }
  }, [deleteDialogState.id, deleteDialogState.region, executeDeleteSubnet, closeDeleteDialog]);

  const [searchQuery, setSearchQuery] = useState('');

  const filterFn = useCallback((subnet: Subnet, filters: FilterValue): boolean => {
    if (filters.state && subnet.state !== filters.state) return false;
    return true;
  }, []);

  const filteredSubnets = useMemo(() => {
    let result = DataProcessor.search(mergedSubnets, searchQuery, {
      keys: ['name', 'id', 'cidr_block', 'state', 'provider'],
      threshold: 0.3,
    });

    result = DataProcessor.filter(result, filters, filterFn);
    
    return result as Array<Subnet & { provider?: CloudProvider; credential_id?: string }>;
  }, [mergedSubnets, searchQuery, filters, filterFn]);

  const isSearching = searchQuery.length > 0;

  const clearSearch = useCallback(() => {
    setSearchQuery('');
  }, []);

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
    if (workspaceId && selectedCredentialIds.length > 0) {
      const newPath = updatePathFilters(window.location.pathname, { vpc_id: vpcId });
      router.replace(newPath, { scroll: false });
    }
  }, [workspaceId, selectedCredentialIds, router]);

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

  const isEmpty = selectedCredentialIds.length === 0 || !selectedVPCId || 
    (useProviderRegionFilterMode ? (!hasSelectedRegions || filteredSubnets.length === 0) : filteredSubnets.length === 0);

  const emptyStateComponent = credentials.length === 0 ? (
    <CredentialRequiredState serviceName={t('network.title')} />
  ) : selectedCredentialIds.length === 0 ? (
    <ResourceEmptyState
      resourceName={t('network.subnets')}
      title={t('credential.selectCredential')}
      description={t('credential.selectCredential')}
      withCard={true}
    />
  ) : !selectedVPCId ? (
    <Card>
      <CardContent className="flex flex-col items-center justify-center py-12">
        <Layers className="h-12 w-12 text-gray-400 mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">
          {t('network.selectVPC')}
        </h3>
        <p className="text-sm text-gray-500 text-center">
          {t('network.selectVPCToViewSubnets')}
        </p>
      </CardContent>
    </Card>
  ) : useProviderRegionFilterMode && !hasSelectedRegions ? (
    <ResourceEmptyState
      resourceName={t('network.subnets')}
      title={t('network.selectRegions') || 'Select Regions'}
      description={t('network.selectRegionsDescription') || 'Please select at least one region for each provider to view subnets.'}
      withCard={true}
    />
  ) : filteredSubnets.length === 0 ? (
    <ResourceEmptyState
      resourceName={t('network.subnets')}
      icon={Layers}
      onCreateClick={handleCreateSubnet}
      description={t('network.noSubnetsFoundForVPC')}
      isSearching={isSearching}
      searchQuery={searchQuery}
      hasFilters={Object.keys(filters).length > 0}
      onClearFilters={() => setFilters({})}
      onClearSearch={clearSearch}
      withCard={true}
    />
  ) : null;

  return (
    <>
    <ResourceListPage
        title={t('network.subnets')}
        resourceName={t('network.subnets')}
        storageKey="subnets-page"
        header={
          <SubnetsPageHeaderSection
            workspaceId={workspaceId || ''}
            workspaceName={currentWorkspace?.name}
            credentials={credentials}
            selectedCredentialIds={selectedCredentialIds}
            onCredentialSelectionChange={setSelectedCredentials}
            selectedProvider={selectedProvider}
            selectedProviders={selectedProviders}
            selectedRegion={selectedRegion}
            onRegionChange={handleRegionChange}
            selectedRegions={useProviderRegionFilterMode ? providerSelectedRegions : undefined}
            onRegionSelectionChange={(regions) => {
              setProviderSelectedRegions(regions);
            }}
            useProviderRegionFilter={useProviderRegionFilterMode}
            vpcs={mergedVPCs}
            selectedVPCId={selectedVPCId}
            onVPCChange={handleVPCChange}
            onRefresh={handleRefresh}
            isRefreshing={isRefreshing}
            lastUpdated={lastUpdated}
            isLoadingCredentials={isLoadingCredentials}
            isLoadingSubnets={isLoadingSubnets}
            subnetErrors={subnetErrors}
            onCreateClick={handleCreateSubnet}
          />
        }
        items={filteredSubnets}
        isLoading={isLoadingSubnets}
        isEmpty={isEmpty}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        onSearchClear={clearSearch}
        isSearching={isSearching}
        searchPlaceholder={t('network.searchSubnetsPlaceholder')}
        filterConfigs={selectedCredentialIds.length > 0 && selectedVPCId && mergedSubnets.length > 0 ? filterConfigs : []}
        filters={filters}
        onFiltersChange={setFilters}
        onFiltersClear={() => setFilters({})}
        showFilters={showFilters}
        onToggleFilters={() => setShowFilters(!showFilters)}
        filterCount={Object.keys(filters).length}
        toolbar={
          selectedCredentialIds.length > 0 && selectedVPCId && filteredSubnets.length > 0 ? (
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
          selectedCredentialIds.length > 0 && selectedVPCId ? (
            <>
              <Button
                onClick={handleCreateSubnet}
                className="flex items-center"
                disabled={!selectedVPCId}
                aria-label={t('network.createSubnet') || 'Create Subnet'}
              >
                <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
                {t('network.createSubnet') || 'Create Subnet'}
              </Button>
              {mergedSubnets.length > 0 && (
                <Button
                  variant="outline"
                  onClick={() => setShowFilters(!showFilters)}
                  className="flex items-center"
                  aria-label={t('common.filter') || 'Filters'}
                >
                  <Filter className="mr-2 h-4 w-4" aria-hidden="true" />
                  {t('common.filter') || 'Filters'}
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
          selectedCredentialIds.length > 0 && selectedVPCId && filteredSubnets.length > 0 ? (
            <>
              <SubnetTable
                subnets={mergedSubnets}
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
                showProviderColumn={isMultiProviderMode}
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

      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => {
          if (open) {
            // Dialog 열기 (이미 openDeleteDialog로 처리됨)
          } else {
            closeDeleteDialog();
          }
        }}
        onConfirm={handleConfirmDelete}
        title={t('network.deleteSubnet')}
        description={t('network.deleteSubnetDescription') || 'Are you sure you want to delete this subnet? This action cannot be undone.'}
        isLoading={deleteSubnetMutation.isPending}
        resourceName={deleteDialogState.name}
        resourceNameLabel={t('network.subnetName') || 'Subnet Name'}
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
