/**
 * Cluster Row Component
 * 개별 클러스터 행 렌더링
 */

'use client';

import * as React from 'react';
import { useRouter } from 'next/navigation';
import { TableCell } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { Download, Settings, Trash2, ExternalLink } from 'lucide-react';
import type { KubernetesCluster } from '@/lib/types';

interface ClusterRowProps {
  cluster: KubernetesCluster;
  isSelected: boolean;
  onSelect: (checked: boolean) => void;
  onDelete: (clusterName: string, region: string) => void;
  onDownloadKubeconfig: (clusterName: string, region: string) => void;
  isDeleting?: boolean;
  isDownloading?: boolean;
}

function getStatusVariant(status: string): 'default' | 'secondary' | 'destructive' {
  if (status === 'ACTIVE') return 'default';
  if (status === 'CREATING') return 'secondary';
  return 'destructive';
}

function ClusterRowComponent({
  cluster,
  isSelected,
  onSelect,
  onDelete,
  onDownloadKubeconfig,
  isDeleting = false,
  isDownloading = false,
}: ClusterRowProps) {
  const router = useRouter();

  return (
    <>
      <TableCell>
        <Checkbox
          checked={isSelected}
          onCheckedChange={onSelect}
        />
      </TableCell>
      <TableCell className="font-medium">
        <Button
          variant="link"
          className="p-0 h-auto font-medium"
          onClick={() => router.push(`/kubernetes/${cluster.name}`)}
        >
          {cluster.name}
        </Button>
      </TableCell>
      <TableCell>{cluster.version}</TableCell>
      <TableCell>
        <Badge variant={getStatusVariant(cluster.status)}>
          {cluster.status}
        </Badge>
      </TableCell>
      <TableCell>{cluster.region}</TableCell>
      <TableCell>
        {cluster.endpoint ? (
          <a
            href={cluster.endpoint}
            target="_blank"
            rel="noopener noreferrer"
            className="text-blue-600 hover:underline flex items-center"
          >
            {cluster.endpoint.substring(0, 30)}...
            <ExternalLink className="ml-1 h-3 w-3" />
          </a>
        ) : (
          <span className="text-gray-400">-</span>
        )}
      </TableCell>
      <TableCell>
        <div className="flex items-center space-x-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onDownloadKubeconfig(cluster.name, cluster.region)}
            disabled={isDownloading}
            aria-label={`Download kubeconfig for ${cluster.name}`}
          >
            <Download className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => router.push(`/kubernetes/${cluster.name}`)}
            aria-label={`View details for ${cluster.name}`}
          >
            <Settings className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onDelete(cluster.name, cluster.region)}
            disabled={isDeleting}
            aria-label={`Delete cluster ${cluster.name}`}
          >
            <Trash2 className="h-4 w-4 text-red-600" />
          </Button>
        </div>
      </TableCell>
    </>
  );
}

export const ClusterRow = React.memo(ClusterRowComponent, (prevProps, nextProps) => {
  // Custom comparison for better memoization
  const clusterId = prevProps.cluster.id || prevProps.cluster.name;
  const nextClusterId = nextProps.cluster.id || nextProps.cluster.name;
  
  return (
    clusterId === nextClusterId &&
    prevProps.cluster.status === nextProps.cluster.status &&
    prevProps.isSelected === nextProps.isSelected &&
    prevProps.isDeleting === nextProps.isDeleting &&
    prevProps.isDownloading === nextProps.isDownloading
  );
});

