/**
 * Cluster Overview Tab Component
 * 클러스터 개요 탭 컴포넌트
 */

'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Label } from '@/components/ui/label';
import { ClusterMetricsChart } from './cluster-metrics-chart';
import type { Cluster } from '@/lib/types';

interface ClusterOverviewTabProps {
  clusterName: string;
  cluster: Cluster | undefined;
}

export function ClusterOverviewTab({ clusterName, cluster }: ClusterOverviewTabProps) {
  return (
    <div className="space-y-4">
      {/* Cluster Metrics Preview */}
      {cluster && (
        <ClusterMetricsChart clusterName={clusterName} />
      )}
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <CardTitle>Network Configuration</CardTitle>
          </CardHeader>
          <CardContent>
            {cluster?.network_config ? (
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-sm text-gray-500">VPC ID:</span>
                  <span className="text-sm">{cluster.network_config.vpc_id || '-'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-500">Subnet ID:</span>
                  <span className="text-sm">{cluster.network_config.subnet_id || '-'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-500">Pod CIDR:</span>
                  <span className="text-sm">{cluster.network_config.pod_cidr || '-'}</span>
                </div>
              </div>
            ) : (
              <p className="text-sm text-gray-500">No network configuration available</p>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Security Configuration</CardTitle>
          </CardHeader>
          <CardContent>
            {cluster?.security_config ? (
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-sm text-gray-500">Workload Identity:</span>
                  <Badge variant={cluster.security_config.workload_identity ? 'default' : 'secondary'}>
                    {cluster.security_config.workload_identity ? 'Enabled' : 'Disabled'}
                  </Badge>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-500">Network Policy:</span>
                  <Badge variant={cluster.security_config.network_policy ? 'default' : 'secondary'}>
                    {cluster.security_config.network_policy ? 'Enabled' : 'Disabled'}
                  </Badge>
                </div>
              </div>
            ) : (
              <p className="text-sm text-gray-500">No security configuration available</p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

