/**
 * Cluster Info Card Component
 * 클러스터 정보 카드 컴포넌트
 */

'use client';

import * as React from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useToast } from '@/hooks/use-toast';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Label } from '@/components/ui/label';
import dynamic from 'next/dynamic';
import type { KubernetesCluster } from '@/lib/types';

const TagManager = dynamic(
  () => import('@/components/common/tag-manager').then(mod => ({ default: mod.TagManager })),
  { 
    ssr: false,
    loading: () => (
      <div className="animate-pulse">
        <div className="h-8 w-24 bg-gray-200 rounded mb-2"></div>
        <div className="h-8 w-32 bg-gray-200 rounded"></div>
      </div>
    ),
  }
);

interface ClusterInfoCardProps {
  cluster: KubernetesCluster | undefined;
  isLoading: boolean;
  selectedProvider?: string;
  clusterName: string;
  selectedCredentialId: string;
  selectedRegion: string;
}

export function ClusterInfoCard({
  cluster,
  isLoading,
  selectedProvider,
  clusterName,
  selectedCredentialId,
  selectedRegion,
}: ClusterInfoCardProps) {
  const queryClient = useQueryClient();
  const { success } = useToast();

  const handleTagsChange = React.useCallback((updatedTags: Record<string, string>) => {
    if (!cluster || !selectedProvider) return;

    queryClient.setQueryData(['kubernetes-cluster', selectedProvider, clusterName, selectedCredentialId, selectedRegion], {
      ...cluster,
      tags: updatedTags,
    });

    success('Tags updated successfully');
  }, [cluster, selectedProvider, clusterName, selectedCredentialId, selectedRegion, queryClient, success]);
  if (isLoading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
            <p className="mt-2 text-gray-600">Loading cluster details...</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!cluster) {
    return null;
  }

  return (
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
          <div className="mt-4 pt-4 border-t md:col-span-3">
            <TagManager
              tags={cluster.tags}
              onTagsChange={handleTagsChange}
            />
          </div>
        )}
      </CardContent>
    </Card>
  );
}

