/**
 * Kubernetes Node Groups Page
 * Kubernetes 노드 그룹 관리 페이지 (AWS EKS용)
 */

'use client';

import * as React from 'react';
import { Suspense } from 'react';
import { useState, useMemo, useCallback, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import Link from 'next/link';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { usePagination } from '@/hooks/use-pagination';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { UI } from '@/lib/constants';
import { useWorkspaceStore } from '@/store/workspace';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Checkbox } from '@/components/ui/checkbox';
import { FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { kubernetesService } from '@/features/kubernetes';
import { useCredentials } from '@/hooks/use-credentials';
import { Plus, Trash2 } from 'lucide-react';
import { CreateNodeGroupForm, NodeGroup } from '@/lib/types';
import { DataProcessor } from '@/lib/data';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useCredentialAutoSelect } from '@/hooks/use-credential-auto-select';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import { useTranslation } from '@/hooks/use-translation';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import { CreateNodeGroupDialog } from '@/features/kubernetes/components/create-node-group-dialog';
import { sseService } from '@/services/sse';
import { log } from '@/lib/logging';
import { useSSEStatus } from '@/hooks/use-sse-status';
import { useEmptyState } from '@/hooks/use-empty-state';
import type { ProviderCluster } from '@/lib/types';
import { isGPUQuotaError, extractGPUQuotaErrorDetails } from '@/lib/error-handling/quota-error-handler';
import { GPUQuotaErrorAlert } from '@/components/common/gpu-quota-error-alert';

function NodeGroupsPageContent() {
  const { t } = useTranslation();
  const { currentWorkspace } = useWorkspaceStore();
  const { success } = useToast();
  const [quotaErrorDetails, setQuotaErrorDetails] = React.useState<ReturnType<typeof extractGPUQuotaErrorDetails> | null>(null);
  const { handleError } = useErrorHandler();
  const queryClient = useQueryClient();

  // Get credential context from global store
  const { selectedCredentialId, selectedRegion, setSelectedCredential, setSelectedRegion } = useCredentialContext();
  
  // Auto-select credential if not selected
  useCredentialAutoSelect({
    enabled: !!currentWorkspace,
    resourceType: 'kubernetes',
    updateUrl: true,
  });

  // SSE 상태 확인
  const { status: sseStatus } = useSSEStatus();

  // SSE 이벤트 구독 (Kubernetes Node Group 실시간 업데이트)
  useEffect(() => {
    if (!sseStatus.isConnected) {
      log.debug('[Node Groups Page] SSE not connected, skipping subscription', {
        isConnected: sseStatus.isConnected,
        readyState: sseStatus.readyState,
      });
      return;
    }

    const filters = {
      credential_ids: selectedCredentialId ? [selectedCredentialId] : undefined,
      regions: selectedRegion ? [selectedRegion] : undefined,
    };

    // Note: Backend currently uses 'kubernetes-node-pool-*' events for both node pools and node groups
    // AWS EKS node groups are handled through the same event system as node pools
    const subscribeToNodeGroupEvents = async () => {
      try {
        await sseService.subscribeToEvent('kubernetes-node-pool-created', filters);
        await sseService.subscribeToEvent('kubernetes-node-pool-updated', filters);
        await sseService.subscribeToEvent('kubernetes-node-pool-deleted', filters);
        
        log.debug('[Node Groups Page] Subscribed to Kubernetes Node Group events (via node-pool events)', { 
          filters,
          clientId: sseService.getClientId(),
        });
      } catch (error) {
        log.error('[Node Groups Page] Failed to subscribe to Kubernetes Node Group events', error, {
          service: 'SSE',
          action: 'subscribeNodeGroupEvents',
        });
      }
    };

    subscribeToNodeGroupEvents();

    // Cleanup: 페이지를 떠날 때 또는 필터가 변경될 때 구독 해제
    return () => {
      const unsubscribe = async () => {
        try {
          await sseService.unsubscribeFromEvent('kubernetes-node-pool-created', filters);
          await sseService.unsubscribeFromEvent('kubernetes-node-pool-updated', filters);
          await sseService.unsubscribeFromEvent('kubernetes-node-pool-deleted', filters);
          
          log.debug('[Node Groups Page] Unsubscribed from Kubernetes Node Group events', { filters });
        } catch (error) {
          log.warn('[Node Groups Page] Failed to unsubscribe from Kubernetes Node Group events', error, {
            service: 'SSE',
            action: 'unsubscribeNodeGroupEvents',
          });
        }
      };
      unsubscribe();
    };
  }, [selectedCredentialId, selectedRegion, sseStatus.isConnected]);

  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [selectedClusterName, setSelectedClusterName] = useState<string>('');
  const [filters, setFilters] = useState<FilterValue>({});
  const [showFilters, setShowFilters] = useState(false);
  const [selectedNodeGroupIds, setSelectedNodeGroupIds] = useState<string[]>([]);
  const [pageSize, setPageSize] = useState(UI.PAGINATION.DEFAULT_PAGE_SIZE);
  const [deleteDialogState, setDeleteDialogState] = useState<{
    open: boolean;
    nodeGroupName: string | null;
    region: string | null;
  }>({
    open: false,
    nodeGroupName: null,
    region: null,
  });

  // Fetch credentials using unified hook
  const { credentials, selectedCredential, selectedProvider } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
  });

  // Ensure AWS provider
  const isAWS = selectedProvider === 'aws';

  // Fetch clusters for selection
  const { data: clusters = [], isLoading: isLoadingClusters } = useQuery({
    queryKey: queryKeys.kubernetesClusters.list(
      currentWorkspace?.id,
      selectedProvider,
      selectedCredentialId || undefined,
      selectedRegion || undefined
    ),
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId || selectedProvider !== 'aws') return [];
      return kubernetesService.listClusters(selectedProvider, selectedCredentialId, selectedRegion || '');
    },
    enabled: !!selectedProvider && !!selectedCredentialId && isAWS && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
  });

  // Fetch selected cluster details for CreateNodeGroupDialog
  const { data: selectedCluster } = useQuery<ProviderCluster | null>({
    queryKey: queryKeys.kubernetesClusters.detail(selectedClusterName),
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId || !selectedClusterName || !selectedRegion || !isAWS) return null;
      return kubernetesService.getCluster(selectedProvider, selectedClusterName, selectedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!selectedClusterName && !!selectedRegion && isAWS && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
  });

  // Fetch Node Groups
  const { data: nodeGroups = [], isLoading: isLoadingNodeGroups } = useQuery({
    queryKey: queryKeys.nodeGroups.list(selectedProvider, selectedCredentialId || undefined, selectedRegion || undefined, selectedClusterName),
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId || !selectedClusterName || !selectedRegion || !isAWS) return [];
      return kubernetesService.listNodeGroups(selectedProvider, selectedClusterName, selectedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!selectedClusterName && !!selectedRegion && isAWS && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: CACHE_TIMES.REALTIME,
  });

  // Search functionality
  const [searchQuery, setSearchQuery] = useState('');

  // Custom filter function for node group filtering (memoized)
  const filterFn = React.useCallback((nodeGroup: NodeGroup, filters: FilterValue): boolean => {
    if (filters.status && nodeGroup.status !== filters.status) return false;
    return true;
  }, []);

  // Filtered node groups (memoized for consistency)
  const filteredNodeGroups = useMemo(() => {
    let result = DataProcessor.search(nodeGroups, searchQuery, {
      keys: ['name', 'cluster_name', 'status'],
      threshold: 0.3,
    });

    result = DataProcessor.filter(result, filters, filterFn);
    
    return result;
  }, [nodeGroups, searchQuery, filters, filterFn]);

  const isSearching = searchQuery.length > 0;

  const clearSearch = useCallback(() => {
    setSearchQuery('');
  }, []);

  // Pagination
  const {
    paginatedItems: paginatedNodeGroups,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredNodeGroups, {
    totalItems: filteredNodeGroups.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  // Create Node Group mutation
  const createNodeGroupMutation = useMutation({
    mutationFn: (data: CreateNodeGroupForm) => {
      if (!selectedProvider || !selectedClusterName || !isAWS) throw new Error('AWS provider and cluster required');
      return kubernetesService.createNodeGroup(selectedProvider, selectedClusterName, data);
    },
    onSuccess: () => {
      success(t('kubernetes.nodeGroupCreationInitiated') || 'Node group creation initiated');
      queryClient.invalidateQueries({ queryKey: queryKeys.kubernetesClusters.all });
      setIsCreateDialogOpen(false);
    },
    onError: (error: unknown) => {
      // GPU quota 에러인지 확인
      if (isGPUQuotaError(error)) {
        const details = extractGPUQuotaErrorDetails(error);
        if (details) {
          setQuotaErrorDetails(details);
          // 일반 에러 처리도 수행 (toast 메시지 표시)
          handleError(error, { operation: 'createNodeGroup', resource: 'NodeGroup' });
          return;
        }
      }
      // 일반 에러 처리
      setQuotaErrorDetails(null);
      handleError(error, { operation: 'createNodeGroup', resource: 'NodeGroup' });
    },
  });

  // Delete Node Group mutation
  const deleteNodeGroupMutation = useMutation({
    mutationFn: async ({ nodeGroupName, credentialId, region }: { nodeGroupName: string; credentialId: string; region: string }) => {
      if (!selectedProvider || !selectedClusterName || !isAWS) throw new Error('AWS provider and cluster required');
      return kubernetesService.deleteNodeGroup(selectedProvider, selectedClusterName, nodeGroupName, credentialId, region);
    },
    onSuccess: () => {
      success(t('kubernetes.nodeGroupDeletionInitiated') || 'Node group deletion initiated');
      queryClient.invalidateQueries({ queryKey: queryKeys.kubernetesClusters.all });
    },
    onError: (error: unknown) => {
      handleError(error, { operation: 'deleteNodeGroup', resource: 'NodeGroup' });
    },
  });

  const handleCreateNodeGroup = useCallback((data: CreateNodeGroupForm) => {
    createNodeGroupMutation.mutate(data);
  }, [createNodeGroupMutation]);

  const handleBulkDeleteNodeGroups = useCallback(async (nodeGroupIds: string[]) => {
    if (!selectedCredentialId || !selectedProvider || !selectedClusterName || !isAWS) return;
    
    const nodeGroupsToDelete = filteredNodeGroups.filter(ng => nodeGroupIds.includes(ng.id));
    const deletePromises = nodeGroupsToDelete.map(nodeGroup =>
      deleteNodeGroupMutation.mutateAsync({
        nodeGroupName: nodeGroup.name,
        credentialId: selectedCredentialId,
        region: nodeGroup.region,
      })
    );

    try {
      await Promise.all(deletePromises);
      success(`Successfully initiated deletion of ${nodeGroupIds.length} node group(s)`);
      setSelectedNodeGroupIds([]);
    } catch (error) {
      handleError(error, { operation: 'bulkDeleteNodeGroups', resource: 'NodeGroup' });
    }
  }, [selectedCredentialId, selectedProvider, selectedClusterName, filteredNodeGroups, deleteNodeGroupMutation, success, handleError, isAWS]);

  const handleDeleteNodeGroup = (nodeGroupName: string, region: string) => {
    if (!selectedCredentialId || !selectedProvider || !selectedClusterName || !isAWS) return;
    setDeleteDialogState({ open: true, nodeGroupName, region });
  };

  const handleConfirmDelete = () => {
    if (!deleteDialogState.nodeGroupName || !deleteDialogState.region || !selectedCredentialId || !selectedProvider || !selectedClusterName || !isAWS) return;
    
    deleteNodeGroupMutation.mutate({
      nodeGroupName: deleteDialogState.nodeGroupName,
      credentialId: selectedCredentialId,
      region: deleteDialogState.region,
    }, {
      onSuccess: () => {
        setDeleteDialogState({ open: false, nodeGroupName: null, region: null });
      },
    });
  };

  const handlePageSizeChange = (newSize: number) => {
    setPageSize(newSize);
    setPaginationPageSize(newSize);
  };

  // Filter configurations
  const filterConfigs: FilterConfig[] = useMemo(() => {
    if (!selectedCredentialId || !selectedClusterName || !nodeGroups.length) return [];
    
    const statuses = Array.from(new Set(nodeGroups.map(ng => ng.status).filter(Boolean)));
    return [
      {
        id: 'status',
        label: t('common.status'),
        type: 'select',
        options: statuses.map(status => ({ value: status, label: status })),
      },
    ];
  }, [selectedCredentialId, selectedClusterName, nodeGroups, t]);

  // Empty state
  const { isEmpty, emptyStateComponent } = useEmptyState({
    credentials,
    selectedProvider,
    selectedCredentialId,
    filteredItems: filteredNodeGroups,
    resourceName: t('kubernetes.nodeGroups'),
    serviceName: t('nav.kubernetes'),
    onCreateClick: () => setIsCreateDialogOpen(true),
    additionalConditions: [
      {
        value: !!selectedRegion,
        title: t('kubernetes.selectClusterAndRegion') || 'Select Cluster and Region',
        description: t('kubernetes.selectClusterAndRegionMessage') || 'Please select a cluster and region to view node groups',
      },
      {
        value: !!selectedClusterName,
        title: t('kubernetes.selectClusterToViewNodeGroups') || 'Select Cluster',
        description: t('kubernetes.selectClusterToViewNodeGroups') || 'Select a cluster to view node groups',
      },
    ],
  });

  // Header component
  const header = (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">{t('kubernetes.nodeGroups')}</h1>
          <p className="text-gray-600 mt-1">
            {currentWorkspace 
              ? t('kubernetes.manageNodeGroupsWithWorkspace', { workspaceName: currentWorkspace.name }) || `Manage node groups in ${currentWorkspace.name}`
              : t('kubernetes.manageNodeGroups') || 'Manage Kubernetes node groups'
            }
          </p>
        </div>
        <div className="flex items-center space-x-2">
          {selectedClusterName && (
            <Button onClick={() => setIsCreateDialogOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              {t('common.create')} {t('kubernetes.nodeGroups')}
            </Button>
          )}
        </div>
      </div>
      
      {/* Configuration Card - Cluster Selection */}
      {isAWS && selectedCredentialId && (
        <Card>
          <CardHeader>
            <CardTitle>{t('common.configuration')}</CardTitle>
            <CardDescription>{t('kubernetes.selectClusterToViewNodeGroups') || 'Select a cluster to view node groups'}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <Label>{t('kubernetes.cluster')} *</Label>
              <Select
                value={selectedClusterName}
                onValueChange={(value) => {
                  setSelectedClusterName(value);
                }}
                disabled={isLoadingClusters}
              >
                <SelectTrigger>
                  <SelectValue placeholder={isLoadingClusters ? t('common.loading') : t('kubernetes.selectCluster')} />
                </SelectTrigger>
                <SelectContent>
                  {clusters.length === 0 ? (
                    <div className="px-2 py-6 text-center text-sm text-muted-foreground">
                      {t('kubernetes.noClustersFound') || 'No clusters found'}
                    </div>
                  ) : (
                    clusters.map((cluster) => (
                      <SelectItem key={cluster.name} value={cluster.name}>
                        {cluster.name}
                      </SelectItem>
                    ))
                  )}
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>
      )}

      {!isAWS && selectedCredentialId && (
        <Card>
          <CardContent className="py-6">
            <p className="text-sm text-muted-foreground text-center">
              {t('kubernetes.nodeGroupsOnlyForAWS') || 'Node Groups are only available for AWS EKS clusters. Please select an AWS credential.'}
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  );

  // Bulk actions
  const bulkActions = selectedNodeGroupIds.length > 0 ? (
    <BulkActionsToolbar
      selectedCount={selectedNodeGroupIds.length}
      onDelete={() => handleBulkDeleteNodeGroups(selectedNodeGroupIds)}
      isDeleting={deleteNodeGroupMutation.isPending}
    />
  ) : null;

  // Table component
  const table = selectedProvider === 'aws' && selectedCredentialId && selectedClusterName && selectedRegion && filteredNodeGroups.length > 0 ? (
    <Card>
      <CardHeader>
        <CardTitle>{t('kubernetes.nodeGroups')}</CardTitle>
        <CardDescription>
          {filteredNodeGroups.length} of {nodeGroups.length} node group{nodeGroups.length !== 1 ? 's' : ''} 
          {isSearching && ` (${searchQuery})`}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12">
                <Checkbox
                  checked={selectedNodeGroupIds.length === filteredNodeGroups.length && filteredNodeGroups.length > 0}
                  onCheckedChange={(checked) => {
                    if (checked) {
                      setSelectedNodeGroupIds(filteredNodeGroups.map(ng => ng.id));
                    } else {
                      setSelectedNodeGroupIds([]);
                    }
                  }}
                />
              </TableHead>
              <TableHead>{t('common.name')}</TableHead>
              <TableHead>{t('common.instanceType')}</TableHead>
              <TableHead>{t('common.nodes')}</TableHead>
              <TableHead>{t('common.minMax')}</TableHead>
              <TableHead>{t('common.status')}</TableHead>
              <TableHead>{t('common.actions')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {paginatedNodeGroups.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                  {t('common.noResults') || 'No results found'}
                </TableCell>
              </TableRow>
            ) : (
              paginatedNodeGroups.map((ng) => (
                <TableRow key={ng.id || ng.name}>
                  <TableCell>
                    <Checkbox
                      checked={selectedNodeGroupIds.includes(ng.id)}
                      onCheckedChange={(checked) => {
                        if (checked) {
                          setSelectedNodeGroupIds([...selectedNodeGroupIds, ng.id]);
                        } else {
                          setSelectedNodeGroupIds(selectedNodeGroupIds.filter(id => id !== ng.id));
                        }
                      }}
                    />
                  </TableCell>
                  <TableCell className="font-medium">
                    <Link
                      href={`/kubernetes/node-groups/${ng.name}?cluster=${ng.cluster_name}`}
                      className="text-blue-600 hover:text-blue-800 hover:underline"
                    >
                      {ng.name}
                    </Link>
                  </TableCell>
                  <TableCell>{ng.instance_type}</TableCell>
                  <TableCell>{ng.node_count}</TableCell>
                  <TableCell>{ng.min_size}/{ng.max_size}</TableCell>
                  <TableCell>
                    <Badge variant={ng.status === 'ACTIVE' ? 'default' : 'secondary'}>
                      {ng.status}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleDeleteNodeGroup(ng.name, ng.region)}
                      disabled={deleteNodeGroupMutation.isPending}
                    >
                      <Trash2 className="h-4 w-4 text-red-600" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  ) : null;

  return (
    <>
      <ResourceListPage
        title={t('kubernetes.nodeGroups')}
        resourceName={t('kubernetes.nodeGroups')}
        storageKey="kubernetes-node-groups-page"
        header={header}
        items={filteredNodeGroups}
        isLoading={isLoadingNodeGroups || isLoadingClusters}
        isEmpty={isEmpty}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        onSearchClear={clearSearch}
        isSearching={isSearching}
        searchPlaceholder={t('kubernetes.searchNodeGroupsPlaceholder') || 'Search node groups by name, cluster, or status...'}
        filterConfigs={selectedCredentialId && selectedClusterName && nodeGroups.length > 0 ? filterConfigs : []}
        filters={filters}
        onFiltersChange={setFilters}
        onFiltersClear={() => setFilters({})}
        showFilters={showFilters}
        onToggleFilters={() => setShowFilters(!showFilters)}
        filterCount={Object.keys(filters).length}
        toolbar={bulkActions}
        emptyState={emptyStateComponent}
        content={table}
        pageSize={pageSize}
        onPageSizeChange={handlePageSizeChange}
        searchResultsCount={filteredNodeGroups.length}
        skeletonColumns={7}
      />

      {/* GPU Quota Error Alert */}
      {quotaErrorDetails && (
        <div className="fixed bottom-4 right-4 z-50 max-w-2xl">
          <GPUQuotaErrorAlert
            errorDetails={quotaErrorDetails}
            onRegionChange={(region) => {
              setSelectedRegion(region);
              setQuotaErrorDetails(null);
            }}
          />
        </div>
      )}

      {/* Create Node Group Dialog */}
      {selectedClusterName && selectedCredentialId && selectedRegion && (
        <CreateNodeGroupDialog
          open={isCreateDialogOpen}
          onOpenChange={(open) => {
            setIsCreateDialogOpen(open);
            if (!open) {
              // 다이얼로그가 닫힐 때 quota 에러 상태 초기화
              setQuotaErrorDetails(null);
            }
          }}
          clusterName={selectedClusterName}
          cluster={selectedCluster || null}
          defaultRegion={selectedRegion}
          defaultCredentialId={selectedCredentialId}
          onSubmit={handleCreateNodeGroup}
          onCredentialIdChange={setSelectedCredential}
          onRegionChange={setSelectedRegion}
          isPending={createNodeGroupMutation.isPending}
        />
      )}

      {/* Delete Confirmation Dialog */}
      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => setDeleteDialogState({ ...deleteDialogState, open })}
        onConfirm={handleConfirmDelete}
        title={t('kubernetes.deleteNodeGroup') || 'Delete Node Group'}
        description={deleteDialogState.nodeGroupName ? t('kubernetes.confirmDeleteNodeGroup', { nodeGroupName: deleteDialogState.nodeGroupName }) || `Are you sure you want to delete node group ${deleteDialogState.nodeGroupName}? This action cannot be undone.` : ''}
        isLoading={deleteNodeGroupMutation.isPending}
        resourceName={deleteDialogState.nodeGroupName || undefined}
      />
    </>
  );
}

export default function NodeGroupsPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
      </div>
    }>
      <NodeGroupsPageContent />
    </Suspense>
  );
}

