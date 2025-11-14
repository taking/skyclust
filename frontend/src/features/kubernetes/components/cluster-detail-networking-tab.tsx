/**
 * Cluster Detail Networking Tab Component
 * 클러스터 상세 네트워킹 탭 컴포넌트 (AWS EKS)
 */

'use client';

import { useRouter } from 'next/navigation';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ExternalLink } from 'lucide-react';
import type { ProviderCluster, BaseCluster } from '@/lib/types';
import { isAWSCluster } from '@/lib/types';
import { useCredentialContext } from '@/hooks/use-credential-context';

interface ClusterDetailNetworkingTabProps {
  cluster: ProviderCluster | BaseCluster;
  selectedProvider?: string;
}

export function ClusterDetailNetworkingTab({
  cluster,
  selectedProvider = 'aws',
}: ClusterDetailNetworkingTabProps) {
  const router = useRouter();
  const { selectedCredentialId, selectedRegion } = useCredentialContext();

  const handleVPCClick = (vpcId: string) => {
    if (!selectedCredentialId || !selectedRegion) return;
    const params = new URLSearchParams({
      credentialId: selectedCredentialId,
      region: selectedRegion,
    });
    router.push(`/networks/vpcs?${params.toString()}&vpc_id=${encodeURIComponent(vpcId)}`);
  };

  const handleSubnetClick = (subnetId: string) => {
    if (!selectedCredentialId || !selectedRegion) return;
    const params = new URLSearchParams({
      credentialId: selectedCredentialId,
      region: selectedRegion,
    });
    router.push(`/networks/subnets?${params.toString()}&subnet_id=${encodeURIComponent(subnetId)}`);
  };

  const handleSecurityGroupClick = (sgId: string) => {
    if (!selectedCredentialId || !selectedRegion) return;
    const params = new URLSearchParams({
      credentialId: selectedCredentialId,
      region: selectedRegion,
    });
    router.push(`/networks/security-groups?${params.toString()}&sg_id=${encodeURIComponent(sgId)}`);
  };

  if (!isAWSCluster(cluster)) {
    return (
      <Card>
        <CardContent className="py-8 text-center text-muted-foreground">
          Networking details are only available for AWS EKS clusters.
        </CardContent>
      </Card>
    );
  }

  const awsCluster = cluster;

  return (
    <div className="space-y-4">
      {/* VPC */}
      {awsCluster.resources_vpc_config?.vpc_id && (
        <Card>
          <CardHeader>
            <CardTitle>VPC</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <span className="text-sm font-mono">{awsCluster.resources_vpc_config.vpc_id}</span>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleVPCClick(awsCluster.resources_vpc_config!.vpc_id)}
                className="h-8"
              >
                <ExternalLink className="h-4 w-4 mr-2" />
                View VPC
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 클러스터 IP 주소 패밀리 */}
      {awsCluster.kubernetes_network_config?.ip_family && (
        <Card>
          <CardHeader>
            <CardTitle>클러스터 IP 주소 패밀리</CardTitle>
          </CardHeader>
          <CardContent>
            <Badge variant="outline">{awsCluster.kubernetes_network_config.ip_family}</Badge>
          </CardContent>
        </Card>
      )}

      {/* 서비스 IPv4 범위 */}
      {awsCluster.kubernetes_network_config?.service_ipv4_cidr && (
        <Card>
          <CardHeader>
            <CardTitle>서비스 IPv4 범위</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm font-mono">{awsCluster.kubernetes_network_config.service_ipv4_cidr}</p>
          </CardContent>
        </Card>
      )}

      {/* 서브넷 */}
      {awsCluster.resources_vpc_config?.subnet_ids && awsCluster.resources_vpc_config.subnet_ids.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>서브넷</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {awsCluster.resources_vpc_config.subnet_ids.map((subnetId, index) => (
              <div key={index} className="flex items-center justify-between py-2 border-b last:border-0">
                <span className="text-sm font-mono">{subnetId}</span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleSubnetClick(subnetId)}
                  className="h-8"
                >
                  <ExternalLink className="h-4 w-4 mr-2" />
                  View Subnet
                </Button>
              </div>
            ))}
          </CardContent>
        </Card>
      )}

      {/* 클러스터 보안 그룹 */}
      {awsCluster.resources_vpc_config?.cluster_security_group_id && (
        <Card>
          <CardHeader>
            <CardTitle>클러스터 보안 그룹</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <span className="text-sm font-mono">
                {awsCluster.resources_vpc_config.cluster_security_group_id}
              </span>
              <Button
                variant="ghost"
                size="sm"
                onClick={() =>
                  handleSecurityGroupClick(awsCluster.resources_vpc_config!.cluster_security_group_id!)
                }
                className="h-8"
              >
                <ExternalLink className="h-4 w-4 mr-2" />
                View Security Group
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 추가 보안 그룹 */}
      {awsCluster.resources_vpc_config?.security_group_ids &&
        awsCluster.resources_vpc_config.security_group_ids.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle>추가 보안 그룹</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {awsCluster.resources_vpc_config.security_group_ids.map((sgId, index) => (
                <div key={index} className="flex items-center justify-between py-2 border-b last:border-0">
                  <span className="text-sm font-mono">{sgId}</span>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleSecurityGroupClick(sgId)}
                    className="h-8"
                  >
                    <ExternalLink className="h-4 w-4 mr-2" />
                    View Security Group
                  </Button>
                </div>
              ))}
            </CardContent>
          </Card>
        )}

      {/* API 서버 엔드포인트 액세스 */}
      <Card>
        <CardHeader>
          <CardTitle>API 서버 엔드포인트 액세스</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">퍼블릭 액세스</span>
            <Badge
              variant={awsCluster.resources_vpc_config?.endpoint_public_access ? 'default' : 'secondary'}
            >
              {awsCluster.resources_vpc_config?.endpoint_public_access ? 'Enabled' : 'Disabled'}
            </Badge>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">프라이빗 액세스</span>
            <Badge
              variant={awsCluster.resources_vpc_config?.endpoint_private_access ? 'default' : 'secondary'}
            >
              {awsCluster.resources_vpc_config?.endpoint_private_access ? 'Enabled' : 'Disabled'}
            </Badge>
          </div>
        </CardContent>
      </Card>

      {/* 퍼블릭 액세스 소스 허용 목록 */}
      {awsCluster.resources_vpc_config?.public_access_cidrs &&
        awsCluster.resources_vpc_config.public_access_cidrs.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle>퍼블릭 액세스 소스 허용 목록</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {awsCluster.resources_vpc_config.public_access_cidrs.map((cidr, index) => (
                <div key={index} className="text-sm font-mono py-1">
                  {cidr}
                </div>
              ))}
            </CardContent>
          </Card>
        )}
    </div>
  );
}

