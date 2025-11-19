/**
 * Node Pools/Groups Page Content
 * 
 * Node Pool/Group 페이지의 메인 콘텐츠 컴포넌트
 * 클러스터 페이지와 일관성 있는 구조
 */

'use client';

import * as React from 'react';
import dynamic from 'next/dynamic';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import type { NodePool, NodeGroup, CloudProvider } from '@/lib/types/kubernetes';

// Dynamic imports for heavy components
const NodePoolsGroupsTable = dynamic(
  () => import('@/features/kubernetes').then(mod => ({ default: mod.NodePoolsGroupsTable })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={10} rows={5} showCheckbox={true} />,
  }
);

type NodePoolOrGroup = (NodePool | NodeGroup) & {
  cluster_name: string;
  cluster_id: string;
  provider: CloudProvider;
  resource_type: 'node-pool' | 'node-group';
  credential_id?: string;
};

export interface NodePoolsGroupsPageContentProps {
  items: NodePoolOrGroup[];
  filteredItems: NodePoolOrGroup[];
  paginatedItems: NodePoolOrGroup[];
  selectedProvider: CloudProvider | undefined;
  selectedItemIds: string[];
  onSelectionChange: (ids: string[] | ((prev: string[]) => string[])) => void;
  onDelete: (name: string, clusterName: string, region: string) => void;
  isDeleting: boolean;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
  isSearching: boolean;
  searchQuery: string;
  isMultiProviderMode: boolean;
  selectedProviders: CloudProvider[];
  resourceName: string; // "Node Pools" or "Node Groups" or "Node Pools & Groups"
  workspaceId: string;
  credentialId: string;
}

export function NodePoolsGroupsPageContent({
  items,
  filteredItems,
  paginatedItems,
  selectedProvider,
  selectedItemIds,
  onSelectionChange,
  onDelete,
  isDeleting,
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
  isSearching,
  searchQuery,
  isMultiProviderMode,
  selectedProviders,
  resourceName,
  workspaceId,
  credentialId,
}: NodePoolsGroupsPageContentProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{resourceName}</CardTitle>
        <CardDescription>
          {filteredItems.length} of {items.length} {resourceName.toLowerCase()}{items.length !== 1 ? 's' : ''} 
          {isSearching && ` (${searchQuery})`}
          {isMultiProviderMode && ` • ${selectedProviders.length} provider${selectedProviders.length !== 1 ? 's' : ''}`}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <NodePoolsGroupsTable
          items={paginatedItems}
          selectedIds={selectedItemIds}
          onSelectionChange={onSelectionChange}
          onDelete={onDelete}
          isDeleting={isDeleting}
          page={page}
          pageSize={pageSize}
          total={total}
          onPageChange={onPageChange}
          onPageSizeChange={onPageSizeChange}
          showProviderColumn={isMultiProviderMode}
          workspaceId={workspaceId}
          credentialId={credentialId}
        />
      </CardContent>
    </Card>
  );
}



