/**
 * SSE Status Badge Component
 * 
 * SSE 연결 상태를 표시하는 Badge 컴포넌트
 */

'use client';

import { useMemo } from 'react';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { useSSEStatus } from '@/hooks/use-sse-status';
import { useTranslation } from '@/hooks/use-translation';
import { formatDistanceToNow } from 'date-fns';
import { ko } from 'date-fns/locale';

/**
 * SSE 상태 Badge 컴포넌트
 * 
 * @example
 * ```tsx
 * <SSEStatusBadge />
 * ```
 */
export function SSEStatusBadge() {
  const { status } = useSSEStatus();
  const { t, locale } = useTranslation();

  // 상태에 따른 스타일 결정
  const badgeStyle = useMemo(() => {
    if (status.isConnecting) {
      return {
        variant: 'outline' as const,
        className: 'border-yellow-500 text-yellow-700 dark:text-yellow-400',
        dotClassName: 'bg-yellow-500 animate-pulse',
        label: t('sse.connecting') || 'Connecting...',
      };
    }

    if (status.isConnected) {
      return {
        variant: 'outline' as const,
        className: 'border-green-500 text-green-700 dark:text-green-400',
        dotClassName: 'bg-green-500 animate-pulse',
        label: t('sse.realTime') || 'Real-time',
      };
    }

    return {
      variant: 'outline' as const,
      className: 'border-red-500 text-red-700 dark:text-red-400',
      dotClassName: 'bg-red-500',
      label: t('sse.offline') || 'Offline',
    };
  }, [status.isConnected, status.isConnecting, t]);

  // 마지막 업데이트 시간 포맷팅
  const lastUpdateText = useMemo(() => {
    if (!status.lastUpdateTime) {
      return null;
    }

    try {
      const localeObj = locale === 'ko' ? ko : undefined;
      return formatDistanceToNow(status.lastUpdateTime, {
        addSuffix: true,
        locale: localeObj,
      });
    } catch {
      // date-fns 오류 시 fallback
      const seconds = Math.floor((Date.now() - status.lastUpdateTime.getTime()) / 1000);
      if (seconds < 60) {
        return `${seconds}${locale === 'ko' ? '초 전' : 's ago'}`;
      }
      const minutes = Math.floor(seconds / 60);
      return `${minutes}${locale === 'ko' ? '분 전' : 'm ago'}`;
    }
  }, [status.lastUpdateTime, locale]);

  return (
    <Badge
      variant={badgeStyle.variant}
      className={cn(
        'flex items-center gap-1.5 px-2 py-1 text-xs font-normal',
        badgeStyle.className
      )}
    >
      <div
        className={cn(
          'h-2 w-2 rounded-full',
          badgeStyle.dotClassName
        )}
        aria-hidden="true"
      />
      <span className="hidden sm:inline">{badgeStyle.label}</span>
      {status.isConnected && lastUpdateText && (
        <span className="hidden md:inline text-muted-foreground">
          • {lastUpdateText}
        </span>
      )}
    </Badge>
  );
}


