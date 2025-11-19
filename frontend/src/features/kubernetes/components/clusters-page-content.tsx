/**
 * Clusters Page Content
 * 
 * 클러스터 페이지의 메인 콘텐츠 컴포넌트
 * 테이블, 다이얼로그를 포함합니다.
 */

'use client';

import * as React from 'react';
import dynamic from 'next/dynamic';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import { BulkTagDialog } from './bulk-tag-dialog';
import type { KubernetesCluster, CloudProvider } from '@/lib/types/kubernetes';

// Dynamic imports for heavy components
const ClusterTable = dynamic(
  () => import('@/features/kubernetes').then(mod => ({ default: mod.ClusterTable })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={8} rows={5} showCheckbox={true} />,
  }
);

export interface ClustersPageContentProps {
  clusters: (KubernetesCluster & { provider?: CloudProvider; credential_id?: string })[];
  filteredClusters: (KubernetesCluster & { provider?: CloudProvider; credential_id?: string })[];
  paginatedClusters: (KubernetesCluster & { provider?: CloudProvider; credential_id?: string })[];
  selectedProvider: CloudProvider | undefined;
  selectedClusterIds: string[];
  onSelectionChange: (ids: string[] | ((prev: string[]) => string[])) => void;
  onDelete: (clusterName: string, clusterRegion: string) => void;
  onDownloadKubeconfig: (clusterName: string, clusterRegion: string) => void;
  isDeleting: boolean;
  isDownloading: boolean;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
  isSearching: boolean;
  searchQuery: string;
  isMultiProviderMode: boolean;
  selectedProviders: CloudProvider[];
  isTagDialogOpen: boolean;
  onTagDialogOpenChange: (open: boolean) => void;
  onBulkTagSubmit: () => void;
  bulkTagKey: string;
  bulkTagValue: string;
  onBulkTagKeyChange: (key: string) => void;
  onBulkTagValueChange: (value: string) => void;
}

export function ClustersPageContent({
  clusters,
  filteredClusters,
  paginatedClusters,
  selectedProvider,
  selectedClusterIds,
  onSelectionChange,
  onDelete,
  onDownloadKubeconfig,
  isDeleting,
  isDownloading,
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
  isSearching,
  searchQuery,
  isMultiProviderMode,
  selectedProviders,
  isTagDialogOpen,
  onTagDialogOpenChange,
  onBulkTagSubmit,
  bulkTagKey,
  bulkTagValue,
  onBulkTagKeyChange,
  onBulkTagValueChange,
}: ClustersPageContentProps) {
  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Clusters</CardTitle>
          <CardDescription>
            {filteredClusters.length} of {clusters.length} cluster{clusters.length !== 1 ? 's' : ''} 
            {isSearching && ` (${searchQuery})`}
            {isMultiProviderMode && ` • ${selectedProviders.length} provider${selectedProviders.length !== 1 ? 's' : ''}`}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <ClusterTable
            clusters={paginatedClusters}
            provider={selectedProvider}
            selectedIds={selectedClusterIds}
            onSelectionChange={onSelectionChange}
            onDelete={onDelete}
            onDownloadKubeconfig={onDownloadKubeconfig}
            isDeleting={isDeleting}
            isDownloading={isDownloading}
            page={page}
            pageSize={pageSize}
            total={total}
            onPageChange={onPageChange}
            onPageSizeChange={onPageSizeChange}
            showProviderColumn={isMultiProviderMode}
          />
        </CardContent>
      </Card>

      {/* Bulk Tag Dialog */}
      <BulkTagDialog
        open={isTagDialogOpen}
        onOpenChange={onTagDialogOpenChange}
        onSubmit={onBulkTagSubmit}
        tagKey={bulkTagKey}
        tagValue={bulkTagValue}
        onTagKeyChange={onBulkTagKeyChange}
        onTagValueChange={onBulkTagValueChange}
        selectedCount={selectedClusterIds.length}
      />
    </>
  );
}

