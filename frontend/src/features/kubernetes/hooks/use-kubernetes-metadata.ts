/**
 * Kubernetes Metadata Hook
 * Kubernetes 버전, Region, Availability Zone 목록 조회
 */

import { useQuery } from '@tanstack/react-query';
import { kubernetesService } from '../services/kubernetes';
import type { CloudProvider } from '@/lib/types';
import { queryKeys } from '@/lib/query-keys';
import { CACHE_TIMES, GC_TIMES } from '@/lib/query-client';

interface UseKubernetesMetadataOptions {
  provider?: CloudProvider;
  credentialId?: string;
  region?: string;
}

/**
 * EKS Kubernetes 버전 목록 조회
 */
export function useEKSVersions({
  provider,
  credentialId,
  region,
}: UseKubernetesMetadataOptions) {
  return useQuery({
    queryKey: queryKeys.kubernetesMetadata.versions(provider, credentialId, region),
    queryFn: async () => {
      if (!provider || !credentialId || !region || provider !== 'aws') {
        return [];
      }
      return kubernetesService.getEKSVersions(provider, credentialId, region);
    },
    enabled: !!provider && !!credentialId && !!region && provider === 'aws',
    staleTime: CACHE_TIMES.STATIC, // 1시간 - 버전 목록은 자주 변하지 않음
    gcTime: GC_TIMES.LONG, // 24시간
  });
}

/**
 * AWS Region 목록 조회
 */
export function useAWSRegions({
  provider,
  credentialId,
}: UseKubernetesMetadataOptions) {
  return useQuery({
    queryKey: queryKeys.kubernetesMetadata.regions(provider, credentialId),
    queryFn: async () => {
      if (!provider || !credentialId || provider !== 'aws') {
        return [];
      }
      return kubernetesService.getAWSRegions(provider, credentialId);
    },
    enabled: !!provider && !!credentialId && provider === 'aws',
    staleTime: CACHE_TIMES.STATIC, // 1시간 - Region 목록은 매우 자주 변하지 않음
    gcTime: GC_TIMES.LONG, // 24시간
  });
}

/**
 * Availability Zone 목록 조회
 */
export function useAvailabilityZones({
  provider,
  credentialId,
  region,
}: UseKubernetesMetadataOptions) {
  return useQuery({
    queryKey: queryKeys.kubernetesMetadata.availabilityZones(provider, credentialId, region),
    queryFn: async () => {
      if (!provider || !credentialId || !region || provider !== 'aws') {
        return [];
      }
      return kubernetesService.getAvailabilityZones(provider, credentialId, region);
    },
    enabled: !!provider && !!credentialId && !!region && provider === 'aws',
    staleTime: CACHE_TIMES.STATIC, // 1시간 - Zone 목록은 자주 변하지 않음
    gcTime: GC_TIMES.LONG, // 24시간
  });
}

