/**
 * Cluster Overview Header Component
 * 클러스터 개요 헤더 컴포넌트 (Status, Version, 공급자)
 */

'use client';

import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Label } from '@/components/ui/label';
import type { BaseCluster, CloudProvider } from '@/lib/types';

interface ClusterOverviewHeaderProps {
  cluster: BaseCluster | undefined;
  selectedProvider?: CloudProvider;
  isLoading?: boolean;
}

export function ClusterOverviewHeader({
  cluster,
  selectedProvider,
  isLoading = false,
}: ClusterOverviewHeaderProps) {
  if (isLoading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
        </CardContent>
      </Card>
    );
  }

  if (!cluster) {
    return null;
  }

  const getStatusVariant = (status: string) => {
    switch (status?.toUpperCase()) {
      case 'ACTIVE':
      case 'RUNNING':
        return 'default';
      case 'CREATING':
      case 'UPDATING':
      case 'PROVISIONING':
        return 'secondary';
      case 'DELETING':
      case 'STOPPING':
        return 'destructive';
      default:
        return 'outline';
    }
  };

  const getProviderLabel = (provider?: CloudProvider) => {
    if (!provider) return '-';
    return provider.toUpperCase();
  };

  return (
    <Card>
      <CardContent className="py-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {/* Status */}
          <div className="space-y-2">
            <Label className="text-sm text-muted-foreground">Status</Label>
            <div>
              <Badge variant={getStatusVariant(cluster.status)} className="text-sm">
                {cluster.status || '-'}
              </Badge>
            </div>
          </div>

          {/* Version */}
          <div className="space-y-2">
            <Label className="text-sm text-muted-foreground">Version (Kubernetes Version)</Label>
            <p className="text-lg font-semibold">{cluster.version || '-'}</p>
          </div>

          {/* Provider */}
          <div className="space-y-2">
            <Label className="text-sm text-muted-foreground">공급자</Label>
            <p className="text-lg font-semibold">{getProviderLabel(selectedProvider)}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

