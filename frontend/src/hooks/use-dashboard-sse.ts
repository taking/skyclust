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
import { useSSEStatus } from '@/hooks/use-sse-status';
import type {
  KubernetesClusterEventData,
  NetworkVPCEventData,
  NetworkSubnetEventData,
  NetworkSecurityGroupEventData,
  VMEventData,
  DashboardSummaryEventData,
} from '@/lib/types/sse-events';
import { applyDashboardSummaryUpdatedUpdate } from '@/lib/sse/query-updates';

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
  const { status: sseStatus } = useSSEStatus();
  const previousRequiredEventsRef = useRef<Set<string>>(new Set());
  const previousFiltersRef = useRef<{
    credentialId?: string;
    region?: string;
  }>({});

  // Dashboard Summary 이벤트 구독 (통합 이벤트 방식)
  useEffect(() => {
    if (!enabled || !includeSummary) {
      return;
    }

    // SSE 연결 완료 확인
    if (!sseStatus.isConnected) {
      log.debug('[Dashboard SSE] SSE not connected, skipping dashboard summary subscription', {
        isConnected: sseStatus.isConnected,
        readyState: sseStatus.readyState,
      });
      return;
    }

    const subscribeToDashboardSummary = async () => {
      try {
        await sseService.subscribeToEvent('dashboard-summary-updated', {
          credential_ids: credentialId ? [credentialId] : undefined,
          regions: region ? [region] : undefined,
        });
        log.debug('[Dashboard SSE] Subscribed to dashboard-summary-updated', {
          credentialId,
          region,
          clientId: sseService.getClientId(),
        });
      } catch (error) {
        logger.logError(error, {
          service: 'SSE',
          action: 'subscribeDashboardSummary',
          credentialId,
          region,
        });
      }
    };

    subscribeToDashboardSummary();

    // Cleanup: 구독 해제
    return () => {
      const unsubscribe = async () => {
        try {
          await sseService.unsubscribeFromEvent('dashboard-summary-updated', {
            credential_ids: credentialId ? [credentialId] : undefined,
            regions: region ? [region] : undefined,
          });
          log.debug('[Dashboard SSE] Unsubscribed from dashboard-summary-updated', {
            credentialId,
            region,
          });
        } catch (error) {
          log.warn('[Dashboard SSE] Failed to unsubscribe from dashboard-summary-updated', error, {
            service: 'SSE',
            action: 'unsubscribeDashboardSummary',
          });
        }
      };
      unsubscribe();
    };
  }, [enabled, includeSummary, credentialId, region, sseStatus.isConnected]);

  // Widget별 이벤트 구독 (필요한 경우에만)
  useEffect(() => {
    if (!enabled || widgets.length === 0) {
      // 위젯이 없으면 이전 구독 정리
      if (previousRequiredEventsRef.current.size > 0) {
        const syncSubscriptions = async () => {
          try {
            if (sseStatus.isConnected) {
              await sseService.syncSubscriptions(new Set<string>(), {
                credential_ids: credentialId ? [credentialId] : undefined,
                regions: region ? [region] : undefined,
              });
              previousRequiredEventsRef.current = new Set();
              previousFiltersRef.current = {};
              log.debug('[Dashboard SSE] Cleared widget subscriptions (no widgets)');
            }
          } catch (error) {
            log.warn('[Dashboard SSE] Failed to clear widget subscriptions', error);
          }
        };
        syncSubscriptions();
      }
      return;
    }

    // SSE 연결 확인
    if (!sseStatus.isConnected) {
      log.debug('[Dashboard SSE] SSE not connected, skipping widget subscription sync');
      return;
    }

    // 필요한 이벤트 계산 (dashboard-summary 이벤트 제외)
    const requiredEvents = getAllRequiredEvents(widgets, false);

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
        log.debug('[Dashboard SSE] Widget subscription sync completed', {
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
  }, [widgets, credentialId, region, enabled, sseStatus.isConnected]);

  // Dashboard Summary 이벤트 핸들러 등록 (실시간 업데이트)
  useEffect(() => {
    if (!enabled || !includeSummary || !sseStatus.isConnected) {
      return;
    }

    const callbacks = {
      // Dashboard Summary 이벤트: 직접 캐시 업데이트
      onDashboardSummaryUpdated: (data: unknown) => {
        const eventData = data as DashboardSummaryEventData;
        
        // 필터와 일치하는 경우에만 업데이트
        if (
          credentialId &&
          eventData.credential_id &&
          credentialId !== eventData.credential_id
        ) {
          return;
        }
        if (region && eventData.region && region !== eventData.region) {
          return;
        }

        try {
          // 실시간 업데이트 시도
          applyDashboardSummaryUpdatedUpdate(queryClient, eventData);
          log.debug('[Dashboard SSE] Real-time updated dashboard summary', {
            workspaceId: eventData.workspace_id,
            credentialId: eventData.credential_id,
            region: eventData.region,
          });
        } catch (error) {
          log.warn('[Dashboard SSE] Failed to apply real-time dashboard summary update, falling back to invalidation', error);
          // Fallback: 무효화
          queryClient.invalidateQueries({
            queryKey: queryKeys.dashboard.summary(
              eventData.workspace_id,
              eventData.credential_id,
              eventData.region
            ),
          });
        }
      },
    };

    // 기존 콜백에 추가
    sseService.updateCallbacks(callbacks);

    // cleanup: 컴포넌트 언마운트 시 콜백 제거하지 않음 (다른 컴포넌트에서도 사용 가능)
  }, [enabled, includeSummary, credentialId, region, queryClient, sseStatus.isConnected]);
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

