/**
 * SSE Events Hook
 * 
 * React Query와 통합하여 SSE 이벤트 수신 시 자동으로 쿼리를 무효화하고 재조회합니다.
 */

import { useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { sseService } from '@/services/sse';
import { queryKeys } from '@/lib/query';
import { log } from '@/lib/logging';
import type { SSECallbacks } from '@/lib/types/sse';
import type {
  KubernetesClusterEventData,
  NetworkVPCEventData,
  NetworkSubnetEventData,
  NetworkSecurityGroupEventData,
  VMEventData,
} from '@/lib/types/sse-events';
import {
  applyVMCreatedUpdate,
  applyVMUpdatedUpdate,
  applyVMDeletedUpdate,
  applyKubernetesClusterCreatedUpdate,
  applyKubernetesClusterUpdatedUpdate,
  applyKubernetesClusterDeletedUpdate,
  applyVPCCreatedUpdate,
  applyVPCUpdatedUpdate,
  applyVPCDeletedUpdate,
  applySubnetCreatedUpdate,
  applySubnetUpdatedUpdate,
  applySubnetDeletedUpdate,
  applySecurityGroupCreatedUpdate,
  applySecurityGroupUpdatedUpdate,
  applySecurityGroupDeletedUpdate,
} from '@/lib/sse/query-updates';

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
      onConnected: async (data) => {
        log.debug('[SSE] Connected', { data });
        
        // 연결 후 기본 시스템 이벤트만 구독
        // 리소스 관련 이벤트는 각 페이지에서 동적으로 구독하도록 변경
        // 이를 통해 불필요한 이벤트 수신을 방지하고 네트워크 트래픽을 최적화
        try {
          // 시스템 이벤트만 기본 구독 (모든 페이지에서 필요)
          await sseService.subscribeToEvent('system-notification');
          await sseService.subscribeToEvent('system-alert');
          log.debug('[SSE] Auto-subscribed to system events');
        } catch (error) {
          log.error('[SSE] Failed to subscribe to system events', error, {
            service: 'SSE',
            action: 'auto-subscribe',
          });
        }
      },

      // Kubernetes 클러스터 이벤트 (실시간 업데이트)
      onKubernetesClusterCreated: (data) => {
        log.debug('[SSE] Kubernetes cluster created', { data });
        const eventData = data as KubernetesClusterEventData;
        try {
          // 실시간 업데이트 시도
          applyKubernetesClusterCreatedUpdate(queryClient, eventData);
        } catch (error) {
          log.warn('[SSE] Failed to apply real-time cluster created update, falling back to invalidation', error);
          // Fallback: 무효화
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
        }
        // 대시보드 무효화 (항상 수행)
        queryClient.invalidateQueries({
          queryKey: queryKeys.dashboard.all,
        });
      },

      onKubernetesClusterUpdated: (data) => {
        log.debug('[SSE] Kubernetes cluster updated', { data });
        const eventData = data as KubernetesClusterEventData;
        try {
          // 실시간 업데이트 시도
          applyKubernetesClusterUpdatedUpdate(queryClient, eventData);
        } catch (error) {
          log.warn('[SSE] Failed to apply real-time cluster updated update, falling back to invalidation', error);
          // Fallback: 무효화
          const { provider, credentialId, region, clusterId } = eventData;
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
          queryClient.invalidateQueries({
            queryKey: queryKeys.clusters.list(provider, credentialId, region),
          });
        }
        // 대시보드 무효화 (항상 수행)
        queryClient.invalidateQueries({
          queryKey: queryKeys.dashboard.all,
        });
      },

      onKubernetesClusterDeleted: (data) => {
        log.debug('[SSE] Kubernetes cluster deleted', { data });
        const eventData = data as KubernetesClusterEventData;
        try {
          // 실시간 업데이트 시도
          applyKubernetesClusterDeletedUpdate(queryClient, eventData);
        } catch (error) {
          log.warn('[SSE] Failed to apply real-time cluster deleted update, falling back to invalidation', error);
          // Fallback: 무효화
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
        }
        // 대시보드 무효화 (항상 수행)
        queryClient.invalidateQueries({
          queryKey: queryKeys.dashboard.all,
        });
      },

      onKubernetesClusterList: (data) => {
        log.debug('[SSE] Kubernetes cluster list updated', { data });
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

      // Network VPC 이벤트 (실시간 업데이트)
      onNetworkVPCCreated: (data) => {
        log.debug('[SSE] Network VPC created', { data });
        const eventData = data as NetworkVPCEventData;
        try {
          // 실시간 업데이트 시도
          applyVPCCreatedUpdate(queryClient, eventData);
        } catch (error) {
          log.warn('[SSE] Failed to apply real-time VPC created update, falling back to invalidation', error);
          // Fallback: 무효화
          const { provider, credentialId, region } = eventData;
          queryClient.invalidateQueries({
            queryKey: queryKeys.vpcs.list(provider, credentialId, region),
          });
        }
        // 대시보드 무효화 (항상 수행)
        queryClient.invalidateQueries({
          queryKey: queryKeys.dashboard.all,
        });
      },

      onNetworkVPCUpdated: (data) => {
        log.debug('[SSE] Network VPC updated', { data });
        const eventData = data as NetworkVPCEventData;
        try {
          // 실시간 업데이트 시도
          applyVPCUpdatedUpdate(queryClient, eventData);
        } catch (error) {
          log.warn('[SSE] Failed to apply real-time VPC updated update, falling back to invalidation', error);
          // Fallback: 무효화
          const { provider, credentialId, region, vpcId } = eventData;
          queryClient.invalidateQueries({
            queryKey: queryKeys.vpcs.list(provider, credentialId, region),
          });
          if (vpcId) {
            queryClient.invalidateQueries({
              queryKey: queryKeys.vpcs.detail(vpcId),
            });
          }
        }
        // 대시보드 무효화 (항상 수행)
        queryClient.invalidateQueries({
          queryKey: queryKeys.dashboard.all,
        });
      },

      onNetworkVPCDeleted: (data) => {
        log.debug('[SSE] Network VPC deleted', { data });
        const eventData = data as NetworkVPCEventData;
        try {
          // 실시간 업데이트 시도
          applyVPCDeletedUpdate(queryClient, eventData);
        } catch (error) {
          log.warn('[SSE] Failed to apply real-time VPC deleted update, falling back to invalidation', error);
          // Fallback: 무효화
          const { provider, credentialId, region } = eventData;
          queryClient.invalidateQueries({
            queryKey: queryKeys.vpcs.list(provider, credentialId, region),
          });
        }
        // 대시보드 무효화 (항상 수행)
        queryClient.invalidateQueries({
          queryKey: queryKeys.dashboard.all,
        });
      },

      onNetworkVPCList: (data) => {
        log.debug('[SSE] Network VPC list updated', { data });
        const eventData = data as NetworkVPCEventData;
        const { provider, credentialId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.vpcs.list(provider, credentialId, region),
        });
      },

      // Network Subnet 이벤트 (실시간 업데이트)
      onNetworkSubnetCreated: (data) => {
        const eventData = data as NetworkSubnetEventData;
        try {
          // 실시간 업데이트 시도
          applySubnetCreatedUpdate(queryClient, eventData);
        } catch (error) {
          log.warn('[SSE] Failed to apply real-time subnet created update, falling back to invalidation', error);
          // Fallback: 무효화
          const { provider, credentialId, vpcId, region } = eventData;
          queryClient.invalidateQueries({
            queryKey: queryKeys.subnets.list(provider, credentialId, vpcId, region),
          });
        }
        // 대시보드 무효화 (항상 수행)
        queryClient.invalidateQueries({
          queryKey: queryKeys.dashboard.all,
        });
      },

      onNetworkSubnetUpdated: (data) => {
        const eventData = data as NetworkSubnetEventData;
        try {
          // 실시간 업데이트 시도
          applySubnetUpdatedUpdate(queryClient, eventData);
        } catch (error) {
          log.warn('[SSE] Failed to apply real-time subnet updated update, falling back to invalidation', error);
          // Fallback: 무효화
          const { provider, credentialId, vpcId, region } = eventData;
          queryClient.invalidateQueries({
            queryKey: queryKeys.subnets.list(provider, credentialId, vpcId, region),
          });
        }
      },

      onNetworkSubnetDeleted: (data) => {
        const eventData = data as NetworkSubnetEventData;
        try {
          // 실시간 업데이트 시도
          applySubnetDeletedUpdate(queryClient, eventData);
        } catch (error) {
          log.warn('[SSE] Failed to apply real-time subnet deleted update, falling back to invalidation', error);
          // Fallback: 무효화
          const { provider, credentialId, vpcId, region } = eventData;
          queryClient.invalidateQueries({
            queryKey: queryKeys.subnets.list(provider, credentialId, vpcId, region),
          });
        }
        // 대시보드 무효화 (항상 수행)
        queryClient.invalidateQueries({
          queryKey: queryKeys.dashboard.all,
        });
      },

      onNetworkSubnetList: (data) => {
        const eventData = data as NetworkSubnetEventData;
        const { provider, credentialId, vpcId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.subnets.list(provider, credentialId, vpcId, region),
        });
      },

      // Network Security Group 이벤트 (실시간 업데이트)
      onNetworkSecurityGroupCreated: (data) => {
        const eventData = data as NetworkSecurityGroupEventData;
        try {
          // 실시간 업데이트 시도
          applySecurityGroupCreatedUpdate(queryClient, eventData);
        } catch (error) {
          log.warn('[SSE] Failed to apply real-time security group created update, falling back to invalidation', error);
          // Fallback: 무효화
          const { provider, credentialId, region } = eventData;
          queryClient.invalidateQueries({
            queryKey: queryKeys.securityGroups.list(provider, credentialId, undefined, region),
          });
        }
        // 대시보드 무효화 (항상 수행)
        queryClient.invalidateQueries({
          queryKey: queryKeys.dashboard.all,
        });
      },

      onNetworkSecurityGroupUpdated: (data) => {
        const eventData = data as NetworkSecurityGroupEventData;
        try {
          // 실시간 업데이트 시도
          applySecurityGroupUpdatedUpdate(queryClient, eventData);
        } catch (error) {
          log.warn('[SSE] Failed to apply real-time security group updated update, falling back to invalidation', error);
          // Fallback: 무효화
          const { provider, credentialId, region } = eventData;
          queryClient.invalidateQueries({
            queryKey: queryKeys.securityGroups.list(provider, credentialId, undefined, region),
          });
        }
      },

      onNetworkSecurityGroupDeleted: (data) => {
        const eventData = data as NetworkSecurityGroupEventData;
        try {
          // 실시간 업데이트 시도
          applySecurityGroupDeletedUpdate(queryClient, eventData);
        } catch (error) {
          log.warn('[SSE] Failed to apply real-time security group deleted update, falling back to invalidation', error);
          // Fallback: 무효화
          const { provider, credentialId, region } = eventData;
          queryClient.invalidateQueries({
            queryKey: queryKeys.securityGroups.list(provider, credentialId, undefined, region),
          });
        }
      },

      onNetworkSecurityGroupList: (data) => {
        const eventData = data as NetworkSecurityGroupEventData;
        const { provider, credentialId, region } = eventData;
        queryClient.invalidateQueries({
          queryKey: queryKeys.securityGroups.list(provider, credentialId, undefined, region),
        });
      },

      // VM 이벤트 (실시간 업데이트)
      onVMCreated: (data) => {
        log.debug('[SSE] VM created', { data });
        const eventData = data as VMEventData;
        try {
          // 실시간 업데이트 시도
          applyVMCreatedUpdate(queryClient, eventData);
        } catch (error) {
          log.warn('[SSE] Failed to apply real-time VM created update, falling back to invalidation', error);
          // Fallback: 무효화
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
        }
        // 대시보드 무효화 (항상 수행)
        queryClient.invalidateQueries({
          queryKey: queryKeys.dashboard.all,
        });
      },

      onVMUpdated: (data) => {
        log.debug('[SSE] VM updated', { data });
        const eventData = data as VMEventData;
        try {
          // 실시간 업데이트 시도
          applyVMUpdatedUpdate(queryClient, eventData);
        } catch (error) {
          log.warn('[SSE] Failed to apply real-time VM updated update, falling back to invalidation', error);
          // Fallback: 무효화
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
        }
        // 대시보드 무효화 (항상 수행)
        queryClient.invalidateQueries({
          queryKey: queryKeys.dashboard.all,
        });
      },

      onVMDeleted: (data) => {
        log.debug('[SSE] VM deleted', { data });
        const eventData = data as VMEventData;
        try {
          // 실시간 업데이트 시도
          applyVMDeletedUpdate(queryClient, eventData);
        } catch (error) {
          log.warn('[SSE] Failed to apply real-time VM deleted update, falling back to invalidation', error);
          // Fallback: 무효화
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
        }
        // 대시보드 무효화 (항상 수행)
        queryClient.invalidateQueries({
          queryKey: queryKeys.dashboard.all,
        });
      },

      onVMList: (data) => {
        log.debug('[SSE] VM list updated', { data });
        const eventData = data as VMEventData;
        const { workspaceId } = eventData;
        if (workspaceId) {
          queryClient.invalidateQueries({
            queryKey: queryKeys.vms.list(workspaceId),
          });
          // 대시보드 쿼리 무효화 (모든 credentialId/region 조합)
          queryClient.invalidateQueries({
            queryKey: queryKeys.dashboard.all,
          });
        } else {
          queryClient.invalidateQueries({
            queryKey: queryKeys.vms.all,
          });
        }
      },

      onError: (event) => {
        // SSEErrorInfo 타입인지 확인
        if (event && typeof event === 'object' && 'type' in event && event.type === 'SSE') {
          // SSEErrorInfo 타입인 경우
          const errorInfo = event as { type: string; readyState?: number; url?: string; timestamp?: string; message?: string };
          const readyStateText = errorInfo.readyState === 0 ? 'CONNECTING' : 
                                errorInfo.readyState === 1 ? 'OPEN' : 
                                errorInfo.readyState === 2 ? 'CLOSED' : 'UNKNOWN';
          
          const errorMessage = errorInfo.message || 'SSE connection error';
          const error = new Error(errorMessage);
          error.name = 'SSEError';
          
          log.error('[SSE] Error', error, {
            type: errorInfo.type,
            readyState: errorInfo.readyState,
            readyStateText,
            url: errorInfo.url,
            timestamp: errorInfo.timestamp,
          });
        } else {
          // Event 타입이거나 다른 타입인 경우
          const errorInfo = typeof event === 'object' && event !== null
            ? { ...event, timestamp: new Date().toISOString() }
            : { error: event, timestamp: new Date().toISOString() };
          
          const error = new Error('SSE connection error');
          error.name = 'SSEError';
          
          log.error('[SSE] Error', error, errorInfo);
        }
      },
    };

    sseService.connect(token, callbacks);

    return () => {
      sseService.disconnect();
    };
  }, [token, queryClient]);
}

