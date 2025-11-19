/**
 * SSE Query Invalidation
 * 
 * SSE 이벤트를 받아 세밀하게 React Query 쿼리를 무효화하는 유틸리티
 * 실시간 업데이트가 실패하거나, 이벤트 데이터에 리소스 객체가 포함되지 않은 경우 사용
 */

import { QueryClient } from '@tanstack/react-query';
import { queryKeys } from '@/lib/query';
import type {
  KubernetesClusterEventData,
  NetworkVPCEventData,
  NetworkSubnetEventData,
  NetworkSecurityGroupEventData,
  VMEventData,
} from '@/lib/types/sse-events';
import { log } from '@/lib/logging';

/**
 * VM 이벤트에 대한 세밀한 쿼리 무효화
 */
export function invalidateVMQueries(
  queryClient: QueryClient,
  eventData: VMEventData,
  action: 'created' | 'updated' | 'deleted' | 'list'
): void {
  const { workspaceId, vmId, provider, region } = eventData;

  // workspaceId가 있으면 해당 workspace의 VM 목록만 무효화
  if (workspaceId) {
    queryClient.invalidateQueries({
      queryKey: queryKeys.vms.list(workspaceId),
      exact: false, // 하위 쿼리도 포함
    });

    // 특정 VM 상세 정보 무효화
    if (vmId && (action === 'updated' || action === 'deleted')) {
      queryClient.invalidateQueries({
        queryKey: queryKeys.vms.detail(vmId),
        exact: true,
      });
    }
  } else {
    // workspaceId가 없으면 모든 VM 쿼리 무효화 (fallback)
    log.warn('[SSE Query Invalidation] VM event missing workspaceId, invalidating all VM queries', { eventData });
    queryClient.invalidateQueries({
      queryKey: queryKeys.vms.all,
    });
  }

  // 대시보드 무효화 (항상 수행)
  queryClient.invalidateQueries({
    queryKey: queryKeys.dashboard.all,
  });
}

/**
 * Kubernetes 클러스터 이벤트에 대한 세밀한 쿼리 무효화
 */
export function invalidateKubernetesClusterQueries(
  queryClient: QueryClient,
  eventData: KubernetesClusterEventData,
  action: 'created' | 'updated' | 'deleted' | 'list'
): void {
  const { provider, credentialId, region, clusterId, cluster_name } = eventData;

  // 특정 provider, credentialId, region 조합의 클러스터 목록만 무효화
  if (provider && credentialId && region) {
    queryClient.invalidateQueries({
      queryKey: queryKeys.kubernetesClusters.list(
        undefined,
        provider,
        credentialId,
        region
      ),
      exact: false,
    });

    // 간단한 clusters 목록도 무효화
    queryClient.invalidateQueries({
      queryKey: queryKeys.clusters.list(provider, credentialId, region),
      exact: false,
    });
  } else {
    // 필수 파라미터가 없으면 해당 provider의 모든 클러스터 무효화
    if (provider) {
      log.warn('[SSE Query Invalidation] Cluster event missing credentialId/region, invalidating all cluster queries for provider', {
        provider,
        eventData,
      });
      queryClient.invalidateQueries({
        queryKey: [...queryKeys.kubernetesClusters.all, provider],
        exact: false,
      });
    } else {
      // provider도 없으면 모든 클러스터 무효화
      log.warn('[SSE Query Invalidation] Cluster event missing provider, invalidating all cluster queries', { eventData });
      queryClient.invalidateQueries({
        queryKey: queryKeys.kubernetesClusters.all,
      });
    }
  }

  // 특정 클러스터 상세 정보 무효화
  const detailKey = clusterId || cluster_name;
  if (detailKey && (action === 'updated' || action === 'deleted')) {
    queryClient.invalidateQueries({
      queryKey: queryKeys.kubernetesClusters.detail(detailKey),
      exact: true,
    });

    // 클러스터 관련 하위 리소스도 무효화
    if (provider && credentialId && region) {
      queryClient.invalidateQueries({
        queryKey: queryKeys.kubernetesClusters.nodePools(detailKey, provider, credentialId, region),
        exact: false,
      });
      queryClient.invalidateQueries({
        queryKey: queryKeys.kubernetesClusters.nodes(detailKey, provider, credentialId, region),
        exact: false,
      });
    }
  }

  // 대시보드 무효화 (항상 수행)
  queryClient.invalidateQueries({
    queryKey: queryKeys.dashboard.all,
  });
}

/**
 * VPC 이벤트에 대한 세밀한 쿼리 무효화
 */
export function invalidateVPCQueries(
  queryClient: QueryClient,
  eventData: NetworkVPCEventData,
  action: 'created' | 'updated' | 'deleted' | 'list'
): void {
  const { provider, credentialId, region, vpcId } = eventData;

  // 특정 provider, credentialId, region 조합의 VPC 목록만 무효화
  if (provider && credentialId && region) {
    queryClient.invalidateQueries({
      queryKey: queryKeys.vpcs.list(provider, credentialId, region),
      exact: false,
    });
  } else {
    // 필수 파라미터가 없으면 해당 provider의 모든 VPC 무효화
    if (provider) {
      log.warn('[SSE Query Invalidation] VPC event missing credentialId/region, invalidating all VPC queries for provider', {
        provider,
        eventData,
      });
      queryClient.invalidateQueries({
        queryKey: [...queryKeys.vpcs.all, provider],
        exact: false,
      });
    } else {
      log.warn('[SSE Query Invalidation] VPC event missing provider, invalidating all VPC queries', { eventData });
      queryClient.invalidateQueries({
        queryKey: queryKeys.vpcs.all,
      });
    }
  }

  // 특정 VPC 상세 정보 무효화
  if (vpcId && (action === 'updated' || action === 'deleted')) {
    queryClient.invalidateQueries({
      queryKey: queryKeys.vpcs.detail(vpcId),
      exact: true,
    });

    // VPC 관련 하위 리소스도 무효화 (Subnets, Security Groups)
    if (provider && credentialId && region) {
      queryClient.invalidateQueries({
        queryKey: queryKeys.subnets.list(provider, credentialId, vpcId, region),
        exact: false,
      });
      queryClient.invalidateQueries({
        queryKey: queryKeys.securityGroups.list(provider, credentialId, vpcId, region),
        exact: false,
      });
    }
  }

  // 대시보드 무효화 (항상 수행)
  queryClient.invalidateQueries({
    queryKey: queryKeys.dashboard.all,
  });
}

/**
 * Subnet 이벤트에 대한 세밀한 쿼리 무효화
 */
export function invalidateSubnetQueries(
  queryClient: QueryClient,
  eventData: NetworkSubnetEventData,
  action: 'created' | 'updated' | 'deleted' | 'list'
): void {
  const { provider, credentialId, vpcId, region, subnetId } = eventData;

  // 특정 provider, credentialId, vpcId, region 조합의 Subnet 목록만 무효화
  if (provider && credentialId && vpcId && region) {
    queryClient.invalidateQueries({
      queryKey: queryKeys.subnets.list(provider, credentialId, vpcId, region),
      exact: false,
    });
  } else {
    // 필수 파라미터가 없으면 해당 provider의 모든 Subnet 무효화
    if (provider) {
      log.warn('[SSE Query Invalidation] Subnet event missing credentialId/vpcId/region, invalidating all Subnet queries for provider', {
        provider,
        eventData,
      });
      queryClient.invalidateQueries({
        queryKey: [...queryKeys.subnets.all, provider],
        exact: false,
      });
    } else {
      log.warn('[SSE Query Invalidation] Subnet event missing provider, invalidating all Subnet queries', { eventData });
      queryClient.invalidateQueries({
        queryKey: queryKeys.subnets.all,
      });
    }
  }

  // 특정 Subnet 상세 정보 무효화
  if (subnetId && (action === 'updated' || action === 'deleted')) {
    queryClient.invalidateQueries({
      queryKey: queryKeys.subnets.detail(subnetId),
      exact: true,
    });
  }

  // 대시보드 무효화 (항상 수행)
  queryClient.invalidateQueries({
    queryKey: queryKeys.dashboard.all,
  });
}

/**
 * Security Group 이벤트에 대한 세밀한 쿼리 무효화
 */
export function invalidateSecurityGroupQueries(
  queryClient: QueryClient,
  eventData: NetworkSecurityGroupEventData,
  action: 'created' | 'updated' | 'deleted' | 'list'
): void {
  const { provider, credentialId, region, securityGroupId } = eventData;

  // 특정 provider, credentialId, region 조합의 Security Group 목록만 무효화
  // vpcId는 선택적이므로 undefined로 전달
  if (provider && credentialId && region) {
    queryClient.invalidateQueries({
      queryKey: queryKeys.securityGroups.list(provider, credentialId, undefined, region),
      exact: false,
    });
  } else {
    // 필수 파라미터가 없으면 해당 provider의 모든 Security Group 무효화
    if (provider) {
      log.warn('[SSE Query Invalidation] Security Group event missing credentialId/region, invalidating all Security Group queries for provider', {
        provider,
        eventData,
      });
      queryClient.invalidateQueries({
        queryKey: [...queryKeys.securityGroups.all, provider],
        exact: false,
      });
    } else {
      log.warn('[SSE Query Invalidation] Security Group event missing provider, invalidating all Security Group queries', { eventData });
      queryClient.invalidateQueries({
        queryKey: queryKeys.securityGroups.all,
      });
    }
  }

  // 특정 Security Group 상세 정보 무효화
  if (securityGroupId && (action === 'updated' || action === 'deleted')) {
    queryClient.invalidateQueries({
      queryKey: queryKeys.securityGroups.detail(securityGroupId),
      exact: true,
    });
  }
}

