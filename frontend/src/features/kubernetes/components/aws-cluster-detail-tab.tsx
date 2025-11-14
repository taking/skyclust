/**
 * AWS EKS Cluster Detail Tab Component
 * AWS EKS 클러스터 상세 정보 탭 컴포넌트
 */

'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import type { AWSCluster } from '@/lib/types';

interface AWSClusterDetailTabProps {
  clusterName: string;
  cluster: AWSCluster;
}

export function AWSClusterDetailTab({ clusterName, cluster }: AWSClusterDetailTabProps) {
  return (
    <div className="space-y-4">
      {/* VPC Configuration */}
      {cluster.resources_vpc_config && (
        <Card>
          <CardHeader>
            <CardTitle>VPC Configuration</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">VPC ID:</span>
                  <span className="text-sm font-medium">{cluster.resources_vpc_config.vpc_id || '-'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Cluster Security Group:</span>
                  <span className="text-sm font-medium">{cluster.resources_vpc_config.cluster_security_group_id || '-'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Endpoint Public Access:</span>
                  <Badge variant={cluster.resources_vpc_config.endpoint_public_access ? 'default' : 'secondary'}>
                    {cluster.resources_vpc_config.endpoint_public_access ? 'Enabled' : 'Disabled'}
                  </Badge>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Endpoint Private Access:</span>
                  <Badge variant={cluster.resources_vpc_config.endpoint_private_access ? 'default' : 'secondary'}>
                    {cluster.resources_vpc_config.endpoint_private_access ? 'Enabled' : 'Disabled'}
                  </Badge>
                </div>
              </div>
              <div className="space-y-2">
                <div>
                  <span className="text-sm text-muted-foreground">Subnet IDs:</span>
                  <div className="mt-1 space-y-1">
                    {cluster.resources_vpc_config.subnet_ids.length > 0 ? (
                      cluster.resources_vpc_config.subnet_ids.map((subnetId, index) => (
                        <div key={index} className="text-sm font-mono text-xs bg-muted px-2 py-1 rounded">
                          {subnetId}
                        </div>
                      ))
                    ) : (
                      <span className="text-sm text-muted-foreground">-</span>
                    )}
                  </div>
                </div>
                <div>
                  <span className="text-sm text-muted-foreground">Security Group IDs:</span>
                  <div className="mt-1 space-y-1">
                    {cluster.resources_vpc_config.security_group_ids.length > 0 ? (
                      cluster.resources_vpc_config.security_group_ids.map((sgId, index) => (
                        <div key={index} className="text-sm font-mono text-xs bg-muted px-2 py-1 rounded">
                          {sgId}
                        </div>
                      ))
                    ) : (
                      <span className="text-sm text-muted-foreground">-</span>
                    )}
                  </div>
                </div>
                {cluster.resources_vpc_config.public_access_cidrs.length > 0 && (
                  <div>
                    <span className="text-sm text-muted-foreground">Public Access CIDRs:</span>
                    <div className="mt-1 space-y-1">
                      {cluster.resources_vpc_config.public_access_cidrs.map((cidr, index) => (
                        <div key={index} className="text-sm font-mono text-xs bg-muted px-2 py-1 rounded">
                          {cidr}
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Kubernetes Network Configuration */}
      {cluster.kubernetes_network_config && (
        <Card>
          <CardHeader>
            <CardTitle>Kubernetes Network Configuration</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Service IPv4 CIDR:</span>
                  <span className="text-sm font-medium">{cluster.kubernetes_network_config.service_ipv4_cidr || '-'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Service IPv6 CIDR:</span>
                  <span className="text-sm font-medium">{cluster.kubernetes_network_config.service_ipv6_cidr || '-'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">IP Family:</span>
                  <Badge variant="outline">{cluster.kubernetes_network_config.ip_family || '-'}</Badge>
                </div>
              </div>
              {cluster.kubernetes_network_config.elastic_load_balancing && (
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-sm text-muted-foreground">Elastic Load Balancing:</span>
                    <Badge variant={cluster.kubernetes_network_config.elastic_load_balancing.enabled ? 'default' : 'secondary'}>
                      {cluster.kubernetes_network_config.elastic_load_balancing.enabled ? 'Enabled' : 'Disabled'}
                    </Badge>
                  </div>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Access Configuration */}
      {cluster.access_config && (
        <Card>
          <CardHeader>
            <CardTitle>Access Configuration</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Authentication Mode:</span>
              <Badge variant="outline">{cluster.access_config.authentication_mode || '-'}</Badge>
            </div>
            {cluster.access_config.bootstrap_cluster_creator_admin_permissions !== undefined && (
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Bootstrap Cluster Creator Admin:</span>
                <Badge variant={cluster.access_config.bootstrap_cluster_creator_admin_permissions ? 'default' : 'secondary'}>
                  {cluster.access_config.bootstrap_cluster_creator_admin_permissions ? 'Enabled' : 'Disabled'}
                </Badge>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Cluster Information */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <CardTitle>Cluster Information</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Role ARN:</span>
              <span className="text-sm font-mono text-xs">{cluster.role_arn || '-'}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Platform Version:</span>
              <span className="text-sm font-medium">{cluster.platform_version || '-'}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Deletion Protection:</span>
              <Badge variant={cluster.deletion_protection ? 'destructive' : 'secondary'}>
                {cluster.deletion_protection ? 'Enabled' : 'Disabled'}
              </Badge>
            </div>
          </CardContent>
        </Card>

        {cluster.upgrade_policy && (
          <Card>
            <CardHeader>
              <CardTitle>Upgrade Policy</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Support Type:</span>
                <Badge variant="outline">{cluster.upgrade_policy.support_type || '-'}</Badge>
              </div>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Common Network Configuration */}
      {cluster.network_config && (
        <Card>
          <CardHeader>
            <CardTitle>Network Configuration</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">VPC ID:</span>
              <span className="text-sm font-medium">{cluster.network_config.vpc_id || '-'}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Subnet ID:</span>
              <span className="text-sm font-medium">{cluster.network_config.subnet_id || '-'}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Pod CIDR:</span>
              <span className="text-sm font-medium">{cluster.network_config.pod_cidr || '-'}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Service CIDR:</span>
              <span className="text-sm font-medium">{cluster.network_config.service_cidr || '-'}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Private Endpoint:</span>
              <Badge variant={cluster.network_config.private_endpoint ? 'default' : 'secondary'}>
                {cluster.network_config.private_endpoint ? 'Enabled' : 'Disabled'}
              </Badge>
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
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Workload Identity:</span>
              <Badge variant={cluster.security_config.workload_identity ? 'default' : 'secondary'}>
                {cluster.security_config.workload_identity ? 'Enabled' : 'Disabled'}
              </Badge>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Network Policy:</span>
              <Badge variant={cluster.security_config.network_policy ? 'default' : 'secondary'}>
                {cluster.security_config.network_policy ? 'Enabled' : 'Disabled'}
              </Badge>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

