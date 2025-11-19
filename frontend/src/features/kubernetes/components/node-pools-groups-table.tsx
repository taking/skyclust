/**
 * Node Pools/Groups Table Component
 * Kubernetes Node Pool/Group 목록 테이블
 */

'use client';

import * as React from 'react';
import { useMemo, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Checkbox } from '@/components/ui/checkbox';
import { Pagination } from '@/components/ui/pagination';
import { VirtualizedTable } from '@/components/common/virtualized-table';
import { NodePoolGroupRow } from './node-pool-group-row';
import { useTranslation } from '@/hooks/use-translation';
import { buildCredentialResourceDetailPath } from '@/lib/routing/helpers';
import type { NodePool, NodeGroup, CloudProvider } from '@/lib/types/kubernetes';

type NodePoolOrGroup = (NodePool | NodeGroup) & {
  cluster_name: string;
  cluster_id: string;
  provider: CloudProvider;
  resource_type: 'node-pool' | 'node-group';
  credential_id?: string;
};

interface NodePoolsGroupsTableProps {
  items: NodePoolOrGroup[];
  selectedIds: string[];
  onSelectionChange: (ids: string[]) => void;
  onDelete: (name: string, clusterName: string, region: string) => void;
  isDeleting?: boolean;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
  showProviderColumn?: boolean;
  workspaceId: string;
  credentialId: string;
}

export function NodePoolsGroupsTable({
  items,
  selectedIds,
  onSelectionChange,
  onDelete,
  isDeleting = false,
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
  showProviderColumn = false,
  workspaceId,
  credentialId,
}: NodePoolsGroupsTableProps) {
  const { t } = useTranslation();
  const router = useRouter();
  
  const shouldUseVirtualScrolling = items.length >= 50;
  const allSelected = useMemo(
    () => selectedIds.length === items.length && items.length > 0,
    [selectedIds.length, items.length]
  );

  const handleSelectAll = useCallback((checked: boolean) => {
    onSelectionChange(checked ? items.map(item => item.id || item.name) : []);
  }, [items, onSelectionChange]);

  const handleSelectItem = useCallback((itemId: string, checked: boolean) => {
    if (checked) {
      onSelectionChange([...selectedIds, itemId]);
    } else {
      onSelectionChange(selectedIds.filter(id => id !== itemId));
    }
  }, [selectedIds, onSelectionChange]);

  const handleRowClick = useCallback((item: NodePoolOrGroup) => {
    const path = buildCredentialResourceDetailPath(
      workspaceId,
      credentialId,
      'k8s',
      item.resource_type === 'node-group' ? 'node-groups' : 'node-pools',
      item.name,
      { region: item.region }
    );
    router.push(path);
  }, [workspaceId, credentialId, router]);

  const renderRow = useCallback((item: NodePoolOrGroup) => {
    const itemId = item.id || item.name;
    const isSelected = selectedIds.includes(itemId);
    
    return (
      <TableRow
        key={`${item.cluster_id}-${itemId}`}
        className="cursor-pointer hover:bg-accent"
        onClick={() => handleRowClick(item)}
      >
        <NodePoolGroupRow
          item={item}
          isSelected={isSelected}
          onSelect={(checked) => handleSelectItem(itemId, checked)}
          onDelete={onDelete}
          isDeleting={isDeleting}
          showProvider={showProviderColumn}
          workspaceId={workspaceId}
          credentialId={credentialId}
        />
      </TableRow>
    );
  }, [selectedIds, onDelete, isDeleting, showProviderColumn, workspaceId, credentialId, handleSelectItem, handleRowClick]);

  return (
    <div className="space-y-4">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-12">
              <Checkbox
                checked={allSelected}
                onCheckedChange={handleSelectAll}
                aria-label={t('common.selectAll') || 'Select all'}
              />
            </TableHead>
            {showProviderColumn && (
              <TableHead>{t('common.provider') || 'Provider'}</TableHead>
            )}
            <TableHead>{t('kubernetes.cluster') || 'Cluster'}</TableHead>
            <TableHead>{t('common.name') || 'Name'}</TableHead>
            <TableHead>{t('common.instanceType') || 'Instance Type'}</TableHead>
            <TableHead>{t('common.nodes') || 'Nodes'}</TableHead>
            <TableHead>{t('common.minMax') || 'Min/Max'}</TableHead>
            <TableHead>{t('common.status') || 'Status'}</TableHead>
            <TableHead>{t('common.region') || 'Region'}</TableHead>
            <TableHead>{t('common.actions') || 'Actions'}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {shouldUseVirtualScrolling ? (
            <VirtualizedTable
              items={items}
              renderRow={renderRow}
              rowHeight={60}
              overscan={5}
            />
          ) : (
            items.map(item => renderRow(item))
          )}
        </TableBody>
      </Table>

      {total > pageSize && (
        <Pagination
          currentPage={page}
          totalPages={Math.ceil(total / pageSize)}
          onPageChange={onPageChange}
          pageSize={pageSize}
          onPageSizeChange={onPageSizeChange}
          totalItems={total}
        />
      )}
    </div>
  );
}

