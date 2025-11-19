/**
 * Cluster Table Component
 * Kubernetes 클러스터 목록 테이블
 */

'use client';

import * as React from 'react';
import { useMemo, useCallback } from 'react';
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Checkbox } from '@/components/ui/checkbox';
import { Pagination } from '@/components/ui/pagination';
import { VirtualizedTable } from '@/components/common/virtualized-table';
import { ClusterRow } from './cluster-row';
import type { KubernetesCluster, CloudProvider } from '@/lib/types';

interface ClusterWithProvider extends KubernetesCluster {
  provider?: CloudProvider;
  credential_id?: string;
}

interface ClusterTableProps {
  clusters: ClusterWithProvider[];
  provider?: CloudProvider;
  selectedIds: string[];
  onSelectionChange: (ids: string[]) => void;
  onDelete: (clusterName: string, region: string) => void;
  onDownloadKubeconfig: (clusterName: string, region: string) => void;
  isDeleting?: boolean;
  isDownloading?: boolean;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
  showProviderColumn?: boolean;
}

function ClusterTableComponent({
  clusters,
  provider,
  selectedIds,
  onSelectionChange,
  onDelete,
  onDownloadKubeconfig,
  isDeleting = false,
  isDownloading = false,
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
  showProviderColumn = false,
}: ClusterTableProps) {
  const isAzure = provider === 'azure';
  const hasMultipleProviders = showProviderColumn || clusters.some(c => c.provider && c.provider !== provider);
  // Virtual scrolling은 50개 이상일 때만 활성화
  const shouldUseVirtualScrolling = clusters.length >= 50;
  const allSelected = useMemo(
    () => selectedIds.length === clusters.length && clusters.length > 0,
    [selectedIds.length, clusters.length]
  );

  const handleSelectAll = useCallback((checked: boolean) => {
    onSelectionChange(checked ? clusters.map(c => c.id || c.name) : []);
  }, [clusters, onSelectionChange]);

  const handleSelectCluster = useCallback((clusterId: string, checked: boolean) => {
    if (checked) {
      onSelectionChange([...selectedIds, clusterId]);
    } else {
      onSelectionChange(selectedIds.filter(id => id !== clusterId));
    }
  }, [selectedIds, onSelectionChange]);

  // renderRow를 조건문 밖에서 정의 (React Hook 규칙 준수)
  const renderRow = useCallback((item: ClusterWithProvider & { id?: string; [key: string]: unknown }) => {
    const cluster = item as ClusterWithProvider;
    const clusterId = cluster.id || cluster.name;
    const isSelected = selectedIds.includes(clusterId);
    const clusterProvider = cluster.provider || provider;
    
    return (
      <TableRow key={clusterId}>
        <ClusterRow
          cluster={cluster}
          provider={clusterProvider}
          isSelected={isSelected}
          onSelect={(checked) => handleSelectCluster(clusterId, checked)}
          onDelete={onDelete}
          onDownloadKubeconfig={onDownloadKubeconfig}
          isDeleting={isDeleting}
          isDownloading={isDownloading}
          showProvider={hasMultipleProviders}
        />
      </TableRow>
    );
  }, [selectedIds, handleSelectCluster, onDelete, onDownloadKubeconfig, isDeleting, isDownloading, provider, hasMultipleProviders]);

  if (shouldUseVirtualScrolling) {
    return (
      <>
        <VirtualizedTable
          data={clusters as Array<KubernetesCluster & { id?: string; [key: string]: unknown }>}
          minItems={50}
          containerHeight="600px"
          estimateSize={60}
          renderHeader={() => (
            <TableRow>
              <TableHead className="w-12">
                <Checkbox
                  checked={allSelected}
                  onCheckedChange={handleSelectAll}
                />
              </TableHead>
              {hasMultipleProviders && <TableHead>Provider</TableHead>}
              <TableHead>Name</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Version</TableHead>
              <TableHead>Region</TableHead>
              {isAzure && <TableHead>Resource Group</TableHead>}
              <TableHead>Endpoint</TableHead>
              <TableHead>Created</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          )}
          renderRow={renderRow}
        />
        {total > 0 && (
          <div className="border-t">
            <Pagination
              total={total}
              page={page}
              pageSize={pageSize}
              onPageChange={onPageChange}
              onPageSizeChange={onPageSizeChange}
              pageSizeOptions={[10, 20, 50, 100]}
              showPageSizeSelector={true}
            />
          </div>
        )}
      </>
    );
  }

  return (
    <>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-12">
              <Checkbox
                checked={allSelected}
                onCheckedChange={handleSelectAll}
              />
            </TableHead>
            {hasMultipleProviders && <TableHead>Provider</TableHead>}
            <TableHead>Name</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Version</TableHead>
            <TableHead>Region</TableHead>
            {isAzure && <TableHead>Resource Group</TableHead>}
            <TableHead>Endpoint</TableHead>
            <TableHead>Created</TableHead>
            <TableHead>Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {clusters.map((cluster) => {
            const clusterId = cluster.id || cluster.name;
            const isSelected = selectedIds.includes(clusterId);
            const clusterProvider = cluster.provider || provider;
            
            return (
              <TableRow key={clusterId}>
                <ClusterRow
                  cluster={cluster}
                  provider={clusterProvider}
                  isSelected={isSelected}
                  onSelect={(checked) => handleSelectCluster(clusterId, checked)}
                  onDelete={onDelete}
                  onDownloadKubeconfig={onDownloadKubeconfig}
                  isDeleting={isDeleting}
                  isDownloading={isDownloading}
                  showProvider={hasMultipleProviders}
                />
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
      {total > 0 && (
        <div className="border-t">
          <Pagination
            total={total}
            page={page}
            pageSize={pageSize}
            onPageChange={onPageChange}
            onPageSizeChange={onPageSizeChange}
            pageSizeOptions={[10, 20, 50, 100]}
            showPageSizeSelector={true}
          />
        </div>
      )}
    </>
  );
}

export const ClusterTable = React.memo(ClusterTableComponent);

