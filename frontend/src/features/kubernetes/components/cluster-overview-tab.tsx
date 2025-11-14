/**
 * Cluster Overview Tab Component
 * 클러스터 개요 탭 컴포넌트 - Provider별 상세 정보 표시
 */

'use client';

import { ClusterMetricsChart } from './cluster-metrics-chart';
import { AWSClusterDetailTab } from './aws-cluster-detail-tab';
import { GCPClusterDetailTab } from './gcp-cluster-detail-tab';
import { AzureClusterDetailTab } from './azure-cluster-detail-tab';
import type { ProviderCluster, BaseCluster } from '@/lib/types';
import { isAWSCluster, isGCPCluster, isAzureCluster } from '@/lib/types';

interface ClusterOverviewTabProps {
  clusterName: string;
  cluster: ProviderCluster | BaseCluster | undefined;
}

export function ClusterOverviewTab({ clusterName, cluster }: ClusterOverviewTabProps) {
  if (!cluster) {
    return (
      <div className="space-y-4">
        <p className="text-sm text-muted-foreground">Loading cluster information...</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Cluster Metrics Preview */}
      <ClusterMetricsChart clusterName={clusterName} />

      {/* Provider-specific detail tabs */}
      {isAWSCluster(cluster) && (
        <AWSClusterDetailTab clusterName={clusterName} cluster={cluster} />
      )}

      {isGCPCluster(cluster) && (
        <GCPClusterDetailTab clusterName={clusterName} cluster={cluster} />
      )}

      {isAzureCluster(cluster) && (
        <AzureClusterDetailTab clusterName={clusterName} cluster={cluster} />
      )}

      {/* Fallback for unknown provider or base cluster */}
      {!isAWSCluster(cluster) && !isGCPCluster(cluster) && !isAzureCluster(cluster) && (
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">Provider-specific details not available. Showing basic information.</p>
        </div>
      )}
    </div>
  );
}

