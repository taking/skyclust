/**
 * Kubernetes Cluster Detail Page (Refactored)
 * 클러스터 상세 페이지 - 리팩토링된 버전
 */

'use client';

import { useState, Suspense, useEffect } from 'react';
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
  ClusterUpgradeStatusCard,
  ClusterOverviewHeader,
  ClusterDetailOverviewTab,
  ClusterDetailResourcesTab,
  ClusterDetailComputingTab,
  ClusterDetailNetworkingTab,
  ClusterDetailAccessTab,
  ClusterDetailTagsTab,
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

function KubernetesClusterDetailPageContent() {
  const params = useParams();
  const router = useRouter();
  const { isLoading: authLoading } = useRequireAuth();
  useSSEMonitoring();

  const clusterName = params.name as string;
  const [activeTab, setActiveTab] = useState<
    'overview' | 'resources' | 'computing' | 'networking' | 'access' | 'tags' | 'metrics'
  >('overview');
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


  // 클러스터가 삭제된 경우 목록 페이지로 리다이렉트
  useEffect(() => {
    if (!isLoadingCluster && !cluster && clusterName && currentWorkspace && selectedProvider) {
      // 클러스터가 로드되지 않았고, 로딩이 완료된 경우 (삭제되었거나 존재하지 않음)
      router.push('/kubernetes/clusters');
    }
  }, [isLoadingCluster, cluster, clusterName, currentWorkspace, selectedProvider, router]);

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

        <ClusterOverviewHeader
          cluster={cluster}
          selectedProvider={selectedProvider}
          isLoading={isLoadingCluster}
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

        {selectedCredentialId && cluster && (
          <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as typeof activeTab)} className="space-y-4">
            <div className="flex items-center justify-between">
              <TabsList>
                <TabsTrigger value="overview">개요</TabsTrigger>
                <TabsTrigger value="resources">리소스</TabsTrigger>
                <TabsTrigger value="computing">컴퓨팅</TabsTrigger>
                <TabsTrigger value="networking">네트워킹</TabsTrigger>
                <TabsTrigger value="access">액세스</TabsTrigger>
                <TabsTrigger value="tags">태그</TabsTrigger>
                <TabsTrigger value="metrics">Metrics</TabsTrigger>
              </TabsList>
            </div>

            <TabsContent value="overview" className="space-y-4">
              <ClusterDetailOverviewTab cluster={cluster as any} />
            </TabsContent>

            <TabsContent value="resources" className="space-y-4">
              <ClusterDetailResourcesTab />
            </TabsContent>

            <TabsContent value="computing" className="space-y-4">
              <ClusterDetailComputingTab
                nodePools={nodePools}
                nodeGroups={nodeGroups}
                nodes={nodes}
                isLoadingNodePools={isLoadingNodePools}
                isLoadingNodeGroups={isLoadingNodeGroups}
                isLoadingNodes={isLoadingNodes}
                selectedProvider={selectedProvider}
                onCreateNodePoolClick={() => setIsCreateDialogOpen(true)}
                onCreateNodeGroupClick={() => setIsCreateDialogOpen(true)}
                onScaleNodePoolClick={handleScaleNodePool}
                onDeleteNodePoolClick={(name) => deleteNodePoolMutation.mutate({ nodePoolName: name })}
                onDeleteNodeGroupClick={(name) => deleteNodeGroupMutation.mutate({ nodeGroupName: name })}
                isDeletingNodePool={deleteNodePoolMutation.isPending}
                isDeletingNodeGroup={deleteNodeGroupMutation.isPending}
              />
            </TabsContent>

            <TabsContent value="networking" className="space-y-4">
              <ClusterDetailNetworkingTab
                cluster={cluster as any}
                selectedProvider={selectedProvider as 'aws' | 'gcp' | 'azure' | undefined}
              />
            </TabsContent>

            <TabsContent value="access" className="space-y-4">
              <ClusterDetailAccessTab />
            </TabsContent>

            <TabsContent value="tags" className="space-y-4">
              <ClusterDetailTagsTab
                cluster={cluster}
                selectedProvider={selectedProvider}
                clusterName={clusterName}
                selectedCredentialId={selectedCredentialId}
                selectedRegion={selectedRegion}
              />
            </TabsContent>

            <TabsContent value="metrics" className="space-y-4">
              <ClusterMetricsTab clusterName={clusterName} nodes={nodes} />
            </TabsContent>
          </Tabs>
        )}

        {activeTab === 'computing' && selectedProvider !== 'aws' && (
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

        {activeTab === 'computing' && selectedProvider === 'aws' && (
          <CreateNodeGroupDialog
            open={isCreateDialogOpen}
            onOpenChange={setIsCreateDialogOpen}
            clusterName={clusterName}
            cluster={cluster}
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

export default function KubernetesClusterDetailPage() {
  return (
    <Suspense fallback={
      <Layout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
        </div>
      </Layout>
    }>
      <KubernetesClusterDetailPageContent />
    </Suspense>
  );
}

