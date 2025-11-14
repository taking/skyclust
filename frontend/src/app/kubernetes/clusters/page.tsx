/**
 * Kubernetes Clusters Page
 * Kubernetes 클러스터 관리 페이지
 * 
 * ResourceListPage 템플릿을 사용한 리팩토링 버전
 */

'use client';

import { Suspense, useState, useCallback, useRef } from 'react';
import { useMemo, useEffect } from 'react';
import * as React from 'react';
import { useRouter } from 'next/navigation';
import dynamic from 'next/dynamic';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { TagFilter } from '@/components/common/tag-filter';
import { BulkOperationProgress } from '@/components/common/bulk-operation-progress';
import { usePagination } from '@/hooks/use-pagination';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { UI } from '@/lib/constants';
import { useWorkspaceStore } from '@/store/workspace';
import { Filter } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useCredentialAutoSelect } from '@/hooks/use-credential-auto-select';
import { useTranslation } from '@/hooks/use-translation';
import { useResourceListState } from '@/hooks/use-resource-list-state';
import { sseService } from '@/services/sse';
import { log } from '@/lib/logging';
import { useSSEStatus } from '@/hooks/use-sse-status';
import { useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '@/lib/query';
import {
  ClusterPageHeader,
  useKubernetesClusters,
  useClusterFilters,
  useClusterBulkActions,
  useClusterTagDialog,
} from '@/features/kubernetes';
import { downloadKubeconfig } from '@/utils/kubeconfig';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { Breadcrumb } from '@/components/common/breadcrumb';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import type { CreateClusterForm, CloudProvider } from '@/lib/types/kubernetes';

// Dynamic imports for heavy components
const BulkTagDialog = dynamic(
  () => import('@/features/kubernetes').then(mod => ({ default: mod.BulkTagDialog })),
  { 
    ssr: false,
    loading: () => null,
  }
);

const ClusterTable = dynamic(
  () => import('@/features/kubernetes').then(mod => ({ default: mod.ClusterTable })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={8} rows={5} showCheckbox={true} />, // Updated: 8 columns (Name, Status, Version, Region, Resource Group (Azure), Endpoint, Created, Actions)
  }
);

function KubernetesClustersPageContent() {
  const { success } = useToast();
  const { handleError } = useErrorHandler();
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  // Get workspace from store (consistent with other pages)
  const { currentWorkspace } = useWorkspaceStore();

  // Get credential context from global store
  const { selectedCredentialId, selectedRegion } = useCredentialContext();
  
  // Auto-select credential if not selected
  useCredentialAutoSelect({
    enabled: !!currentWorkspace,
    resourceType: 'kubernetes',
    updateUrl: true,
  });

  // SSE 상태 확인
  const { status: sseStatus } = useSSEStatus();

  // Refresh 상태 관리
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const lastRefreshTimeRef = useRef<number>(0);
  const REFRESH_DEBOUNCE_MS = 2000; // 2초 debounce

  // SSE 이벤트 구독 (Kubernetes 클러스터 실시간 업데이트)
  useEffect(() => {
    // SSE 연결 완료 확인 (clientId는 subscribeToEvent 내부에서 대기 처리)
    if (!sseStatus.isConnected) {
      log.debug('[Kubernetes Clusters Page] SSE not connected, skipping subscription', {
        isConnected: sseStatus.isConnected,
        readyState: sseStatus.readyState,
      });
      return;
    }

    const filters = {
      credential_ids: selectedCredentialId ? [selectedCredentialId] : undefined,
      regions: selectedRegion ? [selectedRegion] : undefined,
    };

    const subscribeToKubernetesEvents = async () => {
      try {
        await sseService.subscribeToEvent('kubernetes-cluster-created', filters);
        await sseService.subscribeToEvent('kubernetes-cluster-updated', filters);
        await sseService.subscribeToEvent('kubernetes-cluster-deleted', filters);
        await sseService.subscribeToEvent('kubernetes-cluster-list', filters);
        
        log.debug('[Kubernetes Clusters Page] Subscribed to Kubernetes cluster events', { 
          filters,
          clientId: sseService.getClientId(),
        });
      } catch (error) {
        log.error('[Kubernetes Clusters Page] Failed to subscribe to Kubernetes cluster events', error, {
          service: 'SSE',
          action: 'subscribeKubernetesEvents',
        });
      }
    };

    subscribeToKubernetesEvents();

    // Cleanup: 페이지를 떠날 때 또는 필터가 변경될 때 구독 해제
    return () => {
      const unsubscribe = async () => {
        try {
          await sseService.unsubscribeFromEvent('kubernetes-cluster-created', filters);
          await sseService.unsubscribeFromEvent('kubernetes-cluster-updated', filters);
          await sseService.unsubscribeFromEvent('kubernetes-cluster-deleted', filters);
          await sseService.unsubscribeFromEvent('kubernetes-cluster-list', filters);
          
          log.debug('[Kubernetes Clusters Page] Unsubscribed from Kubernetes cluster events', { filters });
        } catch (error) {
          log.warn('[Kubernetes Clusters Page] Failed to unsubscribe from Kubernetes cluster events', error, {
            service: 'SSE',
            action: 'unsubscribeKubernetesEvents',
          });
        }
      };
      unsubscribe();
    };
  }, [selectedCredentialId, selectedRegion, sseStatus.isConnected]);

  // Local state
  const router = useRouter();
  
  // 공통 리스트 상태 관리
  const {
    filters,
    setFilters,
    showFilters,
    setShowFilters,
    selectedIds: selectedClusterIds,
    setSelectedIds: setSelectedClusterIds,
    pageSize,
    setPageSize,
    deleteDialogState,
    openDeleteDialog,
    closeDeleteDialog,
  } = useResourceListState({
    storageKey: 'kubernetes-clusters-page',
  });
  
  const [tagFilters, setTagFilters] = React.useState<Record<string, string[]>>({});

  // Tag dialog hook
  const {
    isOpen: isTagDialogOpen,
    tagKey: bulkTagKey,
    tagValue: bulkTagValue,
    openDialog: openTagDialog,
    closeDialog: closeTagDialog,
    setTagKey: setBulkTagKey,
    setTagValue: setBulkTagValue,
    reset: resetTagDialog,
  } = useClusterTagDialog();

  // Kubernetes clusters hook
  const {
    credentials,
    clusters,
    isLoading,
    createClusterMutation,
    deleteClusterMutation,
    downloadKubeconfigMutation,
  } = useKubernetesClusters({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || '',
    selectedRegion: selectedRegion || '',
  });

  // Get selected credential and provider
  const selectedCredential = credentials.find(c => c.id === selectedCredentialId);
  const selectedProvider = selectedCredential?.provider as CloudProvider | undefined;

  // Refresh 핸들러 (debouncing 적용)
  const handleRefresh = useCallback(async () => {
    const now = Date.now();
    const timeSinceLastRefresh = now - lastRefreshTimeRef.current;

    // Debounce: 2초 이내 재요청 방지
    if (timeSinceLastRefresh < REFRESH_DEBOUNCE_MS) {
      return;
    }

    lastRefreshTimeRef.current = now;
    setIsRefreshing(true);

    try {
      // 클러스터 목록 쿼리 무효화 및 재요청
      await queryClient.invalidateQueries({
        queryKey: queryKeys.kubernetesClusters.list(
          currentWorkspace?.id,
          selectedProvider,
          selectedCredentialId,
          selectedRegion
        ),
      });
      
      // 관련 쿼리도 무효화 (상세 정보 등)
      await queryClient.invalidateQueries({
        queryKey: queryKeys.kubernetesClusters.all,
      });

      // 재요청
      await queryClient.refetchQueries({
        queryKey: queryKeys.kubernetesClusters.list(
          currentWorkspace?.id,
          selectedProvider,
          selectedCredentialId,
          selectedRegion
        ),
      });

      setLastUpdated(new Date());
      success(t('kubernetes.clustersRefreshed') || '클러스터 목록을 새로고침했습니다');
    } catch (error) {
      handleError(error, { operation: 'refreshClusters', resource: 'Cluster' });
    } finally {
      setIsRefreshing(false);
    }
  }, [
    queryClient,
    currentWorkspace?.id,
    selectedProvider,
    selectedCredentialId,
    selectedRegion,
    success,
    t,
    handleError,
  ]);

  // 클러스터 데이터가 업데이트될 때마다 lastUpdated 갱신 (SSE 업데이트 포함)
  useEffect(() => {
    if (clusters && clusters.length >= 0 && !isLoading) {
      // SSE 업데이트나 수동 새로고침이 아닌 경우에만 갱신 (초기 로드 시)
      if (!lastUpdated) {
        setLastUpdated(new Date());
      }
    }
  }, [clusters, isLoading, lastUpdated]);

  // Cluster filters hook
  const {
    searchQuery,
    setSearchQuery,
    isSearching,
    clearSearch,
    availableTags,
    filteredClusters,
  } = useClusterFilters({
    clusters,
    filters,
    tagFilters,
  });

  // Bulk actions hook
  const {
    bulkOperationProgress,
    handleBulkDelete,
    handleBulkTag,
    handleCancelOperation,
    clearProgress,
  } = useClusterBulkActions(selectedProvider, selectedCredentialId || undefined);

  // Pagination
  const {
    page,
    paginatedItems: paginatedClusters,
    setPage,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredClusters, {
    totalItems: filteredClusters.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  // Filter configurations (memoized)
  const filterConfigs: FilterConfig[] = useMemo(() => [
    {
      id: 'status',
      label: t('filters.status'),
      type: 'select',
      options: [
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
      options: Array.from(new Set(clusters.map(c => c.region)))
        .filter(Boolean)
        .map((r, idx) => ({ 
          id: `region-${idx}`, 
          value: r, 
          label: r 
        })),
    },
  ], [clusters, t]);

  // Event handlers

  const handleDeleteCluster = (clusterName: string, region: string) => {
    if (!selectedCredentialId || !selectedProvider) return;
    openDeleteDialog(clusterName, region);
  };

  const handleConfirmDelete = () => {
    if (!deleteDialogState.id || !deleteDialogState.region || !selectedCredentialId || !selectedProvider) return;
    
    deleteClusterMutation.mutate(
      {
        provider: selectedProvider,
        clusterName: deleteDialogState.id,
        credentialId: selectedCredentialId,
        region: deleteDialogState.region,
      },
      {
        onSuccess: () => {
          success(t('kubernetes.clusterDeletionInitiated'));
          closeDeleteDialog();
        },
        onError: (error: unknown) => {
          handleError(error, { operation: 'deleteCluster', resource: 'Cluster' });
        },
      }
    );
  };

  const handleDownloadKubeconfig = (clusterName: string, region: string) => {
    if (!selectedCredentialId || !selectedProvider) return;
    downloadKubeconfigMutation.mutate(
      {
        provider: selectedProvider,
        clusterName,
        credentialId: selectedCredentialId,
        region,
      },
      {
        onSuccess: (kubeconfig) => {
          downloadKubeconfig(kubeconfig, clusterName);
          success(t('kubernetes.downloadKubeconfig'));
        },
        onError: (error: unknown) => {
          handleError(error, { operation: 'downloadKubeconfig', resource: 'Cluster' });
        },
      }
    );
  };

  const handleBulkDeleteSubmit = (clusterIds: string[]) => {
    handleBulkDelete(clusterIds, filteredClusters, success, (error: unknown) => 
      handleError(error, { operation: 'bulkDeleteClusters', resource: 'Cluster' })
    );
  };

  const handleBulkTagSubmit = () => {
    if (!bulkTagKey.trim() || !bulkTagValue.trim()) return;
    handleBulkTag(
      selectedClusterIds,
      filteredClusters,
      bulkTagKey,
      bulkTagValue,
      success,
      (error: unknown) => handleError(error, { operation: 'bulkTagClusters', resource: 'Cluster' })
    );
    resetTagDialog();
    setSelectedClusterIds([]);
  };

  const handlePageSizeChange = (newSize: number) => {
    setPageSize(newSize);
    setPaginationPageSize(newSize);
  };

  // Determine empty state
  const isEmpty = !selectedProvider || !selectedCredentialId || filteredClusters.length === 0;

  // Empty state component
  const emptyStateComponent = credentials.length === 0 ? (
    <CredentialRequiredState serviceName={t('kubernetes.title')} />
  ) : !selectedProvider ? (
    <ResourceEmptyState
      resourceName={t('kubernetes.clusters')}
      title={t('credential.selectCredential')}
      description={t('credential.selectCredential')}
      withCard={true}
    />
  ) : !selectedCredentialId ? (
    <CredentialRequiredState
      title={t('credential.selectCredential')}
      description={t('credential.selectCredential')}
      serviceName={t('kubernetes.title')}
    />
  ) : filteredClusters.length === 0 ? (
    <ResourceEmptyState
      resourceName={t('kubernetes.clusters')}
      title={t('kubernetes.noClustersFound')}
      description={t('kubernetes.createFirst')}
      onCreateClick={() => router.push('/kubernetes/clusters/create')}
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
        title={t('kubernetes.clusters')}
        resourceName={t('kubernetes.clusters')}
        storageKey="kubernetes-clusters-page"
        header={
          <ClusterPageHeader
            workspaceName={currentWorkspace?.name}
            credentials={credentials}
            selectedCredentialId={selectedCredentialId || ''}
            onCredentialChange={() => {}} // Handled by Header
            selectedRegion={selectedRegion || ''}
            onRegionChange={() => {}} // Handled by Header
            selectedProvider={selectedProvider}
            onRefresh={handleRefresh}
            isRefreshing={isRefreshing}
            lastUpdated={lastUpdated}
          />
        }
      items={filteredClusters}
      isLoading={isLoading}
      isEmpty={isEmpty}
      searchQuery={searchQuery}
      onSearchChange={setSearchQuery}
      onSearchClear={clearSearch}
      isSearching={isSearching}
      searchPlaceholder="Search clusters by name, version, status, or region..."
      filterConfigs={selectedCredentialId && clusters.length > 0 ? filterConfigs : []}
      filters={filters}
      onFiltersChange={setFilters}
      onFiltersClear={() => setFilters({})}
      showFilters={showFilters}
      onToggleFilters={() => setShowFilters(!showFilters)}
      filterCount={Object.keys(filters).length}
      toolbar={
        <>
          {bulkOperationProgress && (
            <BulkOperationProgress
              {...bulkOperationProgress}
              onDismiss={clearProgress}
              onCancel={handleCancelOperation}
            />
          )}
          
          {selectedCredentialId && filteredClusters.length > 0 && (
            <BulkActionsToolbar
              items={filteredClusters}
              selectedIds={selectedClusterIds}
              onSelectionChange={setSelectedClusterIds}
              onBulkDelete={handleBulkDeleteSubmit}
              onBulkTag={openTagDialog}
              getItemDisplayName={(cluster) => cluster.name}
            />
          )}
        </>
      }
      additionalControls={
        selectedCredentialId && clusters.length > 0 ? (
          <>
            <TagFilter
              availableTags={availableTags}
              selectedTags={tagFilters}
              onTagsChange={setTagFilters}
            />
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
          </>
        ) : null
      }
      emptyState={emptyStateComponent}
      content={
        selectedCredentialId && filteredClusters.length > 0 ? (
          <>
            <Card>
              <CardHeader>
                <CardTitle>Clusters</CardTitle>
                <CardDescription>
                  {filteredClusters.length} of {clusters.length} cluster{clusters.length !== 1 ? 's' : ''} 
                  {isSearching && ` (${searchQuery})`}
                </CardDescription>
              </CardHeader>
              <CardContent>
            <ClusterTable
              clusters={paginatedClusters}
              provider={selectedProvider}
              selectedIds={selectedClusterIds}
              onSelectionChange={setSelectedClusterIds}
              onDelete={handleDeleteCluster}
              onDownloadKubeconfig={handleDownloadKubeconfig}
              isDeleting={deleteClusterMutation.isPending}
              isDownloading={downloadKubeconfigMutation.isPending}
              page={page}
              pageSize={pageSize}
              total={filteredClusters.length}
              onPageChange={setPage}
              onPageSizeChange={handlePageSizeChange}
            />
              </CardContent>
            </Card>

            {/* Bulk Tag Dialog */}
            <BulkTagDialog
              open={isTagDialogOpen}
              onOpenChange={(open) => open ? openTagDialog() : closeTagDialog()}
              onSubmit={handleBulkTagSubmit}
              tagKey={bulkTagKey}
              tagValue={bulkTagValue}
              onTagKeyChange={setBulkTagKey}
              onTagValueChange={setBulkTagValue}
              selectedCount={selectedClusterIds.length}
            />
          </>
        ) : emptyStateComponent
      }
      pageSize={pageSize}
      onPageSizeChange={handlePageSizeChange}
      searchResultsCount={filteredClusters.length}
      skeletonColumns={6}
      skeletonRows={5}
      skeletonShowCheckbox={true}
      showFilterButton={false}
      showSearchResultsInfo={false}
      />

      {/* Delete Confirmation Dialog */}
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
        title={t('kubernetes.deleteCluster')}
        description={deleteDialogState.id ? t('kubernetes.confirmDeleteCluster', { clusterName: deleteDialogState.id }) : ''}
        isLoading={deleteClusterMutation.isPending}
        resourceName={deleteDialogState.name || deleteDialogState.id || undefined}
        resourceNameLabel="클러스터 이름"
      />
    </>
  );
}

export default function KubernetesClustersPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <KubernetesClustersPageContent />
    </Suspense>
  );
}

