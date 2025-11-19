/**
 * SSE Subscription Hook
 * 
 * SSE 이벤트 구독을 안전하게 관리하는 훅
 * 컴포넌트 unmount 시 자동으로 구독 해제를 보장합니다.
 * 
 * SSE Subscription Manager를 사용하여 중앙화된 구독 관리를 제공합니다.
 */

'use client';

import { useEffect, useRef } from 'react';
import { sseSubscriptionManager } from '@/services/sse-subscription-manager';
import { log } from '@/lib/logging';
import { useSSEStatus } from './use-sse-status';
import type { SubscriptionFilters } from '@/services/sse-subscription-manager';

export type { SubscriptionFilters };

export interface UseSSESubscriptionOptions {
  /**
   * 구독할 이벤트 타입 배열
   */
  eventTypes: string[];
  
  /**
   * 구독 필터
   */
  filters?: SubscriptionFilters;
  
  /**
   * 구독 활성화 여부
   */
  enabled?: boolean;
  
  /**
   * 구독 성공 콜백
   */
  onSubscribed?: (eventTypes: string[]) => void;
  
  /**
   * 구독 실패 콜백
   */
  onSubscriptionError?: (error: unknown, eventTypes: string[]) => void;
}

/**
 * useSSESubscription Hook
 * 
 * SSE 이벤트를 안전하게 구독하고, 컴포넌트 unmount 시 자동으로 해제합니다.
 * SSE Subscription Manager를 사용하여 중복 구독을 방지하고 참조 카운팅을 관리합니다.
 * 
 * @example
 * ```tsx
 * useSSESubscription({
 *   eventTypes: ['kubernetes-cluster-created', 'kubernetes-cluster-updated'],
 *   filters: { credential_ids: [credentialId], regions: [region] },
 *   enabled: !!credentialId && !!region,
 * });
 * ```
 */
export function useSSESubscription({
  eventTypes,
  filters,
  enabled = true,
  onSubscribed,
  onSubscriptionError,
}: UseSSESubscriptionOptions): void {
  const { status: sseStatus } = useSSEStatus();
  const subscriberIdRef = useRef<string | null>(null);
  const isUnmountingRef = useRef(false);
  const previousEventTypesRef = useRef<string[]>([]);
  const previousFiltersRef = useRef<SubscriptionFilters | undefined>(undefined);

  useEffect(() => {
    isUnmountingRef.current = false;
    
    return () => {
      isUnmountingRef.current = true;
      
      // 컴포넌트 unmount 시 구독 해제
      if (subscriberIdRef.current) {
        sseSubscriptionManager.unsubscribe(subscriberIdRef.current).catch((error) => {
          log.error('[SSE Subscription] Failed to unsubscribe on unmount', error, {
            service: 'SSE',
            action: 'unsubscribe',
            subscriberId: subscriberIdRef.current,
          });
        });
        subscriberIdRef.current = null;
      }
    };
  }, []);

  useEffect(() => {
    if (!enabled || eventTypes.length === 0) {
      // 구독 비활성화 시 기존 구독 해제
      if (subscriberIdRef.current) {
        sseSubscriptionManager.unsubscribe(subscriberIdRef.current).catch((error) => {
          log.error('[SSE Subscription] Failed to unsubscribe when disabled', error, {
            service: 'SSE',
            action: 'unsubscribe',
            subscriberId: subscriberIdRef.current,
          });
        });
        subscriberIdRef.current = null;
      }
      return;
    }

    // SSE가 연결되지 않았으면 대기 (연결되면 자동으로 구독 생성)
    if (!sseStatus.isConnected) {
      log.debug('[SSE Subscription] SSE not connected, waiting for connection', {
        eventTypes,
        filters,
      });
      return;
    }

    let isCancelled = false;

    const subscribe = async () => {
      if (isUnmountingRef.current || isCancelled) {
        return;
      }

      // 이벤트 타입이나 필터가 변경되었는지 확인
      const eventTypesChanged = 
        previousEventTypesRef.current.length !== eventTypes.length ||
        !previousEventTypesRef.current.every((et, i) => et === eventTypes[i]);
      
      // 필터 비교 최적화: 배열을 정렬하여 비교
      const normalizeFilters = (f?: SubscriptionFilters): string => {
        if (!f) return '';
        const normalized = {
          providers: f.providers ? [...f.providers].sort() : undefined,
          credential_ids: f.credential_ids ? [...f.credential_ids].sort() : undefined,
          regions: f.regions ? [...f.regions].sort() : undefined,
        };
        return JSON.stringify(normalized);
      };
      
      const filtersChanged = normalizeFilters(previousFiltersRef.current) !== normalizeFilters(filters);

      // 기존 구독자가 있고 변경사항이 있으면 업데이트
      if (subscriberIdRef.current && (eventTypesChanged || filtersChanged)) {
        try {
          await sseSubscriptionManager.updateSubscription(
            subscriberIdRef.current,
            eventTypes,
            filters
          );
          
          previousEventTypesRef.current = [...eventTypes];
          previousFiltersRef.current = filters ? { ...filters } : undefined;
          
          if (onSubscribed && !isUnmountingRef.current && !isCancelled) {
            onSubscribed(eventTypes);
          }
        } catch (error) {
          log.error('[SSE Subscription] Failed to update subscription', error, {
            service: 'SSE',
            action: 'updateSubscription',
            eventTypes,
            filters,
            subscriberId: subscriberIdRef.current,
          });
          
          if (onSubscriptionError && !isUnmountingRef.current && !isCancelled) {
            onSubscriptionError(error, eventTypes);
          }
        }
        return;
      }

      // 새로운 구독 생성
      if (!subscriberIdRef.current) {
        try {
          const subscriberId = await sseSubscriptionManager.subscribe(
            eventTypes,
            filters
          );

          if (isUnmountingRef.current || isCancelled) {
            // 구독 후 즉시 취소된 경우 해제
            await sseSubscriptionManager.unsubscribe(subscriberId);
            return;
          }

          subscriberIdRef.current = subscriberId;
          previousEventTypesRef.current = [...eventTypes];
          previousFiltersRef.current = filters ? { ...filters } : undefined;

          log.debug('[SSE Subscription] Subscribed to events via manager', {
            eventTypes,
            filters,
            subscriberId,
          });

          if (onSubscribed && !isUnmountingRef.current && !isCancelled) {
            onSubscribed(eventTypes);
          }
        } catch (error) {
          log.error('[SSE Subscription] Failed to subscribe via manager', error, {
            service: 'SSE',
            action: 'subscribe',
            eventTypes,
            filters,
          });

          if (onSubscriptionError && !isUnmountingRef.current && !isCancelled) {
            onSubscriptionError(error, eventTypes);
          }
        }
      }
    };

    subscribe();

    return () => {
      isCancelled = true;
    };
  }, [eventTypes, filters, enabled, sseStatus.isConnected, onSubscribed, onSubscriptionError]);
}

