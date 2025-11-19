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

import { useState, useMemo } from 'react';
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
  
  /**
   * Props로 전달받은 credential ID (Context 대신 사용)
   */
  credentialId?: string;
  
  /**
   * Props로 전달받은 region (Context 대신 사용)
   */
  region?: string;
  
  /**
   * Props로 전달받은 zone (Subnet 필터링용)
   */
  zone?: string;
  
  /**
   * Context 대신 Props 사용 여부
   */
  useProps?: boolean;
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
  const { 
    resourceType, 
    requireVPC = false, 
    initialVPCId,
    credentialId: propsCredentialId,
    region: propsRegion,
    zone: propsZone,
    useProps = false,
  } = options;
  
  const { currentWorkspace } = useWorkspaceStore();
  const context = useCredentialContext();
  const [selectedVPCId, setSelectedVPCId] = useState<string>(initialVPCId || '');

  // Props 우선, 없으면 Context 사용
  const effectiveCredentialId = useProps && propsCredentialId 
    ? propsCredentialId 
    : context.selectedCredentialId || '';
    
  const effectiveRegion = useProps && propsRegion
    ? propsRegion
    : context.selectedRegion || '';

  const { credentials, selectedCredential, selectedProvider } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: effectiveCredentialId || undefined,
  });

  const watchedCredentialId = effectiveCredentialId;
  const watchedRegion = effectiveRegion;

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
  const { data: rawSubnets = [], isLoading: isLoadingSubnets } = useQuery({
    queryKey: queryKeys.subnets.list(selectedProvider, watchedCredentialId, selectedVPCId, watchedRegion),
    queryFn: async () => {
      if (!selectedProvider || !watchedCredentialId || !selectedVPCId || !watchedRegion) return [];
      return networkService.listSubnets(selectedProvider, watchedCredentialId, selectedVPCId, watchedRegion);
    },
    enabled: resourceType === 'subnets' && !!selectedProvider && !!watchedCredentialId && !!selectedVPCId && !!watchedRegion && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
  });

  // Zone 필터링 제거: AWS EKS는 최소 2개의 다른 AZ에 서브넷이 필요하므로 전체 zone의 subnets를 표시
  // GCP, Azure도 전체 zone의 subnets를 조회하여 사용자가 선택할 수 있도록 함
  const subnets = rawSubnets;

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

