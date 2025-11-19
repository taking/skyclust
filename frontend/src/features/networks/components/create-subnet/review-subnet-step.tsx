/**
 * Review Subnet Step
 * Step 3: 최종 확인 및 생성
 */

'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { 
  FileText, 
  Network, 
  Settings, 
  Globe,
  MapPin,
  Tag,
  Cloud
} from 'lucide-react';
import type { CreateSubnetForm, CloudProvider } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';
import { useNetworkResources } from '@/features/networks/hooks/use-network-resources';

interface ReviewSubnetStepProps {
  formData: CreateSubnetForm;
  selectedProvider?: CloudProvider;
}

export function ReviewSubnetStep({
  formData,
  selectedProvider,
}: ReviewSubnetStepProps) {
  const { t } = useTranslation();
  const { vpcs } = useNetworkResources({ resourceType: 'vpcs' });
  const selectedVPC = vpcs.find(v => v.id === formData.vpc_id);

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
      <div className="bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
        <p className="text-sm text-blue-800 dark:text-blue-200">
          {t('network.review.alertMessage')}
        </p>
      </div>

      <Separator />

      {/* Basic Configuration */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <FileText className="h-5 w-5 text-muted-foreground" />
            <CardTitle className="text-lg font-semibold">{t('network.review.basicConfiguration')}</CardTitle>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                <Network className="h-4 w-4" />
                {t('network.review.subnetName')}
              </div>
              <p className="text-sm font-semibold ml-6">{formData.name || 'N/A'}</p>
            </div>
            {selectedVPC && (
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <Network className="h-4 w-4" />
                  {t('network.review.vpc')}
                </div>
                <div className="ml-6">
                  <Badge variant="secondary">
                    {selectedVPC.name || selectedVPC.id}
                  </Badge>
                </div>
              </div>
            )}
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                <Network className="h-4 w-4" />
                {t('network.review.cidrBlock')}
              </div>
              <p className="text-sm font-semibold ml-6">{formData.cidr_block || 'N/A'}</p>
            </div>
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                <Globe className="h-4 w-4" />
                {t('network.review.region')}
              </div>
              <p className="text-sm font-semibold ml-6">{formData.region || 'N/A'}</p>
            </div>
            {(formData.availability_zone || formData.zone) && (
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <MapPin className="h-4 w-4" />
                  {selectedProvider === 'gcp' ? t('network.review.zone') : t('network.review.availabilityZone')}
                </div>
                <p className="text-sm font-semibold ml-6">
                  {formData.availability_zone || formData.zone || 'N/A'}
                </p>
              </div>
            )}
            {selectedProvider && (
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <Cloud className="h-4 w-4" />
                  {t('network.review.provider')}
                </div>
                <div className="ml-6">
                  <Badge className={getProviderColor(selectedProvider)}>
                    {selectedProvider.toUpperCase()}
                  </Badge>
                </div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      <Separator />

      {/* Advanced Settings */}
      {(formData.tags && Object.keys(formData.tags).length > 0) || 
       formData.private_ip_google_access || 
       formData.flow_logs ? (
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Settings className="h-5 w-5 text-muted-foreground" />
              <CardTitle className="text-lg font-semibold">{t('network.review.advancedSettings')}</CardTitle>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            {formData.tags && Object.keys(formData.tags).length > 0 && (
              <div>
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground mb-2">
                  <Tag className="h-4 w-4" />
                  {t('network.review.tags')} ({Object.keys(formData.tags).length})
                </div>
                <div className="flex flex-wrap gap-2 ml-6">
                  {Object.entries(formData.tags).map(([key, value]) => (
                    <Badge key={key} variant="secondary" className="text-xs">
                      {key}: {value}
                    </Badge>
                  ))}
                </div>
              </div>
            )}
            {formData.private_ip_google_access && (
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <Network className="h-4 w-4" />
                  {t('network.review.privateIPGoogleAccess')}
                </div>
                <p className="text-sm font-semibold ml-6">{t('network.review.enabled')}</p>
              </div>
            )}
            {formData.flow_logs && (
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <Network className="h-4 w-4" />
                  {t('network.review.flowLogs')}
                </div>
                <p className="text-sm font-semibold ml-6">{t('network.review.enabled')}</p>
              </div>
            )}
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}

