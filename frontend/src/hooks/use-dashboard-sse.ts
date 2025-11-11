/**
 * useDashboardSSE Hook
 * 대시보드 위젯과 요약 정보에 필요한 SSE 이벤트를 동적으로 구독/해제합니다.
 */

import { useEffect, useRef } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { sseService } from '@/services/sse';
import {
  getAllRequiredEvents,
  getDashboardSummaryEvents,
} from '@/lib/sse/widget-events';
import type { WidgetData } from '@/lib/widgets';
import { log } from '@/lib/logging';
import { logger } from '@/lib/logging/logger';
import { queryKeys } from '@/lib/query';
import type {
  KubernetesClusterEventData,
  NetworkVPCEventData,
  NetworkSubnetEventData,
  NetworkSecurityGroupEventData,
  VMEventData,
} from '@/lib/types/sse-events';

interface UseDashboardSSEOptions {
  widgets: WidgetData[];
  credentialId?: string;
  region?: string;
  includeSummary?: boolean;
  enabled?: boolean;
}

/**
 * 대시보드 SSE 동적 구독 관리 훅
 * 
 * 위젯 목록과 필터 옵션에 따라 필요한 이벤트만 구독하고,
 * 위젯이 추가/제거되거나 필터가 변경될 때 자동으로 구독을 동기화합니다.
 * 
 * @param options - 구독 옵션
 * @example
 * ```tsx
 * const { widgets } = useDashboard();
 * const { selectedCredentialId, selectedRegion } = useCredentialContext();
 * 
 * useDashboardSSE({
 *   widgets,
 *   credentialId: selectedCredentialId,
 *   region: selectedRegion,
 *   includeSummary: true,
 * });
 * ```
 */
export function useDashboardSSE({
  widgets,
  credentialId,
  region,
  includeSummary = true,
  enabled = true,
}: UseDashboardSSEOptions): void {
  const queryClient = useQueryClient();
  const previousRequiredEventsRef = useRef<Set<string>>(new Set());
  const previousFiltersRef = useRef<{
    credentialId?: string;
    region?: string;
  }>({});

  useEffect(() => {
    if (!enabled) {
      return;
    }

    // SSE 연결 확인
    if (!sseService.isConnected()) {
      log.debug('[Dashboard SSE] SSE not connected, skipping subscription sync');
      return;
    }

    // 필요한 이벤트 계산
    const requiredEvents = includeSummary
      ? getAllRequiredEvents(widgets, true)
      : getAllRequiredEvents(widgets, false);

    // 필터 옵션 준비
    const filters = {
      credential_ids: credentialId ? [credentialId] : undefined,
      regions: region ? [region] : undefined,
    };

    // 이전 상태와 비교하여 변경사항 확인
    const eventsChanged =
      !setsEqual(requiredEvents, previousRequiredEventsRef.current);
    const filtersChanged =
      previousFiltersRef.current.credentialId !== credentialId ||
      previousFiltersRef.current.region !== region;

    // 변경사항이 없으면 스킵
    if (!eventsChanged && !filtersChanged) {
      return;
    }

    // 구독 동기화
    const syncSubscriptions = async () => {
      try {
        await sseService.syncSubscriptions(requiredEvents, filters);
        previousRequiredEventsRef.current = requiredEvents;
        previousFiltersRef.current = { credentialId, region };
        log.debug('[Dashboard SSE] Subscription sync completed', {
          eventCount: requiredEvents.size,
          filters,
        });
      } catch (error) {
        logger.logError(error, {
          service: 'SSE',
          action: 'syncDashboardSubscriptions',
          widgets: widgets.map((w) => w.type),
          filters,
        });
      }
    };

    syncSubscriptions();
  }, [widgets, credentialId, region, includeSummary, enabled]);

  // SSE 이벤트 리스너 등록: 대시보드 요약 정보 무효화
  useEffect(() => {
    if (!enabled || !sseService.isConnected()) {
      return;
    }

    const callbacks = {
      // VM 이벤트: 대시보드 요약 정보 무효화
      onVMCreated: (data: unknown) => {
        const eventData = data as VMEventData;
        invalidateDashboardSummary(eventData.credentialId, eventData.region);
      },
      onVMUpdated: (data: unknown) => {
        const eventData = data as VMEventData;
        invalidateDashboardSummary(eventData.credentialId, eventData.region);
      },
      onVMDeleted: (data: unknown) => {
        const eventData = data as VMEventData;
        invalidateDashboardSummary(eventData.credentialId, eventData.region);
      },

      // Kubernetes 이벤트: 대시보드 요약 정보 무효화
      onKubernetesClusterCreated: (data: unknown) => {
        const eventData = data as KubernetesClusterEventData;
        invalidateDashboardSummary(eventData.credentialId, eventData.region);
      },
      onKubernetesClusterUpdated: (data: unknown) => {
        const eventData = data as KubernetesClusterEventData;
        invalidateDashboardSummary(eventData.credentialId, eventData.region);
      },
      onKubernetesClusterDeleted: (data: unknown) => {
        const eventData = data as KubernetesClusterEventData;
        invalidateDashboardSummary(eventData.credentialId, eventData.region);
      },

      // Network 이벤트: 대시보드 요약 정보 무효화
      onNetworkVPCCreated: (data: unknown) => {
        const eventData = data as NetworkVPCEventData;
        invalidateDashboardSummary(eventData.credentialId, eventData.region);
      },
      onNetworkVPCUpdated: (data: unknown) => {
        const eventData = data as NetworkVPCEventData;
        invalidateDashboardSummary(eventData.credentialId, eventData.region);
      },
      onNetworkVPCDeleted: (data: unknown) => {
        const eventData = data as NetworkVPCEventData;
        invalidateDashboardSummary(eventData.credentialId, eventData.region);
      },
      onNetworkSubnetCreated: (data: unknown) => {
        const eventData = data as NetworkSubnetEventData;
        invalidateDashboardSummary(eventData.credentialId, eventData.region);
      },
      onNetworkSubnetUpdated: (data: unknown) => {
        const eventData = data as NetworkSubnetEventData;
        invalidateDashboardSummary(eventData.credentialId, eventData.region);
      },
      onNetworkSubnetDeleted: (data: unknown) => {
        const eventData = data as NetworkSubnetEventData;
        invalidateDashboardSummary(eventData.credentialId, eventData.region);
      },
      onNetworkSecurityGroupCreated: (data: unknown) => {
        const eventData = data as NetworkSecurityGroupEventData;
        invalidateDashboardSummary(eventData.credentialId, eventData.region);
      },
      onNetworkSecurityGroupUpdated: (data: unknown) => {
        const eventData = data as NetworkSecurityGroupEventData;
        invalidateDashboardSummary(eventData.credentialId, eventData.region);
      },
      onNetworkSecurityGroupDeleted: (data: unknown) => {
        const eventData = data as NetworkSecurityGroupEventData;
        invalidateDashboardSummary(eventData.credentialId, eventData.region);
      },
    };

    // 대시보드 요약 정보 무효화 헬퍼 함수
    function invalidateDashboardSummary(
      eventCredentialId?: string,
      eventRegion?: string
    ) {
      // 필터와 일치하는 경우에만 무효화
      if (
        credentialId &&
        eventCredentialId &&
        credentialId !== eventCredentialId
      ) {
        return;
      }
      if (region && eventRegion && region !== eventRegion) {
        return;
      }

      // 대시보드 요약 정보 무효화
      queryClient.invalidateQueries({
        queryKey: queryKeys.dashboard.all,
      });
      log.debug('[Dashboard SSE] Invalidated dashboard summary', {
        eventCredentialId,
        eventRegion,
        filterCredentialId: credentialId,
        filterRegion: region,
      });
    }

    // 기존 콜백에 추가 (기존 콜백은 유지)
    sseService.updateCallbacks(callbacks);

    // cleanup: 컴포넌트 언마운트 시 콜백 제거하지 않음 (다른 컴포넌트에서도 사용 가능)
    // 대신 필터 조건을 확인하여 무효화 여부를 결정
  }, [enabled, credentialId, region, queryClient]);
}

/**
 * 두 Set이 동일한지 확인합니다.
 */
function setsEqual<T>(set1: Set<T>, set2: Set<T>): boolean {
  if (set1.size !== set2.size) {
    return false;
  }
  for (const item of set1) {
    if (!set2.has(item)) {
      return false;
    }
  }
  return true;
}

