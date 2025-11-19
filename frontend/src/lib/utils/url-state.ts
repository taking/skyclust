/**
 * URL State Utilities
 * Credential 및 Region 선택 상태를 URL query string으로 인코딩/디코딩
 */

import type { CloudProvider } from '@/lib/types/kubernetes';
import type { ProviderRegionSelection } from '@/hooks/use-provider-region-filter';

/**
 * Multi-credential 선택을 URL query string으로 인코딩
 * 
 * @example
 * encodeCredentials(['cred1', 'cred2', 'cred3']) // 'cred1,cred2,cred3'
 */
export function encodeCredentials(credentialIds: string[]): string {
  return credentialIds.join(',');
}

/**
 * URL query string에서 Multi-credential 선택 디코딩
 * 
 * @example
 * decodeCredentials('cred1,cred2,cred3') // ['cred1', 'cred2', 'cred3']
 * decodeCredentials(null) // []
 */
export function decodeCredentials(credentialsParam: string | null): string[] {
  if (!credentialsParam) return [];
  return credentialsParam.split(',').filter(Boolean);
}

/**
 * Provider별 Region 선택을 URL query string으로 인코딩
 * 
 * @example
 * encodeProviderRegions({
 *   aws: ['us-east-1', 'us-west-2'],
 *   gcp: ['asia-northeast3'],
 *   azure: []
 * })
 * // { aws_regions: 'us-east-1,us-west-2', gcp_regions: 'asia-northeast3' }
 */
export function encodeProviderRegions(
  selectedRegions: ProviderRegionSelection
): Record<string, string> {
  const params: Record<string, string> = {};
  
  Object.entries(selectedRegions).forEach(([provider, regions]) => {
    if (regions.length > 0) {
      params[`${provider}_regions`] = regions.join(',');
    }
  });
  
  return params;
}

/**
 * URL query string에서 Provider별 Region 선택 디코딩
 * 
 * @example
 * const searchParams = new URLSearchParams('aws_regions=us-east-1,us-west-2&gcp_regions=asia-northeast3');
 * decodeProviderRegions(searchParams)
 * // { aws: ['us-east-1', 'us-west-2'], gcp: ['asia-northeast3'], azure: [] }
 */
export function decodeProviderRegions(
  searchParams: URLSearchParams
): ProviderRegionSelection {
  const result: ProviderRegionSelection = {
    aws: [],
    gcp: [],
    azure: [],
  };
  
  (['aws', 'gcp', 'azure'] as CloudProvider[]).forEach(provider => {
    const param = searchParams.get(`${provider}_regions`);
    if (param) {
      result[provider] = param.split(',').filter(Boolean);
    }
  });
  
  return result;
}

/**
 * URL 길이 제한 확인
 * 브라우저 URL 길이 제한은 보통 2000자 정도이므로 이를 기준으로 함
 */
export const MAX_URL_LENGTH = 2000;

/**
 * URL이 너무 긴지 확인
 */
export function isUrlTooLong(url: string): boolean {
  return url.length > MAX_URL_LENGTH;
}

