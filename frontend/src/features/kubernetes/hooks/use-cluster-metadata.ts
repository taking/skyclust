/**
 * Cluster Metadata Hook
 * 클러스터 생성 시 필요한 메타데이터 (버전, 리전, 존) 로딩 훅
 * 
 * AWS 클러스터 생성 시 필요한 Kubernetes 버전, 리전, 가용 영역 정보를 통합 관리합니다.
 * 
 * @example
 * ```tsx
 * const {
 *   versions,
 *   regions,
 *   zones,
 *   isLoadingVersions,
 *   isLoadingRegions,
 *   isLoadingZones,
 *   versionsError,
 *   regionsError,
 *   zonesError,
 * } = useClusterMetadata({
 *   provider: 'aws',
 *   credentialId: 'cred-123',
 *   region: 'ap-northeast-2',
 * });
 * ```
 */

import { useEKSVersions, useAWSRegions, useAvailabilityZones } from './use-kubernetes-metadata';
import type { CloudProvider } from '@/lib/types';
import type { RegionOption } from '@/lib/regions';

export interface UseClusterMetadataOptions {
  /** 클라우드 프로바이더 */
  provider?: CloudProvider;
  /** Credential ID */
  credentialId: string;
  /** 선택된 리전 */
  region?: string;
}

export interface UseClusterMetadataReturn {
  /** Kubernetes 버전 목록 */
  versions: string[];
  /** AWS 리전 목록 */
  regions: Array<{ value: string; label: string }>;
  /** 가용 영역 목록 */
  zones: string[];
  /** 버전 로딩 중 여부 */
  isLoadingVersions: boolean;
  /** 리전 로딩 중 여부 */
  isLoadingRegions: boolean;
  /** 존 로딩 중 여부 */
  isLoadingZones: boolean;
  /** 버전 로딩 에러 */
  versionsError: Error | null;
  /** 리전 로딩 에러 */
  regionsError: Error | null;
  /** 존 로딩 에러 */
  zonesError: Error | null;
  /** 메타데이터 로딩 가능 여부 (AWS이고 credentialId가 있는 경우) */
  canLoadMetadata: boolean;
}

/**
 * 클러스터 메타데이터 로딩 훅
 */
export function useClusterMetadata(
  options: UseClusterMetadataOptions
): UseClusterMetadataReturn {
  const { provider, credentialId, region = '' } = options;

  const isAWS = provider === 'aws';
  const canLoadMetadata = isAWS && !!credentialId;

  // Fetch Kubernetes versions (AWS only)
  const {
    data: versions = [],
    isLoading: isLoadingVersions,
    isError: isVersionsError,
    error: versionsError,
  } = useEKSVersions({
    provider,
    credentialId,
    region,
  });

  // Fetch AWS regions (AWS only)
  const {
    data: awsRegionsData = [],
    isLoading: isLoadingRegions,
    isError: isRegionsError,
    error: regionsError,
  } = useAWSRegions({
    provider,
    credentialId,
  });

  // Convert string[] to RegionOption[]
  const regions: RegionOption[] = awsRegionsData.map(region => ({
    value: region,
    label: region,
  }));

  // Fetch availability zones (AWS only, when region is selected)
  const {
    data: zones = [],
    isLoading: isLoadingZones,
    isError: isZonesError,
    error: zonesError,
  } = useAvailabilityZones({
    provider,
    credentialId,
    region,
  });

  return {
    versions,
    regions,
    zones,
    isLoadingVersions,
    isLoadingRegions,
    isLoadingZones,
    versionsError: isVersionsError ? (versionsError as Error) : null,
    regionsError: isRegionsError ? (regionsError as Error) : null,
    zonesError: isZonesError ? (zonesError as Error) : null,
    canLoadMetadata,
  };
}

