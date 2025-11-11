/**
 * SSE Status Dialog Component
 * 
 * SSE 연결 상태 상세 정보를 표시하는 Dialog/Sheet 컴포넌트
 */

'use client';

import { useMemo } from 'react';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useSSEStatus } from '@/hooks/use-sse-status';
import { useTranslation } from '@/hooks/use-translation';
import { formatDistanceToNow, format } from 'date-fns';
import { ko } from 'date-fns/locale';
import { RefreshCw, CheckCircle2, XCircle, Clock, Activity } from 'lucide-react';
import { cn } from '@/lib/utils';

interface SSEStatusDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/**
 * SSE 상태 상세 정보 Dialog
 * 
 * @example
 * ```tsx
 * <SSEStatusDialog open={isOpen} onOpenChange={setIsOpen} />
 * ```
 */
export function SSEStatusDialog({ open, onOpenChange }: SSEStatusDialogProps) {
  const { status, reconnect } = useSSEStatus();
  const { t, locale } = useTranslation();

  const localeObj = locale === 'ko' ? ko : undefined;

  // 연결 시간 포맷팅
  const connectedDuration = useMemo(() => {
    if (!status.connectedAt) {
      return null;
    }

    try {
      return formatDistanceToNow(status.connectedAt, {
        addSuffix: false,
        locale: localeObj,
      });
    } catch {
      const seconds = Math.floor((Date.now() - status.connectedAt.getTime()) / 1000);
      const minutes = Math.floor(seconds / 60);
      const hours = Math.floor(minutes / 60);
      
      if (hours > 0) {
        return `${hours}${locale === 'ko' ? '시간' : 'h'} ${minutes % 60}${locale === 'ko' ? '분' : 'm'}`;
      }
      if (minutes > 0) {
        return `${minutes}${locale === 'ko' ? '분' : 'm'}`;
      }
      return `${seconds}${locale === 'ko' ? '초' : 's'}`;
    }
  }, [status.connectedAt, locale, localeObj]);

  // 마지막 업데이트 시간 포맷팅
  const lastUpdateFormatted = useMemo(() => {
    if (!status.lastUpdateTime) {
      return null;
    }

    try {
      return format(status.lastUpdateTime, locale === 'ko' ? 'yyyy-MM-dd HH:mm:ss' : 'yyyy-MM-dd hh:mm:ss a', {
        locale: localeObj,
      });
    } catch {
      return status.lastUpdateTime.toLocaleString();
    }
  }, [status.lastUpdateTime, locale, localeObj]);

  // 연결 시간 포맷팅
  const connectedAtFormatted = useMemo(() => {
    if (!status.connectedAt) {
      return null;
    }

    try {
      return format(status.connectedAt, locale === 'ko' ? 'yyyy-MM-dd HH:mm:ss' : 'yyyy-MM-dd hh:mm:ss a', {
        locale: localeObj,
      });
    } catch {
      return status.connectedAt.toLocaleString();
    }
  }, [status.connectedAt, locale, localeObj]);

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

  const handleReconnect = async () => {
    await reconnect();
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="sm:max-w-md w-full overflow-y-auto">
        <SheetHeader className="text-left px-6">
          <SheetTitle className="flex items-center gap-2 text-lg">
            <Activity className="h-5 w-5" />
            {t('sse.connectionStatus') || 'SSE Connection Status'}
          </SheetTitle>
          <SheetDescription className="text-left">
            {t('sse.connectionStatusDescription') || 'Real-time update connection status and statistics'}
          </SheetDescription>
        </SheetHeader>

        <div className="mt-6 space-y-6 px-6 pb-6">
          {/* 연결 상태 */}
          <div className="space-y-3">
            <div className="grid grid-cols-[auto_1fr] items-center gap-x-4 gap-y-3">
              <span className="text-sm font-medium text-foreground whitespace-nowrap">
                {t('sse.status') || 'Status'}:
              </span>
              <div className="flex justify-end">
                <Badge
                  variant={status.isConnected ? 'default' : 'destructive'}
                  className={cn(
                    'flex items-center gap-1.5 shrink-0',
                    status.isConnected && 'bg-green-500 hover:bg-green-600 text-white',
                    status.isConnecting && 'bg-yellow-500 hover:bg-yellow-600 text-white'
                  )}
                >
                  {status.isConnected ? (
                    <CheckCircle2 className="h-3 w-3" />
                  ) : status.isConnecting ? (
                    <RefreshCw className="h-3 w-3 animate-spin" />
                  ) : (
                    <XCircle className="h-3 w-3" />
                  )}
                  {status.isConnected
                    ? t('sse.connected') || 'Connected'
                    : status.isConnecting
                    ? t('sse.connecting') || 'Connecting...'
                    : t('sse.disconnected') || 'Disconnected'}
                </Badge>
              </div>

              {status.connectedAt && connectedDuration && (
                <>
                  <span className="text-muted-foreground flex items-center gap-1.5 whitespace-nowrap text-sm">
                    <Clock className="h-4 w-4 shrink-0" />
                    {t('sse.connectedFor') || 'Connected for'}:
                  </span>
                  <span className="font-semibold text-foreground text-right text-sm">{connectedDuration}</span>
                </>
              )}

              {status.connectedAt && connectedAtFormatted && (
                <>
                  <span className="text-muted-foreground whitespace-nowrap text-sm">
                    {t('sse.connectedAt') || 'Connected at'}:
                  </span>
                  <span className="font-semibold text-foreground text-xs text-right break-words">{connectedAtFormatted}</span>
                </>
              )}

              {status.lastUpdateTime && lastUpdateFormatted && (
                <>
                  <span className="text-muted-foreground whitespace-nowrap text-sm">
                    {t('sse.lastUpdate') || 'Last update'}:
                  </span>
                  <span className="font-semibold text-foreground text-xs text-right break-words">{lastUpdateFormatted}</span>
                </>
              )}

              <span className="text-muted-foreground whitespace-nowrap text-sm">
                {t('sse.readyState') || 'Ready State'}:
              </span>
              <span className="font-semibold text-foreground text-right text-sm">{readyStateText}</span>
            </div>
          </div>

          {/* 이벤트 통계 */}
          {status.isConnected && (
            <div className="space-y-3 pt-4 border-t border-border">
              <h3 className="text-sm font-semibold text-foreground">
                {t('sse.eventStatistics') || 'Event Statistics'}
              </h3>
              
              <div className="grid grid-cols-[auto_1fr] items-center gap-x-4 gap-y-2">
                <span className="text-muted-foreground whitespace-nowrap text-sm">
                  {t('sse.eventsLastMinute') || 'Events (last minute)'}:
                </span>
                <span className="font-semibold text-foreground text-right text-sm">{status.eventCountLastMinute}</span>
                
                <span className="text-muted-foreground whitespace-nowrap text-sm">
                  {t('sse.eventsLast5Minutes') || 'Events (last 5 minutes)'}:
                </span>
                <span className="font-semibold text-foreground text-right text-sm">{status.eventCountLast5Minutes}</span>
              </div>
            </div>
          )}

          {/* 구독 중인 이벤트 */}
          {status.subscribedEvents.length > 0 && (
            <div className="space-y-3 pt-4 border-t border-border">
              <h3 className="text-sm font-semibold text-foreground">
                {t('sse.subscribedEvents') || 'Subscribed Events'}
              </h3>
              <div className="flex flex-wrap gap-2">
                {status.subscribedEvents.map((event) => (
                  <Badge 
                    key={event} 
                    variant="secondary" 
                    className="text-xs font-semibold bg-blue-500/20 text-blue-700 dark:bg-blue-400/30 dark:text-blue-300 border border-blue-500/30 dark:border-blue-400/40"
                  >
                    {event}
                  </Badge>
                ))}
              </div>
            </div>
          )}

          {/* 재연결 버튼 */}
          {!status.isConnected && !status.isConnecting && (
            <div className="pt-4 border-t border-border">
              <Button
                onClick={handleReconnect}
                variant="outline"
                className="w-full"
                disabled={status.isConnecting}
              >
                <RefreshCw className={cn('h-4 w-4 mr-2', status.isConnecting && 'animate-spin')} />
                {t('sse.reconnect') || 'Reconnect'}
              </Button>
            </div>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}


