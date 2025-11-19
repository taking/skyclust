/**
 * Review Step
 * Step 4: 최종 확인 및 생성
 */

'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { 
  FileText, 
  Server, 
  Network, 
  Globe, 
  MapPin, 
  Shield, 
  ChevronRight,
  Cloud,
  Layers
} from 'lucide-react';
import { useNetworkResources } from '@/features/networks/hooks/use-network-resources';
import { useTranslation } from '@/hooks/use-translation';
import type { CreateClusterForm, CloudProvider } from '@/lib/types';
import { AWSReviewConfig } from './providers/aws/aws-review-config';
import { GCPReviewConfig } from './providers/gcp/gcp-review-config';
import { AzureReviewConfig } from './providers/azure/azure-review-config';

interface ReviewStepProps {
  formData: CreateClusterForm;
  selectedProvider?: CloudProvider;
  selectedProjectId?: string;
  onCreate: () => void;
  isPending: boolean;
}

export function ReviewStep({
  formData,
  selectedProvider,
  selectedProjectId,
  onCreate,
  isPending,
}: ReviewStepProps) {
  const { t } = useTranslation();
  // VPC와 Subnet 정보를 가져오기 위해 사용
  const { vpcs } = useNetworkResources({ resourceType: 'vpcs' });
  const { subnets = [] } = useNetworkResources({ resourceType: 'subnets', requireVPC: true });

  // 선택된 Subnet 정보 찾기
  const selectedSubnets = formData.subnet_ids
    ? formData.subnet_ids.map(id => subnets.find(s => s.id === id)).filter(Boolean)
    : [];

  // 선택된 VPC 찾기 (subnet_ids에서 첫 번째 subnet의 vpc_id 사용)
  const firstSubnet = selectedSubnets.length > 0 ? selectedSubnets[0] : null;
  const vpcIdFromSubnet = firstSubnet?.vpc_id;
  const selectedVPC = vpcIdFromSubnet 
    ? vpcs.find(v => v.id === vpcIdFromSubnet) 
    : (formData.vpc_id ? vpcs.find(v => v.id === formData.vpc_id) : null);

  // Provider별 색상 설정
  const getProviderColor = (provider?: CloudProvider) => {
    switch (provider) {
      case 'aws':
        return 'bg-orange-100 text-orange-800 border-orange-200';
      case 'gcp':
        return 'bg-blue-100 text-blue-800 border-blue-200';
      case 'azure':
        return 'bg-cyan-100 text-cyan-800 border-cyan-200';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };


  return (
    <div className="space-y-6">
      {/* Alert Banner */}
      <div className="bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 rounded-lg p-4 flex items-start gap-3">
        <Shield className="h-5 w-5 text-blue-600 dark:text-blue-400 mt-0.5 flex-shrink-0" />
        <p className="text-sm text-blue-800 dark:text-blue-200">
          {t('kubernetes.review.alertMessage')}
        </p>
      </div>

      <Separator />

      {/* Basic Configuration */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <FileText className="h-5 w-5 text-muted-foreground" />
            <CardTitle className="text-lg font-semibold">{t('kubernetes.review.basicConfiguration')}</CardTitle>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                <Server className="h-4 w-4" />
                {t('kubernetes.review.clusterName')}
              </div>
              <p className="text-sm font-semibold ml-6">{formData.name || 'N/A'}</p>
            </div>
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                <Layers className="h-4 w-4" />
                {t('kubernetes.review.kubernetesVersion')}
              </div>
              <p className="text-sm font-semibold ml-6">{formData.version || 'N/A'}</p>
            </div>
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                <Globe className="h-4 w-4" />
                {t('kubernetes.review.region')}
              </div>
              <p className="text-sm font-semibold ml-6">{formData.region || 'N/A'}</p>
            </div>
            {formData.zone && (
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <MapPin className="h-4 w-4" />
                  {t('kubernetes.review.zone')}
                </div>
                <p className="text-sm font-semibold ml-6">{formData.zone}</p>
              </div>
            )}
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                <Cloud className="h-4 w-4" />
                {t('kubernetes.review.provider')}
              </div>
              <div className="ml-6">
                <Badge className={getProviderColor(selectedProvider)}>
                  {selectedProvider?.toUpperCase() || 'N/A'}
                </Badge>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <Separator />

      {/* Network Configuration */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Network className="h-5 w-5 text-muted-foreground" />
            <CardTitle className="text-lg font-semibold">{t('kubernetes.review.networkConfiguration')}</CardTitle>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {selectedVPC ? (
            <div className="space-y-3">
              {/* VPC 트리 구조 */}
              <div className="space-y-2">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <Network className="h-4 w-4" />
                  {t('kubernetes.review.vpc')}
                </div>
                <div className="ml-6 space-y-1">
                  <div className="flex items-center gap-2">
                    <Badge variant="secondary" className="font-medium">
                      {selectedVPC.name || selectedVPC.id}
                    </Badge>
                    {selectedVPC.cidr_block && (
                      <span className="text-xs text-muted-foreground">
                        {selectedVPC.cidr_block}
                      </span>
                    )}
                  </div>
                  
                  {/* Subnets 트리 구조 */}
                  {selectedSubnets.length > 0 && (
                    <div className="mt-3 space-y-2">
                      <div className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
                        <ChevronRight className="h-3 w-3" />
                        {t('kubernetes.review.subnets')} ({selectedSubnets.length})
                      </div>
                      <div className="ml-4 space-y-1.5">
                        {selectedSubnets.map((subnet, index) => (
                          <div key={subnet?.id} className="flex items-center gap-2">
                            <div className="flex items-center gap-1.5">
                              <div className="h-1.5 w-1.5 rounded-full bg-muted-foreground/40" />
                              <Badge variant="outline" className="text-xs">
                                {subnet?.name || subnet?.id}
                              </Badge>
                            </div>
                            {subnet?.availability_zone && (
                              <Badge variant="secondary" className="text-xs">
                                {subnet.availability_zone}
                              </Badge>
                            )}
                            {subnet?.cidr_block && (
                              <span className="text-xs text-muted-foreground">
                                {subnet.cidr_block}
                              </span>
                            )}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>
          ) : selectedSubnets.length > 0 ? (
            <div className="space-y-2">
              <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                <Layers className="h-4 w-4" />
                {t('kubernetes.review.subnets')} ({selectedSubnets.length})
              </div>
              <div className="ml-6 space-y-1.5">
                {selectedSubnets.map((subnet) => (
                  <div key={subnet?.id} className="flex items-center gap-2">
                    <Badge variant="outline">
                      {subnet?.name || subnet?.id}
                    </Badge>
                    {subnet?.availability_zone && (
                      <Badge variant="secondary" className="text-xs">
                        {subnet.availability_zone}
                      </Badge>
                    )}
                  </div>
                ))}
              </div>
            </div>
          ) : formData.subnet_ids && formData.subnet_ids.length > 0 ? (
            <div className="flex flex-wrap gap-2">
              {formData.subnet_ids.map((subnetId) => (
                <Badge key={subnetId} variant="secondary">
                  {subnetId}
                </Badge>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">{t('kubernetes.review.subnets')}: {t('emptyState.noResource', { resource: t('network.subnets') })}</p>
          )}
        </CardContent>
      </Card>

      {/* Provider별 Review Configuration */}
      {selectedProvider === 'aws' && (
        <AWSReviewConfig formData={formData} />
      )}
      {selectedProvider === 'gcp' && (
        <GCPReviewConfig formData={formData} selectedProjectId={selectedProjectId} />
      )}
      {selectedProvider === 'azure' && (
        <AzureReviewConfig formData={formData} />
      )}
    </div>
  );
}

