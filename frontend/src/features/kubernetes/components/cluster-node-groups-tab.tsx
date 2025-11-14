/**
 * Cluster Node Groups Tab Component
 * 클러스터 노드 그룹 탭 컴포넌트 (EKS용)
 */

'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import { Server, Trash2 } from 'lucide-react';
import type { NodeGroup } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

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
  const { t } = useTranslation();
  const [deleteDialogState, setDeleteDialogState] = useState<{
    open: boolean;
    nodeGroupName: string | null;
  }>({
    open: false,
    nodeGroupName: null,
  });

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
        resourceName={t('kubernetes.nodeGroups')}
        icon={Server}
        onCreateClick={onCreateClick}
        withCard={true}
      />
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('kubernetes.nodeGroups')}</CardTitle>
        <CardDescription>{t('kubernetes.nodeGroupCount', { count: nodeGroups.length })}</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('common.name')}</TableHead>
              <TableHead>{t('common.instanceType')}</TableHead>
              <TableHead>{t('common.nodes')}</TableHead>
              <TableHead>{t('common.minMax')}</TableHead>
              <TableHead>{t('common.status')}</TableHead>
              <TableHead>{t('common.actions')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {nodeGroups.map((ng) => (
              <TableRow key={ng.id || ng.name}>
                <TableCell className="font-medium">
                  <Link
                    href={`/kubernetes/node-groups/${ng.name}?cluster=${ng.cluster_name}`}
                    className="text-blue-600 hover:text-blue-800 hover:underline"
                  >
                    {ng.name}
                  </Link>
                </TableCell>
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
                      setDeleteDialogState({ open: true, nodeGroupName: ng.name });
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

      {/* Delete Confirmation Dialog */}
      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => setDeleteDialogState({ ...deleteDialogState, open })}
        onConfirm={() => {
          if (deleteDialogState.nodeGroupName) {
            onDeleteClick(deleteDialogState.nodeGroupName);
            setDeleteDialogState({ open: false, nodeGroupName: null });
          }
        }}
        title={t('kubernetes.deleteNodeGroup')}
        description={deleteDialogState.nodeGroupName ? t('kubernetes.confirmDeleteNodeGroup', { nodeGroupName: deleteDialogState.nodeGroupName }) : ''}
        isLoading={isDeleting}
        resourceName={deleteDialogState.nodeGroupName || undefined}
        resourceNameLabel="노드 그룹 이름"
      />
    </Card>
  );
}

