/**
 * Kubernetes Clusters Page
 * Kubernetes 클러스터 관리 페이지
 * 
 * 새로운 라우팅 구조: /{workspaceId}/{credentialId}/kubernetes/clusters
 */

'use client';

import { Suspense, useState, useCallback, useRef } from 'react';
import { useMemo, useEffect } from 'react';
import * as React from 'react';
import { useRouter } from 'next/navigation';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { TagFilter } from '@/components/common/tag-filter';
import { BulkOperationProgress } from '@/components/common/bulk-operation-progress';
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
import { useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '@/lib/query';
import { usePageRefresh } from '@/hooks/use-page-refresh';
import { useRequiredResourceContext } from '@/hooks/use-resource-context';
import { buildCredentialResourceDetailPath, buildCredentialResourceCreatePath } from '@/lib/routing/helpers';
import {
  ClustersPageHeaderSection,
  ClustersPageContent,
  useKubernetesClusters,
  useClusterFilters,
  useClusterBulkActions,
  useClusterTagDialog,
} from '@/features/kubernetes';
import { useCredentialContextStore } from '@/store/credential-context';
import { useWorkspaceStore } from '@/store/workspace';
import { useMultiProviderClusters } from '@/hooks/use-multi-provider-clusters';
import { useMultiProviderClustersWithRegions } from '@/hooks/use-multi-provider-clusters-with-regions';
import { useRegionFilter } from '@/hooks/use-region-filter';
import { useProviderRegionFilter, type ProviderRegionSelection } from '@/hooks/use-provider-region-filter';
import { downloadKubeconfig } from '@/utils/kubeconfig';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import { useCredentials } from '@/hooks/use-credentials';
import { useCredentialRegionSync } from '@/hooks/use-credential-region-sync';
import type { CreateClusterForm, CloudProvider, KubernetesCluster } from '@/lib/types/kubernetes';

function KubernetesClustersPageContent() {
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
    toggleCredential,
  } = useCredentialContextStore();

  // Credentials 조회
  const { credentials, isLoading: isLoadingCredentials } = useCredentials({
    workspaceId: workspaceId || '',
    enabled: !!workspaceId,
  });

  // Zustand persist hydration 완료 여부 추적
  // Store가 초기화되면 hydration이 완료된 것으로 간주
  const isInitializedRef = React.useRef<boolean>(false);
  const hydrationCheckRef = React.useRef<boolean>(false);

  // Store hydration 완료 확인 (초기 마운트 시 한 번만)
  React.useEffect(() => {
    if (hydrationCheckRef.current) return;
    
    // Store가 접근 가능하면 hydration 완료로 간주
    // 약간의 지연을 두어 persist hydration이 완료될 시간을 줌
    const timeoutId = setTimeout(() => {
      hydrationCheckRef.current = true;
    }, 100);

    return () => clearTimeout(timeoutId);
  }, []);

  // 초기화 로직: URL 우선, 없으면 pathCredentialId 또는 첫 번째 credential
  React.useEffect(() => {
    // Hydration 체크가 완료되지 않았거나 로딩 중이면 스킵
    if (!hydrationCheckRef.current || isLoadingCredentials) {
      return;
    }

    // 이미 초기화되었으면 스킵
    if (isInitializedRef.current) {
      return;
    }

    // Credentials가 없으면 스킵
    if (credentials.length === 0) {
      isInitializedRef.current = true;
      return;
    }

    // URL에서 읽은 credentials가 있으면 useCredentialRegionSync가 처리하므로 스킵
    const urlParams = new URLSearchParams(window.location.search);
    const urlCredentials = urlParams.get('credentials');
    
    if (urlCredentials) {
      // URL에 credentials가 있으면 useCredentialRegionSync가 처리
      // selectedCredentialIds가 업데이트될 때까지 대기
      if (selectedCredentialIds.length > 0) {
        isInitializedRef.current = true;
      }
      return;
    }

    // URL에 credentials가 없고, 선택된 credential도 없으면 pathCredentialId만 사용 (자동 선택 제거)
    if (selectedCredentialIds.length === 0 && pathCredentialId) {
      setSelectedCredentials([pathCredentialId]);
    }

    isInitializedRef.current = true;
  }, [isLoadingCredentials, credentials, selectedCredentialIds.length, pathCredentialId, setSelectedCredentials]);

  // Region 자동 설정
  useEffect(() => {
    if (pathRegion && !selectedRegion) {
      setSelectedRegion(pathRegion);
    }
  }, [pathRegion, selectedRegion, setSelectedRegion]);

  // Credentials 정규화 (메모이제이션)
  const normalizedCredentials = useMemo(() => 
    credentials.map((c) => ({ id: c.id, provider: c.provider as CloudProvider })),
    [credentials]
  );

  // 선택된 providers 추출
  const selectedProviders = useMemo(() => {
    const providers = new Set<CloudProvider>();
    selectedCredentialIds.forEach((credentialId: string) => {
      const credential = normalizedCredentials.find((c) => c.id === credentialId);
      if (credential) {
        providers.add(credential.provider);
      }
    });
    return Array.from(providers);
  }, [selectedCredentialIds, normalizedCredentials]);

  // Store에서 Provider별 Region 선택 가져오기
  const { providerSelectedRegions: providerSelectedRegionsFromStore, setProviderSelectedRegions } = useCredentialContextStore();
  
  // Provider가 변경되면 Region 선택 상태 초기화 (초기 마운트 제외)
  const prevProvidersRef = React.useRef<string>('');
  const isInitialMountRef = React.useRef<boolean>(true);
  React.useEffect(() => {
    // 초기 마운트 시에는 스킵
    if (isInitialMountRef.current) {
      isInitialMountRef.current = false;
      prevProvidersRef.current = selectedProviders.sort().join(',');
      return;
    }
    
    const currentProvidersKey = selectedProviders.sort().join(',');
    if (prevProvidersRef.current && prevProvidersRef.current !== currentProvidersKey) {
      // Provider가 변경되면 Region 선택 초기화
      setProviderSelectedRegions({
        aws: [],
        gcp: [],
        azure: [],
      });
    }
    prevProvidersRef.current = currentProvidersKey;
  }, [selectedProviders, setProviderSelectedRegions]);
  
  // Region 선택 변경 핸들러 (메모이제이션)
  const handleRegionSelectionChange = useCallback((regions: ProviderRegionSelection) => {
    setProviderSelectedRegions(regions);
    // 단일 Provider일 때 selectedRegion 동기화
    if (selectedProviders.length === 1) {
      const singleProvider = selectedProviders[0];
      const providerRegions = (regions as Record<CloudProvider, string[]>)[singleProvider] || [];
      if (providerRegions.length === 1) {
        setSelectedRegion(providerRegions[0]);
      } else if (providerRegions.length === 0) {
        setSelectedRegion(null);
      }
    } else {
      setSelectedRegion(null);
    }
  }, [selectedProviders, setProviderSelectedRegions, setSelectedRegion]);
  
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
    onRegionSelectionChange: handleRegionSelectionChange,
  });

  // 항상 Provider별 Region 지원 클러스터 조회 사용
  const {
    clusters: mergedClusters,
    isLoading: isLoadingClusters,
    errors: clusterErrors,
    hasError: hasClusterError,
  } = useMultiProviderClustersWithRegions({
    workspaceId: workspaceId || '',
    credentialIds: selectedCredentialIds,
    credentials: normalizedCredentials,
    selectedRegions: providerSelectedRegions,
    enabled: selectedCredentialIds.length > 0 && hasSelectedRegions,
  });

  // Region 필터링 훅 (단일 Region 모드용)
  const {
    availableRegions,
    setSelectedRegion: handleRegionChange,
  } = useRegionFilter({
    providers: selectedProviders,
    selectedRegion: selectedRegion || undefined,
    onRegionChange: setSelectedRegion,
  });

  // 기존 hook 사용 (mutations용)
  const {
    createClusterMutation,
    deleteClusterMutation,
    downloadKubeconfigMutation,
  } = useKubernetesClusters({
    workspaceId: workspaceId || '',
    selectedCredentialId: selectedCredentialIds[0] || '',
    selectedRegion: selectedRegion || '',
  });

  // Get selected credential and provider (첫 번째 credential 기준, 메모이제이션)
  const selectedCredential = useMemo(() => 
    normalizedCredentials.find((c) => c.id === selectedCredentialIds[0]),
    [normalizedCredentials, selectedCredentialIds]
  );
  const selectedProvider = selectedCredential?.provider;

  // Multi-provider 모드 여부 (메모이제이션)
  const isMultiProviderMode = useMemo(() => 
    selectedCredentialIds.length > 1 || selectedProviders.length > 1,
    [selectedCredentialIds.length, selectedProviders.length]
  );

  // 페이지 새로고침 시 쿼리 무효화 및 재요청 (queryKeys 메모이제이션)
  const refreshQueryKeys = useMemo(() => 
    selectedCredentialIds.map((credId) => {
      const cred = normalizedCredentials.find((c) => c.id === credId);
      return queryKeys.kubernetesClusters.list(
        workspaceId,
        cred?.provider,
        credId,
        selectedRegion || undefined
      );
    }),
    [selectedCredentialIds, normalizedCredentials, workspaceId, selectedRegion]
  );

  usePageRefresh({
    queryKeys: refreshQueryKeys,
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

  // SSE 이벤트 구독 필터 최적화: 안정적인 키 생성
  const subscriptionFilters = React.useMemo(() => {
    if (selectedCredentialIds.length === 0) {
      return undefined;
    }

    const sortedCredentialIds = [...selectedCredentialIds].sort();
    
    if (hasSelectedRegions) {
      const allSelectedRegions = Object.values(providerSelectedRegions)
        .flat()
        .filter(Boolean)
        .sort();
      
      return {
        credential_ids: sortedCredentialIds,
        regions: allSelectedRegions.length > 0 ? allSelectedRegions : undefined,
      };
    }
    
    return {
      credential_ids: sortedCredentialIds,
      regions: undefined,
    };
  }, [
    // 안정적인 의존성: 배열을 문자열로 변환하여 비교
    selectedCredentialIds.sort().join(','),
    hasSelectedRegions,
    // Provider regions를 정규화하여 비교
    JSON.stringify({
      aws: [...((providerSelectedRegions as Record<string, string[]>).aws || [])].sort(),
      gcp: [...((providerSelectedRegions as Record<string, string[]>).gcp || [])].sort(),
      azure: [...((providerSelectedRegions as Record<string, string[]>).azure || [])].sort(),
    }),
  ]);

  useSSESubscription({
    eventTypes: [
      'kubernetes-cluster-created',
      'kubernetes-cluster-updated',
      'kubernetes-cluster-deleted',
      'kubernetes-cluster-list',
    ],
    filters: subscriptionFilters,
    enabled: selectedCredentialIds.length > 0 || !!selectedRegion,
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
      // 모든 선택된 credential에 대한 쿼리 무효화
      selectedCredentialIds.forEach(credId => {
        const cred = normalizedCredentials.find(c => c.id === credId);
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
      success(t('kubernetes.clustersRefreshed') || '클러스터 목록을 새로고침했습니다');
    } catch (error) {
      handleError(error, { operation: 'refreshClusters', resource: 'Cluster' });
    } finally {
      setIsRefreshing(false);
    }
  }, [
    queryClient,
    workspaceId,
    selectedCredentialIds,
    normalizedCredentials,
    selectedRegion,
    success,
    t,
    handleError,
  ]);

  // Combined loading state
  const isLoadingCombined = isLoadingCredentials || isLoadingClusters;

  // 클러스터 데이터가 업데이트될 때마다 lastUpdated 갱신
  useEffect(() => {
    if (mergedClusters.length > 0 && !isLoadingCombined && !lastUpdated) {
      setLastUpdated(new Date());
    }
  }, [mergedClusters.length, isLoadingCombined, lastUpdated]);

  // Cluster filters hook
  const {
    searchQuery,
    setSearchQuery,
    isSearching,
    clearSearch,
    availableTags,
    filteredClusters,
  } = useClusterFilters({
    clusters: mergedClusters,
    filters,
    tagFilters,
  });

  // Bulk actions hook (첫 번째 credential 기준)
  const {
    bulkOperationProgress,
    handleBulkDelete,
    handleBulkTag,
    handleCancelOperation,
    clearProgress,
  } = useClusterBulkActions(selectedProvider, selectedCredentialIds[0] || '');

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

  // Region 옵션 메모이제이션
  const regionOptions = useMemo(() => 
    Array.from(new Set(mergedClusters.map((c) => c.region)))
      .filter(Boolean)
      .map((r, idx) => ({ 
        id: `region-${idx}`, 
        value: r, 
        label: r 
      })),
    [mergedClusters]
  );

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
      options: regionOptions,
    },
  ], [regionOptions, t]);

  // Event handlers (memoized with useCallback to prevent unnecessary re-renders)
  const handleDeleteCluster = useCallback((clusterName: string, clusterRegion: string) => {
    const firstCredentialId = selectedCredentialIds[0];
    if (!firstCredentialId || !selectedProvider) return;
    openDeleteDialog(clusterName, clusterRegion);
  }, [selectedCredentialIds, selectedProvider, openDeleteDialog]);

  const handleConfirmDelete = useCallback(() => {
    const firstCredentialId = selectedCredentialIds[0];
    if (!deleteDialogState.id || !deleteDialogState.region || !firstCredentialId || !selectedProvider) return;
    
    deleteClusterMutation.mutate(
      {
        provider: selectedProvider,
        clusterName: deleteDialogState.id,
        credentialId: firstCredentialId,
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
  }, [
    selectedCredentialIds,
    selectedProvider,
    deleteDialogState.id,
    deleteDialogState.region,
    deleteClusterMutation,
    success,
    t,
    closeDeleteDialog,
    handleError,
  ]);

  const handleDownloadKubeconfig = useCallback((clusterName: string, clusterRegion: string) => {
    const firstCredentialId = selectedCredentialIds[0];
    if (!firstCredentialId || !selectedProvider) return;
    downloadKubeconfigMutation.mutate(
      {
        provider: selectedProvider,
        clusterName,
        credentialId: firstCredentialId,
        region: clusterRegion,
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
  }, [
    selectedCredentialIds,
    selectedProvider,
    downloadKubeconfigMutation,
    success,
    t,
    handleError,
  ]);

  const handleBulkDeleteSubmit = useCallback((clusterIds: string[]) => {
    handleBulkDelete(clusterIds, filteredClusters, success, (error: unknown) => 
      handleError(error, { operation: 'bulkDeleteClusters', resource: 'Cluster' })
    );
  }, [handleBulkDelete, filteredClusters, success, handleError]);

  const handleBulkTagSubmit = useCallback(() => {
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
  }, [
    bulkTagKey,
    bulkTagValue,
    handleBulkTag,
    selectedClusterIds,
    filteredClusters,
    success,
    handleError,
    resetTagDialog,
    setSelectedClusterIds,
  ]);

  const handlePageSizeChange = useCallback((newSize: number) => {
    setPageSize(newSize);
    setPaginationPageSize(newSize);
  }, [setPageSize, setPaginationPageSize]);

  const handleClusterClick = useCallback((cluster: KubernetesCluster & { provider?: CloudProvider; credential_id?: string }) => {
    if (!workspaceId || !cluster.credential_id) return;
    const path = buildCredentialResourceDetailPath(
      workspaceId,
      cluster.credential_id,
      'k8s',
      'clusters',
      cluster.name,
      { region: cluster.region || selectedRegion || undefined }
    );
    router.push(path);
  }, [workspaceId, selectedRegion, router]);

  const handleCreateCluster = useCallback(() => {
    if (!workspaceId || selectedCredentialIds.length === 0) return;
    // 첫 번째 credential으로 생성 페이지 이동
    const firstCredentialId = selectedCredentialIds[0];
    const path = buildCredentialResourceCreatePath(
      workspaceId,
      firstCredentialId,
      'k8s',
      'clusters',
      { region: selectedRegion || undefined }
    );
    router.push(path);
  }, [workspaceId, selectedCredentialIds, selectedRegion, router]);

  // 필터 관련 핸들러들 (메모이제이션)
  const handleFiltersClear = useCallback(() => {
    setFilters({});
  }, [setFilters]);

  const handleToggleFilters = useCallback(() => {
    setShowFilters(prev => !prev);
  }, [setShowFilters]);

  // 조건부 렌더링 변수들 (메모이제이션)
  const shouldShowFilters = useMemo(() => 
    selectedCredentialIds.length > 0 && mergedClusters.length > 0,
    [selectedCredentialIds.length, mergedClusters.length]
  );

  const shouldShowToolbar = useMemo(() => 
    selectedCredentialIds.length > 0 && filteredClusters.length > 0,
    [selectedCredentialIds.length, filteredClusters.length]
  );

  // Determine empty state (메모이제이션)
  const isEmpty = useMemo(() => 
    selectedCredentialIds.length === 0 || 
    (!hasSelectedRegions || filteredClusters.length === 0),
    [selectedCredentialIds.length, hasSelectedRegions, filteredClusters.length]
  );

  // Empty state component
  const emptyStateComponent = credentials.length === 0 ? (
    <CredentialRequiredState serviceName={t('kubernetes.title')} />
  ) : selectedCredentialIds.length === 0 ? (
    <ResourceEmptyState
      resourceName={t('kubernetes.clusters.label')}
      title={t('credential.selectCredential')}
      description={t('credential.selectCredential')}
      withCard={true}
    />
  ) : !hasSelectedRegions ? (
    <ResourceEmptyState
      resourceName={t('kubernetes.clusters.label')}
      title={t('kubernetes.selectRegions') || 'Select Regions'}
      description={t('kubernetes.selectRegionsDescription') || 'Please select at least one region for each provider to view clusters.'}
      withCard={true}
    />
  ) : filteredClusters.length === 0 ? (
    <ResourceEmptyState
      resourceName={t('kubernetes.clusters.label')}
      title={t('kubernetes.noClustersFound')}
      description={t('kubernetes.createFirst')}
      onCreateClick={handleCreateCluster}
      isSearching={isSearching}
      searchQuery={searchQuery}
      hasFilters={Object.keys(filters).length > 0}
      onClearFilters={handleFiltersClear}
      onClearSearch={clearSearch}
      withCard={true}
    />
  ) : null;

  return (
    <>
    <ResourceListPage
        title={t('kubernetes.clusters.label')}
        resourceName={t('kubernetes.clusters.label')}
        storageKey="kubernetes-clusters-page"
        header={
          <ClustersPageHeaderSection
            workspaceId={workspaceId || ''}
            workspaceName={currentWorkspace?.name}
            credentials={credentials}
            selectedCredentialIds={selectedCredentialIds}
            onCredentialSelectionChange={setSelectedCredentials}
            selectedProvider={selectedProvider}
            selectedProviders={selectedProviders}
            selectedRegion={selectedRegion}
            onRegionChange={handleRegionChange}
            selectedRegions={providerSelectedRegions}
            onRegionSelectionChange={handleRegionSelectionChange}
            onRefresh={handleRefresh}
            isRefreshing={isRefreshing}
            lastUpdated={lastUpdated}
            isLoadingCredentials={isLoadingCredentials}
            isLoadingClusters={isLoadingClusters}
            clusterErrors={clusterErrors}
            onCreateClick={handleCreateCluster}
          />
        }
      items={filteredClusters}
      isLoading={isLoadingCombined}
      isEmpty={isEmpty}
      searchQuery={searchQuery}
      onSearchChange={setSearchQuery}
      onSearchClear={clearSearch}
      isSearching={isSearching}
      searchPlaceholder="Search clusters by name, version, status, or region..."
      filterConfigs={shouldShowFilters ? filterConfigs : []}
      filters={filters}
      onFiltersChange={setFilters}
      onFiltersClear={handleFiltersClear}
      showFilters={showFilters}
      onToggleFilters={handleToggleFilters}
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
          
          {shouldShowToolbar && (
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
        shouldShowFilters ? (
          <>
            <TagFilter
              availableTags={availableTags}
              selectedTags={tagFilters}
              onTagsChange={setTagFilters}
            />
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
        selectedCredentialIds.length > 0 && filteredClusters.length > 0 ? (
          <ClustersPageContent
            clusters={mergedClusters}
            filteredClusters={filteredClusters}
            paginatedClusters={paginatedClusters}
            selectedProvider={selectedProvider}
            selectedClusterIds={selectedClusterIds}
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
            isSearching={isSearching}
            searchQuery={searchQuery}
            isMultiProviderMode={isMultiProviderMode}
            selectedProviders={selectedProviders}
            isTagDialogOpen={isTagDialogOpen}
            onTagDialogOpenChange={(open) => open ? openTagDialog() : closeTagDialog()}
            onBulkTagSubmit={handleBulkTagSubmit}
            bulkTagKey={bulkTagKey}
            bulkTagValue={bulkTagValue}
            onBulkTagKeyChange={setBulkTagKey}
            onBulkTagValueChange={setBulkTagValue}
          />
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

