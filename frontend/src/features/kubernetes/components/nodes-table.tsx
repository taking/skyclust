/**
 * Nodes Table Component
 * Kubernetes 노드 목록 테이블
 */

'use client';

import * as React from 'react';
import { useMemo, useCallback } from 'react';
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Checkbox } from '@/components/ui/checkbox';
import { Pagination } from '@/components/ui/pagination';
import { VirtualizedTable } from '@/components/common/virtualized-table';
import { Badge } from '@/components/ui/badge';
import { TableCell } from '@/components/ui/table';
import { useTranslation } from '@/hooks/use-translation';
import type { Node, CloudProvider } from '@/lib/types';

interface NodeWithMetadata extends Node {
  cluster_name: string;
  cluster_id: string;
  provider?: CloudProvider;
  credential_id?: string;
}

interface NodesTableProps {
  nodes: NodeWithMetadata[];
  provider?: CloudProvider;
  selectedIds: string[];
  onSelectionChange: (ids: string[]) => void;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
  showProviderColumn?: boolean;
}

function NodesTableComponent({
  nodes,
  provider,
  selectedIds,
  onSelectionChange,
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
  showProviderColumn = false,
}: NodesTableProps) {
  const { t } = useTranslation();
  const hasMultipleProviders = showProviderColumn || nodes.some(n => n.provider && n.provider !== provider);
  const shouldUseVirtualScrolling = nodes.length >= 50;
  const allSelected = useMemo(
    () => selectedIds.length === nodes.length && nodes.length > 0,
    [selectedIds.length, nodes.length]
  );

  const handleSelectAll = useCallback((checked: boolean) => {
    onSelectionChange(checked ? nodes.map(n => n.id || n.name) : []);
  }, [nodes, onSelectionChange]);

  const handleSelectNode = useCallback((nodeId: string, checked: boolean) => {
    if (checked) {
      onSelectionChange([...selectedIds, nodeId]);
    } else {
      onSelectionChange(selectedIds.filter(id => id !== nodeId));
    }
  }, [selectedIds, onSelectionChange]);

  const renderRow = useCallback((item: NodeWithMetadata & { id?: string; [key: string]: unknown }) => {
    const node = item as NodeWithMetadata;
    const nodeId = node.id || node.name;
    const isSelected = selectedIds.includes(nodeId);
    const nodeProvider = node.provider || provider;
    
    return (
      <TableRow key={nodeId}>
        <TableCell className="w-12">
          <Checkbox
            checked={isSelected}
            onCheckedChange={(checked) => handleSelectNode(nodeId, checked as boolean)}
            aria-label={`Select node ${node.name}`}
          />
        </TableCell>
        {hasMultipleProviders && (
          <TableCell>
            <Badge variant="outline">{nodeProvider}</Badge>
          </TableCell>
        )}
        <TableCell className="font-medium">{node.cluster_name}</TableCell>
        <TableCell className="font-medium">{node.name}</TableCell>
        <TableCell>{node.instance_type || '-'}</TableCell>
        <TableCell>{node.zone || '-'}</TableCell>
        <TableCell>
          <Badge variant={node.status === 'Ready' ? 'default' : 'secondary'}>
            {node.status}
          </Badge>
        </TableCell>
        <TableCell>{node.private_ip || '-'}</TableCell>
      </TableRow>
    );
  }, [selectedIds, handleSelectNode, provider, hasMultipleProviders]);

  if (shouldUseVirtualScrolling) {
    return (
      <div className="space-y-4">
        <div className="flex items-center space-x-2 border-b pb-2">
          <Checkbox
            checked={allSelected}
            onCheckedChange={handleSelectAll}
            aria-label="Select all nodes"
          />
          <span className="text-sm text-muted-foreground">
            {selectedIds.length} of {nodes.length} selected
          </span>
        </div>
        <VirtualizedTable
          items={nodes}
          renderRow={renderRow}
          header={
            <TableHeader>
              <TableRow>
                <TableHead className="w-12"></TableHead>
                {hasMultipleProviders && (
                  <TableHead>{t('common.provider') || 'Provider'}</TableHead>
                )}
                <TableHead>{t('kubernetes.cluster') || 'Cluster'}</TableHead>
                <TableHead>{t('common.name') || 'Name'}</TableHead>
                <TableHead>{t('common.instanceType') || 'Instance Type'}</TableHead>
                <TableHead>{t('common.zone') || 'Zone'}</TableHead>
                <TableHead>{t('common.status') || 'Status'}</TableHead>
                <TableHead>{t('common.privateIP') || 'Private IP'}</TableHead>
              </TableRow>
            </TableHeader>
          }
          itemHeight={48}
          containerHeight={600}
        />
        <Pagination
          page={page}
          pageSize={pageSize}
          total={total}
          onPageChange={onPageChange}
          onPageSizeChange={onPageSizeChange}
        />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-12">
              <Checkbox
                checked={allSelected}
                onCheckedChange={handleSelectAll}
                aria-label="Select all nodes"
              />
            </TableHead>
            {hasMultipleProviders && (
              <TableHead>{t('common.provider') || 'Provider'}</TableHead>
            )}
            <TableHead>{t('kubernetes.cluster') || 'Cluster'}</TableHead>
            <TableHead>{t('common.name') || 'Name'}</TableHead>
            <TableHead>{t('common.instanceType') || 'Instance Type'}</TableHead>
            <TableHead>{t('common.zone') || 'Zone'}</TableHead>
            <TableHead>{t('common.status') || 'Status'}</TableHead>
            <TableHead>{t('common.privateIP') || 'Private IP'}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {nodes.map((node) => renderRow(node))}
        </TableBody>
      </Table>
      <Pagination
        page={page}
        pageSize={pageSize}
        total={total}
        onPageChange={onPageChange}
        onPageSizeChange={onPageSizeChange}
      />
    </div>
  );
}

export const NodesTable = React.memo(NodesTableComponent);

