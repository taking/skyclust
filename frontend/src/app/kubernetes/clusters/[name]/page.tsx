/**
 * Kubernetes Cluster Detail Page (Refactored)
 * 클러스터 상세 페이지 - 리팩토링된 버전
 */

'use client';

import { useState } from 'react';
import dynamic from 'next/dynamic';
import { useParams, useRouter } from 'next/navigation';
import { Layout } from '@/components/layout/layout';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Breadcrumb } from '@/components/common/breadcrumb';
import { Plus } from 'lucide-react';
import { useRequireAuth } from '@/hooks/use-auth';
import { useSSEMonitoring } from '@/hooks/use-sse-monitoring';
import type { CreateNodePoolForm, CreateNodeGroupForm } from '@/lib/types';
import {
  useClusterDetail,
  ClusterHeader,
  ClusterConfigurationCard,
  ClusterInfoCard,
  ClusterUpgradeStatusCard,
  ClusterOverviewTab,
  ClusterNodePoolsTab,
  ClusterNodeGroupsTab,
  ClusterNodesTab,
} from '@/features/kubernetes';

// Dynamic imports for Dialog components
const UpgradeClusterDialog = dynamic(
  () => import('@/features/kubernetes').then(mod => ({ default: mod.UpgradeClusterDialog })),
  { 
    ssr: false,
    loading: () => null,
  }
);

const CreateNodePoolDialog = dynamic(
  () => import('@/features/kubernetes').then(mod => ({ default: mod.CreateNodePoolDialog })),
  { 
    ssr: false,
    loading: () => null,
  }
);

const CreateNodeGroupDialog = dynamic(
  () => import('@/features/kubernetes').then(mod => ({ default: mod.CreateNodeGroupDialog })),
  { 
    ssr: false,
    loading: () => null,
  }
);

// Dynamic import for ClusterMetricsTab (contains Chart components)
const ClusterMetricsTab = dynamic(
  () => import('@/features/kubernetes').then(mod => ({ default: mod.ClusterMetricsTab })),
  { 
    ssr: false,
    loading: () => (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading metrics...</p>
        </div>
      </div>
    ),
  }
);

export default function KubernetesClusterDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { isLoading: authLoading } = useRequireAuth();
  useSSEMonitoring();

  const clusterName = params.name as string;
  const [activeTab, setActiveTab] = useState<'overview' | 'metrics' | 'nodepools' | 'nodegroups' | 'nodes'>('overview');
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isUpgradeDialogOpen, setIsUpgradeDialogOpen] = useState(false);
  const [upgradeVersion, setUpgradeVersion] = useState<string>('');

  const {
    selectedRegion,
    setSelectedRegion,
    selectedCredentialId,
    setSelectedCredentialId,
    credentials,
    cluster,
    nodePools,
    nodeGroups,
    nodes,
    upgradeStatus,
    isLoadingCluster,
    isLoadingNodePools,
    isLoadingNodeGroups,
    isLoadingNodes,
    createNodePoolMutation,
    createNodeGroupMutation,
    deleteNodePoolMutation,
    deleteNodeGroupMutation,
    scaleNodePoolMutation,
    upgradeClusterMutation,
    downloadKubeconfigMutation,
    selectedProvider,
    currentWorkspace,
  } = useClusterDetail({ clusterName });

  const handleCreateNodePool = (data: CreateNodePoolForm) => {
    createNodePoolMutation.mutate(data, {
      onSuccess: () => {
        setIsCreateDialogOpen(false);
      },
    });
  };

  const handleCreateNodeGroup = (data: CreateNodeGroupForm) => {
    createNodeGroupMutation.mutate(data, {
      onSuccess: () => {
        setIsCreateDialogOpen(false);
      },
    });
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
            <Button onClick={() => router.push(currentWorkspace ? '/kubernetes/clusters' : '/workspaces')}>
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
        <ClusterHeader
          clusterName={clusterName}
          onUpgradeClick={() => setIsUpgradeDialogOpen(true)}
          onDownloadKubeconfigClick={() => downloadKubeconfigMutation.mutate()}
          isUpgradeDisabled={!selectedCredentialId || !cluster}
          isDownloadDisabled={!selectedCredentialId}
          isDownloadPending={downloadKubeconfigMutation.isPending}
        />

        <ClusterConfigurationCard
          credentials={credentials}
          selectedCredentialId={selectedCredentialId}
          onCredentialChange={setSelectedCredentialId}
          selectedRegion={selectedRegion}
          onRegionChange={setSelectedRegion}
        />

        <ClusterInfoCard
          cluster={cluster}
          isLoading={isLoadingCluster}
          selectedProvider={selectedProvider}
          clusterName={clusterName}
          selectedCredentialId={selectedCredentialId}
          selectedRegion={selectedRegion}
        />

        {upgradeStatus && (
          <ClusterUpgradeStatusCard
            upgradeStatus={upgradeStatus}
            currentClusterVersion={cluster?.version}
          />
        )}

        <UpgradeClusterDialog
          open={isUpgradeDialogOpen}
          onOpenChange={setIsUpgradeDialogOpen}
          clusterName={clusterName}
          currentVersion={cluster?.version}
          upgradeVersion={upgradeVersion}
          onUpgradeVersionChange={setUpgradeVersion}
          onUpgrade={(version) => {
            upgradeClusterMutation.mutate(version, {
              onSuccess: () => {
                setIsUpgradeDialogOpen(false);
                setUpgradeVersion('');
              },
            });
          }}
          isPending={upgradeClusterMutation.isPending}
        />

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
                <Button onClick={() => setIsCreateDialogOpen(true)}>
                  <Plus className="mr-2 h-4 w-4" />
                  Create {activeTab === 'nodepools' ? 'Node Pool' : 'Node Group'}
                </Button>
              )}
            </div>

            <TabsContent value="overview" className="space-y-4">
              <ClusterOverviewTab clusterName={clusterName} cluster={cluster} />
            </TabsContent>

            <TabsContent value="metrics" className="space-y-4">
              <ClusterMetricsTab clusterName={clusterName} nodes={nodes} />
            </TabsContent>

            {selectedProvider !== 'aws' && (
              <TabsContent value="nodepools" className="space-y-4">
                <ClusterNodePoolsTab
                  nodePools={nodePools}
                  isLoading={isLoadingNodePools}
                  onCreateClick={() => setIsCreateDialogOpen(true)}
                  onScaleClick={handleScaleNodePool}
                  onDeleteClick={(name) => deleteNodePoolMutation.mutate({ nodePoolName: name })}
                  isDeleting={deleteNodePoolMutation.isPending}
                />
              </TabsContent>
            )}

            {selectedProvider === 'aws' && (
              <TabsContent value="nodegroups" className="space-y-4">
                <ClusterNodeGroupsTab
                  nodeGroups={nodeGroups}
                  isLoading={isLoadingNodeGroups}
                  onCreateClick={() => setIsCreateDialogOpen(true)}
                  onDeleteClick={(name) => deleteNodeGroupMutation.mutate({ nodeGroupName: name })}
                  isDeleting={deleteNodeGroupMutation.isPending}
                />
              </TabsContent>
            )}

            <TabsContent value="nodes" className="space-y-4">
              <ClusterNodesTab nodes={nodes} isLoading={isLoadingNodes} />
            </TabsContent>
          </Tabs>
        )}

        {activeTab === 'nodepools' && selectedProvider !== 'aws' && (
          <CreateNodePoolDialog
            open={isCreateDialogOpen}
            onOpenChange={setIsCreateDialogOpen}
            clusterName={clusterName}
            defaultRegion={selectedRegion}
            defaultCredentialId={selectedCredentialId}
            onSubmit={handleCreateNodePool}
            onCredentialIdChange={setSelectedCredentialId}
            onRegionChange={setSelectedRegion}
            isPending={createNodePoolMutation.isPending}
          />
        )}

        {activeTab === 'nodegroups' && selectedProvider === 'aws' && (
          <CreateNodeGroupDialog
            open={isCreateDialogOpen}
            onOpenChange={setIsCreateDialogOpen}
            clusterName={clusterName}
            defaultRegion={selectedRegion}
            defaultCredentialId={selectedCredentialId}
            onSubmit={handleCreateNodeGroup}
            onCredentialIdChange={setSelectedCredentialId}
            onRegionChange={setSelectedRegion}
            isPending={createNodeGroupMutation.isPending}
          />
        )}
      </div>
    </Layout>
  );
}

