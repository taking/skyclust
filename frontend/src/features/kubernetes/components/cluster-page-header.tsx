/**
 * Cluster Page Header Component
 * Kubernetes 클러스터 페이지 헤더
 */

'use client';

import * as React from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Plus } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import type { Credential, CloudProvider } from '@/lib/types';

interface ClusterPageHeaderProps {
  workspaceName?: string;
  credentials: Credential[];
  selectedCredentialId: string;
  onCredentialChange: (credentialId: string) => void;
  selectedRegion?: string;
  onRegionChange: (region: string) => void;
  selectedProvider?: CloudProvider;
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
    router.push('/kubernetes/clusters/create');
  };

  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">{t('kubernetes.clusters')}</h1>
        <p className="text-gray-600">
          {workspaceName 
            ? t('kubernetes.manageClustersWithWorkspace', { workspaceName }) 
            : t('kubernetes.manageClusters')
          }
        </p>
      </div>
      <div className="flex items-center space-x-2">
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

