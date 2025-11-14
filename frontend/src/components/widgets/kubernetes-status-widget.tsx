'use client';

import * as React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Container, AlertCircle } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { kubernetesService } from '@/features/kubernetes';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import { useCredentials } from '@/hooks/use-credentials';

interface KubernetesStatusWidgetProps {
  credentialId?: string;
  region?: string;
  isLoading?: boolean;
}

function KubernetesStatusWidgetComponent({ credentialId, region, isLoading }: KubernetesStatusWidgetProps) {
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedCredentialId, selectedRegion: contextRegion } = useCredentialContext();

  // Fetch credentials using unified hook
  const { credentials, selectedCredential, selectedProvider } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: credentialId || selectedCredentialId || undefined,
  });

  const activeCredentialId = credentialId || selectedCredentialId || '';
  const activeRegion = region || contextRegion || 'ap-northeast-2';

  const { data: clusters = [], isLoading: isLoadingClusters } = useQuery({
    queryKey: [...queryKeys.kubernetesClusters.all, 'widget', selectedProvider, activeCredentialId, activeRegion],
    queryFn: async () => {
      if (!selectedProvider || !activeCredentialId) return [];
      return kubernetesService.listClusters(selectedProvider, activeCredentialId, activeRegion);
    },
    enabled: !!selectedProvider && !!activeCredentialId && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    // refetchInterval 제거: SSE 이벤트로 자동 업데이트
  });

  const isLoadingData = isLoading || isLoadingClusters;

  if (isLoadingData) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <Container className="mr-2 h-5 w-5" />
            Kubernetes Clusters
          </CardTitle>
          <CardDescription>Loading cluster status...</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="animate-pulse">
              <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
              <div className="h-4 bg-gray-200 rounded w-1/2"></div>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!selectedProvider || !activeCredentialId) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <Container className="mr-2 h-5 w-5" />
            Kubernetes Clusters
          </CardTitle>
          <CardDescription>Select provider and credential</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-4">
            <AlertCircle className="h-8 w-8 text-gray-400" />
          </div>
        </CardContent>
      </Card>
    );
  }

  const totalClusters = clusters.length;
  const activeClusters = clusters.filter(c => c.status === 'ACTIVE' || c.status === 'RUNNING').length;
  const creatingClusters = clusters.filter(c => c.status === 'CREATING').length;
  const totalNodes = clusters.reduce((sum, c) => sum + (c.node_pool_info?.total_nodes || 0), 0);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Container className="mr-2 h-5 w-5" />
          Kubernetes Clusters
        </CardTitle>
        <CardDescription>Cluster status overview</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <div className="text-2xl font-bold">{totalClusters}</div>
              <div className="text-sm text-gray-500">Total Clusters</div>
            </div>
            <div>
              <div className="text-2xl font-bold">{totalNodes}</div>
              <div className="text-sm text-gray-500">Total Nodes</div>
            </div>
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-sm">Active</span>
              <Badge variant="default">{activeClusters}</Badge>
            </div>
            {creatingClusters > 0 && (
              <div className="flex items-center justify-between">
                <span className="text-sm">Creating</span>
                <Badge variant="secondary">{creatingClusters}</Badge>
              </div>
            )}
          </div>

          {clusters.length > 0 && (
            <div className="space-y-2 pt-2 border-t">
              <div className="text-sm font-medium">Recent Clusters</div>
              <div className="space-y-1">
                {clusters.slice(0, 3).map((cluster) => (
                  <div key={cluster.id || cluster.name} className="flex items-center justify-between text-sm">
                    <span className="truncate">{cluster.name}</span>
                    <Badge
                      variant={
                        cluster.status === 'ACTIVE' || cluster.status === 'RUNNING'
                          ? 'default'
                          : cluster.status === 'CREATING'
                          ? 'secondary'
                          : 'destructive'
                      }
                      className="ml-2"
                    >
                      {cluster.status}
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

export const KubernetesStatusWidget = React.memo(KubernetesStatusWidgetComponent);

