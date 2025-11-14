/**
 * VPCs Page (Refactored)
 * Virtual Private Cloud 관리 페이지 - 리팩토링된 버전
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
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { useEmptyState } from '@/hooks/use-empty-state';
import { useCredentialAutoSelect } from '@/hooks/use-credential-auto-select';
import { useWorkspaceStore } from '@/store/workspace';
import { useTranslation } from '@/hooks/use-translation';
import { DataProcessor } from '@/lib/data';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import { useResourceListState } from '@/hooks/use-resource-list-state';
import { sseService } from '@/services/sse';
import { log } from '@/lib/logging';
import { useSSEStatus } from '@/hooks/use-sse-status';
import { useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '@/lib/query';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import {
  useVPCActions,
  VPCsPageHeader,
} from '@/features/networks';
import { useNetworkResources } from '@/features/networks/hooks/use-network-resources';
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
  const queryClient = useQueryClient();
  const { success } = useToast();
  const { handleError } = useErrorHandler();

  const {
    vpcs,
    isLoadingVPCs,
    credentials,
    selectedProvider,
    selectedCredentialId,
    selectedRegion,
  } = useNetworkResources({ resourceType: 'vpcs' });
  
  // Auto-select credential if not selected
  useCredentialAutoSelect({
    enabled: !!currentWorkspace,
    resourceType: 'network',
    updateUrl: true,
  });

  // SSE 상태 확인
  const { status: sseStatus } = useSSEStatus();

  // Refresh 상태 관리
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const lastRefreshTimeRef = useRef<number>(0);
  const REFRESH_DEBOUNCE_MS = 2000; // 2초 debounce

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
      // VPC 목록 쿼리 무효화 및 재요청
      await queryClient.invalidateQueries({
        queryKey: queryKeys.vpcs.list(
          selectedProvider,
          selectedCredentialId,
          selectedRegion
        ),
      });
      
      // 관련 쿼리도 무효화
      await queryClient.invalidateQueries({
        queryKey: queryKeys.vpcs.all,
      });

      // 재요청
      await queryClient.refetchQueries({
        queryKey: queryKeys.vpcs.list(
          selectedProvider,
          selectedCredentialId,
          selectedRegion
        ),
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
    selectedProvider,
    selectedCredentialId,
    selectedRegion,
    success,
    t,
    handleError,
  ]);

  // VPC 데이터가 업데이트될 때마다 lastUpdated 갱신 (SSE 업데이트 포함)
  useEffect(() => {
    if (vpcs && vpcs.length >= 0 && !isLoadingVPCs) {
      // SSE 업데이트나 수동 새로고침이 아닌 경우에만 갱신 (초기 로드 시)
      if (!lastUpdated) {
        setLastUpdated(new Date());
      }
    }
  }, [vpcs, isLoadingVPCs, lastUpdated]);

  // SSE 이벤트 구독 (VPC 실시간 업데이트)
  useEffect(() => {
    // SSE 연결 완료 확인 (clientId는 subscribeToEvent 내부에서 대기 처리)
    if (!sseStatus.isConnected) {
      log.debug('[VPCs Page] SSE not connected, skipping subscription', {
        isConnected: sseStatus.isConnected,
        readyState: sseStatus.readyState,
      });
      return;
    }

    const filters = {
      credential_ids: selectedCredentialId ? [selectedCredentialId] : undefined,
      regions: selectedRegion ? [selectedRegion] : undefined,
    };

    const subscribeToVPCEvents = async () => {
      try {
        await sseService.subscribeToEvent('network-vpc-created', filters);
        await sseService.subscribeToEvent('network-vpc-updated', filters);
        await sseService.subscribeToEvent('network-vpc-deleted', filters);
        await sseService.subscribeToEvent('network-vpc-list', filters);
        
        log.debug('[VPCs Page] Subscribed to VPC events', { 
          filters,
          clientId: sseService.getClientId(),
        });
      } catch (error) {
        log.error('[VPCs Page] Failed to subscribe to VPC events', error, {
          service: 'SSE',
          action: 'subscribeVPCEvents',
        });
      }
    };

    subscribeToVPCEvents();

    // Cleanup: 페이지를 떠날 때 또는 필터가 변경될 때 구독 해제
    return () => {
      const unsubscribe = async () => {
        try {
          await sseService.unsubscribeFromEvent('network-vpc-created', filters);
          await sseService.unsubscribeFromEvent('network-vpc-updated', filters);
          await sseService.unsubscribeFromEvent('network-vpc-deleted', filters);
          await sseService.unsubscribeFromEvent('network-vpc-list', filters);
          
          log.debug('[VPCs Page] Unsubscribed from VPC events', { filters });
        } catch (error) {
          log.warn('[VPCs Page] Failed to unsubscribe from VPC events', error, {
            service: 'SSE',
            action: 'unsubscribeVPCEvents',
          });
        }
      };
      unsubscribe();
    };
  }, [selectedCredentialId, selectedRegion, sseStatus.isConnected]);

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
    selectedCredentialId,
    selectedRegion,
  });

  const handleCreateVPC = useCallback(() => {
    router.push('/networks/vpcs/create');
  }, [router]);

  const handleDeleteVPC = useCallback((vpcId: string, region?: string) => {
    if (!region) return;
    const vpc = vpcs.find(v => v.id === vpcId);
    openDeleteDialog(vpcId, region, vpc?.name || vpc?.id);
  }, [vpcs, openDeleteDialog]);

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

  const handleToggleFilters = useCallback(() => {
    setShowFilters(prev => !prev);
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

  // Empty state
  const { isEmpty, emptyStateComponent } = useEmptyState({
    credentials,
    selectedProvider,
    selectedCredentialId,
    filteredItems: filteredVPCs,
    resourceName: t('network.vpcs'),
    serviceName: t('network.title'),
    onCreateClick: handleCreateVPC,
    icon: Network,
    isSearching,
    searchQuery,
  });

  return (
    <>
    <ResourceListPage
        title={t('network.vpcs')}
        resourceName={t('network.vpcs')}
        storageKey="vpcs-page"
        header={
          <VPCsPageHeader
            onRefresh={handleRefresh}
            isRefreshing={isRefreshing}
            lastUpdated={lastUpdated}
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
        filterConfigs={selectedCredentialId && vpcs.length > 0 ? filterConfigs : []}
        filters={filters}
        onFiltersChange={setFilters}
        onFiltersClear={() => setFilters({})}
        showFilters={showFilters}
        onToggleFilters={handleToggleFilters}
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
        onOpenChange={(open) => {
          if (open && deleteDialogState.id) {
            openDeleteDialog(deleteDialogState.id, deleteDialogState.region || undefined, deleteDialogState.name);
          } else {
            closeDeleteDialog();
          }
        }}
        onConfirm={handleConfirmDelete}
        title={t('network.deleteVPC')}
        description="이 VPC를 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다."
        isLoading={deleteVPCMutation.isPending}
        resourceName={deleteDialogState.name}
        resourceNameLabel="VPC 이름"
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

