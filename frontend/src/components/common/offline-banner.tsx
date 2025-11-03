/**
 * Offline Banner Component
 * 오프라인 상태를 표시하는 배너 컴포넌트
 * 
 * 네트워크 연결이 끊겼을 때 사용자에게 알리고, 연결 복구 시 자동으로 숨김
 */

'use client';

import { useEffect, useState } from 'react';
import { useOffline } from '@/hooks/use-offline';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Wifi, WifiOff, RefreshCw, X } from 'lucide-react';
import { cn } from '@/lib/utils';

export interface OfflineBannerProps {
  /**
   * 배너 위치
   * - 'top': 상단 고정
   * - 'bottom': 하단 고정
   * - 'inline': 인라인 (부모 컨테이너 내)
   */
  position?: 'top' | 'bottom' | 'inline';
  
  /**
   * 자동 숨김 (연결 복구 시)
   */
  autoHide?: boolean;
  
  /**
   * 숨김 가능 여부
   */
  dismissible?: boolean;
  
  /**
   * 수동 새로고침 버튼 표시 여부
   */
  showRefreshButton?: boolean;
  
  /**
   * 추가 클래스명
   */
  className?: string;
  
  /**
   * 오프라인 시간 표시 여부
   */
  showOfflineDuration?: boolean;
}

/**
 * 시간 포맷팅 (초 -> "X분 Y초" 또는 "X초")
 */
function formatDuration(seconds: number): string {
  if (seconds < 60) {
    return `${seconds}초`;
  }
  
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  
  if (remainingSeconds === 0) {
    return `${minutes}분`;
  }
  
  return `${minutes}분 ${remainingSeconds}초`;
}

/**
 * OfflineBanner Component
 * 
 * 네트워크 연결 상태를 감지하고 오프라인 상태를 표시합니다.
 * 
 * @example
 * ```tsx
 * <OfflineBanner position="top" autoHide />
 * ```
 */
export function OfflineBanner({
  position = 'top',
  autoHide = true,
  dismissible = false,
  showRefreshButton = true,
  className,
  showOfflineDuration = true,
}: OfflineBannerProps) {
  const {
    isOffline,
    isOnline,
    offlineSince,
    offlineDuration,
    checkConnection,
  } = useOffline();
  
  const [isDismissed, setIsDismissed] = useState(false);
  const [isChecking, setIsChecking] = useState(false);

  // 연결 복구 시 자동 숨김
  useEffect(() => {
    if (autoHide && isOnline && isOffline) {
      setIsDismissed(false);
    }
  }, [autoHide, isOnline, isOffline]);

  // 수동 연결 확인
  const handleRefresh = async () => {
    setIsChecking(true);
    try {
      await checkConnection();
    } finally {
      setIsChecking(false);
    }
  };

  // 배너 닫기
  const handleDismiss = () => {
    setIsDismissed(true);
  };

  const shouldShow = isOffline && !isDismissed;

  const positionClasses = {
    top: 'fixed top-0 left-0 right-0 z-50',
    bottom: 'fixed bottom-0 left-0 right-0 z-50',
    inline: '',
  };

  const alertContent = (
    <Alert
      variant="destructive"
      className={cn(
        'border-red-500 bg-red-50 dark:bg-red-950',
        position === 'top' || position === 'bottom' ? 'rounded-none border-x-0' : 'rounded-lg',
        className
      )}
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-3 flex-1">
          <WifiOff className="h-5 w-5 text-red-600 flex-shrink-0" />
          <div className="flex-1">
            <AlertTitle className="text-red-900 dark:text-red-100">
              오프라인 상태
            </AlertTitle>
            <AlertDescription className="text-red-800 dark:text-red-200">
              <div className="space-y-1">
                <p>인터넷 연결이 없습니다. 일부 기능이 제한될 수 있습니다.</p>
                {showOfflineDuration && offlineSince && (
                  <p className="text-sm">
                    오프라인 시간: {formatDuration(offlineDuration)}
                  </p>
                )}
              </div>
            </AlertDescription>
          </div>
        </div>
        
        <div className="flex items-center space-x-2 ml-4">
          {showRefreshButton && (
            <Button
              variant="ghost"
              size="sm"
              onClick={handleRefresh}
              disabled={isChecking}
              className="text-red-900 hover:text-red-950 hover:bg-red-100 dark:text-red-100 dark:hover:bg-red-900"
            >
              <RefreshCw
                className={cn(
                  'h-4 w-4 mr-2',
                  isChecking && 'animate-spin'
                )}
              />
              연결 확인
            </Button>
          )}
          
          {dismissible && (
            <Button
              variant="ghost"
              size="sm"
              onClick={handleDismiss}
              className="text-red-900 hover:text-red-950 hover:bg-red-100 dark:text-red-100 dark:hover:bg-red-900"
              aria-label="배너 닫기"
            >
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>
      </div>
    </Alert>
  );

  if (position === 'top' || position === 'bottom') {
    if (!shouldShow) return null;
    
    return (
      <div
        className={cn(
          positionClasses[position],
          'transition-all duration-300 ease-in-out'
        )}
      >
        {alertContent}
      </div>
    );
  }

  // Inline 위치
  if (!shouldShow) return null;
  
  return alertContent;
}

/**
 * OnlineIndicator Component
 * 온라인 상태를 간단히 표시하는 인디케이터
 */
export function OnlineIndicator({ className }: { className?: string }) {
  const { isOnline, connectionQuality } = useOffline();
  
  if (!isOnline) return null;
  
  return (
    <div className={cn('flex items-center space-x-2 text-sm text-gray-600', className)}>
      <Wifi className={cn(
        'h-4 w-4',
        connectionQuality === 'slow' && 'text-yellow-500',
        connectionQuality === 'fast' && 'text-green-500',
        connectionQuality === 'unknown' && 'text-gray-400'
      )} />
      <span>
        {connectionQuality === 'slow' && '느린 연결'}
        {connectionQuality === 'fast' && '연결됨'}
        {connectionQuality === 'unknown' && '온라인'}
      </span>
    </div>
  );
}

