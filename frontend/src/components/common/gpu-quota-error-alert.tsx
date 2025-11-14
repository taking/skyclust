/**
 * GPU Quota Error Alert Component
 * GPU quota 부족 시 상세 정보를 표시하는 Alert 컴포넌트
 */

'use client';

import * as React from 'react';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { AlertCircle, ExternalLink, MapPin } from 'lucide-react';
import type { GPUQuotaErrorDetails } from '@/lib/error-handling/quota-error-handler';
import { useTranslation } from '@/hooks/use-translation';

interface GPUQuotaErrorAlertProps {
  errorDetails: GPUQuotaErrorDetails;
  onRegionChange?: (region: string) => void;
}

export function GPUQuotaErrorAlert({ errorDetails, onRegionChange }: GPUQuotaErrorAlertProps) {
  const { t } = useTranslation();

  return (
    <Alert variant="destructive" className="my-4">
      <AlertCircle className="h-4 w-4" />
      <AlertTitle className="mb-2">
        {t('kubernetes.gpuQuotaInsufficient') || 'GPU Quota Insufficient'}
      </AlertTitle>
      <AlertDescription className="space-y-3">
        <div className="text-sm">
          <p className="mb-2">
            {t('kubernetes.gpuQuotaErrorMessage', {
              instanceType: errorDetails.instance_type,
              region: errorDetails.region,
              available: Math.floor(errorDetails.available_quota),
              required: errorDetails.required_count,
            }) || 
            `GPU instance type ${errorDetails.instance_type} has insufficient quota in region ${errorDetails.region}. Available: ${Math.floor(errorDetails.available_quota)}, Required: ${errorDetails.required_count}.`}
          </p>
          
          <div className="flex flex-wrap gap-2 mb-2">
            <Badge variant="outline" className="text-xs">
              {t('kubernetes.currentQuota') || 'Current Quota'}: {Math.floor(errorDetails.current_quota)}
            </Badge>
            {errorDetails.current_usage !== undefined && (
              <Badge variant="outline" className="text-xs">
                {t('kubernetes.currentUsage') || 'Current Usage'}: {Math.floor(errorDetails.current_usage)}
              </Badge>
            )}
            <Badge variant="outline" className="text-xs">
              {t('kubernetes.availableQuota') || 'Available'}: {Math.floor(errorDetails.available_quota)}
            </Badge>
          </div>

          {errorDetails.quota_increase_url && (
            <div className="mt-3">
              <Button
                variant="outline"
                size="sm"
                onClick={() => window.open(errorDetails.quota_increase_url, '_blank')}
                className="mr-2"
              >
                <ExternalLink className="h-3 w-3 mr-1" />
                {t('kubernetes.requestQuotaIncrease') || 'Request Quota Increase'}
              </Button>
            </div>
          )}

          {errorDetails.available_regions && errorDetails.available_regions.length > 0 && (
            <div className="mt-4 pt-3 border-t">
              <p className="text-sm font-medium mb-2">
                {t('kubernetes.availableRegions') || 'Available Regions'}:
              </p>
              <div className="flex flex-wrap gap-2">
                {errorDetails.available_regions.map((region) => (
                  <Button
                    key={region.region}
                    variant="outline"
                    size="sm"
                    onClick={() => onRegionChange?.(region.region)}
                    className="text-xs"
                  >
                    <MapPin className="h-3 w-3 mr-1" />
                    {region.region}
                    <Badge variant="secondary" className="ml-2 text-xs">
                      {Math.floor(region.available_quota)} {t('kubernetes.available') || 'available'}
                    </Badge>
                  </Button>
                ))}
              </div>
            </div>
          )}
        </div>
      </AlertDescription>
    </Alert>
  );
}

