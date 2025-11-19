/**
 * Nodes Page Content
 * 
 * 노드 페이지의 메인 콘텐츠 컴포넌트
 * 테이블만 담당합니다.
 */

'use client';

import * as React from 'react';
import dynamic from 'next/dynamic';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { TableSkeleton } from '@/components/ui/table-skeleton';
import type { Node, CloudProvider } from '@/lib/types';

const NodesTable = dynamic(
  () => import('@/features/kubernetes').then(mod => ({ default: mod.NodesTable })),
  { 
    ssr: false,
    loading: () => <TableSkeleton columns={7} rows={5} showCheckbox={true} />,
  }
);

export interface NodesPageContentProps {
  nodes: Array<Node & { cluster_name: string; cluster_id: string; provider?: CloudProvider; credential_id?: string }>;
  filteredNodes: Array<Node & { cluster_name: string; cluster_id: string; provider?: CloudProvider; credential_id?: string }>;
  paginatedNodes: Array<Node & { cluster_name: string; cluster_id: string; provider?: CloudProvider; credential_id?: string }>;
  selectedProvider: CloudProvider | undefined;
  selectedNodeIds: string[];
  onSelectionChange: (ids: string[] | ((prev: string[]) => string[])) => void;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
  isSearching: boolean;
  searchQuery: string;
  isMultiProviderMode: boolean;
  selectedProviders: CloudProvider[];
}

export function NodesPageContent({
  nodes,
  filteredNodes,
  paginatedNodes,
  selectedProvider,
  selectedNodeIds,
  onSelectionChange,
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
  isSearching,
  searchQuery,
  isMultiProviderMode,
  selectedProviders,
}: NodesPageContentProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Nodes</CardTitle>
        <CardDescription>
          {filteredNodes.length} of {nodes.length} node{nodes.length !== 1 ? 's' : ''} 
          {isSearching && ` (${searchQuery})`}
          {isMultiProviderMode && ` • ${selectedProviders.length} provider${selectedProviders.length !== 1 ? 's' : ''}`}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <NodesTable
          nodes={paginatedNodes}
          provider={selectedProvider}
          selectedIds={selectedNodeIds}
          onSelectionChange={onSelectionChange}
          page={page}
          pageSize={pageSize}
          total={total}
          onPageChange={onPageChange}
          onPageSizeChange={onPageSizeChange}
          showProviderColumn={isMultiProviderMode}
        />
      </CardContent>
    </Card>
  );
}

