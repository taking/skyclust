/**
 * SSE Status Tooltip Component
 * 
 * SSE 상태 상세 정보를 표시하는 Tooltip 컴포넌트
 */

'use client';

import { useMemo } from 'react';
import { useSSEStatus } from '@/hooks/use-sse-status';
import { useTranslation } from '@/hooks/use-translation';
import { format } from 'date-fns';
import { ko } from 'date-fns/locale';

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

  const localeObj = locale === 'ko' ? ko : undefined;

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


