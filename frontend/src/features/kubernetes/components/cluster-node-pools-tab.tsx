/**
 * Cluster Node Pools Tab Component
 * 클러스터 노드 풀 탭 컴포넌트
 */

'use client';

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { Server, Plus, RefreshCw, Trash2 } from 'lucide-react';
import type { NodePool } from '@/lib/types';

interface ClusterNodePoolsTabProps {
  nodePools: NodePool[];
  isLoading: boolean;
  onCreateClick: () => void;
  onScaleClick: (nodePoolName: string, currentNodes: number) => void;
  onDeleteClick: (nodePoolName: string) => void;
  isDeleting: boolean;
}

export function ClusterNodePoolsTab({
  nodePools,
  isLoading,
  onCreateClick,
  onScaleClick,
  onDeleteClick,
  isDeleting,
}: ClusterNodePoolsTabProps) {
  if (isLoading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
        </CardContent>
      </Card>
    );
  }

  if (nodePools.length === 0) {
    return (
      <ResourceEmptyState
        resourceName="Node Pools"
        icon={Server}
        onCreateClick={onCreateClick}
        withCard={true}
      />
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Node Pools</CardTitle>
        <CardDescription>{nodePools.length} node pool{nodePools.length !== 1 ? 's' : ''}</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Instance Type</TableHead>
              <TableHead>Nodes</TableHead>
              <TableHead>Min/Max</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {nodePools.map((np) => (
              <TableRow key={np.id || np.name}>
                <TableCell className="font-medium">{np.name}</TableCell>
                <TableCell>{np.instance_type}</TableCell>
                <TableCell>{np.node_count}</TableCell>
                <TableCell>{np.min_nodes}/{np.max_nodes}</TableCell>
                <TableCell>
                  <Badge variant={np.status === 'RUNNING' ? 'default' : 'secondary'}>
                    {np.status}
                  </Badge>
                </TableCell>
                <TableCell>
                  <div className="flex items-center space-x-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => onScaleClick(np.name, np.node_count)}
                      disabled={isDeleting}
                    >
                      <RefreshCw className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => {
                        if (confirm(`Delete node pool ${np.name}?`)) {
                          onDeleteClick(np.name);
                        }
                      }}
                      disabled={isDeleting}
                    >
                      <Trash2 className="h-4 w-4 text-red-600" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

