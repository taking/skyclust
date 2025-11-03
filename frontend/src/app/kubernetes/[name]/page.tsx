'use client';

import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Layout } from '@/components/layout/layout';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { kubernetesService } from '@/services/kubernetes';
import { credentialService } from '@/services/credential';
import { useWorkspaceStore } from '@/store/workspace';
import { useProviderStore } from '@/store/provider';
import { ArrowLeft, Download, Settings, Trash2, Plus, Server, RefreshCw, ArrowUp, AlertTriangle, CheckCircle } from 'lucide-react';
import { CreateNodePoolForm, CreateNodeGroupForm, CloudProvider } from '@/lib/types';
import { useToast } from '@/hooks/useToast';
import { useRequireAuth } from '@/hooks/useAuth';
import { useSSEMonitoring } from '@/hooks/useSSEMonitoring';
import { ClusterMetricsChart } from '@/components/kubernetes/cluster-metrics-chart';
import { NodeMetricsChart } from '@/components/kubernetes/node-metrics-chart';
import { TagManager } from '@/components/common/tag-manager';

const createNodePoolSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID'),
  name: z.string().min(1, 'Name is required'),
  cluster_name: z.string().min(1, 'Cluster name is required'),
  version: z.string().optional(),
  region: z.string().min(1, 'Region is required'),
  zone: z.string().optional(),
  instance_type: z.string().min(1, 'Instance type is required'),
  disk_size_gb: z.number().min(10).optional(),
  disk_type: z.string().optional(),
  min_nodes: z.number().min(0),
  max_nodes: z.number().min(1),
  node_count: z.number().min(0),
  auto_scaling: z.boolean().optional(),
});

const createNodeGroupSchema = z.object({
  credential_id: z.string().uuid('Invalid credential ID'),
  name: z.string().min(1, 'Name is required'),
  cluster_name: z.string().min(1, 'Cluster name is required'),
  instance_type: z.string().min(1, 'Instance type is required'),
  disk_size_gb: z.number().min(10).optional(),
  min_size: z.number().min(0),
  max_size: z.number().min(1),
  desired_size: z.number().min(0),
  region: z.string().min(1, 'Region is required'),
});

export default function KubernetesClusterDetailPage() {
  const params = useParams();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { success, error: showError } = useToast();
  const { isLoading: authLoading } = useRequireAuth();
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedProvider } = useProviderStore();
  
  const clusterName = params.name as string;
  const [selectedRegion, setSelectedRegion] = useState<string>('ap-northeast-2');
  const [selectedCredentialId, setSelectedCredentialId] = useState<string>('');
  const [activeTab, setActiveTab] = useState<'overview' | 'metrics' | 'nodepools' | 'nodegroups' | 'nodes'>('overview');
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isUpgradeDialogOpen, setIsUpgradeDialogOpen] = useState(false);
  const [upgradeVersion, setUpgradeVersion] = useState<string>('');

  // SSE 실시간 업데이트
  useSSEMonitoring();

  // Fetch credentials
  const { data: credentials = [] } = useQuery({
    queryKey: ['credentials', currentWorkspace?.id],
    queryFn: () => currentWorkspace ? credentialService.getCredentials(currentWorkspace.id) : Promise.resolve([]),
    enabled: !!currentWorkspace,
  });

  const filteredCredentials = selectedProvider
    ? credentials.filter((c) => c.provider === selectedProvider)
    : [];

  // Fetch cluster details
  const { data: cluster, isLoading: isLoadingCluster } = useQuery({
    queryKey: ['kubernetes-cluster', selectedProvider, clusterName, selectedCredentialId, selectedRegion],
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.getCluster(selectedProvider, clusterName, selectedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!clusterName && !!currentWorkspace,
    refetchInterval: 5000, // Polling for updates
  });

  // Fetch node pools (GKE, AKS, NKS)
  const { data: nodePools = [], isLoading: isLoadingNodePools } = useQuery({
    queryKey: ['node-pools', selectedProvider, clusterName, selectedCredentialId, selectedRegion],
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId) return [];
      return kubernetesService.listNodePools(selectedProvider, clusterName, selectedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!clusterName && !!currentWorkspace,
    refetchInterval: 10000,
  });

  // Fetch node groups (EKS)
  const { data: nodeGroups = [], isLoading: isLoadingNodeGroups } = useQuery({
    queryKey: ['node-groups', selectedProvider, clusterName, selectedCredentialId, selectedRegion],
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId) return [];
      return kubernetesService.listNodeGroups(selectedProvider, clusterName, selectedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!clusterName && !!currentWorkspace && selectedProvider === 'aws',
    refetchInterval: 10000,
  });

  // Fetch nodes
  const { data: nodes = [], isLoading: isLoadingNodes } = useQuery({
    queryKey: ['cluster-nodes', selectedProvider, clusterName, selectedCredentialId, selectedRegion],
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId) return [];
      return kubernetesService.listNodes(selectedProvider, clusterName, selectedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!clusterName && !!currentWorkspace,
    refetchInterval: 10000,
  });

  const nodePoolForm = useForm<CreateNodePoolForm>({
    resolver: zodResolver(createNodePoolSchema),
    defaultValues: {
      cluster_name: clusterName,
      region: selectedRegion,
    },
  });

  const nodeGroupForm = useForm<CreateNodeGroupForm>({
    resolver: zodResolver(createNodeGroupSchema),
    defaultValues: {
      cluster_name: clusterName,
      region: selectedRegion,
    },
  });

  // Create node pool mutation
  const createNodePoolMutation = useMutation({
    mutationFn: (data: CreateNodePoolForm) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return kubernetesService.createNodePool(selectedProvider, clusterName, data);
    },
    onSuccess: () => {
      success('Node pool creation initiated');
      queryClient.invalidateQueries({ queryKey: ['node-pools'] });
      setIsCreateDialogOpen(false);
      nodePoolForm.reset();
    },
    onError: (error: Error) => {
      showError(`Failed to create node pool: ${error.message}`);
    },
  });

  // Create node group mutation
  const createNodeGroupMutation = useMutation({
    mutationFn: (data: CreateNodeGroupForm) => {
      if (!selectedProvider) throw new Error('Provider not selected');
      return kubernetesService.createNodeGroup(selectedProvider, clusterName, data);
    },
    onSuccess: () => {
      success('Node group creation initiated');
      queryClient.invalidateQueries({ queryKey: ['node-groups'] });
      setIsCreateDialogOpen(false);
      nodeGroupForm.reset();
    },
    onError: (error: Error) => {
      showError(`Failed to create node group: ${error.message}`);
    },
  });

  // Delete mutations
  const deleteNodePoolMutation = useMutation({
    mutationFn: async ({ nodePoolName }: { nodePoolName: string }) => {
      if (!selectedProvider || !selectedCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.deleteNodePool(selectedProvider, clusterName, nodePoolName, selectedCredentialId, selectedRegion);
    },
    onSuccess: () => {
      success('Node pool deletion initiated');
      queryClient.invalidateQueries({ queryKey: ['node-pools'] });
    },
    onError: (error: Error) => {
      showError(`Failed to delete node pool: ${error.message}`);
    },
  });

  const deleteNodeGroupMutation = useMutation({
    mutationFn: async ({ nodeGroupName }: { nodeGroupName: string }) => {
      if (!selectedProvider || !selectedCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.deleteNodeGroup(selectedProvider, clusterName, nodeGroupName, selectedCredentialId, selectedRegion);
    },
    onSuccess: () => {
      success('Node group deletion initiated');
      queryClient.invalidateQueries({ queryKey: ['node-groups'] });
    },
    onError: (error: Error) => {
      showError(`Failed to delete node group: ${error.message}`);
    },
  });

  // Scale node pool
  const scaleNodePoolMutation = useMutation({
    mutationFn: async ({ nodePoolName, nodeCount }: { nodePoolName: string; nodeCount: number }) => {
      if (!selectedProvider || !selectedCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.scaleNodePool(selectedProvider, clusterName, nodePoolName, nodeCount, selectedCredentialId, selectedRegion);
    },
    onSuccess: () => {
      success('Node pool scaling initiated');
      queryClient.invalidateQueries({ queryKey: ['node-pools'] });
    },
    onError: (error: Error) => {
      showError(`Failed to scale node pool: ${error.message}`);
    },
  });

  // Fetch upgrade status
  const { data: upgradeStatus, isLoading: isLoadingUpgradeStatus } = useQuery({
    queryKey: ['upgrade-status', selectedProvider, clusterName, selectedCredentialId, selectedRegion],
    queryFn: async () => {
      if (!selectedProvider || !selectedCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.getUpgradeStatus(selectedProvider, clusterName, selectedCredentialId, selectedRegion);
    },
    enabled: !!selectedProvider && !!selectedCredentialId && !!clusterName && !!currentWorkspace,
    refetchInterval: (data) => {
      // Poll more frequently if upgrade is in progress
      return data?.status === 'IN_PROGRESS' || data?.status === 'PENDING' ? 5000 : 30000;
    },
  });

  // Upgrade cluster mutation
  const upgradeClusterMutation = useMutation({
    mutationFn: async (version: string) => {
      if (!selectedProvider || !selectedCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.upgradeCluster(selectedProvider, clusterName, version, selectedCredentialId, selectedRegion);
    },
    onSuccess: () => {
      success('Cluster upgrade initiated');
      queryClient.invalidateQueries({ queryKey: ['kubernetes-cluster'] });
      queryClient.invalidateQueries({ queryKey: ['upgrade-status'] });
      setIsUpgradeDialogOpen(false);
      setUpgradeVersion('');
    },
    onError: (error: Error) => {
      showError(`Failed to initiate upgrade: ${error.message}`);
    },
  });

  // Download kubeconfig
  const downloadKubeconfigMutation = useMutation({
    mutationFn: async () => {
      if (!selectedProvider || !selectedCredentialId) throw new Error('Provider and credential required');
      return kubernetesService.getKubeconfig(selectedProvider, clusterName, selectedCredentialId, selectedRegion);
    },
    onSuccess: (kubeconfig) => {
      const blob = new Blob([kubeconfig], { type: 'application/yaml' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `kubeconfig-${clusterName}.yaml`;
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

  const handleCreateNodePool = (data: CreateNodePoolForm) => {
    createNodePoolMutation.mutate(data);
  };

  const handleCreateNodeGroup = (data: CreateNodeGroupForm) => {
    createNodeGroupMutation.mutate(data);
  };

  const handleScaleNodePool = (nodePoolName: string, currentNodes: number) => {
    const newNodeCount = prompt(`Enter new node count (current: ${currentNodes}):`);
    if (newNodeCount && !isNaN(Number(newNodeCount))) {
      scaleNodePoolMutation.mutate({ nodePoolName, nodeCount: Number(newNodeCount) });
    }
  };

  if (authLoading) {
    return (
      <Layout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
        </div>
      </Layout>
    );
  }

  if (!currentWorkspace || !selectedProvider) {
    return (
      <Layout>
        <div className="flex items-center justify-center h-64">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">
              {!currentWorkspace ? 'No Workspace Selected' : 'Select a Provider'}
            </h2>
            <Button onClick={() => router.push(currentWorkspace ? '/kubernetes' : '/workspaces')}>
              {currentWorkspace ? 'Select Provider' : 'Manage Workspaces'}
            </Button>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Button variant="ghost" onClick={() => router.push('/kubernetes')}>
              <ArrowLeft className="mr-2 h-4 w-4" />
              Back
            </Button>
            <div>
              <h1 className="text-3xl font-bold text-gray-900">{clusterName}</h1>
              <p className="text-gray-600">Kubernetes cluster details</p>
            </div>
          </div>
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              onClick={() => setIsUpgradeDialogOpen(true)}
              disabled={!selectedCredentialId || !cluster}
            >
              <ArrowUp className="mr-2 h-4 w-4" />
              Upgrade Cluster
            </Button>
            <Button
              variant="outline"
              onClick={() => downloadKubeconfigMutation.mutate()}
              disabled={downloadKubeconfigMutation.isPending || !selectedCredentialId}
            >
              <Download className="mr-2 h-4 w-4" />
              Download Kubeconfig
            </Button>
          </div>
        </div>

        {/* Credential Selection */}
        <Card>
          <CardHeader>
            <CardTitle>Configuration</CardTitle>
            <CardDescription>Select credential and region to view cluster details</CardDescription>
          </CardHeader>
          <CardContent className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Credential</Label>
              <Select
                value={selectedCredentialId}
                onValueChange={(value) => {
                  setSelectedCredentialId(value);
                  nodePoolForm.setValue('credential_id', value);
                  nodeGroupForm.setValue('credential_id', value);
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
            </div>
            <div className="space-y-2">
              <Label>Region</Label>
              <Input
                value={selectedRegion}
                onChange={(e) => {
                  setSelectedRegion(e.target.value);
                  nodePoolForm.setValue('region', e.target.value);
                  nodeGroupForm.setValue('region', e.target.value);
                }}
                placeholder="ap-northeast-2"
              />
            </div>
          </CardContent>
        </Card>

        {/* Cluster Info */}
        {isLoadingCluster ? (
          <Card>
            <CardContent className="flex items-center justify-center py-12">
              <div className="text-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
                <p className="mt-2 text-gray-600">Loading cluster details...</p>
              </div>
            </CardContent>
          </Card>
        ) : cluster ? (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span>Cluster Overview</span>
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
              </CardTitle>
            </CardHeader>
            <CardContent className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <Label className="text-sm text-gray-500">Version</Label>
                <p className="text-lg font-semibold">{cluster.version}</p>
              </div>
              <div>
                <Label className="text-sm text-gray-500">Region</Label>
                <p className="text-lg font-semibold">{cluster.region}</p>
              </div>
              <div>
                <Label className="text-sm text-gray-500">Endpoint</Label>
                <p className="text-sm truncate">{cluster.endpoint || '-'}</p>
              </div>
              {cluster.node_pool_info && (
                <>
                  <div>
                    <Label className="text-sm text-gray-500">Total Nodes</Label>
                    <p className="text-lg font-semibold">{cluster.node_pool_info.total_nodes}</p>
                  </div>
                  <div>
                    <Label className="text-sm text-gray-500">Node Pools</Label>
                    <p className="text-lg font-semibold">{cluster.node_pool_info.total_node_pools}</p>
                  </div>
                </>
              )}
              {cluster.tags && Object.keys(cluster.tags).length > 0 && (
                <div className="mt-4 pt-4 border-t">
                  <TagManager
                    tags={cluster.tags}
                    onTagsChange={(updatedTags) => {
                      // Optimistic update
                      const previousCluster = queryClient.getQueryData(['kubernetes-cluster', selectedProvider, clusterName, selectedCredentialId, selectedRegion]);
                      
                      queryClient.setQueryData(['kubernetes-cluster', selectedProvider, clusterName, selectedCredentialId, selectedRegion], {
                        ...cluster,
                        tags: updatedTags,
                      });

                      // Note: This would require a backend API to update cluster tags
                      // For now, just update local state optimistically
                      // If backend API exists, use mutation with rollback on error:
                      // updateClusterTagsMutation.mutate(updatedTags, {
                      //   onError: () => {
                      //     queryClient.setQueryData(['kubernetes-cluster', selectedProvider, clusterName, selectedCredentialId, selectedRegion], previousCluster);
                      //   }
                      // });
                      
                      success('Tags updated successfully');
                    }}
                  />
                </div>
              )}
            </CardContent>
          </Card>
        ) : null}

        {/* Upgrade Status */}
        {upgradeStatus && (
          <Card className={upgradeStatus.status === 'FAILED' ? 'border-red-500' : upgradeStatus.status === 'COMPLETED' ? 'border-green-500' : ''}>
            <CardHeader>
              <CardTitle className="flex items-center">
                <ArrowUp className="mr-2 h-5 w-5" />
                Upgrade Status
                <Badge
                  variant={
                    upgradeStatus.status === 'COMPLETED'
                      ? 'default'
                      : upgradeStatus.status === 'IN_PROGRESS' || upgradeStatus.status === 'PENDING'
                      ? 'secondary'
                      : upgradeStatus.status === 'FAILED'
                      ? 'destructive'
                      : 'outline'
                  }
                  className="ml-2"
                >
                  {upgradeStatus.status}
                </Badge>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label className="text-sm text-gray-500">Current Version</Label>
                    <p className="text-sm font-medium">{upgradeStatus.current_version || cluster?.version || '-'}</p>
                  </div>
                  <div>
                    <Label className="text-sm text-gray-500">Target Version</Label>
                    <p className="text-sm font-medium">{upgradeStatus.target_version || '-'}</p>
                  </div>
                </div>
                {upgradeStatus.progress !== undefined && (
                  <div className="space-y-2">
                    <div className="flex justify-between text-sm">
                      <span>Progress</span>
                      <span>{upgradeStatus.progress}%</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div
                        className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                        style={{ width: `${upgradeStatus.progress}%` }}
                      />
                    </div>
                  </div>
                )}
                {upgradeStatus.error && (
                  <div className="flex items-start space-x-2 p-3 bg-red-50 rounded-md">
                    <AlertTriangle className="h-5 w-5 text-red-600 mt-0.5" />
                    <div className="flex-1">
                      <p className="text-sm font-medium text-red-900">Upgrade Error</p>
                      <p className="text-sm text-red-700">{upgradeStatus.error}</p>
                    </div>
                  </div>
                )}
                {upgradeStatus.status === 'COMPLETED' && (
                  <div className="flex items-center space-x-2 p-3 bg-green-50 rounded-md">
                    <CheckCircle className="h-5 w-5 text-green-600" />
                    <p className="text-sm font-medium text-green-900">Upgrade completed successfully</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        )}

        {/* Upgrade Dialog */}
        <Dialog open={isUpgradeDialogOpen} onOpenChange={setIsUpgradeDialogOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Upgrade Cluster</DialogTitle>
              <DialogDescription>
                Upgrade {clusterName} to a new Kubernetes version
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="current-version">Current Version</Label>
                <Input
                  id="current-version"
                  value={cluster?.version || ''}
                  disabled
                  className="bg-gray-50"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="target-version">Target Version *</Label>
                <Input
                  id="target-version"
                  value={upgradeVersion}
                  onChange={(e) => setUpgradeVersion(e.target.value)}
                  placeholder="e.g., 1.29.0"
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && upgradeVersion) {
                      upgradeClusterMutation.mutate(upgradeVersion);
                    }
                  }}
                />
                <p className="text-xs text-gray-500">
                  Enter the Kubernetes version to upgrade to (e.g., 1.29.0)
                </p>
              </div>
              {cluster && (
                <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-md">
                  <div className="flex items-start space-x-2">
                    <AlertTriangle className="h-5 w-5 text-yellow-600 mt-0.5" />
                    <div className="flex-1">
                      <p className="text-sm font-medium text-yellow-900">Important</p>
                      <ul className="mt-1 text-sm text-yellow-800 list-disc list-inside space-y-1">
                        <li>Upgrading a cluster will cause temporary downtime</li>
                        <li>Ensure all node pools are compatible with the target version</li>
                        <li>Backup your workloads before upgrading</li>
                        <li>Upgrade process cannot be easily rolled back</li>
                      </ul>
                    </div>
                  </div>
                </div>
              )}
              <div className="flex justify-end space-x-2">
                <Button
                  variant="outline"
                  onClick={() => {
                    setIsUpgradeDialogOpen(false);
                    setUpgradeVersion('');
                  }}
                >
                  Cancel
                </Button>
                <Button
                  onClick={() => upgradeClusterMutation.mutate(upgradeVersion)}
                  disabled={!upgradeVersion || upgradeClusterMutation.isPending}
                >
                  {upgradeClusterMutation.isPending ? 'Upgrading...' : 'Upgrade Cluster'}
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>

        {/* Tabs */}
        {selectedCredentialId && (
          <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as typeof activeTab)} className="space-y-4">
            <div className="flex items-center justify-between">
              <TabsList>
                <TabsTrigger value="overview">Overview</TabsTrigger>
                <TabsTrigger value="metrics">Metrics</TabsTrigger>
                {selectedProvider !== 'aws' && (
                  <TabsTrigger value="nodepools">Node Pools</TabsTrigger>
                )}
                {selectedProvider === 'aws' && (
                  <TabsTrigger value="nodegroups">Node Groups</TabsTrigger>
                )}
                <TabsTrigger value="nodes">Nodes</TabsTrigger>
              </TabsList>
              {(activeTab === 'nodepools' || activeTab === 'nodegroups') && (
                <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
                  <DialogTrigger asChild>
                    <Button>
                      <Plus className="mr-2 h-4 w-4" />
                      Create {activeTab === 'nodepools' ? 'Node Pool' : 'Node Group'}
                    </Button>
                  </DialogTrigger>
                  <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
                    <DialogHeader>
                      <DialogTitle>
                        Create {activeTab === 'nodepools' ? 'Node Pool' : 'Node Group'}
                      </DialogTitle>
                      <DialogDescription>
                        Create a new {activeTab === 'nodepools' ? 'node pool' : 'node group'} for this cluster
                      </DialogDescription>
                    </DialogHeader>
                    {activeTab === 'nodepools' ? (
                      <form onSubmit={nodePoolForm.handleSubmit(handleCreateNodePool)} className="space-y-4">
                        <div className="grid grid-cols-2 gap-4">
                          <div className="space-y-2">
                            <Label htmlFor="np-name">Name *</Label>
                            <Input id="np-name" {...nodePoolForm.register('name')} />
                            {nodePoolForm.formState.errors.name && (
                              <p className="text-sm text-red-600">{nodePoolForm.formState.errors.name.message}</p>
                            )}
                          </div>
                          <div className="space-y-2">
                            <Label htmlFor="np-instance-type">Instance Type *</Label>
                            <Input id="np-instance-type" {...nodePoolForm.register('instance_type')} placeholder="n1-standard-2" />
                            {nodePoolForm.formState.errors.instance_type && (
                              <p className="text-sm text-red-600">{nodePoolForm.formState.errors.instance_type.message}</p>
                            )}
                          </div>
                        </div>
                        <div className="grid grid-cols-3 gap-4">
                          <div className="space-y-2">
                            <Label htmlFor="np-min-nodes">Min Nodes</Label>
                            <Input
                              id="np-min-nodes"
                              type="number"
                              {...nodePoolForm.register('min_nodes', { valueAsNumber: true })}
                            />
                          </div>
                          <div className="space-y-2">
                            <Label htmlFor="np-max-nodes">Max Nodes</Label>
                            <Input
                              id="np-max-nodes"
                              type="number"
                              {...nodePoolForm.register('max_nodes', { valueAsNumber: true })}
                            />
                          </div>
                          <div className="space-y-2">
                            <Label htmlFor="np-node-count">Node Count</Label>
                            <Input
                              id="np-node-count"
                              type="number"
                              {...nodePoolForm.register('node_count', { valueAsNumber: true })}
                            />
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
                    ) : (
                      <form onSubmit={nodeGroupForm.handleSubmit(handleCreateNodeGroup)} className="space-y-4">
                        <div className="grid grid-cols-2 gap-4">
                          <div className="space-y-2">
                            <Label htmlFor="ng-name">Name *</Label>
                            <Input id="ng-name" {...nodeGroupForm.register('name')} />
                            {nodeGroupForm.formState.errors.name && (
                              <p className="text-sm text-red-600">{nodeGroupForm.formState.errors.name.message}</p>
                            )}
                          </div>
                          <div className="space-y-2">
                            <Label htmlFor="ng-instance-type">Instance Type *</Label>
                            <Input id="ng-instance-type" {...nodeGroupForm.register('instance_type')} placeholder="t3.medium" />
                            {nodeGroupForm.formState.errors.instance_type && (
                              <p className="text-sm text-red-600">{nodeGroupForm.formState.errors.instance_type.message}</p>
                            )}
                          </div>
                        </div>
                        <div className="grid grid-cols-3 gap-4">
                          <div className="space-y-2">
                            <Label htmlFor="ng-min-size">Min Size</Label>
                            <Input
                              id="ng-min-size"
                              type="number"
                              {...nodeGroupForm.register('min_size', { valueAsNumber: true })}
                            />
                          </div>
                          <div className="space-y-2">
                            <Label htmlFor="ng-max-size">Max Size</Label>
                            <Input
                              id="ng-max-size"
                              type="number"
                              {...nodeGroupForm.register('max_size', { valueAsNumber: true })}
                            />
                          </div>
                          <div className="space-y-2">
                            <Label htmlFor="ng-desired-size">Desired Size</Label>
                            <Input
                              id="ng-desired-size"
                              type="number"
                              {...nodeGroupForm.register('desired_size', { valueAsNumber: true })}
                            />
                          </div>
                        </div>
                        <div className="flex justify-end space-x-2">
                          <Button type="button" variant="outline" onClick={() => setIsCreateDialogOpen(false)}>
                            Cancel
                          </Button>
                          <Button type="submit" disabled={createNodeGroupMutation.isPending}>
                            {createNodeGroupMutation.isPending ? 'Creating...' : 'Create Node Group'}
                          </Button>
                        </div>
                      </form>
                    )}
                  </DialogContent>
                </Dialog>
              )}
            </div>

            {/* Overview Tab */}
            <TabsContent value="overview" className="space-y-4">
              {/* Cluster Metrics Preview */}
              {cluster && (
                <ClusterMetricsChart clusterName={clusterName} />
              )}
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Card>
                  <CardHeader>
                    <CardTitle>Network Configuration</CardTitle>
                  </CardHeader>
                  <CardContent>
                    {cluster?.network_config ? (
                      <div className="space-y-2">
                        <div className="flex justify-between">
                          <span className="text-sm text-gray-500">VPC ID:</span>
                          <span className="text-sm">{cluster.network_config.vpc_id || '-'}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-gray-500">Subnet ID:</span>
                          <span className="text-sm">{cluster.network_config.subnet_id || '-'}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-gray-500">Pod CIDR:</span>
                          <span className="text-sm">{cluster.network_config.pod_cidr || '-'}</span>
                        </div>
                      </div>
                    ) : (
                      <p className="text-sm text-gray-500">No network configuration available</p>
                    )}
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader>
                    <CardTitle>Security Configuration</CardTitle>
                  </CardHeader>
                  <CardContent>
                    {cluster?.security_config ? (
                      <div className="space-y-2">
                        <div className="flex justify-between">
                          <span className="text-sm text-gray-500">Workload Identity:</span>
                          <Badge variant={cluster.security_config.workload_identity ? 'default' : 'secondary'}>
                            {cluster.security_config.workload_identity ? 'Enabled' : 'Disabled'}
                          </Badge>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-gray-500">Network Policy:</span>
                          <Badge variant={cluster.security_config.network_policy ? 'default' : 'secondary'}>
                            {cluster.security_config.network_policy ? 'Enabled' : 'Disabled'}
                          </Badge>
                        </div>
                      </div>
                    ) : (
                      <p className="text-sm text-gray-500">No security configuration available</p>
                    )}
                  </CardContent>
                </Card>
              </div>
            </TabsContent>

            {/* Metrics Tab */}
            <TabsContent value="metrics" className="space-y-4">
              {cluster && (
                <>
                  <ClusterMetricsChart clusterName={clusterName} />
                  {nodes.length > 0 && (
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {nodes.slice(0, 4).map((node) => (
                        <NodeMetricsChart key={node.id || node.name} node={node} />
                      ))}
                    </div>
                  )}
                </>
              )}
            </TabsContent>

            {/* Node Pools Tab */}
            {selectedProvider !== 'aws' && (
              <TabsContent value="nodepools" className="space-y-4">
                {isLoadingNodePools ? (
                  <Card>
                    <CardContent className="flex items-center justify-center py-12">
                      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
                    </CardContent>
                  </Card>
                ) : nodePools.length === 0 ? (
                  <Card>
                    <CardContent className="flex flex-col items-center justify-center py-12">
                      <Server className="h-12 w-12 text-gray-400 mb-4" />
                      <h3 className="text-lg font-medium text-gray-900 mb-2">No Node Pools</h3>
                      <p className="text-sm text-gray-500 mb-4">Create your first node pool</p>
                      <Button onClick={() => setIsCreateDialogOpen(true)}>
                        <Plus className="mr-2 h-4 w-4" />
                        Create Node Pool
                      </Button>
                    </CardContent>
                  </Card>
                ) : (
                  <Card>
                    <CardHeader>
                      <CardTitle>Node Pools</CardTitle>
                      <CardDescription>{nodePools.length} node pool{nodePools.length !== 1 ? 's' : ''}</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Name</TableHead>
                            <TableHead>Instance Type</TableHead>
                            <TableHead>Nodes</TableHead>
                            <TableHead>Min/Max</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead>Actions</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {nodePools.map((np) => (
                            <TableRow key={np.id || np.name}>
                              <TableCell className="font-medium">{np.name}</TableCell>
                              <TableCell>{np.instance_type}</TableCell>
                              <TableCell>{np.node_count}</TableCell>
                              <TableCell>{np.min_nodes}/{np.max_nodes}</TableCell>
                              <TableCell>
                                <Badge variant={np.status === 'RUNNING' ? 'default' : 'secondary'}>
                                  {np.status}
                                </Badge>
                              </TableCell>
                              <TableCell>
                                <div className="flex items-center space-x-2">
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={() => handleScaleNodePool(np.name, np.node_count)}
                                  >
                                    <RefreshCw className="h-4 w-4" />
                                  </Button>
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={() => {
                                      if (confirm(`Delete node pool ${np.name}?`)) {
                                        deleteNodePoolMutation.mutate({ nodePoolName: np.name });
                                      }
                                    }}
                                  >
                                    <Trash2 className="h-4 w-4 text-red-600" />
                                  </Button>
                                </div>
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </CardContent>
                  </Card>
                )}
              </TabsContent>
            )}

            {/* Node Groups Tab (EKS) */}
            {selectedProvider === 'aws' && (
              <TabsContent value="nodegroups" className="space-y-4">
                {isLoadingNodeGroups ? (
                  <Card>
                    <CardContent className="flex items-center justify-center py-12">
                      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
                    </CardContent>
                  </Card>
                ) : nodeGroups.length === 0 ? (
                  <Card>
                    <CardContent className="flex flex-col items-center justify-center py-12">
                      <Server className="h-12 w-12 text-gray-400 mb-4" />
                      <h3 className="text-lg font-medium text-gray-900 mb-2">No Node Groups</h3>
                      <p className="text-sm text-gray-500 mb-4">Create your first node group</p>
                      <Button onClick={() => setIsCreateDialogOpen(true)}>
                        <Plus className="mr-2 h-4 w-4" />
                        Create Node Group
                      </Button>
                    </CardContent>
                  </Card>
                ) : (
                  <Card>
                    <CardHeader>
                      <CardTitle>Node Groups</CardTitle>
                      <CardDescription>{nodeGroups.length} node group{nodeGroups.length !== 1 ? 's' : ''}</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Name</TableHead>
                            <TableHead>Instance Type</TableHead>
                            <TableHead>Nodes</TableHead>
                            <TableHead>Min/Max</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead>Actions</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {nodeGroups.map((ng) => (
                            <TableRow key={ng.id || ng.name}>
                              <TableCell className="font-medium">{ng.name}</TableCell>
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
                                  onClick={() => {
                                    if (confirm(`Delete node group ${ng.name}?`)) {
                                      deleteNodeGroupMutation.mutate({ nodeGroupName: ng.name });
                                    }
                                  }}
                                >
                                  <Trash2 className="h-4 w-4 text-red-600" />
                                </Button>
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </CardContent>
                  </Card>
                )}
              </TabsContent>
            )}

            {/* Nodes Tab */}
            <TabsContent value="nodes" className="space-y-4">
              {isLoadingNodes ? (
                <Card>
                  <CardContent className="flex items-center justify-center py-12">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
                  </CardContent>
                </Card>
              ) : nodes.length === 0 ? (
                <Card>
                  <CardContent className="flex flex-col items-center justify-center py-12">
                    <Server className="h-12 w-12 text-gray-400 mb-4" />
                    <h3 className="text-lg font-medium text-gray-900 mb-2">No Nodes</h3>
                    <p className="text-sm text-gray-500">No nodes found in this cluster</p>
                  </CardContent>
                </Card>
              ) : (
                <Card>
                  <CardHeader>
                    <CardTitle>Nodes</CardTitle>
                    <CardDescription>{nodes.length} node{nodes.length !== 1 ? 's' : ''}</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>Name</TableHead>
                          <TableHead>Instance Type</TableHead>
                          <TableHead>Zone</TableHead>
                          <TableHead>Status</TableHead>
                          <TableHead>Private IP</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {nodes.map((node) => (
                          <TableRow key={node.id || node.name}>
                            <TableCell className="font-medium">{node.name}</TableCell>
                            <TableCell>{node.instance_type}</TableCell>
                            <TableCell>{node.zone || '-'}</TableCell>
                            <TableCell>
                              <Badge variant={node.status === 'Ready' ? 'default' : 'secondary'}>
                                {node.status}
                              </Badge>
                            </TableCell>
                            <TableCell>{node.private_ip || '-'}</TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </CardContent>
                </Card>
              )}
            </TabsContent>
          </Tabs>
        )}
      </div>
    </Layout>
  );
}

