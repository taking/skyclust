/**
 * Kubernetes Nodes Page
 * Kubernetes 노드 관리 페이지
 */

'use client';

import { Suspense } from 'react';
import { useState, useMemo, useCallback, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { ResourceListPage } from '@/components/common/resource-list-page';
import { usePagination } from '@/hooks/use-pagination';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { Filter } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { kubernetesService } from '@/features/kubernetes';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentials } from '@/hooks/use-credentials';
import { Building2 } from 'lucide-react';
import { Node } from '@/lib/types';
import { DataProcessor } from '@/lib/data';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useCredentialAutoSelect } from '@/hooks/use-credential-auto-select';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import { useTranslation } from '@/hooks/use-translation';
import { UI } from '@/lib/constants';
import { sseService } from '@/services/sse';
import { log } from '@/lib/logging';
import { useSSEStatus } from '@/hooks/use-sse-status';

function NodesPageContent() {
  const { t } = useTranslation();
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

  // SSE 이벤트 구독 (Kubernetes Node 실시간 업데이트)
  useEffect(() => {
    // SSE 연결 완료 확인 (clientId는 subscribeToEvent 내부에서 대기 처리)
    if (!sseStatus.isConnected) {
      log.debug('[Nodes Page] SSE not connected, skipping subscription', {
        isConnected: sseStatus.isConnected,
        readyState: sseStatus.readyState,
      });
      return;
    }

    const filters = {
      credential_ids: selectedCredentialId ? [selectedCredentialId] : undefined,
      regions: selectedRegion ? [selectedRegion] : undefined,
    };

    const subscribeToNodeEvents = async () => {
      try {
        await sseService.subscribeToEvent('kubernetes-node-created', filters);
        await sseService.subscribeToEvent('kubernetes-node-updated', filters);
        await sseService.subscribeToEvent('kubernetes-node-deleted', filters);
        
        log.debug('[Nodes Page] Subscribed to Kubernetes Node events', { 
          filters,
          clientId: sseService.getClientId(),
        });
      } catch (error) {
        log.error('[Nodes Page] Failed to subscribe to Kubernetes Node events', error, {
          service: 'SSE',
          action: 'subscribeNodeEvents',
        });
      }
    };

    subscribeToNodeEvents();

    // Cleanup: 페이지를 떠날 때 또는 필터가 변경될 때 구독 해제
    return () => {
      const unsubscribe = async () => {
        try {
          await sseService.unsubscribeFromEvent('kubernetes-node-created', filters);
          await sseService.unsubscribeFromEvent('kubernetes-node-updated', filters);
          await sseService.unsubscribeFromEvent('kubernetes-node-deleted', filters);
          
          log.debug('[Nodes Page] Unsubscribed from Kubernetes Node events', { filters });
        } catch (error) {
          log.warn('[Nodes Page] Failed to unsubscribe from Kubernetes Node events', error, {
            service: 'SSE',
            action: 'unsubscribeNodeEvents',
          });
        }
      };
      unsubscribe();
    };
  }, [selectedCredentialId, selectedRegion, sseStatus.isConnected]);

  const [selectedClusterName, setSelectedClusterName] = useState<string>('');
  const [filters, setFilters] = useState<FilterValue>({});
  const [showFilters, setShowFilters] = useState(false);
  const [pageSize, setPageSize] = useState(UI.PAGINATION.DEFAULT_PAGE_SIZE);

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

  // Fetch Nodes
  const { data: nodes = [], isLoading: isLoadingNodes } = useQuery({
    queryKey: queryKeys.nodes.list(selectedProvider, selectedClusterName, selectedCredentialId || undefined, selectedRegion || undefined),
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId || !selectedClusterName || !selectedRegion) return [];
      return kubernetesService.listNodes(selectedProvider, selectedClusterName, selectedCredentialId, selectedRegion || '');
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!selectedClusterName && !!selectedRegion && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: 10000,
  });

  const [searchQuery, setSearchQuery] = useState('');

  // Custom filter function for node filtering (memoized)
  const filterFn = useCallback((node: Node, filters: FilterValue): boolean => {
    if (filters.status && node.status !== filters.status) return false;
    return true;
  }, []);

  // Filtered nodes using DataProcessor (memoized for consistency)
  const filteredNodes = useMemo(() => {
    let result = DataProcessor.search(nodes, searchQuery, {
      keys: ['name', 'cluster_name', 'status', 'private_ip', 'public_ip'],
      threshold: 0.3,
    });

    result = DataProcessor.filter(result, filters, filterFn);
    
    return result;
  }, [nodes, searchQuery, filters, filterFn]);

  const isSearching = searchQuery.length > 0;

  const clearSearch = useCallback(() => {
    setSearchQuery('');
  }, []);

  // Pagination
  const {
    paginatedItems: paginatedNodes,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredNodes, {
    totalItems: filteredNodes.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  const handlePageSizeChange = useCallback((newSize: number) => {
    setPageSize(newSize);
    setPaginationPageSize(newSize);
  }, [setPaginationPageSize]);

  // Filter configurations
  const filterConfigs: FilterConfig[] = useMemo(() => [
    {
      id: 'status',
      label: t('filters.status'),
      type: 'select',
      options: [
        { id: 'Ready', value: 'Ready', label: t('filters.active') },
        { id: 'NotReady', value: 'NotReady', label: t('filters.inactive') },
        { id: 'Unknown', value: 'Unknown', label: t('filters.unknown') },
      ],
    },
  ], [t]);

  // Header component
  const header = (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">{t('kubernetes.nodes')}</h1>
          <p className="text-gray-600 mt-1">
            {currentWorkspace 
              ? t('kubernetes.manageNodesWithWorkspace', { workspaceName: currentWorkspace.name }) 
              : t('kubernetes.manageNodes')
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
            <CardDescription>{t('kubernetes.selectClusterToViewNodes')}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <Label>{t('kubernetes.cluster')} *</Label>
              <Select
                value={selectedClusterName}
                onValueChange={setSelectedClusterName}
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

  // Determine empty state
  const isEmpty = !selectedProvider || !selectedCredentialId || !selectedClusterName || !selectedRegion || filteredNodes.length === 0;

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
        <Building2 className="h-12 w-12 text-gray-400 mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">
          {t('kubernetes.selectClusterAndRegion')}
        </h3>
        <p className="text-sm text-gray-500 text-center">
          {t('kubernetes.selectClusterAndRegionMessage')}
        </p>
      </CardContent>
    </Card>
  ) : filteredNodes.length === 0 ? (
    <ResourceEmptyState
      resourceName={t('kubernetes.nodes')}
      icon={Building2}
      description={t('kubernetes.noNodesFoundForCluster')}
      withCard={true}
    />
  ) : null;

  return (
    <ResourceListPage
        title={t('kubernetes.nodes')}
        resourceName={t('kubernetes.nodes')}
        storageKey="kubernetes-nodes-page"
        header={header}
        items={filteredNodes}
        isLoading={isLoadingNodes}
        isEmpty={isEmpty}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        onSearchClear={clearSearch}
        isSearching={isSearching}
        searchPlaceholder="Search nodes by name, cluster, status, or IP..."
        filterConfigs={selectedCredentialId && selectedClusterName && nodes.length > 0 ? filterConfigs : []}
        filters={filters}
        onFiltersChange={setFilters}
        onFiltersClear={() => setFilters({})}
        showFilters={showFilters}
        onToggleFilters={() => setShowFilters(!showFilters)}
        filterCount={Object.keys(filters).length}
        additionalControls={
          selectedCredentialId && selectedClusterName && nodes.length > 0 ? (
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
          selectedProvider && selectedCredentialId && selectedClusterName && selectedRegion && filteredNodes.length > 0 ? (
            <>
              <Card>
                <CardHeader>
                  <CardTitle>Nodes</CardTitle>
                  <CardDescription>
                    {filteredNodes.length} of {nodes.length} node{nodes.length !== 1 ? 's' : ''} 
                    {isSearching && ` (${searchQuery})`}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Name</TableHead>
                        <TableHead>Cluster</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Instance Type</TableHead>
                        <TableHead>Zone</TableHead>
                        <TableHead>Private IP</TableHead>
                        <TableHead>Public IP</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {paginatedNodes.map((node) => (
                        <TableRow key={node.id}>
                          <TableCell className="font-medium">{node.name}</TableCell>
                          <TableCell>{node.cluster_name}</TableCell>
                          <TableCell>
                            <Badge variant={node.status === 'Ready' ? 'default' : 'secondary'}>
                              {node.status}
                            </Badge>
                          </TableCell>
                          <TableCell>{node.instance_type}</TableCell>
                          <TableCell>{node.zone || '-'}</TableCell>
                          <TableCell>{node.private_ip || '-'}</TableCell>
                          <TableCell>{node.public_ip || '-'}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </CardContent>
              </Card>
            </>
          ) : emptyStateComponent
        }
        pageSize={pageSize}
        onPageSizeChange={handlePageSizeChange}
        searchResultsCount={filteredNodes.length}
        skeletonColumns={7}
        skeletonRows={5}
      showFilterButton={false}
      showSearchResultsInfo={false}
      />
  );
}

export default function NodesPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <NodesPageContent />
    </Suspense>
  );
}

