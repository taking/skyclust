/**
 * SSE Subscription Manager
 * 
 * SSE 구독을 중앙에서 관리하는 매니저
 * - 중복 구독 방지
 * - 구독 참조 카운팅
 * - 자동 정리
 * - 전역 구독 상태 관리
 */

import { sseService } from './sse';
import { log } from '@/lib/logging';

export interface SubscriptionFilters {
  providers?: string[];
  credential_ids?: string[];
  regions?: string[];
}

interface SubscriptionEntry {
  eventType: string;
  filters?: SubscriptionFilters;
  refCount: number;
  subscribers: Set<string>; // subscriber ID set
  timestamp: Date;
}

interface SubscriberInfo {
  id: string;
  eventTypes: Set<string>;
  filters?: SubscriptionFilters;
}

/**
 * SSE 구독 관리자
 * 
 * 전역적으로 SSE 구독을 관리하여:
 * - 중복 구독 방지
 * - 구독 참조 카운팅
 * - 자동 정리
 */
class SSESubscriptionManager {
  private subscriptions = new Map<string, SubscriptionEntry>();
  private subscribers = new Map<string, SubscriberInfo>();
  private subscriptionKeyCache = new Map<string, string>(); // subscriberId + eventType -> subscriptionKey

  /**
   * 구독 키 생성
   * eventType + filters의 해시를 기반으로 고유 키 생성
   */
  private getSubscriptionKey(eventType: string, filters?: SubscriptionFilters): string {
    if (!filters || Object.keys(filters).length === 0) {
      return eventType;
    }

    // 필터를 정렬하여 일관된 키 생성
    const sortedFilters = JSON.stringify({
      providers: filters.providers?.sort(),
      credential_ids: filters.credential_ids?.sort(),
      regions: filters.regions?.sort(),
    });

    return `${eventType}:${sortedFilters}`;
  }

  /**
   * 구독자 ID 생성 (컴포넌트 인스턴스별 고유 ID)
   */
  private generateSubscriberId(): string {
    return `subscriber_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
  }

  /**
   * 구독 추가
   * 
   * @param eventTypes - 구독할 이벤트 타입 배열
   * @param filters - 구독 필터
   * @param subscriberId - 구독자 ID (옵션, 없으면 자동 생성)
   * @returns 구독자 ID
   */
  async subscribe(
    eventTypes: string[],
    filters?: SubscriptionFilters,
    subscriberId?: string
  ): Promise<string> {
    const id = subscriberId || this.generateSubscriberId();
    
    // 구독자 정보 저장
    const subscriberInfo: SubscriberInfo = {
      id,
      eventTypes: new Set(eventTypes),
      filters,
    };
    this.subscribers.set(id, subscriberInfo);

    // 각 이벤트 타입에 대해 구독 처리
    const subscriptionResults: Array<{ eventType: string; success: boolean; error?: unknown }> = [];

    for (const eventType of eventTypes) {
      const subscriptionKey = this.getSubscriptionKey(eventType, filters);
      const cacheKey = `${id}:${eventType}`;
      
      // 캐시에 저장
      this.subscriptionKeyCache.set(cacheKey, subscriptionKey);

      try {
        // 기존 구독이 있는지 확인
        const existingSubscription = this.subscriptions.get(subscriptionKey);

        if (existingSubscription) {
          // 기존 구독이 있으면 참조 카운트만 증가
          existingSubscription.refCount++;
          existingSubscription.subscribers.add(id);
          
          log.debug('[SSE Subscription Manager] Incremented ref count for existing subscription', {
            subscriptionKey,
            eventType,
            refCount: existingSubscription.refCount,
            subscriberId: id,
          });
          
          subscriptionResults.push({ eventType, success: true });
        } else {
          // 새로운 구독 생성
          await sseService.subscribeToEvent(eventType, filters);

          const newSubscription: SubscriptionEntry = {
            eventType,
            filters,
            refCount: 1,
            subscribers: new Set([id]),
            timestamp: new Date(),
          };

          this.subscriptions.set(subscriptionKey, newSubscription);

          log.debug('[SSE Subscription Manager] Created new subscription', {
            subscriptionKey,
            eventType,
            refCount: 1,
            subscriberId: id,
          });

          subscriptionResults.push({ eventType, success: true });
        }
      } catch (error) {
        log.error('[SSE Subscription Manager] Failed to subscribe', error, {
          service: 'SSE',
          action: 'subscribe',
          eventType,
          filters,
          subscriberId: id,
        });

        subscriptionResults.push({ eventType, success: false, error });
      }
    }

    // 실패한 구독이 있으면 로깅
    const failedSubscriptions = subscriptionResults.filter(r => !r.success);
    if (failedSubscriptions.length > 0) {
      log.warn('[SSE Subscription Manager] Some subscriptions failed', {
        failedCount: failedSubscriptions.length,
        totalCount: eventTypes.length,
        failedEvents: failedSubscriptions.map(r => r.eventType),
        subscriberId: id,
      });
    }

    return id;
  }

  /**
   * 구독 해제
   * 
   * @param subscriberId - 구독자 ID
   */
  async unsubscribe(subscriberId: string): Promise<void> {
    const subscriberInfo = this.subscribers.get(subscriberId);
    
    if (!subscriberInfo) {
      log.debug('[SSE Subscription Manager] Subscriber not found, skipping unsubscribe', {
        subscriberId,
      });
      return;
    }

    // 구독자 정보에서 이벤트 타입 가져오기
    const eventTypes = Array.from(subscriberInfo.eventTypes);
    const filters = subscriberInfo.filters;

    // 각 이벤트 타입에 대해 구독 해제 처리
    for (const eventType of eventTypes) {
      const cacheKey = `${subscriberId}:${eventType}`;
      const subscriptionKey = this.subscriptionKeyCache.get(cacheKey);

      if (!subscriptionKey) {
        log.warn('[SSE Subscription Manager] Subscription key not found in cache', {
          subscriberId,
          eventType,
          cacheKey,
        });
        continue;
      }

      const subscription = this.subscriptions.get(subscriptionKey);

      if (!subscription) {
        log.warn('[SSE Subscription Manager] Subscription not found', {
          subscriptionKey,
          eventType,
          subscriberId,
        });
        continue;
      }

      // 참조 카운트 감소
      subscription.refCount--;
      subscription.subscribers.delete(subscriberId);

      // 참조 카운트가 0이 되면 실제 구독 해제
      if (subscription.refCount <= 0) {
        try {
          await sseService.unsubscribeFromEvent(eventType, filters);
          this.subscriptions.delete(subscriptionKey);

          log.debug('[SSE Subscription Manager] Unsubscribed and removed subscription', {
            subscriptionKey,
            eventType,
            subscriberId,
          });
        } catch (error) {
          log.error('[SSE Subscription Manager] Failed to unsubscribe', error, {
            service: 'SSE',
            action: 'unsubscribe',
            eventType,
            filters,
            subscriptionKey,
            subscriberId,
          });

          // 에러가 발생해도 구독 정보는 제거 (정리)
          this.subscriptions.delete(subscriptionKey);
        }
      } else {
        log.debug('[SSE Subscription Manager] Decremented ref count, keeping subscription', {
          subscriptionKey,
          eventType,
          refCount: subscription.refCount,
          subscriberId,
        });
      }

      // 캐시에서 제거
      this.subscriptionKeyCache.delete(cacheKey);
    }

    // 구독자 정보 제거
    this.subscribers.delete(subscriberId);

    log.debug('[SSE Subscription Manager] Unsubscribed all events for subscriber', {
      subscriberId,
      eventTypes,
    });
  }

  /**
   * 구독 업데이트
   * 
   * 기존 구독을 해제하고 새로운 구독을 추가합니다.
   * 
   * @param subscriberId - 구독자 ID
   * @param eventTypes - 새로운 이벤트 타입 배열
   * @param filters - 새로운 구독 필터
   */
  async updateSubscription(
    subscriberId: string,
    eventTypes: string[],
    filters?: SubscriptionFilters
  ): Promise<void> {
    // 기존 구독 해제
    await this.unsubscribe(subscriberId);

    // 새로운 구독 추가
    await this.subscribe(eventTypes, filters, subscriberId);
  }

  /**
   * 모든 구독 정리
   * 
   * 모든 구독을 해제하고 상태를 초기화합니다.
   */
  async clearAll(): Promise<void> {
    const subscriberIds = Array.from(this.subscribers.keys());

    for (const subscriberId of subscriberIds) {
      await this.unsubscribe(subscriberId);
    }

    this.subscriptions.clear();
    this.subscribers.clear();
    this.subscriptionKeyCache.clear();

    log.info('[SSE Subscription Manager] Cleared all subscriptions');
  }

  /**
   * 현재 구독 상태 조회
   */
  getSubscriptionStatus(): {
    totalSubscriptions: number;
    totalSubscribers: number;
    subscriptions: Array<{
      eventType: string;
      filters?: SubscriptionFilters;
      refCount: number;
      subscriberCount: number;
      timestamp: Date;
    }>;
    subscribers: Array<{
      id: string;
      eventTypes: string[];
      filters?: SubscriptionFilters;
    }>;
  } {
    return {
      totalSubscriptions: this.subscriptions.size,
      totalSubscribers: this.subscribers.size,
      subscriptions: Array.from(this.subscriptions.entries()).map(([key, entry]) => ({
        eventType: entry.eventType,
        filters: entry.filters,
        refCount: entry.refCount,
        subscriberCount: entry.subscribers.size,
        timestamp: entry.timestamp,
      })),
      subscribers: Array.from(this.subscribers.entries()).map(([id, info]) => ({
        id,
        eventTypes: Array.from(info.eventTypes),
        filters: info.filters,
      })),
    };
  }

  /**
   * 특정 구독자 정보 조회
   */
  getSubscriberInfo(subscriberId: string): SubscriberInfo | undefined {
    return this.subscribers.get(subscriberId);
  }

  /**
   * 특정 구독 정보 조회
   */
  getSubscriptionInfo(eventType: string, filters?: SubscriptionFilters): SubscriptionEntry | undefined {
    const subscriptionKey = this.getSubscriptionKey(eventType, filters);
    return this.subscriptions.get(subscriptionKey);
  }
}

// 싱글톤 인스턴스
export const sseSubscriptionManager = new SSESubscriptionManager();

