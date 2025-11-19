/**
 * Node Pool/Group Row Component
 * 개별 Node Pool 또는 Node Group 행 렌더링
 */

'use client';

import * as React from 'react';
import { useRouter } from 'next/navigation';
import { TableCell } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { Settings, Trash2 } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { ProviderBadge } from './provider-badge';
import { buildCredentialResourceDetailPath } from '@/lib/routing/helpers';
import type { NodePool, NodeGroup, CloudProvider } from '@/lib/types/kubernetes';

type NodePoolOrGroup = (NodePool | NodeGroup) & {
  cluster_name: string;
  cluster_id: string;
  provider: CloudProvider;
  resource_type: 'node-pool' | 'node-group';
  credential_id?: string;
};

interface NodePoolGroupRowProps {
  item: NodePoolOrGroup;
  isSelected: boolean;
  onSelect: (checked: boolean) => void;
  onDelete: (name: string, clusterName: string, region: string) => void;
  isDeleting?: boolean;
  showProvider?: boolean;
  workspaceId: string;
  credentialId: string;
}

function getStatusVariant(status: string): 'default' | 'secondary' | 'destructive' {
  if (status === 'RUNNING' || status === 'ACTIVE') return 'default';
  if (status === 'CREATING' || status === 'UPDATING') return 'secondary';
  return 'destructive';
}

export function NodePoolGroupRow({
  item,
  isSelected,
  onSelect,
  onDelete,
  isDeleting = false,
  showProvider = false,
  workspaceId,
  credentialId,
}: NodePoolGroupRowProps) {
  const router = useRouter();
  const { t } = useTranslation();

  const handleRowClick = React.useCallback((e?: React.MouseEvent) => {
    if (e) {
      e.stopPropagation();
    }
    const path = buildCredentialResourceDetailPath(
      workspaceId,
      credentialId,
      'k8s',
      item.resource_type === 'node-group' ? 'node-groups' : 'node-pools',
      item.name,
      { region: item.region }
    );
    router.push(path);
  }, [workspaceId, credentialId, item, router]);

  const handleDelete = React.useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    onDelete(item.name, item.cluster_name, item.region);
  }, [item, onDelete]);

  const nodeCount = 'node_count' in item ? item.node_count : item.desired_size;
  const minNodes = 'min_nodes' in item ? item.min_nodes : item.min_size;
  const maxNodes = 'max_nodes' in item ? item.max_nodes : item.max_size;
  const instanceType = item.instance_type || (item.instance_types?.join(', ') || '-');

  return (
    <>
      <TableCell>
        <Checkbox
          checked={isSelected}
          onCheckedChange={onSelect}
          onClick={(e) => e.stopPropagation()}
        />
      </TableCell>
      {showProvider && (
        <TableCell>
          <ProviderBadge provider={item.provider} />
        </TableCell>
      )}
      <TableCell className="font-medium">
        <Button
          variant="link"
          className="p-0 h-auto font-medium"
          onClick={(e) => handleRowClick(e)}
        >
          {item.cluster_name}
        </Button>
      </TableCell>
      <TableCell className="font-medium">
        <Button
          variant="link"
          className="p-0 h-auto font-medium"
          onClick={(e) => handleRowClick(e)}
        >
          {item.name}
        </Button>
      </TableCell>
      <TableCell>{instanceType}</TableCell>
      <TableCell>{nodeCount ?? '-'}</TableCell>
      <TableCell>{minNodes}/{maxNodes}</TableCell>
      <TableCell>
        <Badge variant={getStatusVariant(item.status)}>
          {item.status}
        </Badge>
      </TableCell>
      <TableCell>{item.region}</TableCell>
      <TableCell>
        <div className="flex items-center space-x-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={(e) => handleRowClick(e)}
            aria-label={t('common.settings') || 'Settings'}
          >
            <Settings className="h-4 w-4" aria-hidden="true" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleDelete}
            disabled={isDeleting}
            aria-label={t('common.delete') || 'Delete'}
          >
            <Trash2 className="h-4 w-4 text-destructive" aria-hidden="true" />
          </Button>
        </div>
      </TableCell>
    </>
  );
}

