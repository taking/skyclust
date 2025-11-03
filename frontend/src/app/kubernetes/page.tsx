'use client';

import { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Layout } from '@/components/layout/layout';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useRouter } from 'next/navigation';
import * as z from 'zod';
import { kubernetesService } from '@/services/kubernetes';
import { credentialService } from '@/services/credential';
import { useWorkspaceStore } from '@/store/workspace';
import { useProviderStore } from '@/store/provider';
import { WorkspaceRequired } from '@/components/common/workspace-required';
import { Plus, Trash2, ExternalLink, Download, Settings, Server, Search, Filter } from 'lucide-react';
import { CreateClusterForm, CloudProvider } from '@/lib/types';
import { useToast } from '@/hooks/useToast';
import { useRequireAuth } from '@/hooks/useAuth';
import { useSearch } from '@/hooks/useSearch';
import { useSSEMonitoring } from '@/hooks/useSSEMonitoring';
import { SearchBar } from '@/components/ui/search-bar';
import { FilterPanel, FilterConfig, FilterValue } from '@/components/ui/filter-panel';
import { BulkActionsToolbar } from '@/components/common/bulk-actions-toolbar';
import { Checkbox } from '@/components/ui/checkbox';
import { TagFilter } from '@/components/common/tag-filter';
import { BulkOperationProgress } from '@/components/common/bulk-operation-progress';
import { Pagination } from '@/components/ui/pagination';
import { usePagination } from '@/hooks/usePagination';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { VirtualizedTable } from '@/components/common/virtualized-table';

const createClusterSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID'),
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters'),
  version: z.string().min(1, 'Version is required'),
  region: z.string().min(1, 'Region is required'),
  zone: z.string().optional(),
  subnet_ids: z.array(z.string()).min(1, 'At least one subnet is required'),
  vpc_id: z.string().optional(),
  role_arn: z.string().optional(),
  tags: z.record(z.string(), z.string()).optional(),
});

export default function KubernetesPage() {
  const { currentWorkspace } = useWorkspaceStore();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { success, error: showError } = useToast();
  const { isLoading: authLoading } = useRequireAuth();
  
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [selectedRegion, setSelectedRegion] = useState<string>('');
  const [filters, setFilters] = useState<FilterValue>({});
  const [showFilters, setShowFilters] = useState(false);
  const [selectedClusterIds, setSelectedClusterIds] = useState<string[]>([]);
  const [isTagDialogOpen, setIsTagDialogOpen] = useState(false);
  const [bulkTagKey, setBulkTagKey] = useState('');
  const [bulkTagValue, setBulkTagValue] = useState('');
  const [tagFilters, setTagFilters] = useState<Record<string, string[]>>({});
  const [bulkOperationProgress, setBulkOperationProgress] = useState<{
    operation: 'delete' | 'tag';
    total: number;
    completed: number;
    failed: number;
    cancelled: number;
    isComplete: boolean;
    isCancelled: boolean;
  } | null>(null);
  const [isOperationCancelled, setIsOperationCancelled] = useState(false);
  
  // Selected credential (provider)
  const [selectedCredentialId, setSelectedCredentialId] = useState<string>('');
  
  // Pagination
  const [pageSize, setPageSize] = useState(20);

  // SSE 실시간 업데이트
  useSSEMonitoring();

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    watch,
    setValue,
  } = useForm<CreateClusterForm>({
    resolver: zodResolver(createClusterSchema),
  });

  // Use selectedCredentialId for consistency
  // const watchedCredentialId = watch('credential_id');
  const watchedCredentialId = selectedCredentialId; // Use selectedCredentialId instead

  // Fetch credentials for selected workspace
  const { data: credentials = [] } = useQuery({
    queryKey: ['credentials', currentWorkspace?.id],
    queryFn: () => currentWorkspace ? credentialService.getCredentials(currentWorkspace.id) : Promise.resolve([]),
    enabled: !!currentWorkspace,
  });

  // Filter credentials by selected credential (no additional filtering needed)
  const filteredCredentials = credentials;
  
  // Get selected credential and provider (after credentials is loaded)
  const selectedCredential = credentials.find(c => c.id === selectedCredentialId);
  const selectedProvider = selectedCredential?.provider as CloudProvider | undefined;

  // Fetch clusters
  const { data: clusters = [], isLoading } = useQuery({
    queryKey: ['kubernetes-clusters', selectedProvider, selectedCredentialId, selectedRegion],
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId || !currentWorkspace) {
        return [];
      }
      return kubernetesService.listClusters(selectedProvider, selectedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!currentWorkspace,
    refetchInterval: 30000, // Poll every 30 seconds
  });

  // Search functionality
  const {
    query: searchQuery,
    setQuery: setSearchQuery,
    results: searchResults,
    isSearching,
    clearSearch,
  } = useSearch(clusters, {
    keys: ['name', 'version', 'status', 'region'],
    threshold: 0.3,
  });

  // Filter configurations
  const filterConfigs: FilterConfig[] = [
    {
      key: 'status',
      label: 'Status',
      type: 'select',
      options: [
        { value: 'ACTIVE', label: 'Active' },
        { value: 'CREATING', label: 'Creating' },
        { value: 'UPDATING', label: 'Updating' },
        { value: 'DELETING', label: 'Deleting' },
        { value: 'FAILED', label: 'Failed' },
      ],
    },
    {
      key: 'region',
      label: 'Region',
      type: 'select',
      options: Array.from(new Set(clusters.map(c => c.region))).map(r => ({ value: r, label: r })),
    },
  ];

  // Extract available tags from clusters
  const availableTags = useMemo(() => {
    const tagMap: Record<string, Set<string>> = {};
    clusters.forEach(cluster => {
      if (cluster.tags) {
        Object.entries(cluster.tags).forEach(([key, value]) => {
          if (!tagMap[key]) {
            tagMap[key] = new Set();
          }
          tagMap[key].add(value);
        });
      }
    });
    
    const result: Record<string, string[]> = {};
    Object.entries(tagMap).forEach(([key, valueSet]) => {
      result[key] = Array.from(valueSet).sort();
    });
    return result;
  }, [clusters]);

  // Apply filters including tag filters
  const filteredClusters = useMemo(() => {
    let result = searchResults.filter((cluster) => {
      if (filters.status && cluster.status !== filters.status) return false;
      if (filters.region && cluster.region !== filters.region) return false;
      
      // Apply tag filters
      if (Object.keys(tagFilters).length > 0 && cluster.tags) {
        for (const [tagKey, tagValues] of Object.entries(tagFilters)) {
          if (tagValues && tagValues.length > 0) {
            const clusterTagValue = cluster.tags[tagKey];
            if (!clusterTagValue || !tagValues.includes(clusterTagValue)) {
              return false;
            }
          }
        }
      }
      
      return true;
    });
    
    return result;
  }, [searchResults, filters, tagFilters]);

  // Apply pagination to filtered clusters
  const {
    page,
    totalPages,
    paginatedItems: paginatedClusters,
    setPage,
    setPageSize: setPaginationPageSize,
  } = usePagination(filteredClusters, {
    totalItems: filteredClusters.length,
    initialPage: 1,
    initialPageSize: pageSize,
  });

  // Auto-enable virtual scrolling when page has 50+ items
  const shouldUseVirtualScrolling = filteredClusters.length >= 50;

  // Create cluster mutation
  const createClusterMutation = useMutation({
    mutationFn: (data: CreateClusterForm) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return kubernetesService.createCluster(selectedProvider, data);
    },
    onSuccess: () => {
      success('Cluster creation initiated');
      queryClient.invalidateQueries({ queryKey: ['kubernetes-clusters'] });
      setIsCreateDialogOpen(false);
      reset();
    },
    onError: (error: Error) => {
      showError(`Failed to create cluster: ${error.message}`);
    },
  });

  // Delete cluster mutation
  const deleteClusterMutation = useMutation({
    mutationFn: async ({ clusterName, credentialId, region }: { clusterName: string; credentialId: string; region: string }) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return kubernetesService.deleteCluster(selectedProvider, clusterName, credentialId, region);
    },
    onSuccess: () => {
      success('Cluster deletion initiated');
      queryClient.invalidateQueries({ queryKey: ['kubernetes-clusters'] });
    },
    onError: (error: Error) => {
      showError(`Failed to delete cluster: ${error.message}`);
    },
  });

  // Bulk delete clusters with progress tracking
  const handleBulkDelete = async (clusterIds: string[]) => {
    if (!watchedCredentialId || !selectedProvider) return;
    
    const clustersToDelete = filteredClusters.filter(c => clusterIds.includes(c.id || c.name));
    const total = clustersToDelete.length;
    let completed = 0;
    let failed = 0;

    // Initialize progress tracking
    setBulkOperationProgress({
      operation: 'delete',
      total,
      completed: 0,
      failed: 0,
      cancelled: 0,
      isComplete: false,
      isCancelled: false,
    });

    let cancelled = 0;
    const deletePromises = clustersToDelete.map(async (cluster) => {
      // Check if operation was cancelled
      if (isOperationCancelled) {
        cancelled++;
        setBulkOperationProgress(prev => prev ? {
          ...prev,
          cancelled,
          isComplete: completed + failed + cancelled === total,
        } : null);
        return;
      }
      try {
        await deleteClusterMutation.mutateAsync({
          clusterName: cluster.name,
          credentialId: watchedCredentialId,
          region: cluster.region,
        });
        completed++;
        setBulkOperationProgress(prev => prev ? {
          ...prev,
          completed,
          isComplete: completed + failed === total,
        } : null);
      } catch (error) {
        failed++;
        setBulkOperationProgress(prev => prev ? {
          ...prev,
          failed,
          isComplete: completed + failed === total,
        } : null);
      }
    });

    try {
      await Promise.allSettled(deletePromises);
      
      // Mark as complete (or cancelled)
      setBulkOperationProgress(prev => prev ? {
        ...prev,
        isComplete: true,
        isCancelled: isOperationCancelled,
      } : null);
      
      if (isOperationCancelled) {
        showError(`Operation cancelled: ${completed} completed, ${failed} failed, ${cancelled} cancelled`);
      } else if (failed === 0) {
        success(`Successfully initiated deletion of ${completed} cluster(s)`);
      } else if (completed > 0) {
        success(`Initiated deletion of ${completed} cluster(s), ${failed} failed`);
      } else {
        showError(`Failed to delete all clusters`);
      }
      
      setSelectedClusterIds([]);
      queryClient.invalidateQueries({ queryKey: ['kubernetes-clusters'] });
      
      // Clear progress after 5 seconds (unless cancelled)
      if (!isOperationCancelled) {
        setTimeout(() => {
          setBulkOperationProgress(null);
        }, 5000);
      }
    } catch (error) {
      showError(`Failed to delete some clusters: ${error instanceof Error ? error.message : 'Unknown error'}`);
      setBulkOperationProgress(null);
    }
  };

  // Handle cancellation
  const handleCancelOperation = () => {
    setIsOperationCancelled(true);
    setBulkOperationProgress(prev => prev ? {
      ...prev,
      isCancelled: true,
    } : null);
  };

  // Bulk tag clusters
  const handleBulkTag = (clusterIds: string[]) => {
    setIsTagDialogOpen(true);
  };

  const handleBulkTagSubmit = async () => {
    if (!bulkTagKey.trim() || !bulkTagValue.trim() || !selectedProvider || !watchedCredentialId) return;
    
    setIsOperationCancelled(false);
    const clustersToTag = filteredClusters.filter(c => selectedClusterIds.includes(c.id || c.name));
    const total = clustersToTag.length;
    let completed = 0;
    let failed = 0;
    let cancelled = 0;

    // Initialize progress tracking
    setBulkOperationProgress({
      operation: 'tag',
      total,
      completed: 0,
      failed: 0,
      cancelled: 0,
      isComplete: false,
      isCancelled: false,
    });

    const tagPromises = clustersToTag.map(async (cluster) => {
      // Check if operation was cancelled
      if (isOperationCancelled) {
        cancelled++;
        setBulkOperationProgress(prev => prev ? {
          ...prev,
          cancelled,
          isComplete: completed + failed + cancelled === total,
        } : null);
        return;
      }
      try {
        const currentTags = cluster.tags || {};
        const updatedTags = {
          ...currentTags,
          [bulkTagKey.trim()]: bulkTagValue.trim(),
        };
        
        await kubernetesService.updateClusterTags(
          selectedProvider,
          cluster.name,
          watchedCredentialId,
          cluster.region,
          updatedTags
        );
        
        completed++;
        setBulkOperationProgress(prev => prev ? {
          ...prev,
          completed,
          isComplete: completed + failed + (prev.cancelled || 0) === total,
        } : null);
      } catch (error) {
        failed++;
        setBulkOperationProgress(prev => prev ? {
          ...prev,
          failed,
          isComplete: completed + failed + (prev.cancelled || 0) === total,
        } : null);
      }
    });

    try {
      await Promise.allSettled(tagPromises);
      
      // Mark as complete (or cancelled)
      setBulkOperationProgress(prev => prev ? {
        ...prev,
        isComplete: true,
        isCancelled: isOperationCancelled,
      } : null);
      
      if (isOperationCancelled) {
        showError(`Operation cancelled: ${completed} completed, ${failed} failed, ${cancelled} cancelled`);
      } else if (failed === 0) {
        success(`Successfully added tag "${bulkTagKey}: ${bulkTagValue}" to ${completed} cluster(s)`);
      } else if (completed > 0) {
        success(`Added tag to ${completed} cluster(s), ${failed} failed`);
      } else {
        showError(`Failed to add tag to all clusters`);
      }
      
      setIsTagDialogOpen(false);
      setBulkTagKey('');
      setBulkTagValue('');
      setSelectedClusterIds([]);
      queryClient.invalidateQueries({ queryKey: ['kubernetes-clusters'] });
      
      // Clear progress after 5 seconds (unless cancelled)
      if (!isOperationCancelled) {
        setTimeout(() => {
          setBulkOperationProgress(null);
        }, 5000);
      }
    } catch (error) {
      showError(`Failed to add tags: ${error instanceof Error ? error.message : 'Unknown error'}`);
      setBulkOperationProgress(null);
    }
  };

  // Download kubeconfig mutation
  const downloadKubeconfigMutation = useMutation({
    mutationFn: async ({ clusterName, credentialId, region }: { clusterName: string; credentialId: string; region: string }) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return kubernetesService.getKubeconfig(selectedProvider, clusterName, credentialId, region);
    },
    onSuccess: (kubeconfig, variables) => {
      // Create download link
      const blob = new Blob([kubeconfig], { type: 'application/yaml' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `kubeconfig-${variables.clusterName}.yaml`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      success('Kubeconfig downloaded');
    },
    onError: (error: Error) => {
      showError(`Failed to download kubeconfig: ${error.message}`);
    },
  });

  const handleCreateCluster = (data: CreateClusterForm) => {
    createClusterMutation.mutate(data);
  };

  const handleDeleteCluster = (clusterName: string, region: string) => {
    if (!watchedCredentialId) return;
    if (confirm(`Are you sure you want to delete cluster ${clusterName}? This action cannot be undone.`)) {
      deleteClusterMutation.mutate({
        clusterName,
        credentialId: watchedCredentialId,
        region,
      });
    }
  };

  const handleDownloadKubeconfig = (clusterName: string, region: string) => {
    if (!watchedCredentialId) return;
    downloadKubeconfigMutation.mutate({
      clusterName,
      credentialId: watchedCredentialId,
      region,
    });
  };

  if (authLoading || isLoading) {
    return (
      <WorkspaceRequired>
        <Layout>
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h1 className="text-3xl font-bold text-gray-900">Kubernetes Clusters</h1>
            </div>
            <Card>
              <CardContent className="pt-6">
                <TableSkeleton columns={6} rows={5} showCheckbox={true} />
              </CardContent>
            </Card>
          </div>
        </Layout>
      </WorkspaceRequired>
    );
  }

  return (
    <WorkspaceRequired>
      <Layout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Kubernetes Clusters</h1>
            <p className="text-gray-600">
              Manage Kubernetes clusters for {currentWorkspace.name}
            </p>
          </div>
          <div className="flex items-center space-x-2">
            <Select
              value={selectedCredentialId}
              onValueChange={(value) => {
                setSelectedCredentialId(value);
                setValue('credential_id', value);
              }}
            >
              <SelectTrigger className="w-[250px]">
                <SelectValue placeholder="Select Credential" />
              </SelectTrigger>
              <SelectContent>
                {credentials.map((credential) => (
                  <SelectItem key={credential.id} value={credential.id}>
                    {credential.name || `${credential.provider.toUpperCase()} (${credential.id.slice(0, 8)})`}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
              <DialogTrigger asChild>
                <Button disabled={!selectedCredentialId || credentials.length === 0}>
                  <Plus className="mr-2 h-4 w-4" />
                  Create Cluster
                </Button>
              </DialogTrigger>
              <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                  <DialogTitle>Create Kubernetes Cluster</DialogTitle>
                  <DialogDescription>
                    Create a new Kubernetes cluster on {selectedProvider?.toUpperCase() || 'your cloud provider'}
                  </DialogDescription>
                </DialogHeader>
                <form onSubmit={handleSubmit(handleCreateCluster)} className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="credential_id">Credential *</Label>
                    <Select
                      value={selectedCredentialId || ''}
                      onValueChange={(value) => {
                        setSelectedCredentialId(value);
                        setValue('credential_id', value);
                      }}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Select credential" />
                      </SelectTrigger>
                      <SelectContent>
                        {filteredCredentials.map((cred) => (
                          <SelectItem key={cred.id} value={cred.id}>
                            {cred.provider} - {cred.id.substring(0, 8)}...
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    {errors.credential_id && (
                      <p className="text-sm text-red-600">{errors.credential_id.message}</p>
                    )}
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="name">Cluster Name *</Label>
                      <Input
                        id="name"
                        {...register('name')}
                        placeholder="my-cluster"
                      />
                      {errors.name && (
                        <p className="text-sm text-red-600">{errors.name.message}</p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="version">Kubernetes Version *</Label>
                      <Input
                        id="version"
                        {...register('version')}
                        placeholder="1.28"
                      />
                      {errors.version && (
                        <p className="text-sm text-red-600">{errors.version.message}</p>
                      )}
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="region">Region *</Label>
                      <Input
                        id="region"
                        {...register('region')}
                        placeholder="ap-northeast-2"
                        onChange={(e) => {
                          setValue('region', e.target.value);
                          setSelectedRegion(e.target.value);
                        }}
                      />
                      {errors.region && (
                        <p className="text-sm text-red-600">{errors.region.message}</p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="zone">Zone (Optional)</Label>
                      <Input
                        id="zone"
                        {...register('zone')}
                        placeholder="ap-northeast-2a"
                      />
                    </div>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="subnet_ids">Subnet IDs *</Label>
                    <Input
                      id="subnet_ids"
                      placeholder="subnet-12345,subnet-67890"
                      onChange={(e) => {
                        const subnets = e.target.value.split(',').map(s => s.trim()).filter(Boolean);
                        setValue('subnet_ids', subnets);
                      }}
                    />
                    <p className="text-sm text-gray-500">Comma-separated list of subnet IDs</p>
                    {errors.subnet_ids && (
                      <p className="text-sm text-red-600">{errors.subnet_ids.message}</p>
                    )}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="vpc_id">VPC ID (Optional)</Label>
                    <Input
                      id="vpc_id"
                      {...register('vpc_id')}
                      placeholder="vpc-12345"
                    />
                  </div>

                  <div className="flex justify-end space-x-2">
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => setIsCreateDialogOpen(false)}
                    >
                      Cancel
                    </Button>
                    <Button
                      type="submit"
                      disabled={createClusterMutation.isPending}
                    >
                      {createClusterMutation.isPending ? 'Creating...' : 'Create Cluster'}
                    </Button>
                  </div>
                </form>
              </DialogContent>
            </Dialog>
          </div>
        </div>

        {/* Search and Filter */}
        {selectedCredentialId && watchedCredentialId && clusters.length > 0 && (
          <Card>
            <CardContent className="pt-6">
              <div className="flex flex-col md:flex-row gap-4">
                <div className="flex-1">
                  <SearchBar
                    value={searchQuery}
                    onChange={setSearchQuery}
                    onClear={clearSearch}
                    placeholder="Search clusters by name, version, status, or region..."
                  />
                </div>
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
                        <Badge variant="secondary" className="ml-2">
                          {Object.keys(filters).length}
                        </Badge>
                      )}
                    </Button>
              </div>
              {showFilters && (
                <div className="mt-4">
                  <FilterPanel
                    configs={filterConfigs}
                    values={filters}
                    onChange={setFilters}
                    onClear={() => setFilters({})}
                  />
                </div>
              )}
            </CardContent>
          </Card>
        )}

        {/* Clusters Table */}
        {!selectedProvider ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Server className="h-12 w-12 text-gray-400 mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">Select a Provider</h3>
              <p className="text-sm text-gray-500 text-center">
                Please select a cloud provider to view Kubernetes clusters
              </p>
            </CardContent>
          </Card>
        ) : !watchedCredentialId ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Server className="h-12 w-12 text-gray-400 mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">Select a Credential</h3>
              <p className="text-sm text-gray-500 text-center">
                Please select a credential to view Kubernetes clusters
              </p>
            </CardContent>
          </Card>
        ) : isLoading ? (
          <Card>
            <CardContent className="pt-6">
              <TableSkeleton columns={6} rows={5} showCheckbox={true} />
            </CardContent>
          </Card>
        ) : filteredClusters.length === 0 ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Server className="h-12 w-12 text-gray-400 mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">No Clusters Found</h3>
              <p className="text-sm text-gray-500 text-center mb-4">
                No Kubernetes clusters found. Create your first cluster to get started.
              </p>
              <Button onClick={() => setIsCreateDialogOpen(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Create Cluster
              </Button>
            </CardContent>
          </Card>
        ) : (
          <>
            {bulkOperationProgress && (
              <BulkOperationProgress
                {...bulkOperationProgress}
                onDismiss={() => {
                  setBulkOperationProgress(null);
                  setIsOperationCancelled(false);
                }}
                onCancel={handleCancelOperation}
              />
            )}
            
            <BulkActionsToolbar
              items={filteredClusters}
              selectedIds={selectedClusterIds}
              onSelectionChange={setSelectedClusterIds}
              onBulkDelete={handleBulkDelete}
              onBulkTag={handleBulkTag}
              getItemDisplayName={(cluster) => cluster.name}
            />
            
            <Card>
              <CardHeader>
                <CardTitle>Clusters</CardTitle>
                <CardDescription>
                  {filteredClusters.length} of {clusters.length} cluster{clusters.length !== 1 ? 's' : ''} 
                  {isSearching && ` (${searchQuery})`}
                </CardDescription>
              </CardHeader>
              <CardContent>
                {shouldUseVirtualScrolling ? (
                  <VirtualizedTable
                    data={paginatedClusters}
                    minItems={0} // Already filtered to paginated items
                    containerHeight="600px"
                    estimateSize={60}
                    renderHeader={() => (
                      <TableRow>
                        <TableHead className="w-12">
                          <Checkbox
                            checked={selectedClusterIds.length === filteredClusters.length && filteredClusters.length > 0}
                            onCheckedChange={(checked) => {
                              if (checked) {
                                setSelectedClusterIds(filteredClusters.map(c => c.id || c.name));
                              } else {
                                setSelectedClusterIds([]);
                              }
                            }}
                          />
                        </TableHead>
                        <TableHead>Name</TableHead>
                        <TableHead>Version</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Region</TableHead>
                        <TableHead>Endpoint</TableHead>
                        <TableHead>Actions</TableHead>
                      </TableRow>
                    )}
                    renderRow={(cluster, index) => {
                      const clusterId = cluster.id || cluster.name;
                      const isSelected = selectedClusterIds.includes(clusterId);
                      
                      return (
                        <>
                          <TableCell>
                            <Checkbox
                              checked={isSelected}
                              onCheckedChange={(checked) => {
                                if (checked) {
                                  setSelectedClusterIds([...selectedClusterIds, clusterId]);
                                } else {
                                  setSelectedClusterIds(selectedClusterIds.filter(id => id !== clusterId));
                                }
                              }}
                            />
                          </TableCell>
                          <TableCell className="font-medium">
                            <Button
                              variant="link"
                              className="p-0 h-auto font-medium"
                              onClick={() => router.push(`/kubernetes/${cluster.name}`)}
                            >
                              {cluster.name}
                            </Button>
                          </TableCell>
                          <TableCell>{cluster.version}</TableCell>
                          <TableCell>
                            <Badge
                              variant={
                                cluster.status === 'ACTIVE'
                                  ? 'default'
                                  : cluster.status === 'CREATING'
                                  ? 'secondary'
                                  : 'destructive'
                              }
                            >
                              {cluster.status}
                            </Badge>
                          </TableCell>
                          <TableCell>{cluster.region}</TableCell>
                          <TableCell>
                            {cluster.endpoint ? (
                              <a
                                href={cluster.endpoint}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-blue-600 hover:underline flex items-center"
                              >
                                {cluster.endpoint.substring(0, 30)}...
                                <ExternalLink className="ml-1 h-3 w-3" />
                              </a>
                            ) : (
                              <span className="text-gray-400">-</span>
                            )}
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center space-x-2">
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleDownloadKubeconfig(cluster.name, cluster.region)}
                                disabled={downloadKubeconfigMutation.isPending}
                              >
                                <Download className="h-4 w-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => router.push(`/kubernetes/${cluster.name}`)}
                              >
                                <Settings className="h-4 w-4" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleDeleteCluster(cluster.name, cluster.region)}
                                disabled={deleteClusterMutation.isPending}
                              >
                                <Trash2 className="h-4 w-4 text-red-600" />
                              </Button>
                            </div>
                          </TableCell>
                        </>
                      );
                    }}
                  />
                ) : (
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-12">
                          <Checkbox
                            checked={selectedClusterIds.length === filteredClusters.length && filteredClusters.length > 0}
                            onCheckedChange={(checked) => {
                              if (checked) {
                                setSelectedClusterIds(filteredClusters.map(c => c.id || c.name));
                              } else {
                                setSelectedClusterIds([]);
                              }
                            }}
                          />
                        </TableHead>
                        <TableHead>Name</TableHead>
                        <TableHead>Version</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Region</TableHead>
                        <TableHead>Endpoint</TableHead>
                        <TableHead>Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {paginatedClusters.map((cluster) => {
                        const clusterId = cluster.id || cluster.name;
                        const isSelected = selectedClusterIds.includes(clusterId);
                        
                        return (
                          <TableRow key={clusterId}>
                            <TableCell>
                              <Checkbox
                                checked={isSelected}
                                onCheckedChange={(checked) => {
                                  if (checked) {
                                    setSelectedClusterIds([...selectedClusterIds, clusterId]);
                                  } else {
                                    setSelectedClusterIds(selectedClusterIds.filter(id => id !== clusterId));
                                  }
                                }}
                              />
                            </TableCell>
                            <TableCell className="font-medium">
                              <Button
                                variant="link"
                                className="p-0 h-auto font-medium"
                                onClick={() => router.push(`/kubernetes/${cluster.name}`)}
                              >
                                {cluster.name}
                              </Button>
                            </TableCell>
                            <TableCell>{cluster.version}</TableCell>
                            <TableCell>
                              <Badge
                                variant={
                                  cluster.status === 'ACTIVE'
                                    ? 'default'
                                    : cluster.status === 'CREATING'
                                    ? 'secondary'
                                    : 'destructive'
                                }
                              >
                                {cluster.status}
                              </Badge>
                            </TableCell>
                            <TableCell>{cluster.region}</TableCell>
                            <TableCell>
                              {cluster.endpoint ? (
                                <a
                                  href={cluster.endpoint}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  className="text-blue-600 hover:underline flex items-center"
                                >
                                  {cluster.endpoint.substring(0, 30)}...
                                  <ExternalLink className="ml-1 h-3 w-3" />
                                </a>
                              ) : (
                                <span className="text-gray-400">-</span>
                              )}
                            </TableCell>
                            <TableCell>
                              <div className="flex items-center space-x-2">
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => handleDownloadKubeconfig(cluster.name, cluster.region)}
                                  disabled={downloadKubeconfigMutation.isPending}
                                >
                                  <Download className="h-4 w-4" />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => router.push(`/kubernetes/${cluster.name}`)}
                                >
                                  <Settings className="h-4 w-4" />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => handleDeleteCluster(cluster.name, cluster.region)}
                                  disabled={deleteClusterMutation.isPending}
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
                )}
                
                {/* Pagination */}
                {filteredClusters.length > 0 && (
                  <div className="border-t">
                    <Pagination
                      total={filteredClusters.length}
                      page={page}
                      pageSize={pageSize}
                      onPageChange={setPage}
                      onPageSizeChange={(newSize) => {
                        setPageSize(newSize);
                        setPaginationPageSize(newSize);
                      }}
                      pageSizeOptions={[10, 20, 50, 100]}
                      showPageSizeSelector={true}
                    />
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Bulk Tag Dialog */}
            <Dialog open={isTagDialogOpen} onOpenChange={setIsTagDialogOpen}>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Add Tags to Selected Clusters</DialogTitle>
                  <DialogDescription>
                    Add the same tag to {selectedClusterIds.length} selected cluster(s)
                  </DialogDescription>
                </DialogHeader>
                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="bulk-tag-key">Tag Key *</Label>
                    <Input
                      id="bulk-tag-key"
                      value={bulkTagKey}
                      onChange={(e) => setBulkTagKey(e.target.value)}
                      placeholder="e.g., Environment"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="bulk-tag-value">Tag Value *</Label>
                    <Input
                      id="bulk-tag-value"
                      value={bulkTagValue}
                      onChange={(e) => setBulkTagValue(e.target.value)}
                      placeholder="e.g., Production"
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' && bulkTagKey && bulkTagValue) {
                          handleBulkTagSubmit();
                        }
                      }}
                    />
                  </div>
                  <div className="flex justify-end space-x-2">
                    <Button variant="outline" onClick={() => setIsTagDialogOpen(false)}>
                      Cancel
                    </Button>
                    <Button onClick={handleBulkTagSubmit} disabled={!bulkTagKey.trim() || !bulkTagValue.trim()}>
                      Add Tag
                    </Button>
                  </div>
                </div>
              </DialogContent>
            </Dialog>
          </>
        )}
      </div>
    </Layout>
    </WorkspaceRequired>
  );
}

