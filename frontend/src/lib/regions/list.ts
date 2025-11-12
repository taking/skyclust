/**
 * 클라우드 프로바이더별 리전 목록 및 유틸리티 함수
 */

export interface RegionOption {
  value: string;
  label: string;
}

/** GCP 리전 목록 */
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

/** AWS 리전 목록 */
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

/** Azure 리전 목록 */
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
 * 특정 프로바이더의 리전 목록 반환
 * @param provider 프로바이더 이름 (gcp, aws, azure)
 * @returns 리전 목록 배열
 */
export function getRegionsByProvider(provider?: string): RegionOption[] {
  // 프로바이더 이름을 소문자로 변환하여 대소문자 구분 없이 비교
  const normalizedProvider = provider?.toLowerCase();
  
  switch (normalizedProvider) {
    case 'gcp':
      return GCP_REGIONS;
    case 'aws':
      return AWS_REGIONS;
    case 'azure':
      return AZURE_REGIONS;
    default:
      // 지원하지 않는 프로바이더이거나 provider가 undefined인 경우 빈 배열 반환
      return [];
  }
}

/**
 * 모든 리전 목록 반환
 */
export function getAllRegions(): RegionOption[] {
  return [...AWS_REGIONS, ...GCP_REGIONS, ...AZURE_REGIONS];
}

/**
 * 프로바이더가 리전 선택을 지원하는지 확인
 * @param provider 프로바이더 이름
 * @returns 리전 선택 지원 여부
 */
export function supportsRegionSelection(provider?: string): boolean {
  // 리전 선택을 지원하는 프로바이더 목록
  const supportedProviders = ['gcp', 'aws', 'azure'];
  
  // 프로바이더 이름을 소문자로 변환하여 지원 목록에 포함되어 있는지 확인
  const normalizedProvider = provider?.toLowerCase() || '';
  return supportedProviders.includes(normalizedProvider);
}

/**
 * 프로바이더의 기본 리전 반환
 * - GCP: asia-northeast3 (서울)
 * - AWS: ap-northeast-3 (오사카)
 * - Azure: koreacentral (한국 중부)
 * @param provider 프로바이더 이름
 * @returns 기본 리전 또는 null
 */
export function getDefaultRegionForProvider(provider?: string): string | null {
  // 프로바이더 이름을 소문자로 변환하여 대소문자 구분 없이 비교
  const normalizedProvider = provider?.toLowerCase();
  
  switch (normalizedProvider) {
    case 'gcp':
      // GCP 기본 리전: 서울 (asia-northeast3)
      return 'asia-northeast3';
    case 'aws':
      // AWS 기본 리전: 오사카 (ap-northeast-3)
      return 'ap-northeast-3';
    case 'azure':
      // Azure 기본 리전: 한국 중부 (koreacentral)
      return 'koreacentral';
    default:
      // 지원하지 않는 프로바이더이거나 provider가 undefined인 경우 null 반환
      return null;
  }
}

