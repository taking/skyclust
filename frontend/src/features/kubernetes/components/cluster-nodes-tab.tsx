/**
 * Cluster Nodes Tab Component
 * 클러스터 노드 탭 컴포넌트
 */

'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { Server } from 'lucide-react';
import type { Node } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

interface ClusterNodesTabProps {
  nodes: Node[];
  isLoading: boolean;
}

export function ClusterNodesTab({ nodes, isLoading }: ClusterNodesTabProps) {
  const { t } = useTranslation();

  if (isLoading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
        </CardContent>
      </Card>
    );
  }

  if (nodes.length === 0) {
    return (
      <ResourceEmptyState
        resourceName={t('kubernetes.nodes')}
        icon={Server}
        description={t('kubernetes.noNodesFoundForCluster')}
        withCard={true}
      />
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('kubernetes.nodes')}</CardTitle>
        <CardDescription>{t('kubernetes.nodeCount', { count: nodes.length })}</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('common.name')}</TableHead>
              <TableHead>{t('common.instanceType')}</TableHead>
              <TableHead>{t('common.zone')}</TableHead>
              <TableHead>{t('common.status')}</TableHead>
              <TableHead>{t('common.privateIP')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {nodes.map((node) => (
              <TableRow key={node.id || node.name}>
                <TableCell className="font-medium">{node.name}</TableCell>
                <TableCell>{node.instance_type}</TableCell>
                <TableCell>{node.zone || '-'}</TableCell>
                <TableCell>
                  <Badge variant={node.status === 'Ready' ? 'default' : 'secondary'}>
                    {node.status}
                  </Badge>
                </TableCell>
                <TableCell>{node.private_ip || '-'}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

