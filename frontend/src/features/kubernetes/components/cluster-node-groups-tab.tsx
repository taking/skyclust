/**
 * Cluster Node Groups Tab Component
 * 클러스터 노드 그룹 탭 컴포넌트 (EKS용)
 */

'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { Server, Plus, Trash2 } from 'lucide-react';
import type { NodeGroup } from '@/lib/types';

interface ClusterNodeGroupsTabProps {
  nodeGroups: NodeGroup[];
  isLoading: boolean;
  onCreateClick: () => void;
  onDeleteClick: (nodeGroupName: string) => void;
  isDeleting: boolean;
}

export function ClusterNodeGroupsTab({
  nodeGroups,
  isLoading,
  onCreateClick,
  onDeleteClick,
  isDeleting,
}: ClusterNodeGroupsTabProps) {
  if (isLoading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
        </CardContent>
      </Card>
    );
  }

  if (nodeGroups.length === 0) {
    return (
      <ResourceEmptyState
        resourceName="Node Groups"
        icon={Server}
        onCreateClick={onCreateClick}
        withCard={true}
      />
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Node Groups</CardTitle>
        <CardDescription>{nodeGroups.length} node group{nodeGroups.length !== 1 ? 's' : ''}</CardDescription>
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
            {nodeGroups.map((ng) => (
              <TableRow key={ng.id || ng.name}>
                <TableCell className="font-medium">{ng.name}</TableCell>
                <TableCell>{ng.instance_type}</TableCell>
                <TableCell>{ng.node_count}</TableCell>
                <TableCell>{ng.min_size}/{ng.max_size}</TableCell>
                <TableCell>
                  <Badge variant={ng.status === 'ACTIVE' ? 'default' : 'secondary'}>
                    {ng.status}
                  </Badge>
                </TableCell>
                <TableCell>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => {
                      if (confirm(`Delete node group ${ng.name}?`)) {
                        onDeleteClick(ng.name);
                      }
                    }}
                    disabled={isDeleting}
                  >
                    <Trash2 className="h-4 w-4 text-red-600" />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

