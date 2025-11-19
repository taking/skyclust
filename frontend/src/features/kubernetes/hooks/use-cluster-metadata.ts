/**
 * Cluster Metadata Hook
 * 클러스터 생성 시 필요한 메타데이터 (버전, 리전, 존) 로딩 훅
 * 
 * 클러스터 생성 시 필요한 Kubernetes 버전, 리전, 가용 영역 정보를 통합 관리합니다.
 * - AWS: 버전, 리전, Zone 지원
 * - GCP: Zone 지원
 * - Azure: Zone 지원
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

import { useKubernetesVersions, useAWSRegions, useAvailabilityZones } from './use-kubernetes-metadata';
import type { CloudProvider } from '@/lib/types';
import type { RegionOption } from '@/lib/regions';
import { AWS_REGIONS } from '@/lib/regions';

export interface UseClusterMetadataOptions {
  /** 클라우드 프로바이더 */
  provider?: CloudProvider;
  /** Credential ID */
  credentialId: string;
  /** 선택된 리전 */
  region?: string;
  /** Workspace ID */
  workspaceId?: string;
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
  /** 메타데이터 로딩 가능 여부 (provider와 credentialId가 있는 경우) */
  canLoadMetadata: boolean;
}

/**
 * 클러스터 메타데이터 로딩 훅
 */
export function useClusterMetadata(
  options: UseClusterMetadataOptions
): UseClusterMetadataReturn {
  const { provider, credentialId, region = '', workspaceId } = options;

  const canLoadMetadata = !!provider && !!credentialId;

  // Fetch Kubernetes versions (AWS, GCP, Azure)
  const {
    data: versions = [],
    isLoading: isLoadingVersions,
    isError: isVersionsError,
    error: versionsError,
  } = useKubernetesVersions({
    provider,
    credentialId,
    region,
    workspaceId,
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
    workspaceId,
  });

  // Convert string[] to RegionOption[] with city names from AWS_REGIONS
  const regions: RegionOption[] = awsRegionsData.map(region => {
    // AWS_REGIONS에서 해당 region을 찾아서 label(도시명) 매핑
    const regionInfo = AWS_REGIONS.find(r => r.value === region);
    return {
      value: region,
      label: regionInfo?.label || region, // 도시명이 있으면 사용, 없으면 region 값 사용
    };
  });

  // Fetch availability zones (AWS, GCP, Azure - when region is selected)
  const {
    data: zones = [],
    isLoading: isLoadingZones,
    isError: isZonesError,
    error: zonesError,
  } = useAvailabilityZones({
    provider,
    credentialId,
    region,
    workspaceId,
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

