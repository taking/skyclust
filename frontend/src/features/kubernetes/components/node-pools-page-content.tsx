/**
 * Node Pools Page Content
 * 
 * Kubernetes Node Pools 페이지의 메인 콘텐츠 컴포넌트
 * 모든 클러스터의 Node Pool을 조회하고 표시합니다.
 */

'use client';

import * as React from 'react';
import { useRouter } from 'next/navigation';
import { useQuery, useQueries } from '@tanstack/react-query';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { Server, Plus } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { kubernetesService } from '@/features/kubernetes';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import { useCredentials } from '@/hooks/use-credentials';
import { buildResourceDetailPath, buildResourceCreatePath } from '@/lib/routing/helpers';
import type { NodePool, CloudProvider } from '@/lib/types';

interface NodePoolsPageContentProps {
  workspaceId: string;
  credentialId: string;
  region?: string;
  onNodePoolClick?: (nodePoolName: string) => void;
  onCreateClick?: () => void;
}

export function NodePoolsPageContent({
  workspaceId,
  credentialId,
  region,
  onNodePoolClick,
  onCreateClick,
}: NodePoolsPageContentProps) {
  const { t } = useTranslation();
  const router = useRouter();
  
  // Credentials 조회
  const { credentials, isLoading: isLoadingCredentials } = useCredentials({
    workspaceId,
    enabled: !!workspaceId,
  });

  // 현재 credential 정보
  const currentCredential = credentials.find(c => c.id === credentialId);
  const provider = currentCredential?.provider as CloudProvider | undefined;

  // 클러스터 목록 조회
  const { data: clusters = [], isLoading: isLoadingClusters } = useQuery({
    queryKey: queryKeys.kubernetesClusters.list(workspaceId, provider, credentialId, region),
    queryFn: async () => {
      if (!provider) return [];
      return kubernetesService.listClusters(provider, credentialId, region);
    },
    enabled: !!provider && !!credentialId && !!workspaceId,
    staleTime: CACHE_TIMES.MONITORING,
    gcTime: GC_TIMES.MEDIUM,
  });

  // 각 클러스터의 Node Pool 조회 (useQueries 사용)
  const nodePoolQueriesResults = useQueries({
    queries: clusters.map(cluster => ({
      queryKey: queryKeys.kubernetesClusters.nodePools(
        cluster.name,
        provider,
        credentialId,
        region
      ),
      queryFn: async () => {
        if (!provider) return [];
        return kubernetesService.listNodePools(provider, cluster.name, credentialId, region || cluster.region);
      },
      enabled: !!provider && !!credentialId && !!cluster.name && clusters.length > 0,
      staleTime: CACHE_TIMES.MONITORING,
      gcTime: GC_TIMES.MEDIUM,
    })),
  });

  // 모든 Node Pool 수집 및 클러스터 정보 추가
  const allNodePools = React.useMemo(() => {
    const nodePools: Array<NodePool & { cluster_name: string; cluster_id: string }> = [];
    
    nodePoolQueriesResults.forEach((result, index) => {
      const cluster = clusters[index];
      if (!cluster || !result.data) return;
      
      result.data.forEach(nodePool => {
        nodePools.push({
          ...nodePool,
          cluster_name: cluster.name,
          cluster_id: cluster.id,
        });
      });
    });
    
    return nodePools;
  }, [nodePoolQueriesResults, clusters]);

  const isLoading = isLoadingCredentials || isLoadingClusters || nodePoolQueriesResults.some(r => r.isLoading);

  const handleNodePoolClick = React.useCallback((nodePool: NodePool & { cluster_name: string; cluster_id: string }) => {
    if (onNodePoolClick) {
      onNodePoolClick(nodePool.name);
    } else {
      const path = buildResourceDetailPath(
        workspaceId,
        credentialId,
        'kubernetes',
        'node-pools',
        nodePool.name,
        { region: region || nodePool.region }
      );
      router.push(path);
    }
  }, [workspaceId, credentialId, region, onNodePoolClick, router]);

  const handleCreateClick = React.useCallback(() => {
    if (onCreateClick) {
      onCreateClick();
    } else {
      const path = buildResourceCreatePath(
        workspaceId,
        credentialId,
        'kubernetes',
        '/node-pools',
        { region: region || undefined }
      );
      router.push(path);
    }
  }, [workspaceId, credentialId, region, onCreateClick, router]);

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>{t('kubernetes.nodePools') || 'Node Pools'}</CardTitle>
          <CardDescription>
            {t('common.loading') || 'Loading node pools...'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <TableSkeleton columns={7} rows={5} />
        </CardContent>
      </Card>
    );
  }

  if (allNodePools.length === 0) {
    return (
      <ResourceEmptyState
        resourceName={t('kubernetes.nodePools')}
        icon={Server}
        description={t('kubernetes.noNodePoolsFound') || 'No node pools found in any cluster'}
        onCreateClick={handleCreateClick}
        withCard={true}
      />
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>{t('kubernetes.nodePools') || 'Node Pools'}</CardTitle>
            <CardDescription>
              {allNodePools.length} {t('kubernetes.nodePoolCount', { count: allNodePools.length }) || 'node pool(s)'} across {clusters.length} {t('kubernetes.clusterCount', { count: clusters.length }) || 'cluster(s)'}
            </CardDescription>
          </div>
          <Button onClick={handleCreateClick} aria-label={t('kubernetes.createNodePool') || 'Create Node Pool'}>
            <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
            {t('kubernetes.createNodePool') || 'Create Node Pool'}
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('kubernetes.cluster') || 'Cluster'}</TableHead>
              <TableHead>{t('common.name') || 'Name'}</TableHead>
              <TableHead>{t('common.instanceType') || 'Instance Type'}</TableHead>
              <TableHead>{t('common.nodes') || 'Nodes'}</TableHead>
              <TableHead>{t('common.minMax') || 'Min/Max'}</TableHead>
              <TableHead>{t('common.status') || 'Status'}</TableHead>
              <TableHead>{t('common.region') || 'Region'}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {allNodePools.map((nodePool) => (
              <TableRow 
                key={`${nodePool.cluster_id}-${nodePool.id || nodePool.name}`}
                className="cursor-pointer hover:bg-accent"
                onClick={() => handleNodePoolClick(nodePool)}
              >
                <TableCell className="font-medium">{nodePool.cluster_name}</TableCell>
                <TableCell className="font-medium">{nodePool.name}</TableCell>
                <TableCell>{nodePool.instance_type || nodePool.instance_types?.join(', ') || '-'}</TableCell>
                <TableCell>{nodePool.node_count || nodePool.desired_size || '-'}</TableCell>
                <TableCell>{nodePool.min_nodes || nodePool.min_size}/{nodePool.max_nodes || nodePool.max_size}</TableCell>
                <TableCell>
                  <Badge variant={nodePool.status === 'RUNNING' || nodePool.status === 'ACTIVE' ? 'default' : 'secondary'}>
                    {nodePool.status}
                  </Badge>
                </TableCell>
                <TableCell>{nodePool.region || '-'}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}



