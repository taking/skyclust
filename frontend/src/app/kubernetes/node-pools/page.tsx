/**
 * Kubernetes Node Pools Page
 * Kubernetes 노드 풀 관리 페이지
 */

'use client';

import * as React from 'react';
import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useStandardMutation } from '@/hooks/use-standard-mutation';
import { useRouter } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { kubernetesService } from '@/features/kubernetes';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentials } from '@/hooks/use-credentials';
import { Plus, Trash2, Edit, Layers, Search } from 'lucide-react';
import { CreateNodePoolForm, CloudProvider, NodePool } from '@/lib/types';
import { useToast } from '@/hooks/use-toast';
import { useRequireAuth } from '@/hooks/use-auth';
import { DataProcessor } from '@/lib/data-processor';
import { SearchBar } from '@/components/ui/search-bar';
import { useSSEMonitoring } from '@/hooks/use-sse-monitoring';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { Checkbox } from '@/components/ui/checkbox';
import { Pagination } from '@/components/ui/pagination';
import { usePagination } from '@/hooks/use-pagination';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { WorkspaceRequired } from '@/components/common/workspace-required';
import { Layout } from '@/components/layout/layout';
import { CredentialRequiredState } from '@/components/common/credential-required-state';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { queryKeys } from '@/lib/query-keys';
import { CACHE_TIMES, GC_TIMES } from '@/lib/query-client';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { UI } from '@/lib/constants';

const createNodePoolSchema = z.object({
  cluster_name: z.string().min(1, 'Cluster name is required'),
  name: z.string().min(1, 'Name is required').max(255),
  region: z.string().min(1, 'Region is required'),
  instance_type: z.string().min(1, 'Instance type is required'),
  node_count: z.number().min(1, 'Node count must be at least 1'),
  min_nodes: z.number().optional(),
  max_nodes: z.number().optional(),
  disk_size_gb: z.number().optional(),
  disk_type: z.string().optional(),
  auto_scaling: z.boolean().optional(),
});

export default function NodePoolsPage() {
  const { currentWorkspace } = useWorkspaceStore();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { success } = useToast();
  const { handleError } = useErrorHandler();
  const { isLoading: authLoading } = useRequireAuth();

  // Get credential context from global store
  const { selectedCredentialId, selectedRegion } = useCredentialContext();

  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [selectedClusterName, setSelectedClusterName] = useState<string>('');
  const [selectedNodePoolIds, setSelectedNodePoolIds] = useState<string[]>([]);
  const [pageSize, setPageSize] = useState(UI.PAGINATION.DEFAULT_PAGE_SIZE);

  useSSEMonitoring();

  const nodePoolForm = useForm<CreateNodePoolForm>({
    resolver: zodResolver(createNodePoolSchema),
    defaultValues: {
      cluster_name: '',
      region: '',
      node_count: 1,
      auto_scaling: false,
    },
  });

  // Fetch credentials using unified hook
  const { credentials, selectedCredential, selectedProvider } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
  });

  // Fetch clusters for selection
  const { data: clusters = [] } = useQuery({
    queryKey: queryKeys.clusters.list(selectedProvider, selectedCredentialId, selectedRegion),
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId) return [];
      return kubernetesService.listClusters(selectedProvider, selectedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!currentWorkspace,
  });

  // Fetch Node Pools
  const { data: nodePools = [], isLoading: isLoadingNodePools } = useQuery({
    queryKey: queryKeys.nodePools.list(selectedProvider, selectedCredentialId, selectedRegion, selectedClusterName),
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId || !selectedClusterName || !selectedRegion) return [];
      return kubernetesService.listNodePools(selectedProvider, selectedClusterName, selectedCredentialId, selectedRegion || '');
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!selectedClusterName && !!selectedRegion && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: 30000,
  });

  // Search functionality
  const [searchQuery, setSearchQuery] = useState('');

  // Filtered node pools (memoized for consistency)
  const filteredNodePools = useMemo(() => {
    if (!searchQuery.trim()) return nodePools;
    return DataProcessor.search(nodePools, searchQuery, {
      keys: ['name', 'cluster_name', 'status'],
      threshold: 0.3,
    });
  }, [nodePools, searchQuery]);

  const clearSearch = () => {
    setSearchQuery('');
  };

  // Pagination
  const {
    page,
    paginatedItems: paginatedNodePools,
    setPage,
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
    successMessage: 'Node pool creation initiated',
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
    successMessage: 'Node pool deletion initiated',
    errorContext: { operation: 'deleteNodePool', resource: 'NodePool' },
  });

  const handleCreateNodePool = (data: CreateNodePoolForm) => {
    createNodePoolMutation.mutate(data);
  };

  const handleBulkDeleteNodePools = async (nodePoolIds: string[]) => {
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
  };

  const handleDeleteNodePool = (nodePoolName: string, region: string) => {
    if (!selectedCredentialId || !selectedProvider || !selectedClusterName) return;
    if (confirm(`Are you sure you want to delete node pool ${nodePoolName}? This action cannot be undone.`)) {
      deleteNodePoolMutation.mutate({
        nodePoolName,
        credentialId: selectedCredentialId,
        region,
      });
    }
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
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Node Pools</h1>
        <p className="text-gray-600 mt-1">
          Manage Kubernetes Node Pools{currentWorkspace ? ` for ${currentWorkspace.name}` : ''}
        </p>
      </div>
      <div className="flex items-center space-x-2">
        {/* Credential selection is now handled in Header */}
      </div>
    </div>
  );

  // Empty state
  const emptyState = credentials.length === 0 ? (
    <CredentialRequiredState serviceName="Kubernetes (Node Pools)" />
  ) : !selectedProvider || !selectedCredentialId ? (
    <CredentialRequiredState
      title="Select a Credential"
      description="Please select a credential to view node pools. If you don't have any credentials, register one first."
      serviceName="Kubernetes (Node Pools)"
    />
  ) : !selectedClusterName || !selectedRegion ? (
    <Card>
      <CardContent className="flex flex-col items-center justify-center py-12">
        <Layers className="h-12 w-12 text-gray-400 mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">
          Select Cluster and Region
        </h3>
        <p className="text-sm text-gray-500 text-center">
          Please select a cluster and region to view node pools
        </p>
      </CardContent>
    </Card>
  ) : filteredNodePools.length === 0 ? (
    <ResourceEmptyState
      resourceName="Node Pools"
      icon={Layers}
      onCreateClick={() => setIsCreateDialogOpen(true)}
      description="No node pools found for the selected cluster. Create your first node pool."
      withCard={true}
    />
  ) : null;

  // Content component
  const content = selectedProvider && selectedCredentialId && selectedClusterName && selectedRegion && filteredNodePools.length > 0 ? (
    <>
      <BulkActionsToolbar
        items={paginatedNodePools}
        selectedIds={selectedNodePoolIds}
        onSelectionChange={setSelectedNodePoolIds}
        onBulkDelete={handleBulkDeleteNodePools}
        getItemDisplayName={(nodePool) => nodePool.name}
      />
      
      <Card>
        <CardHeader>
          <CardTitle>Node Pools</CardTitle>
          <CardDescription>
            {filteredNodePools.length} of {nodePools.length} node pool{nodePools.length !== 1 ? 's' : ''} found
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="mb-4">
            <SearchBar
              value={searchQuery}
              onChange={setSearchQuery}
              onClear={clearSearch}
              placeholder="Search node pools by name, cluster, or status..."
            />
          </div>
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
                        <Badge variant="outline">Enabled</Badge>
                      ) : (
                        <Badge variant="secondary">Disabled</Badge>
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
          
          {filteredNodePools.length > 0 && (
            <div className="border-t mt-4">
              <Pagination
                total={filteredNodePools.length}
                page={page}
                pageSize={pageSize}
                onPageChange={setPage}
                onPageSizeChange={handlePageSizeChange}
                pageSizeOptions={UI.PAGINATION.PAGE_SIZE_OPTIONS}
                showPageSizeSelector={true}
              />
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create Node Pool Dialog */}
      <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
        <DialogTrigger asChild>
          <Button disabled={credentials.length === 0 || !selectedClusterName}>
            <Plus className="mr-2 h-4 w-4" />
            Create Node Pool
          </Button>
        </DialogTrigger>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Create Node Pool</DialogTitle>
            <DialogDescription>
              Create a new node pool on {selectedProvider?.toUpperCase()}
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={nodePoolForm.handleSubmit(handleCreateNodePool)} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="nodepool-name">Name *</Label>
              <Input id="nodepool-name" {...nodePoolForm.register('name')} placeholder="my-node-pool" />
              {nodePoolForm.formState.errors.name && (
                <p className="text-sm text-red-600">{nodePoolForm.formState.errors.name.message}</p>
              )}
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="nodepool-instance-type">Instance Type *</Label>
                <Input id="nodepool-instance-type" {...nodePoolForm.register('instance_type')} placeholder="n1-standard-2" />
                {nodePoolForm.formState.errors.instance_type && (
                  <p className="text-sm text-red-600">{nodePoolForm.formState.errors.instance_type.message}</p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="nodepool-node-count">Node Count *</Label>
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
                Cancel
              </Button>
              <Button type="submit" disabled={createNodePoolMutation.isPending}>
                {createNodePoolMutation.isPending ? 'Creating...' : 'Create Node Pool'}
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </>
  ) : emptyState;

  if (authLoading) {
    return (
      <WorkspaceRequired>
        <Layout>
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
              <p className="mt-2 text-gray-600">Loading...</p>
            </div>
          </div>
        </Layout>
      </WorkspaceRequired>
    );
  }

  return (
    <WorkspaceRequired>
      <Layout>
        <div className="space-y-6">
          {header}
          
          {/* Configuration */}
          {selectedProvider && selectedCredentialId && (
            <Card>
              <CardHeader>
                <CardTitle>Configuration</CardTitle>
                <CardDescription>Select cluster, region, and credential to view node pools</CardDescription>
              </CardHeader>
              <CardContent className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>Cluster *</Label>
                  <Select
                    value={selectedClusterName}
                    onValueChange={(value) => {
                      setSelectedClusterName(value);
                      nodePoolForm.setValue('cluster_name', value);
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select Cluster" />
                    </SelectTrigger>
                    <SelectContent>
                      {clusters.map((cluster) => (
                        <SelectItem key={cluster.name} value={cluster.name}>
                          {cluster.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>Region *</Label>
                  <Input
                    placeholder="e.g., ap-northeast-2"
                    value={selectedRegion || ''}
                    readOnly
                    className="bg-muted"
                  />
                  <p className="text-xs text-muted-foreground">
                    Region selection is now handled in Header
                  </p>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Content */}
          {isLoadingNodePools ? (
            <Card>
              <CardContent className="pt-6">
                <TableSkeleton columns={7} rows={5} showCheckbox={true} />
              </CardContent>
            </Card>
          ) : (
            content
          )}
        </div>
      </Layout>
    </WorkspaceRequired>
  );
}

