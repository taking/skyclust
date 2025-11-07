/**
 * useOfflineQueue Hook
 * 오프라인 큐를 관리하는 React 훅
 */

import { useEffect, useCallback, useRef } from 'react';
import { useOffline } from './use-offline';
import { getOfflineQueue, type QueuedRequest } from '@/lib/offline-queue';
import api from '@/lib/api';

/**
 * useOfflineQueue Hook
 * 
 * 오프라인 상태에서 실패한 요청을 큐에 저장하고,
 * 온라인 복구 시 자동으로 재시도합니다.
 * 
 * @example
 * ```tsx
 * useOfflineQueue(); // 자동으로 큐 처리
 * ```
 */
export function useOfflineQueue(): {
  queue: readonly QueuedRequest[];
  queueSize: number;
  clearQueue: () => void;
} {
  const { isOnline } = useOffline();
  const queueManagerRef = useRef(getOfflineQueue());
  const queueManager = queueManagerRef.current;

  // 요청 실행 함수 (안정적인 참조 유지)
  const executeRequest = useCallback(async (request: QueuedRequest): Promise<Response> => {
    const config = {
      method: request.method.toLowerCase() as 'get' | 'post' | 'put' | 'delete' | 'patch',
      url: request.url,
      headers: request.headers,
    } as {
      method: 'get' | 'post' | 'put' | 'delete' | 'patch';
      url: string;
      headers?: Record<string, string>;
      data?: unknown;
    };

    if (request.data && ['POST', 'PUT', 'PATCH'].includes(request.method)) {
      config.data = request.data;
    }

    const response = await api.request(config);
    return response as unknown as Response;
  }, []);

  // 온라인 복구 시 큐 처리
  useEffect(() => {
    if (isOnline && queueManager.getSize() > 0) {
      queueManager.processQueue(executeRequest);
    }
    // queueManager는 싱글톤이므로 dependency에서 제외
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOnline, executeRequest]);

  return {
    queue: queueManager.getQueue(),
    queueSize: queueManager.getSize(),
    clearQueue: () => queueManager.clear(),
  };
}

