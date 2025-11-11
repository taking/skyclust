/**
 * Network Resources Hook
 * 네트워크 리소스 (VPCs, Subnets, Security Groups) 데이터 fetching 및 상태 관리 통합 훅
 * 
 * 중복된 패턴을 통합하여 일관된 인터페이스를 제공합니다.
 * 
 * @example
 * ```tsx
 * // VPCs만 가져오기
 * const { vpcs, isLoadingVPCs, ... } = useNetworkResources({ resourceType: 'vpcs' });
 * 
 * // Subnets 가져오기 (VPC 선택 필요)
 * const { subnets, isLoadingSubnets, vpcs, selectedVPCId, setSelectedVPCId, ... } = useNetworkResources({ 
 *   resourceType: 'subnets',
 *   requireVPC: true 
 * });
 * 
 * // Security Groups 가져오기 (VPC 선택 필요)
 * const { securityGroups, isLoadingSecurityGroups, vpcs, selectedVPCId, setSelectedVPCId, ... } = useNetworkResources({ 
 *   resourceType: 'securityGroups',
 *   requireVPC: true 
 * });
 * ```
 */

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { networkService } from '@/services/network';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import { useCredentials } from '@/hooks/use-credentials';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useWorkspaceStore } from '@/store/workspace';
import type { VPC, Subnet, SecurityGroup } from '@/lib/types/network';

export type NetworkResourceType = 'vpcs' | 'subnets' | 'securityGroups';

export interface UseNetworkResourcesOptions {
  /**
   * 가져올 리소스 타입
   */
  resourceType: NetworkResourceType;
  
  /**
   * VPC 선택이 필요한지 여부 (subnets, securityGroups의 경우 true)
   */
  requireVPC?: boolean;
  
  /**
   * 초기 VPC ID (선택적)
   */
  initialVPCId?: string;
}

export interface UseNetworkResourcesReturn<TResource = VPC | Subnet | SecurityGroup> {
  // 공통 반환값
  credentials: ReturnType<typeof useCredentials>['credentials'];
  selectedCredential: ReturnType<typeof useCredentials>['selectedCredential'];
  selectedProvider: ReturnType<typeof useCredentials>['selectedProvider'];
  selectedCredentialId: string;
  selectedRegion: string;
  
  // VPCs (항상 포함)
  vpcs: VPC[];
  isLoadingVPCs: boolean;
  
  // VPC 선택 관련 (requireVPC가 true인 경우만)
  selectedVPCId?: string;
  setSelectedVPCId?: (vpcId: string) => void;
  
  // 리소스 타입별 반환값
  subnets?: Subnet[];
  isLoadingSubnets?: boolean;
  securityGroups?: SecurityGroup[];
  isLoadingSecurityGroups?: boolean;
}

/**
 * Network Resources 통합 훅
 */
export function useNetworkResources(
  options: UseNetworkResourcesOptions
): UseNetworkResourcesReturn {
  const { resourceType, requireVPC = false, initialVPCId } = options;
  
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedCredentialId, selectedRegion } = useCredentialContext();
  const [selectedVPCId, setSelectedVPCId] = useState<string>(initialVPCId || '');

  const { credentials, selectedCredential, selectedProvider } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
  });

  const watchedCredentialId = selectedCredentialId || '';
  const watchedRegion = selectedRegion || '';

  // VPCs 가져오기 (항상 필요: 직접 리소스이거나 subnets/securityGroups의 선택을 위해)
  const { data: vpcs = [], isLoading: isLoadingVPCs } = useQuery({
    queryKey: queryKeys.vpcs.list(selectedProvider, watchedCredentialId, watchedRegion),
    queryFn: async () => {
      if (!selectedProvider || !watchedCredentialId) return [];
      return networkService.listVPCs(selectedProvider, watchedCredentialId, watchedRegion);
    },
    enabled: !!selectedProvider && !!watchedCredentialId && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
  });

  // Subnets 가져오기 (resourceType이 'subnets'인 경우)
  const { data: subnets = [], isLoading: isLoadingSubnets } = useQuery({
    queryKey: queryKeys.subnets.list(selectedProvider, watchedCredentialId, selectedVPCId, watchedRegion),
    queryFn: async () => {
      if (!selectedProvider || !watchedCredentialId || !selectedVPCId || !watchedRegion) return [];
      return networkService.listSubnets(selectedProvider, watchedCredentialId, selectedVPCId, watchedRegion);
    },
    enabled: resourceType === 'subnets' && !!selectedProvider && !!watchedCredentialId && !!selectedVPCId && !!watchedRegion && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
  });

  // Security Groups 가져오기 (resourceType이 'securityGroups'인 경우)
  const { data: securityGroups = [], isLoading: isLoadingSecurityGroups } = useQuery({
    queryKey: queryKeys.securityGroups.list(selectedProvider, watchedCredentialId, selectedVPCId, watchedRegion),
    queryFn: async () => {
      if (!selectedProvider || !watchedCredentialId || !currentWorkspace || !selectedVPCId || !watchedRegion) {
        return [];
      }
      return networkService.listSecurityGroups(selectedProvider, watchedCredentialId, selectedVPCId, watchedRegion);
    },
    enabled: resourceType === 'securityGroups' && !!selectedProvider && !!watchedCredentialId && !!currentWorkspace && !!selectedVPCId && !!watchedRegion,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
  });

  // 반환값 구성
  const baseReturn = {
    credentials,
    selectedCredential,
    selectedProvider,
    selectedCredentialId: watchedCredentialId,
    selectedRegion: watchedRegion,
    vpcs,
    isLoadingVPCs,
  };

  // VPC 선택이 필요한 경우
  if (requireVPC) {
    return {
      ...baseReturn,
      selectedVPCId,
      setSelectedVPCId,
      ...(resourceType === 'subnets' && {
        subnets,
        isLoadingSubnets,
      }),
      ...(resourceType === 'securityGroups' && {
        securityGroups,
        isLoadingSecurityGroups,
      }),
    };
  }

  // VPCs만 필요한 경우
  return baseReturn;
}

