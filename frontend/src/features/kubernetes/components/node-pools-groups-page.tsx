/**
 * Node Pools/Groups Page Component
 * 
 * Node Pool/Group 통합 페이지 메인 컴포넌트
 * 클러스터 페이지와 일관성 있는 구조
 */

'use client';

import * as React from 'react';
import { Suspense, useState, useCallback, useRef } from 'react';
import { useMemo, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { usePagination } from '@/hooks/use-pagination';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { Filter } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { useTranslation } from '@/hooks/use-translation';
import { useResourceListState } from '@/hooks/use-resource-list-state';
import { useSSEStatus } from '@/hooks/use-sse-status';
import { useSSESubscription } from '@/hooks/use-sse-subscription';
import { useSSEErrorRecovery } from '@/hooks/use-sse-error-recovery';
import { useQueryClient, useQueries } from '@tanstack/react-query';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import { usePageRefresh } from '@/hooks/use-page-refresh';
import { useRequiredResourceContext } from '@/hooks/use-resource-context';
import { buildCredentialResourceDetailPath, buildCredentialResourceCreatePath } from '@/lib/routing/helpers';
import { NodePoolsGroupsPageHeaderSection } from './node-pools-groups-page-header-section';
import { NodePoolsGroupsPageContent } from './node-pools-groups-page-content';
import { useNodePoolsGroupsFilters } from '../hooks/use-node-pools-groups-filters';
import { useNodePoolsGroups } from '../hooks/use-node-pools-groups';
import { useCredentialContextStore } from '@/store/credential-context';
import { useWorkspaceStore } from '@/store/workspace';
import { useMultiProviderClustersWithRegions } from '@/hooks/use-multi-provider-clusters-with-regions';
import { useProviderRegionFilter, type ProviderRegionSelection } from '@/hooks/use-provider-region-filter';
import { useRegionFilter } from '@/hooks/use-region-filter';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import { useCredentials } from '@/hooks/use-credentials';
import { useCredentialRegionSync } from '@/hooks/use-credential-region-sync';
import { kubernetesService } from '../services/kubernetes';
import type { NodePool, NodeGroup, CloudProvider, KubernetesCluster } from '@/lib/types/kubernetes';

type NodePoolOrGroup = (NodePool | NodeGroup) & {
  cluster_name: string;
  cluster_id: string;
  provider: CloudProvider;
  resource_type: 'node-pool' | 'node-group';
  credential_id?: string;
};

/**
 * Provider별 리소스 이름 결정
 */
function getResourceName(providers: CloudProvider[], t: (key: string) => string): string {
  if (providers.length === 0) {
    return t('kubernetes.nodePools') || 'Node Pools';
  }
  
  if (providers.length === 1) {
    return providers[0] === 'aws' 
      ? (t('kubernetes.nodeGroups') || 'Node Groups')
      : (t('kubernetes.nodePools') || 'Node Pools');
  }
  
  // Multi-provider: 둘 다 포함하는 경우
  const hasAWS = providers.includes('aws');
  const hasOthers = providers.some(p => p !== 'aws');
  
  if (hasAWS && hasOthers) {
    return t('kubernetes.nodePoolsAndGroups') || 'Node Pools & Groups';
  }
  return hasAWS 
    ? (t('kubernetes.nodeGroups') || 'Node Groups')
    : (t('kubernetes.nodePools') || 'Node Pools');
}

function NodePoolsGroupsPageMain() {
  // URL 동기화 Hook 사용 (최우선)
  useCredentialRegionSync();
  const { success } = useToast();
  const { handleError } = useErrorHandler();
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const router = useRouter();

  // Path Parameter에서 컨텍스트 추출
  const { workspaceId, credentialId: pathCredentialId, region: pathRegion } = useRequiredResourceContext();

  // Workspace Store에서 workspace 이름 가져오기
  const { currentWorkspace } = useWorkspaceStore();

  // Credential Context Store에서 multi-select 지원
  const {
    selectedCredentialIds,
    selectedRegion,
    setSelectedCredentials,
    setSelectedRegion,
  } = useCredentialContextStore();

  // Credentials 조회
  const { credentials, isLoading: isLoadingCredentials } = useCredentials({
    workspaceId: workspaceId || '',
    enabled: !!workspaceId,
  });

  // Path parameter의 credentialId 사용 (자동 선택 제거)
  useEffect(() => {
    if (!isLoadingCredentials && credentials.length > 0 && selectedCredentialIds.length === 0 && pathCredentialId) {
      setSelectedCredentials([pathCredentialId]);
    }
  }, [isLoadingCredentials, credentials, selectedCredentialIds.length, pathCredentialId, setSelectedCredentials]);

  // Region 자동 설정
  useEffect(() => {
    if (pathRegion && !selectedRegion) {
      setSelectedRegion(pathRegion);
    }
  }, [pathRegion, selectedRegion, setSelectedRegion]);

  // 선택된 providers 추출
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

  // Provider별 Region 선택 상태
  const useProviderRegionFilterMode = selectedProviders.length > 1;
  const { providerSelectedRegions: providerSelectedRegionsFromStore, setProviderSelectedRegions } = useCredentialContextStore();
  
  // Provider가 변경되면 Region 선택 상태 초기화
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

  // Multi-provider 클러스터 조회
  const {
    clusters: mergedClusters,
    isLoading: isLoadingClusters,
    errors: clusterErrors,
    hasError: hasClusterError,
  } = useMultiProviderClustersWithRegions({
    workspaceId: workspaceId || '',
    credentialIds: selectedCredentialIds,
    credentials: credentials.map((c: { id: string; provider: string }) => ({ id: c.id, provider: c.provider as CloudProvider })),
    selectedRegions: useProviderRegionFilterMode ? providerSelectedRegions : undefined,
    enabled: selectedCredentialIds.length > 0 && (useProviderRegionFilterMode ? hasSelectedRegions : true),
  });

  // 각 클러스터의 Node Pool/Group 조회
  const nodePoolGroupQueriesResults = useQueries({
    queries: mergedClusters.map(cluster => {
      const isAWS = cluster.provider === 'aws';
      return {
        queryKey: isAWS
          ? queryKeys.kubernetesClusters.nodeGroups(cluster.name, cluster.provider, cluster.credential_id, cluster.region)
          : queryKeys.kubernetesClusters.nodePools(cluster.name, cluster.provider, cluster.credential_id, cluster.region),
        queryFn: async () => {
          if (!cluster.provider || !cluster.credential_id) return [];
          if (isAWS) {
            return kubernetesService.listNodeGroups(
              cluster.provider,
              cluster.name,
              cluster.credential_id,
              cluster.region
            );
          }
          return kubernetesService.listNodePools(
            cluster.provider,
            cluster.name,
            cluster.credential_id,
            cluster.region
          );
        },
        enabled: !!cluster.provider && !!cluster.credential_id && !!cluster.name && mergedClusters.length > 0,
        staleTime: CACHE_TIMES.MONITORING,
        gcTime: GC_TIMES.MEDIUM,
      };
    }),
  });

  // 모든 Node Pool/Group 수집
  const allNodePoolsGroups = React.useMemo(() => {
    const items: NodePoolOrGroup[] = [];
    
    nodePoolGroupQueriesResults.forEach((result, index) => {
      const cluster = mergedClusters[index];
      if (!cluster || !result.data) return;
      
      const isAWS = cluster.provider === 'aws';
      result.data.forEach((item: NodePool | NodeGroup) => {
        items.push({
          ...item,
          cluster_name: cluster.name,
          cluster_id: cluster.id || cluster.name,
          provider: cluster.provider,
          resource_type: isAWS ? 'node-group' : 'node-pool',
          credential_id: cluster.credential_id,
        } as NodePoolOrGroup);
      });
    });
    
    return items;
  }, [nodePoolGroupQueriesResults, mergedClusters]);

  const isLoadingNodePoolsGroups = nodePoolGroupQueriesResults.some(r => r.isLoading);
  const isLoadingCombined = isLoadingCredentials || isLoadingClusters || isLoadingNodePoolsGroups;

  // Region 필터링 훅 (단일 Region 모드용)
  const {
    availableRegions,
    setSelectedRegion: handleRegionChange,
  } = useRegionFilter({
    providers: selectedProviders,
    selectedRegion: selectedRegion || undefined,
    onRegionChange: setSelectedRegion,
  });

  // Get selected credential and provider
  const selectedCredential = credentials.find((c: { id: string; provider: string }) => c.id === selectedCredentialIds[0]);
  const selectedProvider = selectedCredential?.provider as CloudProvider | undefined;

  // Multi-provider 모드 여부
  const isMultiProviderMode = selectedCredentialIds.length > 1 || selectedProviders.length > 1;

  // 페이지 새로고침 시 쿼리 무효화 및 재요청
  usePageRefresh({
    queryKeys: selectedCredentialIds.map((credId: string) => {
      const cred = credentials.find((c: { id: string; provider: string }) => c.id === credId);
      return queryKeys.kubernetesClusters.list(
        workspaceId,
        cred?.provider,
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

  // SSE 이벤트 구독
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
  }, [selectedCredentialIds, selectedRegion, useProviderRegionFilterMode, hasSelectedRegions, providerSelectedRegions]);

  useSSESubscription({
    eventTypes: [
      'kubernetes-node-pool-created',
      'kubernetes-node-pool-updated',
      'kubernetes-node-pool-deleted',
      'kubernetes-node-group-created',
      'kubernetes-node-group-updated',
      'kubernetes-node-group-deleted',
    ],
    filters: subscriptionFilters,
    enabled: sseStatus.isConnected && (selectedCredentialIds.length > 0 || !!selectedRegion),
  });

  // SSE 에러 복구
  useSSEErrorRecovery({
    autoReconnect: true,
    showNotifications: true,
  });

  // Local state
  const {
    filters,
    setFilters,
    showFilters,
    setShowFilters,
    selectedIds: selectedItemIds,
    setSelectedIds: setSelectedItemIds,
    pageSize,
    setPageSize,
    deleteDialogState,
    openDeleteDialog,
    closeDeleteDialog,
  } = useResourceListState({
    storageKey: 'kubernetes-node-pools-groups-page',
  });

  // Refresh 핸들러
  const handleRefresh = useCallback(async () => {
    const now = Date.now();
    const timeSinceLastRefresh = now - lastRefreshTimeRef.current;

    if (timeSinceLastRefresh < REFRESH_DEBOUNCE_MS) {
      return;
    }

    lastRefreshTimeRef.current = now;
    setIsRefreshing(true);

    try {
      selectedCredentialIds.forEach(credId => {
        const cred = credentials.find(c => c.id === credId);
        queryClient.invalidateQueries({
          queryKey: queryKeys.kubernetesClusters.list(
            workspaceId,
            cred?.provider,
            credId,
            selectedRegion || undefined
          ),
        });
      });
      
      await queryClient.invalidateQueries({
        queryKey: queryKeys.kubernetesClusters.all,
      });

      setLastUpdated(new Date());
      success(t('kubernetes.nodePoolsGroupsRefreshed') || 'Node Pools/Groups list refreshed');
    } catch (error) {
      handleError(error, { operation: 'refreshNodePoolsGroups', resource: 'NodePoolGroup' });
    } finally {
      setIsRefreshing(false);
    }
  }, [
    queryClient,
    workspaceId,
    selectedCredentialIds,
    credentials,
    selectedRegion,
    success,
    t,
    handleError,
  ]);

  // Combined loading state
  useEffect(() => {
    if (allNodePoolsGroups && allNodePoolsGroups.length >= 0 && !isLoadingCombined) {
      if (!lastUpdated) {
        setLastUpdated(new Date());
      }
    }
  }, [allNodePoolsGroups, isLoadingCombined, lastUpdated]);

  // Filters hook
  const {
    searchQuery,
    setSearchQuery,
    isSearching,
    clearSearch,
    filteredItems,
  } = useNodePoolsGroupsFilters({
    items: allNodePoolsGroups,
    filters,
  });

  // Mutations hook
  const {
    deleteNodePoolGroupMutation,
  } = useNodePoolsGroups({
    workspaceId: workspaceId || '',
    selectedCredentialId: selectedCredentialIds[0] || '',
    selectedRegion: selectedRegion || undefined,
    provider: selectedProvider,
  });

  // Pagination
  const {
    page,
    paginatedItems: paginatedItems,
    setPage,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredItems, {
    totalItems: filteredItems.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  // Filter configurations
  const filterConfigs: FilterConfig[] = useMemo(() => [
    {
      id: 'status',
      label: t('filters.status'),
      type: 'select',
      options: [
        { id: 'running', value: 'RUNNING', label: t('filters.running') },
        { id: 'active', value: 'ACTIVE', label: t('filters.active') },
        { id: 'creating', value: 'CREATING', label: t('kubernetes.creating') },
        { id: 'updating', value: 'UPDATING', label: t('kubernetes.updating') },
        { id: 'deleting', value: 'DELETING', label: t('kubernetes.deleting') },
        { id: 'failed', value: 'FAILED', label: t('kubernetes.failed') },
      ],
    },
    {
      id: 'region',
      label: t('filters.region'),
      type: 'select',
      options: Array.from(new Set(allNodePoolsGroups.map(item => item.region)))
        .filter(Boolean)
        .map((r: string, idx: number) => ({ 
          id: `region-${idx}`, 
          value: r, 
          label: r 
        })),
    },
    {
      id: 'provider',
      label: t('filters.provider'),
      type: 'select',
      options: Array.from(new Set(allNodePoolsGroups.map(item => item.provider)))
        .filter(Boolean)
        .map((p: CloudProvider, idx: number) => ({ 
          id: `provider-${idx}`, 
          value: p, 
          label: p.toUpperCase() 
        })),
    },
    {
      id: 'cluster',
      label: t('kubernetes.cluster'),
      type: 'select',
      options: Array.from(new Set(allNodePoolsGroups.map(item => item.cluster_name)))
        .filter(Boolean)
        .map((c: string, idx: number) => ({ 
          id: `cluster-${idx}`, 
          value: c, 
          label: c 
        })),
    },
  ], [allNodePoolsGroups, t]);

  // Event handlers
  const handleDelete = useCallback((name: string, clusterName: string, region: string) => {
    const item = allNodePoolsGroups.find(item => item.name === name && item.cluster_name === clusterName);
    if (!item || !item.provider) return;
    openDeleteDialog(name, region, name);
  }, [allNodePoolsGroups, openDeleteDialog]);

  const handleBulkDelete = useCallback((ids: string[]) => {
    if (ids.length === 0) return;
    
    const itemsToDelete = allNodePoolsGroups.filter(item => {
      const itemId = item.id || item.name;
      return ids.includes(itemId);
    });

    if (itemsToDelete.length === 0) return;

    const deletePromises = itemsToDelete.map(item => {
      if (!item.provider || !item.cluster_name || !item.credential_id) {
        return Promise.reject(new Error(`Missing required fields for ${item.name}`));
      }

      const isNodeGroup = item.resource_type === 'node-group';
      return deleteNodePoolGroupMutation.mutateAsync({
        provider: item.provider,
        clusterName: item.cluster_name,
        name: item.name,
        isNodeGroup,
        credentialId: item.credential_id,
        region: item.region,
      });
    });

    Promise.allSettled(deletePromises).then(results => {
      const succeeded = results.filter(r => r.status === 'fulfilled').length;
      const failed = results.filter(r => r.status === 'rejected').length;

      if (succeeded > 0) {
        success(
          t('kubernetes.bulkDeleteSuccess') || 
          `Successfully deleted ${succeeded} node pool${succeeded > 1 ? 's' : ''}/group${succeeded > 1 ? 's' : ''}`
        );
      }

      if (failed > 0) {
        handleError(
          new Error(`Failed to delete ${failed} item${failed > 1 ? 's' : ''}`),
          { operation: 'bulkDeleteNodePoolsGroups', resource: 'NodePoolGroup' }
        );
      }

      if (succeeded > 0) {
        setSelectedItemIds([]);
        queryClient.invalidateQueries({
          queryKey: queryKeys.kubernetesClusters.all,
        });
      }
    });
  }, [allNodePoolsGroups, deleteNodePoolGroupMutation, success, t, handleError, setSelectedItemIds, queryClient]);

  const handleConfirmDelete = useCallback(() => {
    if (!deleteDialogState.id || !deleteDialogState.region) return;
    
    const item = allNodePoolsGroups.find(item => item.name === deleteDialogState.id);
    if (!item || !item.provider || !item.cluster_name) return;

    const isNodeGroup = item.resource_type === 'node-group';
    deleteNodePoolGroupMutation.mutate(
      {
        provider: item.provider,
        clusterName: item.cluster_name,
        name: item.name,
        isNodeGroup,
        credentialId: item.credential_id || selectedCredentialIds[0] || '',
        region: deleteDialogState.region,
      },
      {
        onSuccess: () => {
          success(isNodeGroup 
            ? (t('kubernetes.nodeGroupDeletionInitiated') || 'Node Group deletion initiated')
            : (t('kubernetes.nodePoolDeletionInitiated') || 'Node Pool deletion initiated')
          );
          closeDeleteDialog();
        },
        onError: (error: unknown) => {
          handleError(error, { operation: 'deleteNodePoolGroup', resource: 'NodePoolGroup' });
        },
      }
    );
  }, [
    deleteDialogState.id,
    deleteDialogState.region,
    allNodePoolsGroups,
    selectedCredentialIds,
    deleteNodePoolGroupMutation,
    success,
    t,
    closeDeleteDialog,
    handleError,
  ]);

  const handlePageSizeChange = useCallback((newSize: number) => {
    setPageSize(newSize);
    setPaginationPageSize(newSize);
  }, [setPageSize, setPaginationPageSize]);

  const handleCreate = useCallback(() => {
    if (!workspaceId || selectedCredentialIds.length === 0) return;
    const firstCredentialId = selectedCredentialIds[0];
    const path = buildCredentialResourceCreatePath(
      workspaceId,
      firstCredentialId,
      'k8s',
      'node-pools',
      { region: selectedRegion || undefined }
    );
    router.push(path);
  }, [workspaceId, selectedCredentialIds, selectedRegion, router]);

  // Resource name 결정
  const resourceName = useMemo(() => getResourceName(selectedProviders, t), [selectedProviders, t]);

  // Determine empty state
  const isEmpty = selectedCredentialIds.length === 0 || 
    (useProviderRegionFilterMode ? (!hasSelectedRegions || filteredItems.length === 0) : filteredItems.length === 0);

  // Empty state component
  const emptyStateComponent = credentials.length === 0 ? (
    <CredentialRequiredState serviceName={t('kubernetes.title')} />
  ) : selectedCredentialIds.length === 0 ? (
    <ResourceEmptyState
      resourceName={resourceName}
      title={t('credential.selectCredential')}
      description={t('credential.selectCredential')}
      withCard={true}
    />
  ) : useProviderRegionFilterMode && !hasSelectedRegions ? (
    <ResourceEmptyState
      resourceName={resourceName}
      title={t('kubernetes.selectRegions') || 'Select Regions'}
      description={t('kubernetes.selectRegionsDescription') || 'Please select at least one region for each provider to view node pools/groups.'}
      withCard={true}
    />
  ) : filteredItems.length === 0 ? (
    <ResourceEmptyState
      resourceName={resourceName}
      title={t('kubernetes.noNodePoolsGroupsFound') || 'No Node Pools/Groups Found'}
      description={t('kubernetes.createFirst')}
      onCreateClick={handleCreate}
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
        title={resourceName}
        resourceName={resourceName}
        storageKey="kubernetes-node-pools-groups-page"
        header={
          <NodePoolsGroupsPageHeaderSection
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
            isLoadingClusters={isLoadingClusters}
            errors={clusterErrors}
            onCreateClick={handleCreate}
          />
        }
        items={filteredItems}
        isLoading={isLoadingCombined}
        isEmpty={isEmpty}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        onSearchClear={clearSearch}
        isSearching={isSearching}
        searchPlaceholder={t('kubernetes.searchNodePoolsGroups') || 'Search node pools/groups by name, cluster, instance type, status, or region...'}
        filterConfigs={selectedCredentialIds.length > 0 && allNodePoolsGroups.length > 0 ? filterConfigs : []}
        filters={filters}
        onFiltersChange={setFilters}
        onFiltersClear={() => setFilters({})}
        showFilters={showFilters}
        onToggleFilters={() => setShowFilters(!showFilters)}
        filterCount={Object.keys(filters).length}
        toolbar={
          selectedCredentialIds.length > 0 && filteredItems.length > 0 && (
            <BulkActionsToolbar
              items={filteredItems}
              selectedIds={selectedItemIds}
              onSelectionChange={setSelectedItemIds}
              onBulkDelete={handleBulkDelete}
              getItemDisplayName={(item) => item.name}
            />
          )
        }
        additionalControls={
          selectedCredentialIds.length > 0 && allNodePoolsGroups.length > 0 ? (
            <>
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
            </>
          ) : null
        }
        emptyState={emptyStateComponent}
        content={
          selectedCredentialIds.length > 0 && filteredItems.length > 0 ? (
            <NodePoolsGroupsPageContent
              items={allNodePoolsGroups}
              filteredItems={filteredItems}
              paginatedItems={paginatedItems}
              selectedProvider={selectedProvider}
              selectedItemIds={selectedItemIds}
              onSelectionChange={setSelectedItemIds}
              onDelete={handleDelete}
              isDeleting={deleteNodePoolGroupMutation.isPending}
              page={page}
              pageSize={pageSize}
              total={filteredItems.length}
              onPageChange={setPage}
              onPageSizeChange={handlePageSizeChange}
              isSearching={isSearching}
              searchQuery={searchQuery}
              isMultiProviderMode={isMultiProviderMode}
              selectedProviders={selectedProviders}
              resourceName={resourceName}
              workspaceId={workspaceId || ''}
              credentialId={selectedCredentialIds[0] || ''}
            />
          ) : emptyStateComponent
        }
        pageSize={pageSize}
        onPageSizeChange={handlePageSizeChange}
        searchResultsCount={filteredItems.length}
        skeletonColumns={10}
        skeletonRows={5}
        skeletonShowCheckbox={true}
        showFilterButton={false}
        showSearchResultsInfo={false}
      />

      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => {
          if (!open) {
            closeDeleteDialog();
          }
        }}
        onConfirm={handleConfirmDelete}
        title={t('kubernetes.deleteNodePoolGroup') || 'Delete Node Pool/Group'}
        description={t('kubernetes.deleteNodePoolGroupDescription') || 'Are you sure you want to delete this node pool/group? This action cannot be undone.'}
        isLoading={deleteNodePoolGroupMutation.isPending}
        resourceName={deleteDialogState.name}
        resourceNameLabel={t('common.name') || 'Name'}
      />
    </>
  );
}

export function NodePoolsGroupsPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <NodePoolsGroupsPageMain />
    </Suspense>
  );
}

export default NodePoolsGroupsPage;

