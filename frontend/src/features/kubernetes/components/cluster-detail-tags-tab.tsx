/**
 * Cluster Detail Tags Tab Component
 * 클러스터 상세 태그 탭 컴포넌트
 */

'use client';

import * as React from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useToast } from '@/hooks/use-toast';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import dynamic from 'next/dynamic';
import type { BaseCluster, CloudProvider } from '@/lib/types';

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

interface ClusterDetailTagsTabProps {
  cluster: BaseCluster;
  selectedProvider?: CloudProvider;
  clusterName: string;
  selectedCredentialId: string;
  selectedRegion: string;
}

export function ClusterDetailTagsTab({
  cluster,
  selectedProvider,
  clusterName,
  selectedCredentialId,
  selectedRegion,
}: ClusterDetailTagsTabProps) {
  const queryClient = useQueryClient();
  const { success } = useToast();

  const handleTagsChange = React.useCallback(
    (updatedTags: Record<string, string>) => {
      if (!cluster || !selectedProvider) return;

      queryClient.setQueryData(
        ['kubernetes-cluster', selectedProvider, clusterName, selectedCredentialId, selectedRegion],
        {
          ...cluster,
          tags: updatedTags,
        }
      );

      success('Tags updated successfully');
    },
    [cluster, selectedProvider, clusterName, selectedCredentialId, selectedRegion, queryClient, success]
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>태그</CardTitle>
      </CardHeader>
      <CardContent>
        <TagManager tags={cluster.tags || {}} onTagsChange={handleTagsChange} />
      </CardContent>
    </Card>
  );
}

