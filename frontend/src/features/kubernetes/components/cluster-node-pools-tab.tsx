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
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import { Server, RefreshCw, Trash2 } from 'lucide-react';
import type { NodePool } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

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
  const { t } = useTranslation();
  const [deleteDialogState, setDeleteDialogState] = useState<{
    open: boolean;
    nodePoolName: string | null;
  }>({
    open: false,
    nodePoolName: null,
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

  if (nodePools.length === 0) {
    return (
      <ResourceEmptyState
        resourceName={t('kubernetes.nodePools')}
        icon={Server}
        onCreateClick={onCreateClick}
        withCard={true}
      />
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('kubernetes.nodePools')}</CardTitle>
        <CardDescription>{t('kubernetes.nodePoolCount', { count: nodePools.length })}</CardDescription>
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
                        setDeleteDialogState({ open: true, nodePoolName: np.name });
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

      {/* Delete Confirmation Dialog */}
      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => setDeleteDialogState({ ...deleteDialogState, open })}
        onConfirm={() => {
          if (deleteDialogState.nodePoolName) {
            onDeleteClick(deleteDialogState.nodePoolName);
            setDeleteDialogState({ open: false, nodePoolName: null });
          }
        }}
        title={t('kubernetes.deleteNodePool')}
        description={deleteDialogState.nodePoolName ? t('kubernetes.confirmDeleteNodePool', { nodePoolName: deleteDialogState.nodePoolName }) : ''}
        isLoading={isDeleting}
        resourceName={deleteDialogState.nodePoolName || undefined}
        resourceNameLabel="노드 풀 이름"
      />
    </Card>
  );
}

