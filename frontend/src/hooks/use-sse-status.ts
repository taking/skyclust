/**
 * SSE Status Hook
 * 
 * SSE 연결 상태, 마지막 업데이트 시간, 이벤트 통계를 추적하는 훅
 */

import { useState, useEffect, useCallback, useRef } from 'react';
import { sseService } from '@/services/sse';
import type { SSECallbacks } from '@/lib/types/sse';

export interface SSEStatus {
  isConnected: boolean;
  isConnecting: boolean;
  lastUpdateTime: Date | null;
  connectedAt: Date | null;
  subscribedEvents: string[];
  eventCountLastMinute: number;
  eventCountLast5Minutes: number;
  readyState: number | null;
}

interface EventTimestamp {
  timestamp: Date;
}

/**
 * SSE 상태를 추적하는 훅
 * 
 * @example
 * ```tsx
 * const { status, reconnect } = useSSEStatus();
 * 
 * return (
 *   <div>
 *     {status.isConnected ? 'Connected' : 'Disconnected'}
 *     {status.lastUpdateTime && (
 *       <span>Last update: {formatTime(status.lastUpdateTime)}</span>
 *     )}
 *   </div>
 * );
 * ```
 */
export function useSSEStatus() {
  const [status, setStatus] = useState<SSEStatus>({
    isConnected: false,
    isConnecting: false,
    lastUpdateTime: null,
    connectedAt: null,
    subscribedEvents: [],
    eventCountLastMinute: 0,
    eventCountLast5Minutes: 0,
    readyState: null,
  });

  // 이벤트 타임스탬프 추적 (최근 5분간)
  const eventTimestampsRef = useRef<EventTimestamp[]>([]);
  const connectedAtRef = useRef<Date | null>(null);

  // 이벤트 수신 추적
  const trackEvent = useCallback(() => {
    const now = new Date();
    eventTimestampsRef.current.push({ timestamp: now });
    
    // 5분 이전 이벤트 제거
    const fiveMinutesAgo = new Date(now.getTime() - 5 * 60 * 1000);
    eventTimestampsRef.current = eventTimestampsRef.current.filter(
      (event) => event.timestamp > fiveMinutesAgo
    );

    // 통계 계산
    const oneMinuteAgo = new Date(now.getTime() - 60 * 1000);
    const eventCountLastMinute = eventTimestampsRef.current.filter(
      (event) => event.timestamp > oneMinuteAgo
    ).length;
    const eventCountLast5Minutes = eventTimestampsRef.current.length;

    setStatus((prev) => ({
      ...prev,
      lastUpdateTime: now,
      eventCountLastMinute,
      eventCountLast5Minutes,
    }));
  }, []);

  // SSE 이벤트 리스너 설정
  useEffect(() => {
    // 모든 SSE 이벤트를 추적하기 위한 콜백
    const callbacks: SSECallbacks = {
      onConnected: (data) => {
        const now = new Date();
        connectedAtRef.current = now;
        setStatus((prev) => ({
          ...prev,
          isConnected: true,
          isConnecting: false,
          connectedAt: now,
          lastUpdateTime: now,
        }));
        trackEvent();
      },
      onVMStatusUpdate: () => trackEvent(),
      onVMResourceUpdate: () => trackEvent(),
      onProviderStatusUpdate: () => trackEvent(),
      onProviderInstanceUpdate: () => trackEvent(),
      onSystemNotification: () => trackEvent(),
      onSystemAlert: () => trackEvent(),
      onKubernetesClusterCreated: () => trackEvent(),
      onKubernetesClusterUpdated: () => trackEvent(),
      onKubernetesClusterDeleted: () => trackEvent(),
      onKubernetesClusterList: () => trackEvent(),
      onKubernetesNodePoolCreated: () => trackEvent(),
      onKubernetesNodePoolUpdated: () => trackEvent(),
      onKubernetesNodePoolDeleted: () => trackEvent(),
      onKubernetesNodeCreated: () => trackEvent(),
      onKubernetesNodeUpdated: () => trackEvent(),
      onKubernetesNodeDeleted: () => trackEvent(),
      onNetworkVPCCreated: () => trackEvent(),
      onNetworkVPCUpdated: () => trackEvent(),
      onNetworkVPCDeleted: () => trackEvent(),
      onNetworkVPCList: () => trackEvent(),
      onNetworkSubnetCreated: () => trackEvent(),
      onNetworkSubnetUpdated: () => trackEvent(),
      onNetworkSubnetDeleted: () => trackEvent(),
      onNetworkSubnetList: () => trackEvent(),
      onNetworkSecurityGroupCreated: () => trackEvent(),
      onNetworkSecurityGroupUpdated: () => trackEvent(),
      onNetworkSecurityGroupDeleted: () => trackEvent(),
      onNetworkSecurityGroupList: () => trackEvent(),
      onVMCreated: () => trackEvent(),
      onVMUpdated: () => trackEvent(),
      onVMDeleted: () => trackEvent(),
      onVMList: () => trackEvent(),
      onAzureResourceGroupCreated: () => trackEvent(),
      onAzureResourceGroupUpdated: () => trackEvent(),
      onAzureResourceGroupDeleted: () => trackEvent(),
      onAzureResourceGroupList: () => trackEvent(),
      onDashboardSummaryUpdated: () => trackEvent(),
      onError: () => {
        setStatus((prev) => ({
          ...prev,
          isConnected: false,
          isConnecting: false,
        }));
      },
    };

    // 기존 콜백에 추가 (기존 콜백을 덮어쓰지 않음)
    sseService.updateCallbacks(callbacks);

    return () => {
      // 정리 작업은 하지 않음 (다른 곳에서도 콜백을 사용할 수 있음)
    };
  }, [trackEvent]);

  // 상태 주기적 업데이트 (1초마다)
  useEffect(() => {
    const interval = setInterval(() => {
      const isConnected = sseService.isConnected();
      const readyState = sseService.getReadyState();
      const subscribedEvents = Array.from(sseService.getSubscribedEvents());

      setStatus((prev) => {
        // 연결 상태가 변경된 경우
        if (prev.isConnected !== isConnected) {
          if (isConnected && !connectedAtRef.current) {
            // 연결되었지만 connectedAt이 없으면 현재 시간으로 설정
            connectedAtRef.current = new Date();
          } else if (!isConnected) {
            // 연결이 끊긴 경우 초기화
            connectedAtRef.current = null;
            eventTimestampsRef.current = [];
          }
        }

        // 연결되었지만 connectedAt이 없으면 현재 시간으로 설정 (fallback)
        if (isConnected && !connectedAtRef.current) {
          connectedAtRef.current = new Date();
        }

        return {
          ...prev,
          isConnected,
          readyState,
          subscribedEvents,
          connectedAt: connectedAtRef.current,
          // 연결이 끊긴 경우 통계 초기화
          ...(isConnected ? {} : {
            eventCountLastMinute: 0,
            eventCountLast5Minutes: 0,
            lastUpdateTime: null,
          }),
        };
      });

      // 오래된 이벤트 타임스탬프 정리
      const now = new Date();
      const fiveMinutesAgo = new Date(now.getTime() - 5 * 60 * 1000);
      eventTimestampsRef.current = eventTimestampsRef.current.filter(
        (event) => event.timestamp > fiveMinutesAgo
      );

      // 통계 재계산
      if (isConnected) {
        const oneMinuteAgo = new Date(now.getTime() - 60 * 1000);
        const eventCountLastMinute = eventTimestampsRef.current.filter(
          (event) => event.timestamp > oneMinuteAgo
        ).length;
        const eventCountLast5Minutes = eventTimestampsRef.current.length;

        setStatus((prev) => ({
          ...prev,
          eventCountLastMinute,
          eventCountLast5Minutes,
        }));
      }
    }, 1000); // 1초마다 업데이트

    return () => clearInterval(interval);
  }, []);

  // 수동 재연결 함수
  const reconnect = useCallback(async () => {
    if (sseService.isConnected()) {
      return;
    }

    setStatus((prev) => ({ ...prev, isConnecting: true }));

    try {
      // 토큰 가져오기
      const authStorage = localStorage.getItem('auth_storage');
      let token: string | null = null;

      if (authStorage) {
        try {
          const parsed = JSON.parse(authStorage);
          token = parsed?.state?.token || null;
        } catch {
          // Fallback
          token = localStorage.getItem('token');
        }
      } else {
        token = localStorage.getItem('token');
      }

      if (token) {
        sseService.connect(token);
      }
    } catch (error) {
      console.error('Failed to reconnect SSE', error);
      setStatus((prev) => ({ ...prev, isConnecting: false }));
    }
  }, []);

  return {
    status,
    reconnect,
  };
}


