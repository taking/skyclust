/**
 * Review VPC Step
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
  Tag,
  Cloud
} from 'lucide-react';
import type { CreateVPCForm, CloudProvider } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';

interface ReviewVPCStepProps {
  formData: CreateVPCForm;
  selectedProvider?: CloudProvider;
}

export function ReviewVPCStep({
  formData,
  selectedProvider,
}: ReviewVPCStepProps) {
  const { t } = useTranslation();

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
                {t('network.review.vpcName')}
              </div>
              <p className="text-sm font-semibold ml-6">{formData.name || 'N/A'}</p>
            </div>
            {formData.description && (
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <FileText className="h-4 w-4" />
                  {t('network.review.description')}
                </div>
                <p className="text-sm font-semibold ml-6">{formData.description}</p>
              </div>
            )}
            {formData.cidr_block && (
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <Network className="h-4 w-4" />
                  {t('network.review.cidrBlock')}
                </div>
                <p className="text-sm font-semibold ml-6">{formData.cidr_block}</p>
              </div>
            )}
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                <Globe className="h-4 w-4" />
                {t('network.review.region')}
              </div>
              <p className="text-sm font-semibold ml-6">{formData.region || formData.location || 'N/A'}</p>
            </div>
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
            {formData.resource_group && (
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <Network className="h-4 w-4" />
                  {t('network.review.resourceGroup')}
                </div>
                <p className="text-sm font-semibold ml-6">{formData.resource_group}</p>
              </div>
            )}
            {formData.address_space && Array.isArray(formData.address_space) && formData.address_space.length > 0 && (
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <Network className="h-4 w-4" />
                  {t('network.review.addressSpace')}
                </div>
                <p className="text-sm font-semibold ml-6">{formData.address_space.join(', ')}</p>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      <Separator />

      {/* Advanced Settings */}
      {(formData.tags && Object.keys(formData.tags).length > 0) || 
       formData.auto_create_subnets || 
       formData.routing_mode || 
       formData.mtu ? (
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
            {formData.auto_create_subnets && (
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <Network className="h-4 w-4" />
                  {t('network.review.autoCreateSubnets')}
                </div>
                <p className="text-sm font-semibold ml-6">{t('network.review.enabled')}</p>
              </div>
            )}
            {formData.routing_mode && (
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <Network className="h-4 w-4" />
                  {t('network.review.routingMode')}
                </div>
                <p className="text-sm font-semibold ml-6">{formData.routing_mode}</p>
              </div>
            )}
            {formData.mtu && (
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <Network className="h-4 w-4" />
                  {t('network.review.mtu')}
                </div>
                <p className="text-sm font-semibold ml-6">{formData.mtu}</p>
              </div>
            )}
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}

