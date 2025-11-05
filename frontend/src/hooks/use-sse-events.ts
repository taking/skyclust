/**
 * SSE Events Hook
 * 
 * React Query와 통합하여 SSE 이벤트 수신 시 자동으로 쿼리를 무효화하고 재조회합니다.
 */

import { useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { sseService } from '@/services/sse';
import { queryKeys } from '@/lib/query-keys';
import type { SSECallbacks } from '@/lib/types/sse';
import type {
  KubernetesClusterEventData,
  NetworkVPCEventData,
  NetworkSubnetEventData,
  NetworkSecurityGroupEventData,
  VMEventData,
} from '@/lib/types/sse-events';

/**
 * SSE 이벤트를 구독하고 React Query와 통합하는 훅
 */
export function useSSEEvents(token: string | null) {
  const queryClient = useQueryClient();

  useEffect(() => {
    if (!token) {
      return;
    }

    const callbacks: SSECallbacks = {
      onConnected: (data) => {
        if (process.env.NODE_ENV === 'development') {
          console.log('[SSE] Connected:', data);
        }
      },

      // Kubernetes 클러스터 이벤트
      onKubernetesClusterCreated: (data) => {
        if (process.env.NODE_ENV === 'development') {
          console.log('[SSE] Kubernetes cluster created:', data);
        }
        // Backend 필드명과 일치: credentialId, clusterId (not credential_id, cluster_id)
        const eventData = data as KubernetesClusterEventData;
        const { provider, credentialId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.kubernetesClusters.list(
            undefined,
            provider,
            credentialId,
            region
          ),
        });
        queryClient.invalidateQueries({
          queryKey: queryKeys.clusters.list(provider, credentialId, region),
        });
      },

      onKubernetesClusterUpdated: (data) => {
        if (process.env.NODE_ENV === 'development') {
          console.log('[SSE] Kubernetes cluster updated:', data);
        }
        const eventData = data as KubernetesClusterEventData;
        const { provider, credentialId, region, clusterId } = eventData;
        // 목록 쿼리 무효화
        queryClient.invalidateQueries({
          queryKey: queryKeys.kubernetesClusters.list(
            undefined,
            provider,
            credentialId,
            region
          ),
        });
        queryClient.invalidateQueries({
          queryKey: queryKeys.clusters.list(provider, credentialId, region),
        });
        // 개별 클러스터 쿼리 무효화
        if (clusterId) {
          queryClient.invalidateQueries({
            queryKey: queryKeys.kubernetesClusters.detail(clusterId),
          });
        }
      },

      onKubernetesClusterDeleted: (data) => {
        if (process.env.NODE_ENV === 'development') {
          console.log('[SSE] Kubernetes cluster deleted:', data);
        }
        const eventData = data as KubernetesClusterEventData;
        const { provider, credentialId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.kubernetesClusters.list(
            undefined,
            provider,
            credentialId,
            region
          ),
        });
        queryClient.invalidateQueries({
          queryKey: queryKeys.clusters.list(provider, credentialId, region),
        });
      },

      onKubernetesClusterList: (data) => {
        if (process.env.NODE_ENV === 'development') {
          console.log('[SSE] Kubernetes cluster list updated:', data);
        }
        const eventData = data as KubernetesClusterEventData;
        const { provider, credentialId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.kubernetesClusters.list(
            undefined,
            provider,
            credentialId,
            region
          ),
        });
        queryClient.invalidateQueries({
          queryKey: queryKeys.clusters.list(provider, credentialId, region),
        });
      },

      // Kubernetes Node Pool 이벤트
      onKubernetesNodePoolCreated: (data) => {
        const eventData = data as KubernetesClusterEventData;
        const { provider, credentialId, clusterId } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.nodePools.list(provider, credentialId, undefined, clusterId),
        });
      },

      onKubernetesNodePoolUpdated: (data) => {
        const eventData = data as KubernetesClusterEventData;
        const { provider, credentialId, clusterId } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.nodePools.list(provider, credentialId, undefined, clusterId),
        });
      },

      onKubernetesNodePoolDeleted: (data) => {
        const eventData = data as KubernetesClusterEventData;
        const { provider, credentialId, clusterId } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.nodePools.list(provider, credentialId, undefined, clusterId),
        });
      },

      // Kubernetes Node 이벤트
      onKubernetesNodeCreated: (data) => {
        const eventData = data as KubernetesClusterEventData;
        const { provider, credentialId, clusterId } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.nodes.list(provider, clusterId, credentialId),
        });
      },

      onKubernetesNodeUpdated: (data) => {
        const eventData = data as KubernetesClusterEventData;
        const { provider, credentialId, clusterId } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.nodes.list(provider, clusterId, credentialId),
        });
      },

      onKubernetesNodeDeleted: (data) => {
        const eventData = data as KubernetesClusterEventData;
        const { provider, credentialId, clusterId } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.nodes.list(provider, clusterId, credentialId),
        });
      },

      // Network VPC 이벤트
      onNetworkVPCCreated: (data) => {
        if (process.env.NODE_ENV === 'development') {
          console.log('[SSE] Network VPC created:', data);
        }
        const eventData = data as NetworkVPCEventData;
        const { provider, credentialId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.vpcs.list(provider, credentialId, region),
        });
      },

      onNetworkVPCUpdated: (data) => {
        if (process.env.NODE_ENV === 'development') {
          console.log('[SSE] Network VPC updated:', data);
        }
        const eventData = data as NetworkVPCEventData;
        const { provider, credentialId, region, vpcId } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.vpcs.list(provider, credentialId, region),
        });
        if (vpcId) {
          queryClient.invalidateQueries({
            queryKey: queryKeys.vpcs.detail(vpcId),
          });
        }
      },

      onNetworkVPCDeleted: (data) => {
        if (process.env.NODE_ENV === 'development') {
          console.log('[SSE] Network VPC deleted:', data);
        }
        const eventData = data as NetworkVPCEventData;
        const { provider, credentialId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.vpcs.list(provider, credentialId, region),
        });
      },

      onNetworkVPCList: (data) => {
        if (process.env.NODE_ENV === 'development') {
          console.log('[SSE] Network VPC list updated:', data);
        }
        const eventData = data as NetworkVPCEventData;
        const { provider, credentialId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.vpcs.list(provider, credentialId, region),
        });
      },

      // Network Subnet 이벤트
      onNetworkSubnetCreated: (data) => {
        const eventData = data as NetworkSubnetEventData;
        const { provider, credentialId, vpcId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.subnets.list(provider, credentialId, vpcId, region),
        });
      },

      onNetworkSubnetUpdated: (data) => {
        const eventData = data as NetworkSubnetEventData;
        const { provider, credentialId, vpcId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.subnets.list(provider, credentialId, vpcId, region),
        });
      },

      onNetworkSubnetDeleted: (data) => {
        const eventData = data as NetworkSubnetEventData;
        const { provider, credentialId, vpcId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.subnets.list(provider, credentialId, vpcId, region),
        });
      },

      onNetworkSubnetList: (data) => {
        const eventData = data as NetworkSubnetEventData;
        const { provider, credentialId, vpcId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.subnets.list(provider, credentialId, vpcId, region),
        });
      },

      // Network Security Group 이벤트
      onNetworkSecurityGroupCreated: (data) => {
        const eventData = data as NetworkSecurityGroupEventData;
        const { provider, credentialId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.securityGroups.list(provider, credentialId, undefined, region),
        });
      },

      onNetworkSecurityGroupUpdated: (data) => {
        const eventData = data as NetworkSecurityGroupEventData;
        const { provider, credentialId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.securityGroups.list(provider, credentialId, undefined, region),
        });
      },

      onNetworkSecurityGroupDeleted: (data) => {
        const eventData = data as NetworkSecurityGroupEventData;
        const { provider, credentialId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.securityGroups.list(provider, credentialId, undefined, region),
        });
      },

      onNetworkSecurityGroupList: (data) => {
        const eventData = data as NetworkSecurityGroupEventData;
        const { provider, credentialId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.securityGroups.list(provider, credentialId, undefined, region),
        });
      },

      // VM 이벤트
      onVMCreated: (data) => {
        if (process.env.NODE_ENV === 'development') {
          console.log('[SSE] VM created:', data);
        }
        const eventData = data as VMEventData;
        const { workspaceId } = eventData;
        if (workspaceId) {
          queryClient.invalidateQueries({
            queryKey: queryKeys.vms.list(workspaceId),
          });
        } else {
          queryClient.invalidateQueries({
            queryKey: queryKeys.vms.all,
          });
        }
      },

      onVMUpdated: (data) => {
        if (process.env.NODE_ENV === 'development') {
          console.log('[SSE] VM updated:', data);
        }
        const eventData = data as VMEventData;
        const { vmId, workspaceId } = eventData;
        if (workspaceId) {
          queryClient.invalidateQueries({
            queryKey: queryKeys.vms.list(workspaceId),
          });
        } else {
          queryClient.invalidateQueries({
            queryKey: queryKeys.vms.all,
          });
        }
        if (vmId) {
          queryClient.invalidateQueries({
            queryKey: queryKeys.vms.detail(vmId),
          });
        }
      },

      onVMDeleted: (data) => {
        if (process.env.NODE_ENV === 'development') {
          console.log('[SSE] VM deleted:', data);
        }
        const eventData = data as VMEventData;
        const { workspaceId } = eventData;
        if (workspaceId) {
          queryClient.invalidateQueries({
            queryKey: queryKeys.vms.list(workspaceId),
          });
        } else {
          queryClient.invalidateQueries({
            queryKey: queryKeys.vms.all,
          });
        }
      },

      onVMList: (data) => {
        if (process.env.NODE_ENV === 'development') {
          console.log('[SSE] VM list updated:', data);
        }
        const eventData = data as VMEventData;
        const { workspaceId } = eventData;
        if (workspaceId) {
          queryClient.invalidateQueries({
            queryKey: queryKeys.vms.list(workspaceId),
          });
        } else {
          queryClient.invalidateQueries({
            queryKey: queryKeys.vms.all,
          });
        }
      },

      onError: (event) => {
        if (process.env.NODE_ENV === 'development') {
          // SSEErrorInfo 타입인지 확인
          if (event && typeof event === 'object' && 'type' in event && event.type === 'SSE') {
            // SSEErrorInfo 타입인 경우 구조화된 정보로 표시
            const errorInfo = event as { type: string; readyState?: number; url?: string; timestamp?: string; message?: string };
            const readyStateText = errorInfo.readyState === 0 ? 'CONNECTING' : 
                                  errorInfo.readyState === 1 ? 'OPEN' : 
                                  errorInfo.readyState === 2 ? 'CLOSED' : 'UNKNOWN';
            
            const errorDetails = {
              type: errorInfo.type,
              message: errorInfo.message || 'SSE connection error',
              readyState: errorInfo.readyState,
              readyStateText: readyStateText,
              url: errorInfo.url,
              timestamp: errorInfo.timestamp,
            };
            
            // Safari 등 일부 브라우저에서 객체가 제대로 표시되지 않을 수 있으므로 JSON으로 변환
            try {
              console.error('[SSE] Error:', JSON.stringify(errorDetails, null, 2));
            } catch (_e) {
              // JSON.stringify 실패 시 각 속성을 개별 출력
              console.error('[SSE] Error:', errorInfo.message || 'SSE connection error');
              console.error('  Type:', errorInfo.type);
              console.error('  ReadyState:', errorInfo.readyState, `(${readyStateText})`);
              console.error('  URL:', errorInfo.url);
              console.error('  Timestamp:', errorInfo.timestamp);
            }
          } else {
            // Event 타입이거나 다른 타입인 경우 변환
            const errorInfo = typeof event === 'object' && event !== null
              ? { ...event, timestamp: new Date().toISOString() }
              : { error: event, timestamp: new Date().toISOString() };
            
            try {
              console.error('[SSE] Error:', JSON.stringify(errorInfo, null, 2));
            } catch (_e) {
              console.error('[SSE] Error:', String(event));
            }
          }
        }
      },
    };

    sseService.connect(token, callbacks);

    return () => {
      sseService.disconnect();
    };
  }, [token, queryClient]);
}

