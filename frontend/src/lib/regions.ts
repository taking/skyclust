/**
 * Cloud Provider Regions
 * 각 클라우드 프로바이더별 리전 목록
 */

export interface RegionOption {
  value: string;
  label: string;
}

// GCP regions list
export const GCP_REGIONS: RegionOption[] = [
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
export const AWS_REGIONS: RegionOption[] = [
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
export const AZURE_REGIONS: RegionOption[] = [
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

/**
 * Get regions for a specific provider
 */
export function getRegionsForProvider(provider?: string): RegionOption[] {
  switch (provider?.toLowerCase()) {
    case 'gcp':
      return GCP_REGIONS;
    case 'aws':
      return AWS_REGIONS;
    case 'azure':
      return AZURE_REGIONS;
    default:
      return [];
  }
}

/**
 * Check if provider supports region selection
 */
export function supportsRegionSelection(provider?: string): boolean {
  const supportedProviders = ['gcp', 'aws', 'azure'];
  return supportedProviders.includes(provider?.toLowerCase() || '');
}

