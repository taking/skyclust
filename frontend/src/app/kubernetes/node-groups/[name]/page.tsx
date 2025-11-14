/**
 * Kubernetes Node Group Detail Page
 * 노드 그룹 상세 페이지
 */

'use client';

import { useState, Suspense, useCallback } from 'react';
import { useParams, useSearchParams, useRouter } from 'next/navigation';
import { Layout } from '@/components/layout/layout';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Breadcrumb } from '@/components/common/breadcrumb';
import { ArrowLeft, Edit, Trash2 } from 'lucide-react';
import { useRequireAuth } from '@/hooks/use-auth';
import { useSSEMonitoring } from '@/hooks/use-sse-monitoring';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useCredentials } from '@/hooks/use-credentials';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useWorkspaceStore } from '@/store/workspace';
import { kubernetesService } from '@/features/kubernetes';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import type { AWSNodeGroup, NodeGroup, CreateNodeGroupForm } from '@/lib/types';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useTranslation } from '@/hooks/use-translation';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import dynamic from 'next/dynamic';

const CreateNodeGroupDialog = dynamic(
  () => import('@/features/kubernetes').then(mod => ({ default: mod.CreateNodeGroupDialog })),
  { 
    ssr: false,
    loading: () => null,
  }
);

function NodeGroupDetailPageContent() {
  const params = useParams();
  const searchParams = useSearchParams();
  const router = useRouter();
  const { isLoading: authLoading } = useRequireAuth();
  useSSEMonitoring();
  const { t } = useTranslation();

  const nodeGroupName = params.name as string;
  const clusterName = searchParams?.get('cluster') || '';
  const [activeTab, setActiveTab] = useState<'overview' | 'scaling' | 'networking' | 'tags'>('overview');
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const { success } = useToast();
  const { handleError } = useErrorHandler();
  const queryClient = useQueryClient();

  const { currentWorkspace } = useWorkspaceStore();
  const { selectedCredentialId, selectedRegion } = useCredentialContext();
  const { selectedProvider } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
  });

  // Fetch node group details
  const { data: nodeGroup, isLoading: isLoadingNodeGroup } = useQuery({
    queryKey: queryKeys.nodeGroups.detail(nodeGroupName),
    queryFn: async () => {
      if (!selectedProvider || !clusterName || !selectedCredentialId || !selectedRegion) {
        return null;
      }
      return kubernetesService.getNodeGroup(
        selectedProvider,
        clusterName,
        nodeGroupName,
        selectedCredentialId,
        selectedRegion
      );
    },
    enabled: !!selectedProvider && !!clusterName && !!selectedCredentialId && !!selectedRegion && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
  });

  // Update mutation - MUST be called before any early returns
  const updateMutation = useMutation({
    mutationFn: (data: Partial<CreateNodeGroupForm>) => {
      if (!selectedProvider || !clusterName || !selectedCredentialId || !selectedRegion) {
        throw new Error('Missing required parameters');
      }
      return kubernetesService.updateNodeGroup(
        selectedProvider,
        clusterName,
        nodeGroupName,
        data,
        selectedCredentialId,
        selectedRegion
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.nodeGroups.detail(nodeGroupName) });
      queryClient.invalidateQueries({ queryKey: queryKeys.nodeGroups.all });
      setIsEditDialogOpen(false);
      success(t('kubernetes.nodeGroupUpdateInitiated') || 'Node group update initiated');
    },
    onError: (error) => {
      handleError(error, { operation: 'updateNodeGroup', resource: 'NodeGroup' });
    },
  });

  // Delete mutation - MUST be called before any early returns
  const deleteMutation = useMutation({
    mutationFn: () => {
      if (!selectedProvider || !clusterName || !selectedCredentialId || !selectedRegion) {
        throw new Error('Missing required parameters');
      }
      return kubernetesService.deleteNodeGroup(
        selectedProvider,
        clusterName,
        nodeGroupName,
        selectedCredentialId,
        selectedRegion
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.nodeGroups.all });
      router.push('/kubernetes/node-groups');
      success(t('kubernetes.nodeGroupDeletionInitiated') || 'Node group deletion initiated');
    },
    onError: (error) => {
      handleError(error, { operation: 'deleteNodeGroup', resource: 'NodeGroup' });
    },
  });

  // Helper functions - MUST be called before any early returns
  const handleEdit = useCallback(() => {
    setIsEditDialogOpen(true);
  }, []);

  const handleDelete = useCallback(() => {
    setIsDeleteDialogOpen(true);
  }, []);

  const handleConfirmDelete = useCallback(() => {
    deleteMutation.mutate();
    setIsDeleteDialogOpen(false);
  }, [deleteMutation]);

  const handleUpdate = useCallback((data: CreateNodeGroupForm) => {
    updateMutation.mutate(data);
  }, [updateMutation]);

  // Early returns AFTER all hooks
  if (authLoading || isLoadingNodeGroup) {
    return (
      <Layout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
        </div>
      </Layout>
    );
  }

  if (!nodeGroup) {
    return (
      <Layout>
        <div className="flex items-center justify-center h-64">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Node Group Not Found</h2>
            <Button onClick={() => router.push('/kubernetes/node-groups')}>
              Back to Node Groups
            </Button>
          </div>
        </div>
      </Layout>
    );
  }

  // Type guards and computed values - AFTER early returns
  const isAWSNodeGroup = (ng: NodeGroup | AWSNodeGroup): ng is AWSNodeGroup => {
    return 'node_role_arn' in ng || 'ami_type' in ng;
  };

  const awsNodeGroup = isAWSNodeGroup(nodeGroup) ? nodeGroup : null;
  
  // Type guard to ensure tags exist
  const nodeGroupWithTags = nodeGroup as NodeGroup & { tags?: Record<string, string> };

  return (
    <Layout>
      <div className="space-y-6">
        <Breadcrumb
          items={[
            { label: t('nav.kubernetes'), href: '/kubernetes/clusters' },
            { label: t('nav.nodeGroups'), href: '/kubernetes/node-groups' },
            { label: nodeGroupName, href: '#' },
          ]}
        />

        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => router.push('/kubernetes/node-groups')}
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back
            </Button>
            <div>
              <h1 className="text-3xl font-bold text-gray-900">{nodeGroupName}</h1>
              <p className="text-gray-600 mt-1">
                {t('kubernetes.cluster')}: {nodeGroup.cluster_name}
              </p>
            </div>
          </div>
          <div className="flex items-center space-x-2">
            <Badge variant={nodeGroup.status === 'ACTIVE' ? 'default' : 'secondary'}>
              {nodeGroup.status}
            </Badge>
            <Button variant="outline" size="sm" onClick={handleEdit} disabled={updateMutation.isPending}>
              <Edit className="h-4 w-4 mr-2" />
              {t('common.edit')}
            </Button>
            <Button variant="destructive" size="sm" onClick={handleDelete} disabled={deleteMutation.isPending}>
              <Trash2 className="h-4 w-4 mr-2" />
              {t('common.delete')}
            </Button>
          </div>
        </div>

        <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as typeof activeTab)} className="space-y-4">
          <TabsList>
            <TabsTrigger value="overview">{t('common.overview')}</TabsTrigger>
            <TabsTrigger value="scaling">{t('common.scaling')}</TabsTrigger>
            <TabsTrigger value="networking">{t('common.networking')}</TabsTrigger>
            <TabsTrigger value="tags">{t('common.tags')}</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>{t('common.overview')}</CardTitle>
                <CardDescription>{t('kubernetes.nodeGroupDetails')}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm font-medium text-gray-500">{t('common.name')}</p>
                    <p className="text-sm">{nodeGroup.name}</p>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-500">{t('common.status')}</p>
                    <Badge variant={nodeGroup.status === 'ACTIVE' ? 'default' : 'secondary'}>
                      {nodeGroup.status}
                    </Badge>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-500">{t('kubernetes.cluster')}</p>
                    <p className="text-sm">{nodeGroup.cluster_name}</p>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-500">{t('common.region')}</p>
                    <p className="text-sm">{nodeGroup.region}</p>
                  </div>
                  {nodeGroup.version && (
                    <div>
                      <p className="text-sm font-medium text-gray-500">{t('common.version')}</p>
                      <p className="text-sm">{nodeGroup.version}</p>
                    </div>
                  )}
                  {awsNodeGroup?.ami_type && (
                    <div>
                      <p className="text-sm font-medium text-gray-500">AMI Type</p>
                      <p className="text-sm">{awsNodeGroup.ami_type}</p>
                    </div>
                  )}
                  {nodeGroup.instance_types && nodeGroup.instance_types.length > 0 && (
                    <div>
                      <p className="text-sm font-medium text-gray-500">{t('common.instanceType')}</p>
                      <p className="text-sm">{nodeGroup.instance_types.join(', ')}</p>
                    </div>
                  )}
                  {nodeGroup.instance_type && (
                    <div>
                      <p className="text-sm font-medium text-gray-500">{t('common.instanceType')}</p>
                      <p className="text-sm">{nodeGroup.instance_type}</p>
                    </div>
                  )}
                  {(nodeGroup.disk_size || nodeGroup.disk_size_gb) && (
                    <div>
                      <p className="text-sm font-medium text-gray-500">{t('common.diskSize')}</p>
                      <p className="text-sm">{nodeGroup.disk_size || nodeGroup.disk_size_gb} GB</p>
                    </div>
                  )}
                  {nodeGroup.capacity_type && (
                    <div>
                      <p className="text-sm font-medium text-gray-500">{t('common.capacityType')}</p>
                      <p className="text-sm">{nodeGroup.capacity_type}</p>
                    </div>
                  )}
                  {awsNodeGroup?.node_role_arn && (
                    <div>
                      <p className="text-sm font-medium text-gray-500">Node Role ARN</p>
                      <p className="text-sm font-mono text-xs break-all">{awsNodeGroup.node_role_arn}</p>
                    </div>
                  )}
                  {nodeGroup.created_at && (
                    <div>
                      <p className="text-sm font-medium text-gray-500">{t('common.createdAt')}</p>
                      <p className="text-sm">{new Date(nodeGroup.created_at).toLocaleString()}</p>
                    </div>
                  )}
                  {nodeGroup.updated_at && (
                    <div>
                      <p className="text-sm font-medium text-gray-500">{t('common.updatedAt')}</p>
                      <p className="text-sm">{new Date(nodeGroup.updated_at).toLocaleString()}</p>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="scaling" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>{t('common.scaling')}</CardTitle>
                <CardDescription>{t('kubernetes.scalingConfiguration')}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-3 gap-4">
                  <div>
                    <p className="text-sm font-medium text-gray-500">{t('common.minSize')}</p>
                    <p className="text-2xl font-bold">{nodeGroup.min_size}</p>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-500">{t('common.maxSize')}</p>
                    <p className="text-2xl font-bold">{nodeGroup.max_size}</p>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-500">{t('common.desiredSize')}</p>
                    <p className="text-2xl font-bold">{nodeGroup.desired_size || nodeGroup.node_count || 0}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="networking" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>{t('common.networking')}</CardTitle>
                <CardDescription>{t('kubernetes.networkingConfiguration')}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {awsNodeGroup?.subnets && awsNodeGroup.subnets.length > 0 && (
                  <div>
                    <p className="text-sm font-medium text-gray-500 mb-2">Subnets</p>
                    <div className="space-y-1">
                      {awsNodeGroup.subnets.map((subnet, idx) => (
                        <p key={idx} className="text-sm font-mono">{subnet}</p>
                      ))}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="tags" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>{t('common.tags')}</CardTitle>
                <CardDescription>{t('kubernetes.nodeGroupTags')}</CardDescription>
              </CardHeader>
              <CardContent>
                {nodeGroupWithTags.tags && Object.keys(nodeGroupWithTags.tags).length > 0 ? (
                  <div className="space-y-2">
                    {Object.entries(nodeGroupWithTags.tags).map(([key, value]) => (
                      <div key={key} className="flex items-center space-x-2">
                        <span className="text-sm font-medium">{key}:</span>
                        <span className="text-sm">{String(value)}</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-gray-500">{t('common.noTags')}</p>
                )}
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>

        {/* Edit Dialog */}
        {isEditDialogOpen && nodeGroup && (
          <CreateNodeGroupDialog
            open={isEditDialogOpen}
            onOpenChange={setIsEditDialogOpen}
            clusterName={clusterName}
            cluster={null} // Not needed for edit
            defaultRegion={selectedRegion || ''}
            defaultCredentialId={selectedCredentialId || ''}
            onSubmit={handleUpdate}
            onCredentialIdChange={() => {}}
            onRegionChange={() => {}}
            isPending={updateMutation.isPending}
            initialData={{
              name: nodeGroup.name,
              cluster_name: nodeGroup.cluster_name,
              min_size: nodeGroup.min_size,
              max_size: nodeGroup.max_size,
              desired_size: nodeGroup.desired_size || nodeGroup.node_count || 0,
              instance_types: nodeGroup.instance_types || (nodeGroup.instance_type ? [nodeGroup.instance_type] : []),
              ami_type: awsNodeGroup?.ami_type,
              disk_size: nodeGroup.disk_size || nodeGroup.disk_size_gb,
              region: nodeGroup.region,
              subnet_ids: awsNodeGroup?.subnets,
              capacity_type: nodeGroup.capacity_type,
            }}
          />
        )}

        {/* Delete Confirmation Dialog */}
        <DeleteConfirmationDialog
          open={isDeleteDialogOpen}
          onOpenChange={setIsDeleteDialogOpen}
          onConfirm={handleConfirmDelete}
          title={t('kubernetes.deleteNodeGroup') || 'Delete Node Group'}
          description={t('kubernetes.confirmDeleteNodeGroup', { nodeGroupName }) || `Are you sure you want to delete node group ${nodeGroupName}?`}
          isLoading={deleteMutation.isPending}
          resourceName={nodeGroupName}
        />
      </div>
    </Layout>
  );
}

export default function NodeGroupDetailPage() {
  return (
    <Suspense fallback={
      <Layout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
        </div>
      </Layout>
    }>
      <NodeGroupDetailPageContent />
    </Suspense>
  );
}

