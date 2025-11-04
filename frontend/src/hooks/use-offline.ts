/**
 * useOffline Hook
 * 네트워크 연결 상태를 감지하고 관리하는 훅
 * 
 * 오프라인 상태 감지, 온라인 복구 감지, 연결 품질 추적 등 제공
 */

import { useState, useEffect, useCallback, useRef } from 'react';

export interface UseOfflineReturn {
  /**
   * 현재 오프라인 상태
   */
  isOffline: boolean;
  
  /**
   * 온라인 상태 (isOffline의 반대)
   */
  isOnline: boolean;
  
  /**
   * 오프라인으로 전환된 시점
   */
  offlineSince: Date | null;
  
  /**
   * 마지막으로 온라인이었던 시점
   */
  lastOnline: Date | null;
  
  /**
   * 오프라인 상태였던 총 시간 (초)
   */
  offlineDuration: number;
  
  /**
   * 네트워크 연결 품질 (slow, fast, unknown)
   */
  connectionQuality: 'slow' | 'fast' | 'unknown';
  
  /**
   * 네트워크 상태가 변경되었는지 여부
   */
  hasChanged: boolean;
  
  /**
   * 네트워크 상태 변경 이벤트 리스너 등록
   */
  onStatusChange: (callback: (isOnline: boolean) => void) => () => void;
  
  /**
   * 수동으로 온라인 상태 확인 (health check)
   */
  checkConnection: () => Promise<boolean>;
}

/**
 * useOffline Hook
 * 
 * 네트워크 연결 상태를 감지하고 관리합니다.
 * 
 * @example
 * ```tsx
 * const { isOffline, isOnline, checkConnection } = useOffline();
 * 
 * if (isOffline) {
 *   return <OfflineBanner />;
 * }
 * ```
 */
export function useOffline(): UseOfflineReturn {
  const [isOffline, setIsOffline] = useState<boolean>(() => {
    if (typeof window === 'undefined') return false;
    return !navigator.onLine;
  });

  const [offlineSince, setOfflineSince] = useState<Date | null>(null);
  const [lastOnline, setLastOnline] = useState<Date | null>(() => {
    if (typeof window === 'undefined') return null;
    return navigator.onLine ? new Date() : null;
  });
  
  const [connectionQuality, setConnectionQuality] = useState<'slow' | 'fast' | 'unknown'>('unknown');
  const [hasChanged, setHasChanged] = useState(false);
  
  const offlineDurationRef = useRef<number>(0);
  const statusChangeCallbacksRef = useRef<Set<(isOnline: boolean) => void>>(new Set());
  const healthCheckAbortControllerRef = useRef<AbortController | null>(null);

  // 네트워크 상태 확인 (health check)
  const checkConnection = useCallback(async (): Promise<boolean> => {
    // 이전 health check 취소
    if (healthCheckAbortControllerRef.current) {
      healthCheckAbortControllerRef.current.abort();
    }

    const abortController = new AbortController();
    healthCheckAbortControllerRef.current = abortController;

    // 간단한 연결 확인
    // 실제 서버 health check를 시도하되, 실패하면 navigator.onLine 사용
    const apiBaseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
    
    let timeoutId: NodeJS.Timeout | null = null;
    try {
      const controller = new AbortController();
      timeoutId = setTimeout(() => controller.abort(), 3000); // 3초 타임아웃
      
      const response = await fetch(`${apiBaseUrl}/api/v1/health`, {
        method: 'HEAD',
        signal: controller.signal || abortController.signal,
        cache: 'no-cache',
        headers: {
          'Cache-Control': 'no-cache',
        },
      });
      
      if (timeoutId) clearTimeout(timeoutId);
      
      if (abortController.signal.aborted || controller.signal.aborted) {
        return false;
      }
      
      return response.ok;
    } catch {
      // health check 실패 시 navigator.onLine 사용
      // 서버가 다운되었거나 엔드포인트가 없을 수 있음
      if (timeoutId) clearTimeout(timeoutId);
      return navigator.onLine;
    }
  }, []);

  // 연결 품질 확인 (Network Information API 사용)
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const updateConnectionQuality = () => {
      // @ts-expect-error - Network Information API (실험적 기능)
      const connection = navigator.connection || navigator.mozConnection || navigator.webkitConnection;
      
      if (connection) {
        const effectiveType = connection.effectiveType;
        const downlink = connection.downlink;
        
        if (effectiveType === 'slow-2g' || effectiveType === '2g' || downlink < 1) {
          setConnectionQuality('slow');
        } else if (effectiveType === '3g' || effectiveType === '4g') {
          setConnectionQuality('fast');
        } else {
          setConnectionQuality('unknown');
        }
      } else {
        setConnectionQuality('unknown');
      }
    };

    updateConnectionQuality();

    // @ts-expect-error - Network Information API (실험적 기능)
    const connection = navigator.connection || navigator.mozConnection || navigator.webkitConnection;
    if (connection) {
      connection.addEventListener('change', updateConnectionQuality);
      return () => {
        connection.removeEventListener('change', updateConnectionQuality);
      };
    }
  }, []);

  // 네트워크 상태 변경 감지
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const handleOnline = async () => {
      setHasChanged(true);
      
      // 온라인 복구 시 연결 확인 (health check)
      const isHealthy = await checkConnection();
      
      if (isHealthy) {
        setIsOffline(false);
        setOfflineSince(null);
        setLastOnline(new Date());
        offlineDurationRef.current = 0;
        
        // 상태 변경 콜백 호출
        statusChangeCallbacksRef.current.forEach(callback => {
          callback(true);
        });
      } else {
        // health check 실패해도 일단 온라인으로 표시
        // (서버 문제일 수도 있음)
        setIsOffline(false);
        setOfflineSince(null);
        setLastOnline(new Date());
      }
    };

    const handleOffline = () => {
      setHasChanged(true);
      setIsOffline(true);
      setOfflineSince(new Date());
      
      // 상태 변경 콜백 호출
      statusChangeCallbacksRef.current.forEach(callback => {
        callback(false);
      });
    };

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    // 초기 상태 설정
    if (!navigator.onLine) {
      setIsOffline(true);
      setOfflineSince(new Date());
    } else {
      setLastOnline(new Date());
    }

    // 주기적으로 연결 상태 확인 (polling)
    const interval = setInterval(async () => {
      if (navigator.onLine && isOffline) {
        // 오프라인으로 표시되어 있지만 navigator.onLine이 true면 다시 확인
        const isHealthy = await checkConnection();
        if (isHealthy) {
          handleOnline();
        }
      }
    }, 5000); // 5초마다 확인

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
      clearInterval(interval);
      
      if (healthCheckAbortControllerRef.current) {
        healthCheckAbortControllerRef.current.abort();
      }
    };
  }, [isOffline, checkConnection]);

  // 오프라인 시간 계산
  useEffect(() => {
    if (!offlineSince) {
      offlineDurationRef.current = 0;
      return;
    }

    const interval = setInterval(() => {
      const now = new Date();
      offlineDurationRef.current = Math.floor((now.getTime() - offlineSince.getTime()) / 1000);
    }, 1000);

    return () => clearInterval(interval);
  }, [offlineSince]);

  // 상태 변경 콜백 등록 함수
  const onStatusChange = useCallback((callback: (isOnline: boolean) => void) => {
    statusChangeCallbacksRef.current.add(callback);
    
    return () => {
      statusChangeCallbacksRef.current.delete(callback);
    };
  }, []);

  // hasChanged 초기화 (한 번 사용 후 리셋)
  useEffect(() => {
    if (hasChanged) {
      const timer = setTimeout(() => {
        setHasChanged(false);
      }, 1000);
      return () => clearTimeout(timer);
    }
  }, [hasChanged]);

  return {
    isOffline,
    isOnline: !isOffline,
    offlineSince,
    lastOnline,
    offlineDuration: offlineDurationRef.current,
    connectionQuality,
    hasChanged,
    onStatusChange,
    checkConnection,
  };
}

/**
 * useOnline Hook
 * 온라인 상태만 반환하는 간단한 훅
 */
export function useOnline(): boolean {
  const { isOnline } = useOffline();
  return isOnline;
}

