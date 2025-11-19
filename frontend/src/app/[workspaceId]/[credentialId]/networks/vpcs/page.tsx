/**
 * VPCs Page
 * Virtual Private Cloud 관리 페이지
 * 
 * 새로운 라우팅 구조: /{workspaceId}/{credentialId}/networks/vpcs
 */

'use client';

import { Suspense } from 'react';
import { useState, useMemo, useCallback, useEffect, useRef } from 'react';
import * as React from 'react';
import { useRouter } from 'next/navigation';
import dynamic from 'next/dynamic';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { usePagination } from '@/hooks/use-pagination';
import { EVENTS, UI } from '@/lib/constants';
import { Network, Filter, Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { useEmptyState } from '@/hooks/use-empty-state';
import { useTranslation } from '@/hooks/use-translation';
import { DataProcessor } from '@/lib/data';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import { useResourceListState } from '@/hooks/use-resource-list-state';
import { log } from '@/lib/logging';
import { useSSEStatus } from '@/hooks/use-sse-status';
import { useSSESubscription } from '@/hooks/use-sse-subscription';
import { useSSEErrorRecovery } from '@/hooks/use-sse-error-recovery';
import { useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '@/lib/query';
import { usePageRefresh } from '@/hooks/use-page-refresh';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { useRequiredResourceContext } from '@/hooks/use-resource-context';
import { buildResourceCreatePath } from '@/lib/routing/helpers';
import {
  useVPCActions,
  VPCsPageHeader,
  VPCsPageHeaderSection,
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
import type { CloudProvider } from '@/lib/types';
import type { CreateVPCForm, VPC } from '@/lib/types';

const VPCTable = dynamic(
  () => import('@/features/networks').then(mod => ({ default: mod.VPCTable })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={5} rows={5} showCheckbox={true} />,
  }
);

function VPCsPageContent() {
  useCredentialRegionSync();
  const router = useRouter();
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
    errors: vpcErrorsWithRegions,
    hasError: hasVPCErrorWithRegions,
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
    errors: vpcErrorsSingleRegion,
    hasError: hasVPCErrorSingleRegion,
  } = useMultiProviderVPCsSingleRegion({
    workspaceId: workspaceId || '',
    credentialIds: selectedCredentialIds,
    credentials: credentials.map((c: { id: string; provider: string }) => ({ id: c.id, provider: c.provider as CloudProvider })),
    region: selectedRegion || undefined,
    enabled: selectedCredentialIds.length > 0 && !useProviderRegionFilterMode,
  });

  const mergedVPCs = useProviderRegionFilterMode ? mergedVPCsWithRegions : mergedVPCsSingleRegion;
  const isLoadingVPCs = useProviderRegionFilterMode ? isLoadingVPCsWithRegions : isLoadingVPCsSingleRegion;
  const vpcErrors = useProviderRegionFilterMode ? vpcErrorsWithRegions : vpcErrorsSingleRegion;
  const hasVPCError = useProviderRegionFilterMode ? hasVPCErrorWithRegions : hasVPCErrorSingleRegion;

  // 페이지 새로고침 시 쿼리 무효화 및 재요청
  usePageRefresh({
    queryKeys: selectedCredentialIds.map((credId: string) => {
      const cred = credentials.find((c: { id: string; provider: string }) => c.id === credId);
      return queryKeys.vpcs.list(
        cred?.provider as CloudProvider,
        credId,
        selectedRegion || undefined
      );
    }),
    refetch: true,
    trigger: 'mount',
  });

  // SSE 상태 확인
  const { status: sseStatus } = useSSEStatus();

  // Refresh 상태 관리
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const lastRefreshTimeRef = useRef<number>(0);
  const REFRESH_DEBOUNCE_MS = 2000;

  // Refresh 핸들러 (debouncing 적용)
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
            queryKey: queryKeys.vpcs.list(
              cred.provider as CloudProvider,
              credId,
              selectedRegion || undefined
            ),
          });
        })
      );

      await queryClient.invalidateQueries({
        queryKey: queryKeys.vpcs.all,
      });

      await queryClient.refetchQueries({
        predicate: (query) => {
          const key = query.queryKey;
          if (!Array.isArray(key)) return false;
          return key.includes('vpcs');
        },
      });

      setLastUpdated(new Date());
      success(t('network.vpcsRefreshed') || 'VPC 목록을 새로고침했습니다');
    } catch (error) {
      handleError(error, { operation: 'refreshVPCs', resource: 'VPC' });
    } finally {
      setIsRefreshing(false);
    }
  }, [
    queryClient,
    selectedCredentialIds,
    credentials,
    selectedRegion,
    success,
    t,
    handleError,
  ]);

  // VPC 데이터가 업데이트될 때마다 lastUpdated 갱신
  useEffect(() => {
    if (mergedVPCs && mergedVPCs.length >= 0 && !isLoadingVPCs) {
      if (!lastUpdated) {
        setLastUpdated(new Date());
      }
    }
  }, [mergedVPCs, isLoadingVPCs, lastUpdated]);

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
      'network-vpc-created',
      'network-vpc-updated',
      'network-vpc-deleted',
      'network-vpc-list',
    ],
    filters: subscriptionFilters,
    enabled: sseStatus.isConnected && selectedCredentialIds.length > 0,
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
    selectedIds: selectedVPCIds,
    setSelectedIds: setSelectedVPCIds,
    pageSize,
    setPageSize,
    deleteDialogState,
    openDeleteDialog,
    closeDeleteDialog,
  } = useResourceListState({
    storageKey: 'vpcs-page',
  });

  const {
    deleteVPCMutation,
    handleBulkDeleteVPCs: handleBulkDelete,
    executeDeleteVPC,
  } = useVPCActions({
    selectedProvider,
    selectedCredentialId: selectedCredentialIds[0] || '',
    selectedRegion: selectedRegion || '',
  });

  const handleCreateVPC = useCallback(() => {
    if (!workspaceId || selectedCredentialIds.length === 0) return;
    const firstCredentialId = selectedCredentialIds[0];
    const path = buildResourceCreatePath(
      workspaceId,
      firstCredentialId,
      'networks',
      'vpcs',
      { region: selectedRegion || undefined }
    );
    router.push(path);
  }, [workspaceId, selectedCredentialIds, selectedRegion, router]);

  const handleDeleteVPC = useCallback((vpcId: string, vpcRegion?: string) => {
    if (!vpcRegion) return;
    const vpc = mergedVPCs.find(v => v.id === vpcId);
    openDeleteDialog(vpcId, vpcRegion, vpc?.name || vpc?.id);
  }, [mergedVPCs, openDeleteDialog]);

  const handleConfirmDelete = useCallback(() => {
    if (deleteDialogState.id && deleteDialogState.region) {
      executeDeleteVPC(deleteDialogState.id, deleteDialogState.region);
      closeDeleteDialog();
    }
  }, [deleteDialogState.id, deleteDialogState.region, executeDeleteVPC, closeDeleteDialog]);

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
    let result = DataProcessor.search(mergedVPCs, searchQuery, {
      keys: ['name', 'id', 'state', 'provider'],
      threshold: 0.3,
    });

    result = DataProcessor.filter(result, filters, filterFn);
    
    return result as Array<VPC & { provider?: CloudProvider; credential_id?: string }>;
  }, [mergedVPCs, searchQuery, filters, filterFn]);

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
  }, [setPageSize, setPaginationPageSize]);

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

  const isEmpty = selectedCredentialIds.length === 0 || 
    (useProviderRegionFilterMode ? (!hasSelectedRegions || filteredVPCs.length === 0) : filteredVPCs.length === 0);

  const emptyStateComponent = credentials.length === 0 ? (
    <CredentialRequiredState serviceName={t('network.title')} />
  ) : selectedCredentialIds.length === 0 ? (
    <ResourceEmptyState
      resourceName={t('network.vpcs')}
      title={t('credential.selectCredential')}
      description={t('credential.selectCredential')}
      withCard={true}
    />
  ) : useProviderRegionFilterMode && !hasSelectedRegions ? (
    <ResourceEmptyState
      resourceName={t('network.vpcs')}
      title={t('network.selectRegions') || 'Select Regions'}
      description={t('network.selectRegionsDescription') || 'Please select at least one region for each provider to view VPCs.'}
      withCard={true}
    />
  ) : filteredVPCs.length === 0 ? (
    <ResourceEmptyState
      resourceName={t('network.vpcs')}
      title={t('network.noVPCsFound') || 'No VPCs found'}
      description={t('network.createFirst') || 'Create your first VPC to get started'}
      onCreateClick={handleCreateVPC}
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
        title={t('network.vpcs')}
        resourceName={t('network.vpcs')}
        storageKey="vpcs-page"
        header={
          <VPCsPageHeaderSection
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
            onRefresh={handleRefresh}
            isRefreshing={isRefreshing}
            lastUpdated={lastUpdated}
            isLoadingCredentials={isLoadingCredentials}
            isLoadingVPCs={isLoadingVPCs}
            vpcErrors={vpcErrors}
            onCreateClick={handleCreateVPC}
          />
        }
        items={filteredVPCs}
        isLoading={isLoadingVPCs}
        isEmpty={isEmpty}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        onSearchClear={clearSearch}
        isSearching={isSearching}
        searchPlaceholder={t('network.searchVPCsPlaceholder')}
        filterConfigs={selectedCredentialIds.length > 0 && mergedVPCs.length > 0 ? filterConfigs : []}
        filters={filters}
        onFiltersChange={setFilters}
        onFiltersClear={() => setFilters({})}
        showFilters={showFilters}
        onToggleFilters={handleToggleFilters}
        filterCount={Object.keys(filters).length}
        toolbar={
          selectedCredentialIds.length > 0 && filteredVPCs.length > 0 ? (
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
          selectedCredentialIds.length > 0 ? (
            <>
              <Button
                onClick={handleCreateVPC}
                className="flex items-center"
                aria-label={t('network.createVPC') || 'Create VPC'}
              >
                <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
                {t('network.createVPC') || 'Create VPC'}
              </Button>
              {mergedVPCs.length > 0 && (
                <Button
                  variant="outline"
                  onClick={handleToggleFilters}
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
          selectedCredentialIds.length > 0 && filteredVPCs.length > 0 ? (
            <>
              <VPCTable
                vpcs={mergedVPCs}
                filteredVPCs={filteredVPCs}
                paginatedVPCs={paginatedVPCs}
                selectedVPCIds={selectedVPCIds}
                onSelectionChange={setSelectedVPCIds}
                onDelete={handleDeleteVPC}
                selectedRegion={selectedRegion || undefined}
                page={page}
                pageSize={pageSize}
                onPageChange={setPage}
                onPageSizeChange={handlePageSizeChange}
                isDeleting={deleteVPCMutation.isPending}
                showProviderColumn={isMultiProviderMode}
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
        onOpenChange={(open) => {
          if (open && deleteDialogState.id) {
            openDeleteDialog(deleteDialogState.id, deleteDialogState.region || undefined, deleteDialogState.name);
          } else {
            closeDeleteDialog();
          }
        }}
        onConfirm={handleConfirmDelete}
        title={t('network.deleteVPC')}
        description={t('network.deleteVPCDescription') || 'Are you sure you want to delete this VPC? This action cannot be undone.'}
        isLoading={deleteVPCMutation.isPending}
        resourceName={deleteDialogState.name}
        resourceNameLabel={t('network.vpcName') || 'VPC Name'}
      />
    </>
  );
}

const MemoizedVPCsPageContent = React.memo(VPCsPageContent);

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
      <MemoizedVPCsPageContent />
    </Suspense>
  );
}

