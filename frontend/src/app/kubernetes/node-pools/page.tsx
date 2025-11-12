/**
 * Kubernetes Node Pools Page
 * Kubernetes 노드 풀 관리 페이지
 */

'use client';

import * as React from 'react';
import { Suspense } from 'react';
import { useState, useMemo, useCallback, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useStandardMutation } from '@/hooks/use-standard-mutation';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { usePagination } from '@/hooks/use-pagination';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { UI } from '@/lib/constants';
import { useWorkspaceStore } from '@/store/workspace';
import { Filter } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Checkbox } from '@/components/ui/checkbox';
import { FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { kubernetesService } from '@/features/kubernetes';
import { useCredentials } from '@/hooks/use-credentials';
import { Plus, Trash2, Edit, Layers } from 'lucide-react';
import { CreateNodePoolForm, NodePool } from '@/lib/types';
import { DataProcessor } from '@/lib/data';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useCredentialAutoSelect } from '@/hooks/use-credential-auto-select';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import { useTranslation } from '@/hooks/use-translation';
import { createValidationSchemas } from '@/lib/validation';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import { sseService } from '@/services/sse';
import { log } from '@/lib/logging';
import { useSSEStatus } from '@/hooks/use-sse-status';

function NodePoolsPageContent() {
  const { t } = useTranslation();
  const { createNodePoolSchema } = createValidationSchemas(t);
  const { currentWorkspace } = useWorkspaceStore();
  const { success } = useToast();
  const { handleError } = useErrorHandler();

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

  // SSE 이벤트 구독 (Kubernetes Node Pool 실시간 업데이트)
  useEffect(() => {
    // SSE 연결 완료 확인 (clientId는 subscribeToEvent 내부에서 대기 처리)
    if (!sseStatus.isConnected) {
      log.debug('[Node Pools Page] SSE not connected, skipping subscription', {
        isConnected: sseStatus.isConnected,
        readyState: sseStatus.readyState,
      });
      return;
    }

    const filters = {
      credential_ids: selectedCredentialId ? [selectedCredentialId] : undefined,
      regions: selectedRegion ? [selectedRegion] : undefined,
    };

    const subscribeToNodePoolEvents = async () => {
      try {
        await sseService.subscribeToEvent('kubernetes-node-pool-created', filters);
        await sseService.subscribeToEvent('kubernetes-node-pool-updated', filters);
        await sseService.subscribeToEvent('kubernetes-node-pool-deleted', filters);
        
        log.debug('[Node Pools Page] Subscribed to Kubernetes Node Pool events', { 
          filters,
          clientId: sseService.getClientId(),
        });
      } catch (error) {
        log.error('[Node Pools Page] Failed to subscribe to Kubernetes Node Pool events', error, {
          service: 'SSE',
          action: 'subscribeNodePoolEvents',
        });
      }
    };

    subscribeToNodePoolEvents();

    // Cleanup: 페이지를 떠날 때 또는 필터가 변경될 때 구독 해제
    return () => {
      const unsubscribe = async () => {
        try {
          await sseService.unsubscribeFromEvent('kubernetes-node-pool-created', filters);
          await sseService.unsubscribeFromEvent('kubernetes-node-pool-updated', filters);
          await sseService.unsubscribeFromEvent('kubernetes-node-pool-deleted', filters);
          
          log.debug('[Node Pools Page] Unsubscribed from Kubernetes Node Pool events', { filters });
        } catch (error) {
          log.warn('[Node Pools Page] Failed to unsubscribe from Kubernetes Node Pool events', error, {
            service: 'SSE',
            action: 'unsubscribeNodePoolEvents',
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
  const [selectedNodePoolIds, setSelectedNodePoolIds] = useState<string[]>([]);
  const [pageSize, setPageSize] = useState(UI.PAGINATION.DEFAULT_PAGE_SIZE);
  const [deleteDialogState, setDeleteDialogState] = useState<{
    open: boolean;
    nodePoolName: string | null;
    region: string | null;
  }>({
    open: false,
    nodePoolName: null,
    region: null,
  });

  const nodePoolForm = useForm<CreateNodePoolForm>({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    resolver: zodResolver(createNodePoolSchema as any),
    defaultValues: {
      cluster_name: '',
      region: '',
      node_count: 1,
      auto_scaling: false,
    },
  });

  // Fetch credentials using unified hook
  const { credentials, selectedProvider } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
  });

  // Fetch clusters for selection
  const { data: clusters = [] } = useQuery({
    queryKey: queryKeys.clusters.list(selectedProvider, selectedCredentialId || undefined, selectedRegion || undefined),
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId) return [];
      return kubernetesService.listClusters(selectedProvider, selectedCredentialId, selectedRegion || '');
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!currentWorkspace,
  });

  // Fetch Node Pools
  const { data: nodePools = [], isLoading: isLoadingNodePools } = useQuery({
    queryKey: queryKeys.nodePools.list(selectedProvider, selectedCredentialId || undefined, selectedRegion || undefined, selectedClusterName),
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId || !selectedClusterName || !selectedRegion) return [];
      return kubernetesService.listNodePools(selectedProvider, selectedClusterName, selectedCredentialId, selectedRegion || '');
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!selectedClusterName && !!selectedRegion && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: CACHE_TIMES.REALTIME,
  });

  // Search functionality
  const [searchQuery, setSearchQuery] = useState('');

  // Custom filter function for node pool filtering (memoized)
  const filterFn = React.useCallback((nodePool: NodePool, filters: FilterValue): boolean => {
    if (filters.status && nodePool.status !== filters.status) return false;
    return true;
  }, []);

  // Filtered node pools (memoized for consistency)
  const filteredNodePools = useMemo(() => {
    let result = DataProcessor.search(nodePools, searchQuery, {
      keys: ['name', 'cluster_name', 'status'],
      threshold: 0.3,
    });

    result = DataProcessor.filter(result, filters, filterFn);
    
    return result;
  }, [nodePools, searchQuery, filters, filterFn]);

  const isSearching = searchQuery.length > 0;

  const clearSearch = useCallback(() => {
    setSearchQuery('');
  }, []);

  // Pagination
  const {
    paginatedItems: paginatedNodePools,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredNodePools, {
    totalItems: filteredNodePools.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  // Create Node Pool mutation
  const createNodePoolMutation = useStandardMutation({
    mutationFn: (data: CreateNodePoolForm) => {
      if (!selectedProvider || !selectedClusterName) throw new Error('Provider or cluster not selected');
      return kubernetesService.createNodePool(selectedProvider, selectedClusterName, data);
    },
    invalidateQueries: [queryKeys.kubernetesClusters.all],
    successMessage: t('kubernetes.nodePoolCreationInitiated'),
    errorContext: { operation: 'createNodePool', resource: 'NodePool' },
    onSuccess: () => {
      setIsCreateDialogOpen(false);
      nodePoolForm.reset();
    },
  });

  // Delete Node Pool mutation
  const deleteNodePoolMutation = useStandardMutation({
    mutationFn: async ({ nodePoolName, credentialId, region }: { nodePoolName: string; credentialId: string; region: string }) => {
      if (!selectedProvider || !selectedClusterName) throw new Error('Provider or cluster not selected');
      return kubernetesService.deleteNodePool(selectedProvider, selectedClusterName, nodePoolName, credentialId, region);
    },
    invalidateQueries: [queryKeys.kubernetesClusters.all],
    successMessage: t('kubernetes.nodePoolDeletionInitiated'),
    errorContext: { operation: 'deleteNodePool', resource: 'NodePool' },
  });

  const handleCreateNodePool = useCallback((data: CreateNodePoolForm) => {
    createNodePoolMutation.mutate(data);
  }, [createNodePoolMutation]);

  const handleBulkDeleteNodePools = useCallback(async (nodePoolIds: string[]) => {
    if (!selectedCredentialId || !selectedProvider || !selectedClusterName) return;
    
    const nodePoolsToDelete = filteredNodePools.filter(np => nodePoolIds.includes(np.id));
    const deletePromises = nodePoolsToDelete.map(nodePool =>
      deleteNodePoolMutation.mutateAsync({
        nodePoolName: nodePool.name,
        credentialId: selectedCredentialId,
        region: nodePool.region,
      })
    );

    try {
      await Promise.all(deletePromises);
      success(`Successfully initiated deletion of ${nodePoolIds.length} node pool(s)`);
      setSelectedNodePoolIds([]);
    } catch (error) {
      handleError(error, { operation: 'bulkDeleteNodePools', resource: 'NodePool' });
    }
  }, [selectedCredentialId, selectedProvider, selectedClusterName, filteredNodePools, deleteNodePoolMutation, success, handleError]);

  const handleDeleteNodePool = (nodePoolName: string, region: string) => {
    if (!selectedCredentialId || !selectedProvider || !selectedClusterName) return;
    setDeleteDialogState({ open: true, nodePoolName, region });
  };

  const handleConfirmDelete = () => {
    if (!deleteDialogState.nodePoolName || !deleteDialogState.region || !selectedCredentialId || !selectedProvider || !selectedClusterName) return;
    
    deleteNodePoolMutation.mutate({
      nodePoolName: deleteDialogState.nodePoolName,
      credentialId: selectedCredentialId,
      region: deleteDialogState.region,
    }, {
      onSuccess: () => {
        setDeleteDialogState({ open: false, nodePoolName: null, region: null });
      },
    });
  };

  const handlePageSizeChange = (newSize: number) => {
    setPageSize(newSize);
    setPaginationPageSize(newSize);
  };

  // Update form when cluster/region changes
  React.useEffect(() => {
    if (selectedClusterName && selectedRegion) {
      nodePoolForm.setValue('cluster_name', selectedClusterName);
      nodePoolForm.setValue('region', selectedRegion);
    }
  }, [selectedClusterName, selectedRegion, nodePoolForm]);

  // Header component
  const header = (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">{t('kubernetes.nodePools')}</h1>
          <p className="text-gray-600 mt-1">
            {currentWorkspace 
              ? t('kubernetes.manageNodePoolsWithWorkspace', { workspaceName: currentWorkspace.name }) 
              : t('kubernetes.manageNodePools')
            }
          </p>
        </div>
        <div className="flex items-center space-x-2">
          {/* Credential selection is now handled in Header */}
        </div>
      </div>
      
      {/* Configuration Card - Cluster Selection */}
      {selectedProvider && selectedCredentialId && (
        <Card>
          <CardHeader>
            <CardTitle>{t('common.configuration')}</CardTitle>
            <CardDescription>{t('kubernetes.selectClusterToViewNodePools')}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <Label>{t('kubernetes.cluster')} *</Label>
              <Select
                value={selectedClusterName}
                onValueChange={(value) => {
                  setSelectedClusterName(value);
                  nodePoolForm.setValue('cluster_name', value);
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder={t('kubernetes.selectCluster')} />
                </SelectTrigger>
                <SelectContent>
                  {clusters.length === 0 ? (
                    <div className="px-2 py-6 text-center text-sm text-muted-foreground">
                      {t('kubernetes.noClustersFound')}
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
    </div>
  );

  // Filter configurations
  const filterConfigs: FilterConfig[] = useMemo(() => [
    {
      id: 'status',
      label: t('filters.status'),
      type: 'select',
      options: [
        { id: 'RUNNING', value: 'RUNNING', label: t('filters.running') },
        { id: 'CREATING', value: 'CREATING', label: t('kubernetes.creating') },
        { id: 'DELETING', value: 'DELETING', label: t('kubernetes.deleting') },
        { id: 'ERROR', value: 'ERROR', label: t('kubernetes.failed') },
      ],
    },
  ], [t]);

  // Determine empty state
  const isEmpty = !selectedProvider || !selectedCredentialId || !selectedClusterName || !selectedRegion || filteredNodePools.length === 0;

  // Empty state component
  const emptyStateComponent = credentials.length === 0 ? (
    <CredentialRequiredState serviceName={t('kubernetes.title')} />
  ) : !selectedProvider || !selectedCredentialId ? (
    <CredentialRequiredState
      title={t('credential.selectCredential')}
      description={t('credential.selectCredential')}
      serviceName={t('kubernetes.title')}
    />
  ) : !selectedClusterName || !selectedRegion ? (
    <Card>
      <CardContent className="flex flex-col items-center justify-center py-12">
        <Layers className="h-12 w-12 text-gray-400 mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">
          {t('kubernetes.selectClusterAndRegion')}
        </h3>
        <p className="text-sm text-gray-500 text-center">
          {t('kubernetes.selectClusterAndRegionMessage')}
        </p>
      </CardContent>
    </Card>
  ) : filteredNodePools.length === 0 ? (
    <ResourceEmptyState
      resourceName={t('kubernetes.nodePools')}
      icon={Layers}
      onCreateClick={() => setIsCreateDialogOpen(true)}
      description={t('kubernetes.noNodePoolsFoundForCluster')}
      withCard={true}
    />
  ) : null;

  return (
    <>
    <ResourceListPage
        title={t('kubernetes.nodePools')}
        resourceName={t('kubernetes.nodePools')}
        storageKey="kubernetes-node-pools-page"
        header={header}
        items={filteredNodePools}
        isLoading={isLoadingNodePools}
        isEmpty={isEmpty}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        onSearchClear={clearSearch}
        isSearching={isSearching}
        searchPlaceholder={t('kubernetes.searchNodePoolsPlaceholder')}
        filterConfigs={selectedCredentialId && selectedClusterName && nodePools.length > 0 ? filterConfigs : []}
        filters={filters}
        onFiltersChange={setFilters}
        onFiltersClear={() => setFilters({})}
        showFilters={showFilters}
        onToggleFilters={() => setShowFilters(!showFilters)}
        filterCount={Object.keys(filters).length}
        toolbar={
          selectedProvider && selectedCredentialId && selectedClusterName && selectedRegion && filteredNodePools.length > 0 ? (
            <BulkActionsToolbar
              items={filteredNodePools}
              selectedIds={selectedNodePoolIds}
              onSelectionChange={setSelectedNodePoolIds}
              onBulkDelete={handleBulkDeleteNodePools}
              getItemDisplayName={(nodePool) => nodePool.name}
            />
          ) : null
        }
        additionalControls={
          selectedCredentialId && selectedClusterName && nodePools.length > 0 ? (
            <>
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
          selectedProvider && selectedCredentialId && selectedClusterName && selectedRegion && filteredNodePools.length > 0 ? (
            <>
              <Card>
                <CardHeader>
                  <CardTitle>{t('kubernetes.nodePools')}</CardTitle>
                  <CardDescription>
                    {filteredNodePools.length} of {nodePools.length} node pool{nodePools.length !== 1 ? 's' : ''} 
                    {isSearching && ` (${searchQuery})`}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-12">
                          <Checkbox
                            checked={selectedNodePoolIds.length === filteredNodePools.length && filteredNodePools.length > 0}
                            onCheckedChange={(checked) => {
                              if (checked) {
                                setSelectedNodePoolIds(filteredNodePools.map(np => np.id));
                              } else {
                                setSelectedNodePoolIds([]);
                              }
                            }}
                          />
                        </TableHead>
                        <TableHead>Name</TableHead>
                        <TableHead>Cluster</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Node Count</TableHead>
                        <TableHead>Instance Type</TableHead>
                        <TableHead>Auto Scaling</TableHead>
                        <TableHead>Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {paginatedNodePools.map((nodePool) => {
                        const isSelected = selectedNodePoolIds.includes(nodePool.id);
                        
                        return (
                          <TableRow key={nodePool.id}>
                            <TableCell>
                              <Checkbox
                                checked={isSelected}
                                onCheckedChange={(checked) => {
                                  if (checked) {
                                    setSelectedNodePoolIds([...selectedNodePoolIds, nodePool.id]);
                                  } else {
                                    setSelectedNodePoolIds(selectedNodePoolIds.filter(id => id !== nodePool.id));
                                  }
                                }}
                              />
                            </TableCell>
                            <TableCell className="font-medium">{nodePool.name}</TableCell>
                            <TableCell>{nodePool.cluster_name}</TableCell>
                            <TableCell>
                              <Badge variant={nodePool.status === 'RUNNING' ? 'default' : 'secondary'}>
                                {nodePool.status}
                              </Badge>
                            </TableCell>
                            <TableCell>{nodePool.node_count}</TableCell>
                            <TableCell>{nodePool.instance_type}</TableCell>
                            <TableCell>
                              {nodePool.auto_scaling ? (
                                <Badge variant="outline">{t('common.enabled')}</Badge>
                              ) : (
                                <Badge variant="secondary">{t('common.disabled')}</Badge>
                              )}
                            </TableCell>
                            <TableCell>
                              <div className="flex items-center space-x-2">
                                <Button variant="ghost" size="sm">
                                  <Edit className="h-4 w-4" />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => handleDeleteNodePool(nodePool.name, nodePool.region)}
                                  disabled={deleteNodePoolMutation.isPending}
                                >
                                  <Trash2 className="h-4 w-4 text-red-600" />
                                </Button>
                              </div>
                            </TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </CardContent>
              </Card>

              {/* Create Node Pool Dialog */}
              <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
                <DialogTrigger asChild>
                  <Button disabled={credentials.length === 0 || !selectedClusterName}>
                    <Plus className="mr-2 h-4 w-4" />
                    {t('kubernetes.createNodePool')}
                  </Button>
                </DialogTrigger>
                <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
                  <DialogHeader>
                    <DialogTitle>{t('kubernetes.createNodePool')}</DialogTitle>
                    <DialogDescription>
                      {t('kubernetes.createNodePoolOnProvider', { provider: selectedProvider?.toUpperCase() || '' })}
                    </DialogDescription>
                  </DialogHeader>
                  <form onSubmit={nodePoolForm.handleSubmit(handleCreateNodePool)} className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="nodepool-name">{t('vm.name')} *</Label>
                      <Input id="nodepool-name" {...nodePoolForm.register('name')} placeholder="my-node-pool" />
                      {nodePoolForm.formState.errors.name && (
                        <p className="text-sm text-red-600">{nodePoolForm.formState.errors.name.message}</p>
                      )}
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor="nodepool-instance-type">{t('vm.type')} *</Label>
                        <Input id="nodepool-instance-type" {...nodePoolForm.register('instance_type')} placeholder="n1-standard-2" />
                        {nodePoolForm.formState.errors.instance_type && (
                          <p className="text-sm text-red-600">{nodePoolForm.formState.errors.instance_type.message}</p>
                        )}
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="nodepool-node-count">{t('kubernetes.nodeCount')} *</Label>
                        <Input 
                          id="nodepool-node-count" 
                          type="number"
                          {...nodePoolForm.register('node_count', { valueAsNumber: true })} 
                          placeholder="1"
                          min="1"
                        />
                        {nodePoolForm.formState.errors.node_count && (
                          <p className="text-sm text-red-600">{nodePoolForm.formState.errors.node_count.message}</p>
                        )}
                      </div>
                    </div>
                    <div className="flex justify-end space-x-2">
                      <Button type="button" variant="outline" onClick={() => setIsCreateDialogOpen(false)}>
                        {t('common.cancel')}
                      </Button>
                      <Button type="submit" disabled={createNodePoolMutation.isPending}>
                        {createNodePoolMutation.isPending ? t('kubernetes.creatingNodePool') : t('kubernetes.createNodePool')}
                      </Button>
                    </div>
                  </form>
                </DialogContent>
              </Dialog>
            </>
          ) : emptyStateComponent
        }
        pageSize={pageSize}
        onPageSizeChange={handlePageSizeChange}
        searchResultsCount={filteredNodePools.length}
        skeletonColumns={7}
        skeletonRows={5}
        skeletonShowCheckbox={true}
      showFilterButton={false}
      showSearchResultsInfo={false}
      />

      {/* Delete Confirmation Dialog */}
      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => setDeleteDialogState({ ...deleteDialogState, open })}
        onConfirm={handleConfirmDelete}
        title={t('kubernetes.deleteNodePool')}
        description={deleteDialogState.nodePoolName ? t('kubernetes.confirmDeleteNodePool', { nodePoolName: deleteDialogState.nodePoolName }) : ''}
        isLoading={deleteNodePoolMutation.isPending}
        resourceName={deleteDialogState.nodePoolName || undefined}
        resourceNameLabel="노드 풀 이름"
      />
    </>
  );
}

export default function NodePoolsPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <NodePoolsPageContent />
    </Suspense>
  );
}

