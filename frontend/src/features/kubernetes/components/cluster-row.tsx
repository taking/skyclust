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
import { Download, Settings, Trash2, Copy, Check } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import type { KubernetesCluster, CloudProvider } from '@/lib/types';

interface ClusterRowProps {
  cluster: KubernetesCluster;
  provider?: CloudProvider;
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

function formatDate(dateString?: string): string {
  if (!dateString) return '-';
  try {
    return new Date(dateString).toLocaleDateString();
  } catch {
    return '-';
  }
}

function ClusterRowComponent({
  cluster,
  provider,
  isSelected,
  onSelect,
  onDelete,
  onDownloadKubeconfig,
  isDeleting = false,
  isDownloading = false,
}: ClusterRowProps) {
  const router = useRouter();
  const { success } = useToast();
  const [copiedEndpoint, setCopiedEndpoint] = React.useState(false);
  const isAzure = provider === 'azure';

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
          onClick={() => router.push(`/kubernetes/clusters/${cluster.name}`)}
        >
          {cluster.name}
        </Button>
      </TableCell>
      <TableCell>
        <Badge variant={getStatusVariant(cluster.status)}>
          {cluster.status}
        </Badge>
      </TableCell>
      <TableCell>{cluster.version || '-'}</TableCell>
      <TableCell>{cluster.region || '-'}</TableCell>
      {isAzure && (
        <TableCell>{cluster.resource_group || '-'}</TableCell>
      )}
      <TableCell>
        {cluster.endpoint ? (
          <div className="flex items-center gap-2 w-full">
            <a
              href={cluster.endpoint.startsWith('http://') || cluster.endpoint.startsWith('https://') 
                ? cluster.endpoint 
                : `https://${cluster.endpoint}`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-blue-600 hover:underline truncate flex-1"
              title={cluster.endpoint}
            >
              {cluster.endpoint.length > 40 
                ? `${cluster.endpoint.substring(0, 40)}...`
                : cluster.endpoint}
            </a>
            <Button
              variant="ghost"
              size="sm"
              onClick={async (e) => {
                e.preventDefault();
                e.stopPropagation();
                if (cluster.endpoint) {
                  try {
                    await navigator.clipboard.writeText(cluster.endpoint);
                    setCopiedEndpoint(true);
                    success('Endpoint copied to clipboard');
                    setTimeout(() => {
                      setCopiedEndpoint(false);
                    }, 2000);
                  } catch (error) {
                    console.error('Failed to copy endpoint:', error);
                  }
                }
              }}
              className="shrink-0"
              aria-label="Copy endpoint"
            >
              {copiedEndpoint ? (
                <Check className="h-4 w-4 text-green-600" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
            </Button>
          </div>
        ) : (
          <span className="text-muted-foreground">-</span>
        )}
      </TableCell>
      <TableCell>
        {formatDate(cluster.created_at)}
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
            onClick={() => router.push(`/kubernetes/clusters/${cluster.name}`)}
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
    prevProps.cluster.version === nextProps.cluster.version &&
    prevProps.cluster.region === nextProps.cluster.region &&
    prevProps.cluster.endpoint === nextProps.cluster.endpoint &&
    prevProps.cluster.resource_group === nextProps.cluster.resource_group &&
    prevProps.cluster.created_at === nextProps.cluster.created_at &&
    prevProps.provider === nextProps.provider &&
    prevProps.isSelected === nextProps.isSelected &&
    prevProps.isDeleting === nextProps.isDeleting &&
    prevProps.isDownloading === nextProps.isDownloading
  );
});

