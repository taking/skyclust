/**
 * Kubernetes Nodes Page
 * Kubernetes 노드 관리 페이지
 * 
 * 새로운 라우팅 구조: /w/{workspaceId}/c/{credentialId}/k8s/nodes
 */

'use client';

import { Suspense, useState, useCallback, useRef } from 'react';
import { useMemo, useEffect } from 'react';
import * as React from 'react';
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
import {
  NodesPageHeaderSection,
  NodesPageContent,
  useNodesFilters,
} from '@/features/kubernetes';
import { useCredentialContextStore } from '@/store/credential-context';
import { useWorkspaceStore } from '@/store/workspace';
import { useMultiProviderClustersWithRegions } from '@/hooks/use-multi-provider-clusters-with-regions';
import { useMultiProviderClusters } from '@/hooks/use-multi-provider-clusters';
import { useProviderRegionFilter, type ProviderRegionSelection } from '@/hooks/use-provider-region-filter';
import { useRegionFilter } from '@/hooks/use-region-filter';
import { useCredentials } from '@/hooks/use-credentials';
import { useCredentialRegionSync } from '@/hooks/use-credential-region-sync';
import { kubernetesService } from '@/features/kubernetes';
import type { Node, CloudProvider, KubernetesCluster } from '@/lib/types';

type NodeWithMetadata = Node & {
  cluster_name: string;
  cluster_id: string;
  provider?: CloudProvider;
  credential_id?: string;
};

function KubernetesNodesPageContent() {
  useCredentialRegionSync();
  const { success } = useToast();
  const { handleError } = useErrorHandler();
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const router = useRouter();

  const { workspaceId, credentialId: pathCredentialId, region: pathRegion } = useRequiredResourceContext();
  const { currentWorkspace } = useWorkspaceStore();

  const {
    selectedCredentialIds,
    selectedRegion,
    setSelectedCredentials,
    setSelectedRegion,
    toggleCredential,
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
    toggleRegion,
    toggleAllRegionsForProvider,
    clearAllRegions,
    hasSelectedRegions,
    selectedCountsByProvider,
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
    clusters: mergedClustersWithRegions,
    isLoading: isLoadingClustersWithRegions,
    errors: clusterErrorsWithRegions,
    hasError: hasClusterErrorWithRegions,
  } = useMultiProviderClustersWithRegions({
    workspaceId: workspaceId || '',
    credentialIds: selectedCredentialIds,
    credentials: credentials.map((c: { id: string; provider: string }) => ({ id: c.id, provider: c.provider as CloudProvider })),
    selectedRegions: useProviderRegionFilterMode ? providerSelectedRegions : undefined,
    enabled: selectedCredentialIds.length > 0 && (useProviderRegionFilterMode ? hasSelectedRegions : true),
  });

  const {
    clusters: mergedClustersSingleRegion,
    isLoading: isLoadingClustersSingleRegion,
    errors: clusterErrorsSingleRegion,
    hasError: hasClusterErrorSingleRegion,
  } = useMultiProviderClusters({
    workspaceId: workspaceId || '',
    credentialIds: selectedCredentialIds,
    credentials: credentials.map((c: { id: string; provider: string }) => ({ id: c.id, provider: c.provider as CloudProvider })),
    region: selectedRegion || undefined,
    enabled: selectedCredentialIds.length > 0 && !useProviderRegionFilterMode,
  });

  const mergedClusters = useProviderRegionFilterMode ? mergedClustersWithRegions : mergedClustersSingleRegion;
  const isLoadingClusters = useProviderRegionFilterMode ? isLoadingClustersWithRegions : isLoadingClustersSingleRegion;
  const clusterErrors = useProviderRegionFilterMode ? clusterErrorsWithRegions : clusterErrorsSingleRegion;
  const hasClusterError = useProviderRegionFilterMode ? hasClusterErrorWithRegions : hasClusterErrorSingleRegion;

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
      'kubernetes-node-created',
      'kubernetes-node-updated',
      'kubernetes-node-deleted',
      'kubernetes-node-list',
    ],
    filters: subscriptionFilters,
    enabled: sseStatus.isConnected && selectedCredentialIds.length > 0,
  });

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
            queryKey: queryKeys.kubernetesClusters.list(
              workspaceId,
              cred.provider,
              credId,
              selectedRegion || undefined
            ),
          });
        })
      );

      await queryClient.refetchQueries({
        predicate: (query) => {
          const key = query.queryKey;
          if (!Array.isArray(key)) return false;
          return key.includes('kubernetes') && (key.includes('clusters') || key.includes('nodes'));
        },
      });

      setLastUpdated(new Date());
      success(t('kubernetes.nodesRefreshed') || 'Node 목록을 새로고침했습니다');
    } catch (error) {
      handleError(error, { operation: 'refreshNodes', resource: 'Node' });
    } finally {
      setIsRefreshing(false);
    }
  }, [
    queryClient,
    selectedCredentialIds,
    credentials,
    workspaceId,
    selectedRegion,
    success,
    t,
    handleError,
  ]);

  useEffect(() => {
    if (mergedClusters && mergedClusters.length >= 0 && !isLoadingClusters) {
      if (!lastUpdated) {
        setLastUpdated(new Date());
      }
    }
  }, [mergedClusters, isLoadingClusters, lastUpdated]);

  const {
    filters,
    setFilters,
    showFilters,
    setShowFilters,
    selectedIds: selectedNodeIds,
    setSelectedIds: setSelectedNodeIds,
    pageSize,
    setPageSize,
  } = useResourceListState({
    storageKey: 'kubernetes-nodes-page',
  });

  const nodeQueriesResults = useQueries({
    queries: mergedClusters.map((cluster: KubernetesCluster & { provider?: CloudProvider; credential_id?: string }) => {
      const clusterProvider = cluster.provider || selectedProvider;
      const clusterCredentialId = cluster.credential_id || selectedCredentialIds[0];
      const clusterRegion = cluster.region || selectedRegion || undefined;
      
      return {
        queryKey: queryKeys.kubernetesClusters.nodes(
          cluster.name,
          clusterProvider,
          clusterCredentialId,
          clusterRegion
        ),
        queryFn: async () => {
          if (!clusterProvider || !clusterCredentialId) return [];
          return kubernetesService.listNodes(clusterProvider, cluster.name, clusterCredentialId, clusterRegion || cluster.region);
        },
        enabled: !!clusterProvider && !!clusterCredentialId && !!cluster.name && mergedClusters.length > 0,
        staleTime: CACHE_TIMES.MONITORING,
        gcTime: GC_TIMES.MEDIUM,
      };
    }),
  });

  const allNodes = React.useMemo(() => {
    const nodes: NodeWithMetadata[] = [];
    
    nodeQueriesResults.forEach((result, index) => {
      const cluster = mergedClusters[index] as KubernetesCluster & { provider?: CloudProvider; credential_id?: string };
      if (!cluster || !result.data) return;
      
      if (Array.isArray(result.data)) {
        result.data.forEach((node: Node) => {
          nodes.push({
            ...node,
            cluster_name: cluster.name,
            cluster_id: cluster.id || cluster.name,
            provider: cluster.provider,
            credential_id: cluster.credential_id,
          });
        });
      }
    });
    
    return nodes;
  }, [nodeQueriesResults, mergedClusters]);

  const nodeErrors = React.useMemo(() => {
    const errors: Array<{ provider: CloudProvider; credentialId: string; region?: string; error: Error }> = [];
    
    nodeQueriesResults.forEach((result, index) => {
      if (result.error && index < mergedClusters.length) {
        const cluster = mergedClusters[index] as KubernetesCluster & { provider?: CloudProvider; credential_id?: string };
        if (cluster) {
          errors.push({
            provider: (cluster.provider || selectedProvider) as CloudProvider,
            credentialId: cluster.credential_id || selectedCredentialIds[0] || '',
            region: cluster.region,
            error: result.error as Error,
          });
        }
      }
    });
    
    return errors;
  }, [nodeQueriesResults, mergedClusters, selectedProvider, selectedCredentialIds]);

  const {
    searchQuery,
    setSearchQuery,
    isSearching,
    clearSearch,
    filteredNodes,
  } = useNodesFilters({
    nodes: allNodes,
    filters,
  });

  const {
    page,
    paginatedItems: paginatedNodes,
    setPage,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredNodes, {
    totalItems: filteredNodes.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  const handlePageSizeChange = useCallback((newSize: number) => {
    setPageSize(newSize);
    setPaginationPageSize(newSize);
  }, [setPageSize, setPaginationPageSize]);

  const filterConfigs: FilterConfig[] = useMemo(() => [
    {
      id: 'status',
      label: t('filters.status'),
      type: 'select',
      options: [
        { id: 'ready', value: 'Ready', label: t('filters.ready') || 'Ready' },
        { id: 'not-ready', value: 'NotReady', label: t('filters.notReady') || 'Not Ready' },
        { id: 'unknown', value: 'Unknown', label: t('filters.unknown') || 'Unknown' },
      ],
    },
    {
      id: 'cluster',
      label: t('kubernetes.cluster'),
      type: 'select',
      options: Array.from(new Set(allNodes.map((n: NodeWithMetadata) => n.cluster_name)))
        .filter(Boolean)
        .map((clusterName: string, idx: number) => ({ 
          id: `cluster-${idx}`, 
          value: clusterName, 
          label: clusterName 
        })),
    },
    {
      id: 'provider',
      label: t('common.provider'),
      type: 'select',
      options: Array.from(new Set(allNodes.map((n: NodeWithMetadata) => n.provider).filter((p): p is CloudProvider => !!p)))
        .map((provider: CloudProvider, idx: number) => ({ 
          id: `provider-${idx}`, 
          value: provider, 
          label: provider 
        })),
    },
  ], [allNodes, t]);

  const isLoadingCombined = isLoadingCredentials || isLoadingClusters || nodeQueriesResults.some(r => r.isLoading);

  const isEmpty = selectedCredentialIds.length === 0 || 
    (useProviderRegionFilterMode ? (!hasSelectedRegions || filteredNodes.length === 0) : filteredNodes.length === 0);

  const emptyStateComponent = credentials.length === 0 ? (
    <CredentialRequiredState serviceName={t('kubernetes.title')} />
  ) : selectedCredentialIds.length === 0 ? (
    <ResourceEmptyState
      resourceName={t('kubernetes.nodes')}
      title={t('credential.selectCredential')}
      description={t('credential.selectCredential')}
      withCard={true}
    />
  ) : useProviderRegionFilterMode && !hasSelectedRegions ? (
    <ResourceEmptyState
      resourceName={t('kubernetes.nodes')}
      title={t('kubernetes.selectRegions') || 'Select Regions'}
      description={t('kubernetes.selectRegionsDescription') || 'Please select at least one region for each provider to view nodes.'}
      withCard={true}
    />
  ) : filteredNodes.length === 0 ? (
    <ResourceEmptyState
      resourceName={t('kubernetes.nodes')}
      title={t('kubernetes.noNodesFound')}
      description={t('kubernetes.noNodesFoundDescription') || 'No nodes found in any cluster'}
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
        title={t('kubernetes.nodes')}
        resourceName={t('kubernetes.nodes')}
        storageKey="kubernetes-nodes-page"
        header={
          <NodesPageHeaderSection
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
            nodeErrors={nodeErrors}
          />
        }
        items={filteredNodes}
        isLoading={isLoadingCombined}
        isEmpty={isEmpty}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        onSearchClear={clearSearch}
        isSearching={isSearching}
        searchPlaceholder="Search nodes by name, cluster, instance type, status, or zone..."
        filterConfigs={selectedCredentialIds.length > 0 && allNodes.length > 0 ? filterConfigs : []}
        filters={filters}
        onFiltersChange={setFilters}
        onFiltersClear={() => setFilters({})}
        showFilters={showFilters}
        onToggleFilters={() => setShowFilters(!showFilters)}
        filterCount={Object.keys(filters).length}
        toolbar={
          selectedCredentialIds.length > 0 && filteredNodes.length > 0 ? (
            <BulkActionsToolbar
              items={filteredNodes}
              selectedIds={selectedNodeIds}
              onSelectionChange={setSelectedNodeIds}
              getItemDisplayName={(node) => (node as NodeWithMetadata).name}
            />
          ) : null
        }
        additionalControls={
          selectedCredentialIds.length > 0 && allNodes.length > 0 ? (
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
          selectedCredentialIds.length > 0 && filteredNodes.length > 0 ? (
            <NodesPageContent
              nodes={allNodes}
              filteredNodes={filteredNodes}
              paginatedNodes={paginatedNodes}
              selectedProvider={selectedProvider}
              selectedNodeIds={selectedNodeIds}
              onSelectionChange={setSelectedNodeIds}
              page={page}
              pageSize={pageSize}
              total={filteredNodes.length}
              onPageChange={setPage}
              onPageSizeChange={handlePageSizeChange}
              isSearching={isSearching}
              searchQuery={searchQuery}
              isMultiProviderMode={isMultiProviderMode}
              selectedProviders={selectedProviders}
            />
          ) : emptyStateComponent
        }
        pageSize={pageSize}
        onPageSizeChange={handlePageSizeChange}
        searchResultsCount={filteredNodes.length}
        skeletonColumns={7}
        skeletonRows={5}
        skeletonShowCheckbox={true}
        showFilterButton={false}
        showSearchResultsInfo={false}
      />
    </>
  );
}

export default function KubernetesNodesPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <KubernetesNodesPageContent />
    </Suspense>
  );
}
