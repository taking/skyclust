/**
 * Azure AKS Cluster Detail Tab Component
 * Azure AKS 클러스터 상세 정보 탭 컴포넌트
 */

'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import type { AzureCluster } from '@/lib/types';

interface AzureClusterDetailTabProps {
  clusterName: string;
  cluster: AzureCluster;
}

export function AzureClusterDetailTab({ clusterName, cluster }: AzureClusterDetailTabProps) {
  return (
    <div className="space-y-4">
      {/* Network Profile */}
      {cluster.network_profile && (
        <Card>
          <CardHeader>
            <CardTitle>Network Profile</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Network Plugin:</span>
                  <Badge variant="outline">{cluster.network_profile.network_plugin || '-'}</Badge>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Network Policy:</span>
                  <Badge variant="outline">{cluster.network_profile.network_policy || '-'}</Badge>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Load Balancer SKU:</span>
                  <Badge variant="outline">{cluster.network_profile.load_balancer_sku || '-'}</Badge>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Network Mode:</span>
                  <Badge variant="outline">{cluster.network_profile.network_mode || '-'}</Badge>
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Pod CIDR:</span>
                  <span className="text-sm font-medium">{cluster.network_profile.pod_cidr || '-'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Service CIDR:</span>
                  <span className="text-sm font-medium">{cluster.network_profile.service_cidr || '-'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">DNS Service IP:</span>
                  <span className="text-sm font-medium">{cluster.network_profile.dns_service_ip || '-'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Docker Bridge CIDR:</span>
                  <span className="text-sm font-medium">{cluster.network_profile.docker_bridge_cidr || '-'}</span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Service Principal */}
      {cluster.service_principal && (
        <Card>
          <CardHeader>
            <CardTitle>Service Principal</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Client ID:</span>
              <span className="text-sm font-mono text-xs">{cluster.service_principal.client_id || '-'}</span>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Security Configuration */}
      <Card>
        <CardHeader>
          <CardTitle>Security Configuration</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">RBAC:</span>
              <Badge variant={cluster.enable_rbac ? 'default' : 'secondary'}>
                {cluster.enable_rbac ? 'Enabled' : 'Disabled'}
              </Badge>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Pod Security Policy:</span>
              <Badge variant={cluster.enable_pod_security_policy ? 'default' : 'secondary'}>
                {cluster.enable_pod_security_policy ? 'Enabled' : 'Disabled'}
              </Badge>
            </div>
          </div>
          {cluster.api_server_authorized_ip_ranges && cluster.api_server_authorized_ip_ranges.length > 0 && (
            <div className="mt-4">
              <span className="text-sm text-muted-foreground">API Server Authorized IP Ranges:</span>
              <div className="mt-1 space-y-1">
                {cluster.api_server_authorized_ip_ranges.map((ipRange, index) => (
                  <div key={index} className="text-sm font-mono text-xs bg-muted px-2 py-1 rounded">
                    {ipRange}
                  </div>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Addon Profiles */}
      {cluster.addon_profiles && Object.keys(cluster.addon_profiles).length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Addon Profiles</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {Object.entries(cluster.addon_profiles).map(([key, value]) => {
              const enabled = typeof value === 'object' && value !== null && 'enabled' in value
                ? (value as { enabled: boolean }).enabled
                : false;
              return (
                <div key={key} className="flex justify-between">
                  <span className="text-sm text-muted-foreground capitalize">{key.replace(/_/g, ' ')}:</span>
                  <Badge variant={enabled ? 'default' : 'secondary'}>
                    {enabled ? 'Enabled' : 'Disabled'}
                  </Badge>
                </div>
              );
            })}
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
            <span className="text-sm text-muted-foreground">Resource Group:</span>
            <span className="text-sm font-medium">{cluster.resource_group || '-'}</span>
          </div>
        </CardContent>
      </Card>

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
    </div>
  );
}

