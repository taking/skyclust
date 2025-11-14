/**
 * SSE Status Badge Component
 * 
 * SSE 연결 상태를 표시하는 Badge 컴포넌트
 */

'use client';

import { useMemo, useState, useEffect, useRef } from 'react';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { useSSEStatus } from '@/hooks/use-sse-status';
import { useTranslation } from '@/hooks/use-translation';
import { formatDistanceToNow } from 'date-fns';
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
      return `${hours}시간 ${minutes}분`;
    }
    return `${hours}h ${minutes}m`;
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
  const [connectedDuration, setConnectedDuration] = useState<string | null>(null);

  // connectedAt 타임스탬프를 안정적으로 저장 (컴포넌트 마운트 시 한 번만 계산)
  const connectedAtTimestampRef = useRef<number | null>(null);

  // connectedAt이 변경될 때만 타임스탬프 업데이트
  useEffect(() => {
    if (status.connectedAt && status.isConnected) {
      const timestamp = status.connectedAt.getTime();
      if (!isNaN(timestamp) && timestamp > 0) {
        connectedAtTimestampRef.current = timestamp;
      }
    } else {
      connectedAtTimestampRef.current = null;
    }
  }, [status.connectedAt, status.isConnected]);

  // 연결 시간을 1초마다 업데이트 (실시간 증가)
  useEffect(() => {
    if (!connectedAtTimestampRef.current || !status.isConnected) {
      setConnectedDuration(null);
      return;
    }

    const connectedAtTime = connectedAtTimestampRef.current;

    const updateDuration = () => {
      const now = Date.now();
      const seconds = Math.floor((now - connectedAtTime) / 1000);
      
      // 음수이면 null 반환 (계산 오류)
      if (seconds < 0) {
        setConnectedDuration(null);
        return;
      }
      
      setConnectedDuration(formatDuration(seconds, locale));
    };

    // 즉시 업데이트
    updateDuration();

    // 1초마다 업데이트
    const interval = setInterval(updateDuration, 1000);

    return () => clearInterval(interval);
  }, [status.isConnected, locale]);

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
      {status.isConnected && connectedDuration && (
        <span className="hidden md:inline text-muted-foreground">
          • {connectedDuration}
        </span>
      )}
    </Badge>
  );
}
