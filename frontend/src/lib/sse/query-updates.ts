/**
 * SSE Query Updates
 * 
 * SSE 이벤트를 받아 React Query 캐시를 실시간으로 업데이트하는 유틸리티
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
import type { VM } from '@/lib/types/vm';
import type { KubernetesCluster } from '@/lib/types/kubernetes';
import type { VPC, Subnet, SecurityGroup } from '@/lib/types/network';
import { log } from '@/lib/logging';

/**
 * VM 생성 이벤트를 캐시에 반영
 */
export function applyVMCreatedUpdate(
  queryClient: QueryClient,
  eventData: VMEventData
): void {
  if (!eventData.vmId || !eventData.workspaceId) {
    return;
  }

  // 이벤트 데이터에서 VM 객체 추출 시도
  const vm = (eventData as unknown as { vm?: VM }).vm;
  if (!vm) {
    // VM 객체가 없으면 무효화만 수행
    queryClient.invalidateQueries({
      queryKey: queryKeys.vms.list(eventData.workspaceId),
    });
    return;
  }

  // 캐시에 직접 추가
  queryClient.setQueryData<VM[]>(
    queryKeys.vms.list(eventData.workspaceId),
    (oldData) => {
      if (!oldData) return [vm];
      // 중복 체크
      if (oldData.some((v) => v.id === vm.id)) {
        return oldData;
      }
      return [...oldData, vm];
    }
  );
}

/**
 * VM 업데이트 이벤트를 캐시에 반영
 */
export function applyVMUpdatedUpdate(
  queryClient: QueryClient,
  eventData: VMEventData
): void {
  if (!eventData.workspaceId) {
    log.warn('[SSE Query Update] VM updated event missing workspaceId', { eventData });
    return;
  }

  const updatedVM = (eventData as unknown as { vm?: VM; data?: { vm?: VM } }).vm ||
                    (eventData as unknown as { data?: { vm?: VM } }).data?.vm;
  
  if (!updatedVM) {
    // VM 객체가 없으면 무효화만 수행
    log.debug('[SSE Query Update] VM updated event missing VM object, using invalidation', { eventData });
    queryClient.invalidateQueries({
      queryKey: queryKeys.vms.list(eventData.workspaceId),
    });
    if (eventData.vmId) {
      queryClient.invalidateQueries({
        queryKey: queryKeys.vms.detail(eventData.vmId),
      });
    }
    return;
  }

  // 목록 캐시 업데이트 (실시간 업데이트)
  queryClient.setQueryData<VM[]>(
    queryKeys.vms.list(eventData.workspaceId),
    (oldData) => {
      if (!oldData) return [updatedVM];
      const updated = oldData.map((v) => (v.id === updatedVM.id ? updatedVM : v));
      log.debug('[SSE Query Update] Real-time updated VM in list cache', { vmId: updatedVM.id });
      return updated;
    }
  );

  // 상세 캐시 업데이트
  if (eventData.vmId || updatedVM.id) {
    const vmId = eventData.vmId || updatedVM.id;
    queryClient.setQueryData<VM>(
      queryKeys.vms.detail(vmId),
      updatedVM
    );
    log.debug('[SSE Query Update] Real-time updated VM detail cache', { vmId });
  }
}

/**
 * VM 삭제 이벤트를 캐시에 반영
 */
export function applyVMDeletedUpdate(
  queryClient: QueryClient,
  eventData: VMEventData
): void {
  if (!eventData.workspaceId) {
    log.warn('[SSE Query Update] VM deleted event missing workspaceId', { eventData });
    return;
  }

  if (!eventData.vmId) {
    // vmId가 없으면 무효화만 수행
    log.debug('[SSE Query Update] VM deleted event missing vmId, using invalidation', { eventData });
    queryClient.invalidateQueries({
      queryKey: queryKeys.vms.list(eventData.workspaceId),
    });
    return;
  }

  // 목록 캐시에서 제거 (실시간 업데이트)
  queryClient.setQueryData<VM[]>(
    queryKeys.vms.list(eventData.workspaceId),
    (oldData) => {
      if (!oldData) return [];
      const filtered = oldData.filter((v) => v.id !== eventData.vmId);
      log.debug('[SSE Query Update] Real-time removed VM from list cache', { vmId: eventData.vmId });
      return filtered;
    }
  );

  // 상세 캐시 제거
  queryClient.removeQueries({
    queryKey: queryKeys.vms.detail(eventData.vmId),
  });
}

/**
 * Kubernetes 클러스터 생성 이벤트를 캐시에 반영
 */
export function applyKubernetesClusterCreatedUpdate(
  queryClient: QueryClient,
  eventData: KubernetesClusterEventData
): void {
  const { provider, credentialId, region } = eventData;
  const cluster = (eventData as unknown as { cluster?: KubernetesCluster })
    .cluster;

  if (!cluster) {
    // 클러스터 객체가 없으면 무효화만 수행
    queryClient.invalidateQueries({
      queryKey: queryKeys.kubernetesClusters.list(
        undefined,
        provider,
        credentialId,
        region
      ),
    });
    return;
  }

  // 캐시에 직접 추가
  queryClient.setQueryData<KubernetesCluster[]>(
    queryKeys.kubernetesClusters.list(undefined, provider, credentialId, region),
    (oldData) => {
      if (!oldData) return [cluster];
      // 중복 체크
      if (oldData.some((c) => c.id === cluster.id || c.name === cluster.name)) {
        return oldData;
      }
      return [...oldData, cluster];
    }
  );
}

/**
 * Kubernetes 클러스터 업데이트 이벤트를 캐시에 반영
 */
export function applyKubernetesClusterUpdatedUpdate(
  queryClient: QueryClient,
  eventData: KubernetesClusterEventData
): void {
  const { provider, credentialId, region, clusterId } = eventData;
  const updatedCluster = (eventData as unknown as {
    cluster?: KubernetesCluster;
    data?: { cluster?: KubernetesCluster };
  }).cluster ||
  (eventData as unknown as { data?: { cluster?: KubernetesCluster } }).data?.cluster;

  if (!updatedCluster) {
    // 클러스터 객체가 없으면 무효화만 수행
    log.debug('[SSE Query Update] Cluster updated event missing cluster object, using invalidation', { eventData });
    queryClient.invalidateQueries({
      queryKey: queryKeys.kubernetesClusters.list(
        undefined,
        provider,
        credentialId,
        region
      ),
    });
    if (clusterId) {
      queryClient.invalidateQueries({
        queryKey: queryKeys.kubernetesClusters.detail(clusterId),
      });
    }
    return;
  }

  // 목록 캐시 업데이트 (실시간 업데이트)
  queryClient.setQueryData<KubernetesCluster[]>(
    queryKeys.kubernetesClusters.list(undefined, provider, credentialId, region),
    (oldData) => {
      if (!oldData) return [updatedCluster];
      const updated = oldData.map((c) =>
        c.id === updatedCluster.id || c.name === updatedCluster.name
          ? updatedCluster
          : c
      );
      log.debug('[SSE Query Update] Real-time updated cluster in list cache', { clusterId: updatedCluster.id, clusterName: updatedCluster.name });
      return updated;
    }
  );

  // 상세 캐시 업데이트
  const detailKey = clusterId || updatedCluster.name;
  if (detailKey) {
    queryClient.setQueryData<KubernetesCluster>(
      queryKeys.kubernetesClusters.detail(detailKey),
      updatedCluster
    );
    log.debug('[SSE Query Update] Real-time updated cluster detail cache', { clusterId: detailKey });
  }
}

/**
 * Kubernetes 클러스터 삭제 이벤트를 캐시에 반영
 */
export function applyKubernetesClusterDeletedUpdate(
  queryClient: QueryClient,
  eventData: KubernetesClusterEventData
): void {
  const { provider, credentialId, region, clusterId, cluster_name } = eventData;

  if (!clusterId && !cluster_name) {
    // clusterId와 cluster_name이 모두 없으면 무효화만 수행
    log.debug('[SSE Query Update] Cluster deleted event missing clusterId/cluster_name, using invalidation', { eventData });
    queryClient.invalidateQueries({
      queryKey: queryKeys.kubernetesClusters.list(undefined, provider, credentialId, region),
    });
    return;
  }

  // 목록 캐시에서 제거 (실시간 업데이트)
  queryClient.setQueryData<KubernetesCluster[]>(
    queryKeys.kubernetesClusters.list(undefined, provider, credentialId, region),
    (oldData) => {
      if (!oldData) return [];
      const filtered = oldData.filter(
        (c) => c.id !== clusterId && c.name !== cluster_name
      );
      log.debug('[SSE Query Update] Real-time removed cluster from list cache', { clusterId, clusterName: cluster_name });
      return filtered;
    }
  );

  // 상세 캐시 제거
  const detailKey = clusterId || cluster_name;
  if (detailKey) {
    queryClient.removeQueries({
      queryKey: queryKeys.kubernetesClusters.detail(detailKey),
    });
  }
}

/**
 * VPC 생성 이벤트를 캐시에 반영
 */
export function applyVPCCreatedUpdate(
  queryClient: QueryClient,
  eventData: NetworkVPCEventData
): void {
  const { provider, credentialId, region } = eventData;
  const vpc = (eventData as unknown as { vpc?: VPC }).vpc;

  if (!vpc) {
    queryClient.invalidateQueries({
      queryKey: queryKeys.vpcs.list(provider, credentialId, region),
    });
    return;
  }

  queryClient.setQueryData<VPC[]>(
    queryKeys.vpcs.list(provider, credentialId, region),
    (oldData) => {
      if (!oldData) return [vpc];
      if (oldData.some((v) => v.id === vpc.id)) {
        return oldData;
      }
      return [...oldData, vpc];
    }
  );
}

/**
 * VPC 업데이트 이벤트를 캐시에 반영
 */
export function applyVPCUpdatedUpdate(
  queryClient: QueryClient,
  eventData: NetworkVPCEventData
): void {
  const { provider, credentialId, region, vpcId } = eventData;
  const updatedVPC = (eventData as unknown as { vpc?: VPC; data?: { vpc?: VPC } }).vpc ||
                     (eventData as unknown as { data?: { vpc?: VPC } }).data?.vpc;

  if (!updatedVPC) {
    log.debug('[SSE Query Update] VPC updated event missing VPC object, using invalidation', { eventData });
    queryClient.invalidateQueries({
      queryKey: queryKeys.vpcs.list(provider, credentialId, region),
    });
    if (vpcId) {
      queryClient.invalidateQueries({
        queryKey: queryKeys.vpcs.detail(vpcId),
      });
    }
    return;
  }

  queryClient.setQueryData<VPC[]>(
    queryKeys.vpcs.list(provider, credentialId, region),
    (oldData) => {
      if (!oldData) return [updatedVPC];
      const updated = oldData.map((v) => (v.id === updatedVPC.id ? updatedVPC : v));
      log.debug('[SSE Query Update] Real-time updated VPC in list cache', { vpcId: updatedVPC.id });
      return updated;
    }
  );

  const detailKey = vpcId || updatedVPC.id;
  if (detailKey) {
    queryClient.setQueryData<VPC>(queryKeys.vpcs.detail(detailKey), updatedVPC);
    log.debug('[SSE Query Update] Real-time updated VPC detail cache', { vpcId: detailKey });
  }
}

/**
 * VPC 삭제 이벤트를 캐시에 반영
 */
export function applyVPCDeletedUpdate(
  queryClient: QueryClient,
  eventData: NetworkVPCEventData
): void {
  const { provider, credentialId, region, vpcId } = eventData;

  if (!vpcId) {
    log.debug('[SSE Query Update] VPC deleted event missing vpcId, using invalidation', { eventData });
    queryClient.invalidateQueries({
      queryKey: queryKeys.vpcs.list(provider, credentialId, region),
    });
    return;
  }

  queryClient.setQueryData<VPC[]>(
    queryKeys.vpcs.list(provider, credentialId, region),
    (oldData) => {
      if (!oldData) return [];
      const filtered = oldData.filter((v) => v.id !== vpcId);
      log.debug('[SSE Query Update] Real-time removed VPC from list cache', { vpcId });
      return filtered;
    }
  );

  queryClient.removeQueries({
    queryKey: queryKeys.vpcs.detail(vpcId),
  });
}

/**
 * Subnet 생성 이벤트를 캐시에 반영
 */
export function applySubnetCreatedUpdate(
  queryClient: QueryClient,
  eventData: NetworkSubnetEventData
): void {
  const { provider, credentialId, vpcId, region } = eventData;
  const subnet = (eventData as unknown as { subnet?: Subnet }).subnet;

  if (!subnet) {
    queryClient.invalidateQueries({
      queryKey: queryKeys.subnets.list(provider, credentialId, vpcId, region),
    });
    return;
  }

  queryClient.setQueryData<Subnet[]>(
    queryKeys.subnets.list(provider, credentialId, vpcId, region),
    (oldData) => {
      if (!oldData) return [subnet];
      if (oldData.some((s) => s.id === subnet.id)) {
        return oldData;
      }
      return [...oldData, subnet];
    }
  );
}

/**
 * Subnet 업데이트 이벤트를 캐시에 반영
 */
export function applySubnetUpdatedUpdate(
  queryClient: QueryClient,
  eventData: NetworkSubnetEventData
): void {
  const { provider, credentialId, vpcId, region, subnetId } = eventData;
  const updatedSubnet = (eventData as unknown as { subnet?: Subnet; data?: { subnet?: Subnet } }).subnet ||
                        (eventData as unknown as { data?: { subnet?: Subnet } }).data?.subnet;

  if (!updatedSubnet) {
    log.debug('[SSE Query Update] Subnet updated event missing Subnet object, using invalidation', { eventData });
    queryClient.invalidateQueries({
      queryKey: queryKeys.subnets.list(provider, credentialId, vpcId, region),
    });
    return;
  }

  queryClient.setQueryData<Subnet[]>(
    queryKeys.subnets.list(provider, credentialId, vpcId, region),
    (oldData) => {
      if (!oldData) return [updatedSubnet];
      const updated = oldData.map((s) =>
        s.id === updatedSubnet.id ? updatedSubnet : s
      );
      log.debug('[SSE Query Update] Real-time updated Subnet in list cache', { subnetId: updatedSubnet.id });
      return updated;
    }
  );
}

/**
 * Subnet 삭제 이벤트를 캐시에 반영
 */
export function applySubnetDeletedUpdate(
  queryClient: QueryClient,
  eventData: NetworkSubnetEventData
): void {
  const { provider, credentialId, vpcId, region, subnetId } = eventData;

  if (!subnetId) {
    log.debug('[SSE Query Update] Subnet deleted event missing subnetId, using invalidation', { eventData });
    queryClient.invalidateQueries({
      queryKey: queryKeys.subnets.list(provider, credentialId, vpcId, region),
    });
    return;
  }

  queryClient.setQueryData<Subnet[]>(
    queryKeys.subnets.list(provider, credentialId, vpcId, region),
    (oldData) => {
      if (!oldData) return [];
      const filtered = oldData.filter((s) => s.id !== subnetId);
      log.debug('[SSE Query Update] Real-time removed Subnet from list cache', { subnetId });
      return filtered;
    }
  );
}

/**
 * Security Group 생성 이벤트를 캐시에 반영
 */
export function applySecurityGroupCreatedUpdate(
  queryClient: QueryClient,
  eventData: NetworkSecurityGroupEventData
): void {
  const { provider, credentialId, region } = eventData;
  const securityGroup = (eventData as unknown as {
    securityGroup?: SecurityGroup;
    data?: { securityGroup?: SecurityGroup };
  }).securityGroup ||
  (eventData as unknown as { data?: { securityGroup?: SecurityGroup } }).data?.securityGroup;

  if (!securityGroup) {
    log.debug('[SSE Query Update] Security Group created event missing SecurityGroup object, using invalidation', { eventData });
    queryClient.invalidateQueries({
      queryKey: queryKeys.securityGroups.list(
        provider,
        credentialId,
        undefined,
        region
      ),
    });
    return;
  }

  queryClient.setQueryData<SecurityGroup[]>(
    queryKeys.securityGroups.list(provider, credentialId, undefined, region),
    (oldData) => {
      if (!oldData) return [securityGroup];
      if (oldData.some((sg) => sg.id === securityGroup.id)) {
        log.debug('[SSE Query Update] Security Group already exists in cache, skipping', { securityGroupId: securityGroup.id });
        return oldData;
      }
      log.debug('[SSE Query Update] Real-time added Security Group to cache', { securityGroupId: securityGroup.id });
      return [...oldData, securityGroup];
    }
  );
}

/**
 * Security Group 업데이트 이벤트를 캐시에 반영
 */
export function applySecurityGroupUpdatedUpdate(
  queryClient: QueryClient,
  eventData: NetworkSecurityGroupEventData
): void {
  const { provider, credentialId, region, securityGroupId } = eventData;
  const updatedSecurityGroup = (eventData as unknown as {
    securityGroup?: SecurityGroup;
    data?: { securityGroup?: SecurityGroup };
  }).securityGroup ||
  (eventData as unknown as { data?: { securityGroup?: SecurityGroup } }).data?.securityGroup;

  if (!updatedSecurityGroup) {
    log.debug('[SSE Query Update] Security Group updated event missing SecurityGroup object, using invalidation', { eventData });
    queryClient.invalidateQueries({
      queryKey: queryKeys.securityGroups.list(
        provider,
        credentialId,
        undefined,
        region
      ),
    });
    return;
  }

  queryClient.setQueryData<SecurityGroup[]>(
    queryKeys.securityGroups.list(provider, credentialId, undefined, region),
    (oldData) => {
      if (!oldData) return [updatedSecurityGroup];
      const updated = oldData.map((sg) =>
        sg.id === updatedSecurityGroup.id ? updatedSecurityGroup : sg
      );
      log.debug('[SSE Query Update] Real-time updated Security Group in list cache', { securityGroupId: updatedSecurityGroup.id });
      return updated;
    }
  );
}

/**
 * Security Group 삭제 이벤트를 캐시에 반영
 */
export function applySecurityGroupDeletedUpdate(
  queryClient: QueryClient,
  eventData: NetworkSecurityGroupEventData
): void {
  const { provider, credentialId, region, securityGroupId } = eventData;

  if (!securityGroupId) {
    log.debug('[SSE Query Update] Security Group deleted event missing securityGroupId, using invalidation', { eventData });
    queryClient.invalidateQueries({
      queryKey: queryKeys.securityGroups.list(provider, credentialId, undefined, region),
    });
    return;
  }

  queryClient.setQueryData<SecurityGroup[]>(
    queryKeys.securityGroups.list(provider, credentialId, undefined, region),
    (oldData) => {
      if (!oldData) return [];
      const filtered = oldData.filter((sg) => sg.id !== securityGroupId);
      log.debug('[SSE Query Update] Real-time removed Security Group from list cache', { securityGroupId });
      return filtered;
    }
  );
}

