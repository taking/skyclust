/**
 * SSE Status Tooltip Component
 * 
 * SSE 상태 상세 정보를 표시하는 Tooltip 컴포넌트
 */

'use client';

import { useMemo, useState, useEffect } from 'react';
import { useSSEStatus } from '@/hooks/use-sse-status';
import { useTranslation } from '@/hooks/use-translation';
import { format } from 'date-fns';
import { ko } from 'date-fns/locale';

/**
 * 연결 시간을 시분초 형식으로 포맷팅하는 함수
 */
function formatDuration(seconds: number, locale: string): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  if (hours > 0) {
    if (locale === 'ko') {
      return `${hours}시간 ${minutes}분 ${secs}초`;
    }
    return `${hours}h ${minutes}m ${secs}s`;
  }
  if (minutes > 0) {
    if (locale === 'ko') {
      return `${minutes}분 ${secs}초`;
    }
    return `${minutes}m ${secs}s`;
  }
  if (locale === 'ko') {
    return `${secs}초`;
  }
  return `${secs}s`;
}

/**
 * SSE 상태 Tooltip 컴포넌트
 * 
 * @example
 * ```tsx
 * <Tooltip>
 *   <TooltipTrigger>
 *     <SSEStatusBadge />
 *   </TooltipTrigger>
 *   <TooltipContent>
 *     <SSEStatusTooltip />
 *   </TooltipContent>
 * </Tooltip>
 * ```
 */
export function SSEStatusTooltip() {
  const { status } = useSSEStatus();
  const { t, locale } = useTranslation();
  const [connectedDuration, setConnectedDuration] = useState<string | null>(null);

  const localeObj = locale === 'ko' ? ko : undefined;

  // 연결 시간을 1초마다 업데이트 (실시간 증가)
  useEffect(() => {
    if (!status.connectedAt || !status.isConnected) {
      setConnectedDuration(null);
      return;
    }

    const updateDuration = () => {
      const now = Date.now();
      const connectedAt = status.connectedAt!.getTime();
      const seconds = Math.floor((now - connectedAt) / 1000);
      setConnectedDuration(formatDuration(seconds, locale));
    };

    // 즉시 업데이트
    updateDuration();

    // 1초마다 업데이트
    const interval = setInterval(updateDuration, 1000);

    return () => clearInterval(interval);
  }, [status.connectedAt, status.isConnected, locale]);

  // 마지막 업데이트 시간 포맷팅
  const lastUpdateFormatted = useMemo(() => {
    if (!status.lastUpdateTime) {
      return null;
    }

    try {
      return format(status.lastUpdateTime, locale === 'ko' ? 'HH:mm:ss' : 'hh:mm:ss a', {
        locale: localeObj,
      });
    } catch {
      return status.lastUpdateTime.toLocaleTimeString();
    }
  }, [status.lastUpdateTime, locale, localeObj]);

  // ReadyState 텍스트
  const readyStateText = useMemo(() => {
    if (status.readyState === null) {
      return t('sse.unknown') || 'Unknown';
    }

    switch (status.readyState) {
      case 0: // CONNECTING
        return t('sse.connecting') || 'Connecting';
      case 1: // OPEN
        return t('sse.connected') || 'Connected';
      case 2: // CLOSED
        return t('sse.closed') || 'Closed';
      default:
        return t('sse.unknown') || 'Unknown';
    }
  }, [status.readyState, t]);

  return (
    <div className="space-y-3 p-1">
      <div className="font-semibold text-base text-background">
        {t('sse.connectionStatus') || 'SSE Connection Status'}
      </div>
      
      <div className="space-y-2 text-sm">
        <div className="flex items-center justify-between gap-6 min-w-[200px]">
          <span className="text-background/70 whitespace-nowrap">
            {t('sse.status') || 'Status'}:
          </span>
          <span className="font-semibold text-background text-right">
            {status.isConnected
              ? t('sse.connected') || 'Connected'
              : status.isConnecting
              ? t('sse.connecting') || 'Connecting...'
              : t('sse.disconnected') || 'Disconnected'}
          </span>
        </div>

        {status.connectedAt && connectedDuration && (
          <div className="flex items-center justify-between gap-6 min-w-[200px]">
            <span className="text-background/70 whitespace-nowrap">
              {t('sse.connectedFor') || 'Connected'}:
            </span>
            <span className="font-semibold text-background text-right">{connectedDuration}</span>
          </div>
        )}

        {status.lastUpdateTime && lastUpdateFormatted && (
          <div className="flex items-center justify-between gap-6 min-w-[200px]">
            <span className="text-background/70 whitespace-nowrap">
              {t('sse.lastUpdate') || 'Last Update'}:
            </span>
            <span className="font-semibold text-background text-right">{lastUpdateFormatted}</span>
          </div>
        )}

        {status.isConnected && status.eventCountLastMinute > 0 && (
          <div className="flex items-center justify-between gap-6 min-w-[200px]">
            <span className="text-background/70 whitespace-nowrap">
              {t('sse.eventsPerMinute') || 'Events/min'}:
            </span>
            <span className="font-semibold text-background text-right">{status.eventCountLastMinute}</span>
          </div>
        )}

        <div className="flex items-center justify-between gap-6 min-w-[200px]">
          <span className="text-background/70 whitespace-nowrap">
            {t('sse.readyState') || 'Ready State'}:
          </span>
          <span className="font-semibold text-background text-right">{readyStateText}</span>
        </div>

        {status.subscribedEvents.length > 0 && (
          <div className="pt-2 mt-2 border-t border-background/20">
            <div className="text-background/70 mb-2 text-xs font-medium">
              {t('sse.subscribedEvents') || 'Subscribed Events'}:
            </div>
            <div className="flex flex-wrap gap-1.5">
              {status.subscribedEvents.slice(0, 3).map((event) => (
                <span
                  key={event}
                  className="text-xs px-2 py-1 bg-background/20 text-background font-semibold rounded-md border border-background/30"
                >
                  {event}
                </span>
              ))}
              {status.subscribedEvents.length > 3 && (
                <span className="text-xs text-background/70 font-medium px-1">
                  +{status.subscribedEvents.length - 3}
                </span>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}


