/**
 * GCP GKE Cluster Detail Tab Component
 * GCP GKE 클러스터 상세 정보 탭 컴포넌트
 */

'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import type { GCPCluster } from '@/lib/types';

interface GCPClusterDetailTabProps {
  clusterName: string;
  cluster: GCPCluster;
}

export function GCPClusterDetailTab({ clusterName, cluster }: GCPClusterDetailTabProps) {
  return (
    <div className="space-y-4">
      {/* Network Configuration */}
      {cluster.network_config && (
        <Card>
          <CardHeader>
            <CardTitle>Network Configuration</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Network:</span>
                  <span className="text-sm font-medium">{cluster.network_config.network || '-'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Subnetwork:</span>
                  <span className="text-sm font-medium">{cluster.network_config.subnetwork || '-'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Pod CIDR:</span>
                  <span className="text-sm font-medium">{cluster.network_config.pod_cidr || '-'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Service CIDR:</span>
                  <span className="text-sm font-medium">{cluster.network_config.service_cidr || '-'}</span>
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Private Nodes:</span>
                  <Badge variant={cluster.network_config.private_nodes ? 'default' : 'secondary'}>
                    {cluster.network_config.private_nodes ? 'Enabled' : 'Disabled'}
                  </Badge>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Private Endpoint:</span>
                  <Badge variant={cluster.network_config.private_endpoint ? 'default' : 'secondary'}>
                    {cluster.network_config.private_endpoint ? 'Enabled' : 'Disabled'}
                  </Badge>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Private Cluster Configuration */}
      {cluster.private_cluster_config && (
        <Card>
          <CardHeader>
            <CardTitle>Private Cluster Configuration</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Enable Private Nodes:</span>
                <Badge variant={cluster.private_cluster_config.enable_private_nodes ? 'default' : 'secondary'}>
                  {cluster.private_cluster_config.enable_private_nodes ? 'Enabled' : 'Disabled'}
                </Badge>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Enable Private Endpoint:</span>
                <Badge variant={cluster.private_cluster_config.enable_private_endpoint ? 'default' : 'secondary'}>
                  {cluster.private_cluster_config.enable_private_endpoint ? 'Enabled' : 'Disabled'}
                </Badge>
              </div>
              {cluster.private_cluster_config.master_ipv4_cidr && (
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Master IPv4 CIDR:</span>
                  <span className="text-sm font-mono text-xs">{cluster.private_cluster_config.master_ipv4_cidr}</span>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Master Authorized Networks */}
      {cluster.master_authorized_networks_config && cluster.master_authorized_networks_config.enabled && (
        <Card>
          <CardHeader>
            <CardTitle>Master Authorized Networks</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between mb-2">
              <span className="text-sm text-muted-foreground">Enabled:</span>
              <Badge variant="default">Enabled</Badge>
            </div>
            {cluster.master_authorized_networks_config.cidr_blocks && cluster.master_authorized_networks_config.cidr_blocks.length > 0 && (
              <div>
                <span className="text-sm text-muted-foreground">CIDR Blocks:</span>
                <div className="mt-1 space-y-1">
                  {cluster.master_authorized_networks_config.cidr_blocks.map((cidr, index) => (
                    <div key={index} className="text-sm font-mono text-xs bg-muted px-2 py-1 rounded">
                      {cidr}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Workload Identity Configuration */}
      {cluster.workload_identity_config && (
        <Card>
          <CardHeader>
            <CardTitle>Workload Identity Configuration</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Workload Pool:</span>
              <span className="text-sm font-medium">{cluster.workload_identity_config.workload_pool || '-'}</span>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Security Configuration */}
      {cluster.security_config && (
        <Card>
          <CardHeader>
            <CardTitle>Security Configuration</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Workload Identity:</span>
                <Badge variant={cluster.security_config.workload_identity ? 'default' : 'secondary'}>
                  {cluster.security_config.workload_identity ? 'Enabled' : 'Disabled'}
                </Badge>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Binary Authorization:</span>
                <Badge variant={cluster.security_config.binary_authorization ? 'default' : 'secondary'}>
                  {cluster.security_config.binary_authorization ? 'Enabled' : 'Disabled'}
                </Badge>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Network Policy:</span>
                <Badge variant={cluster.security_config.network_policy ? 'default' : 'secondary'}>
                  {cluster.security_config.network_policy ? 'Enabled' : 'Disabled'}
                </Badge>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Pod Security Policy:</span>
                <Badge variant={cluster.security_config.pod_security_policy ? 'default' : 'secondary'}>
                  {cluster.security_config.pod_security_policy ? 'Enabled' : 'Disabled'}
                </Badge>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Cluster Information */}
      <Card>
        <CardHeader>
          <CardTitle>Cluster Information</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          <div className="flex justify-between">
            <span className="text-sm text-muted-foreground">Project ID:</span>
            <span className="text-sm font-medium">{cluster.project_id || '-'}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-sm text-muted-foreground">Zone:</span>
            <span className="text-sm font-medium">{cluster.zone || '-'}</span>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

