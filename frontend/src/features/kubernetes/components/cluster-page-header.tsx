/**
 * Cluster Page Header Component
 * Kubernetes 클러스터 페이지 헤더
 */

'use client';

import * as React from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Plus, RefreshCw, Info } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { cn } from '@/lib/utils';
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip';
import type { Credential, CloudProvider } from '@/lib/types';

interface ClusterPageHeaderProps {
  workspaceName?: string;
  credentials: Credential[];
  selectedCredentialId: string;
  onCredentialChange: (credentialId: string) => void;
  selectedRegion?: string;
  onRegionChange: (region: string) => void;
  selectedProvider?: CloudProvider;
  onRefresh?: () => void;
  isRefreshing?: boolean;
  lastUpdated?: Date | null;
  onCreateClick?: () => void;
}

// GCP regions list
const GCP_REGIONS = [
  { value: 'asia-east1', label: 'Asia East (Taiwan)' },
  { value: 'asia-northeast1', label: 'Asia Northeast (Tokyo)' },
  { value: 'asia-northeast2', label: 'Asia Northeast 2 (Osaka)' },
  { value: 'asia-northeast3', label: 'Asia Northeast 3 (Seoul)' },
  { value: 'asia-south1', label: 'Asia South (Mumbai)' },
  { value: 'asia-southeast1', label: 'Asia Southeast (Singapore)' },
  { value: 'australia-southeast1', label: 'Australia Southeast (Sydney)' },
  { value: 'europe-west1', label: 'Europe West (Belgium)' },
  { value: 'europe-west4', label: 'Europe West 4 (Netherlands)' },
  { value: 'europe-west6', label: 'Europe West 6 (Zurich)' },
  { value: 'northamerica-northeast1', label: 'North America Northeast (Montreal)' },
  { value: 'southamerica-east1', label: 'South America East (São Paulo)' },
  { value: 'us-central1', label: 'US Central (Iowa)' },
  { value: 'us-east1', label: 'US East (South Carolina)' },
  { value: 'us-east4', label: 'US East 4 (Northern Virginia)' },
  { value: 'us-west1', label: 'US West (Oregon)' },
  { value: 'us-west2', label: 'US West 2 (Los Angeles)' },
  { value: 'us-west3', label: 'US West 3 (Salt Lake City)' },
  { value: 'us-west4', label: 'US West 4 (Las Vegas)' },
];

// AWS regions list
const AWS_REGIONS = [
  { value: 'us-east-1', label: 'US East (N. Virginia)' },
  { value: 'us-east-2', label: 'US East (Ohio)' },
  { value: 'us-west-1', label: 'US West (N. California)' },
  { value: 'us-west-2', label: 'US West (Oregon)' },
  { value: 'af-south-1', label: 'Africa (Cape Town)' },
  { value: 'ap-east-1', label: 'Asia Pacific (Hong Kong)' },
  { value: 'ap-south-1', label: 'Asia Pacific (Mumbai)' },
  { value: 'ap-south-2', label: 'Asia Pacific (Hyderabad)' },
  { value: 'ap-southeast-1', label: 'Asia Pacific (Singapore)' },
  { value: 'ap-southeast-2', label: 'Asia Pacific (Sydney)' },
  { value: 'ap-southeast-3', label: 'Asia Pacific (Jakarta)' },
  { value: 'ap-southeast-4', label: 'Asia Pacific (Melbourne)' },
  { value: 'ap-northeast-1', label: 'Asia Pacific (Tokyo)' },
  { value: 'ap-northeast-2', label: 'Asia Pacific (Seoul)' },
  { value: 'ap-northeast-3', label: 'Asia Pacific (Osaka)' },
  { value: 'ca-central-1', label: 'Canada (Central)' },
  { value: 'eu-central-1', label: 'Europe (Frankfurt)' },
  { value: 'eu-central-2', label: 'Europe (Zurich)' },
  { value: 'eu-west-1', label: 'Europe (Ireland)' },
  { value: 'eu-west-2', label: 'Europe (London)' },
  { value: 'eu-west-3', label: 'Europe (Paris)' },
  { value: 'eu-south-1', label: 'Europe (Milan)' },
  { value: 'eu-south-2', label: 'Europe (Spain)' },
  { value: 'eu-north-1', label: 'Europe (Stockholm)' },
  { value: 'me-south-1', label: 'Middle East (Bahrain)' },
  { value: 'me-central-1', label: 'Middle East (UAE)' },
  { value: 'sa-east-1', label: 'South America (São Paulo)' },
];

// Azure regions list
const AZURE_REGIONS = [
  { value: 'eastus', label: 'East US' },
  { value: 'eastus2', label: 'East US 2' },
  { value: 'southcentralus', label: 'South Central US' },
  { value: 'westus2', label: 'West US 2' },
  { value: 'westus3', label: 'West US 3' },
  { value: 'australiaeast', label: 'Australia East' },
  { value: 'southeastasia', label: 'Southeast Asia' },
  { value: 'northeurope', label: 'North Europe' },
  { value: 'swedencentral', label: 'Sweden Central' },
  { value: 'uksouth', label: 'UK South' },
  { value: 'westeurope', label: 'West Europe' },
  { value: 'centralus', label: 'Central US' },
  { value: 'southafricanorth', label: 'South Africa North' },
  { value: 'centralindia', label: 'Central India' },
  { value: 'eastasia', label: 'East Asia' },
  { value: 'japaneast', label: 'Japan East' },
  { value: 'koreacentral', label: 'Korea Central' },
  { value: 'canadacentral', label: 'Canada Central' },
  { value: 'francecentral', label: 'France Central' },
  { value: 'germanywestcentral', label: 'Germany West Central' },
  { value: 'italynorth', label: 'Italy North' },
  { value: 'norwayeast', label: 'Norway East' },
  { value: 'polandcentral', label: 'Poland Central' },
  { value: 'switzerlandnorth', label: 'Switzerland North' },
  { value: 'uaenorth', label: 'UAE North' },
  { value: 'brazilsouth', label: 'Brazil South' },
  { value: 'israelcentral', label: 'Israel Central' },
  { value: 'qatarcentral', label: 'Qatar Central' },
  { value: 'centralusstage', label: 'Central US (Stage)' },
  { value: 'eastusstage', label: 'East US (Stage)' },
];

function ClusterPageHeaderComponent({
  workspaceName,
  credentials,
  selectedCredentialId,
  onCredentialChange: _onCredentialChange,
  selectedRegion: _selectedRegion = '',
  onRegionChange: _onRegionChange,
  selectedProvider,
  onRefresh,
  isRefreshing = false,
  lastUpdated,
  onCreateClick,
}: ClusterPageHeaderProps) {
  const { t } = useTranslation();
  const router = useRouter();
  
  // Show region selector for GCP, AWS, and Azure
  const _showRegionSelector = selectedProvider === 'gcp' || selectedProvider === 'aws' || selectedProvider === 'azure';

  // Get regions based on provider
  const getRegions = () => {
    switch (selectedProvider) {
      case 'gcp':
        return GCP_REGIONS;
      case 'aws':
        return AWS_REGIONS;
      case 'azure':
        return AZURE_REGIONS;
      default:
        return [];
    }
  };

  const _regions = getRegions();

  const handleCreateClick = () => {
    if (onCreateClick) {
      onCreateClick();
    } else {
      // Fallback: use legacy path if onCreateClick not provided
      router.push('/kubernetes/clusters/create');
    }
  };

  // 마지막 업데이트 시간 포맷팅
  const formatLastUpdated = (date: Date | null | undefined): string => {
    if (!date) return '';
    
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);
    
    if (diffSec < 60) {
      return t('common.justNow') || '방금 전';
    } else if (diffMin < 60) {
      return t('common.minutesAgo', { minutes: diffMin }) || `${diffMin}분 전`;
    } else if (diffHour < 24) {
      return t('common.hoursAgo', { hours: diffHour }) || `${diffHour}시간 전`;
    } else {
      const diffDay = Math.floor(diffHour / 24);
      return t('common.daysAgo', { days: diffDay }) || `${diffDay}일 전`;
    }
  };

  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">{t('kubernetes.clusters.label')}</h1>
        <div className="flex items-center gap-2">
          <p className="text-gray-600">
            {workspaceName 
              ? t('kubernetes.manageClustersWithWorkspace', { workspaceName }) 
              : t('kubernetes.manageClusters')
            }
          </p>
          {lastUpdated && (
            <div className="flex items-center gap-1.5">
              <span className="text-sm text-gray-500">
                ({t('common.lastUpdated')}: {formatLastUpdated(lastUpdated)})
              </span>
              <Tooltip delayDuration={200}>
                <TooltipTrigger asChild>
                  <button
                    type="button"
                    className="focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-offset-1 rounded-sm transition-colors"
                    aria-label={t('sse.syncInfo') || 'Sync information'}
                  >
                    <Info className="h-3.5 w-3.5 text-gray-400 hover:text-gray-600 transition-colors" aria-hidden="true" />
                  </button>
                </TooltipTrigger>
                <TooltipContent 
                  side="right" 
                  className="max-w-xs bg-gray-900 text-white border-gray-700 z-50"
                >
                  <div className="space-y-2 text-xs">
                    <div className="font-semibold text-sm mb-1.5 pb-1.5 border-b border-gray-700">
                      {t('sse.dataSyncInfo') || '데이터 동기화 정보'}
                    </div>
                    
                    <div className="space-y-1.5">
                      <div>
                        <span className="font-medium text-gray-100">
                          {t('sse.syncInterval') || '동기화 주기'}:
                        </span>
                        <span className="ml-1.5 text-gray-300">
                          {t('sse.syncEvery5Minutes') || '5분마다'}
                        </span>
                      </div>
                      
                      <div>
                        <span className="font-medium text-gray-100">
                          {t('sse.lastSyncTime') || '마지막 동기화'}:
                        </span>
                        <span className="ml-1.5 text-gray-300">
                          {formatLastUpdated(lastUpdated)}
                        </span>
                      </div>
                    </div>
                    
                    <div className="pt-1.5 mt-1.5 border-t border-gray-700">
                      <p className="text-gray-300 leading-relaxed">
                        {t('sse.autoUpdateDescription') || '백엔드에서 5분마다 클라우드 서비스 제공자(CSP) API를 호출하여 최신 데이터를 조회합니다. 변경사항이 감지되면 자동으로 SSE 이벤트를 발행하여 실시간으로 업데이트됩니다.'}
                      </p>
                    </div>
                  </div>
                </TooltipContent>
              </Tooltip>
            </div>
          )}
        </div>
      </div>
      <div className="flex items-center space-x-2">
        {onRefresh && (
          <Button
            variant="outline"
            size="sm"
            onClick={onRefresh}
            disabled={isRefreshing || !selectedCredentialId}
          >
            <RefreshCw className={cn('mr-2 h-4 w-4', isRefreshing && 'animate-spin')} />
            {isRefreshing ? (t('common.refreshing') || '새로고침 중...') : (t('common.refresh') || '새로고침')}
          </Button>
        )}
        <Button
          onClick={handleCreateClick}
          disabled={!selectedCredentialId || credentials.length === 0}
        >
          <Plus className="mr-2 h-4 w-4" />
          {t('kubernetes.createCluster')}
        </Button>
      </div>
    </div>
  );
}

export const ClusterPageHeader = React.memo(ClusterPageHeaderComponent);

