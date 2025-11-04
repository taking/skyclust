/**
 * Offline Queue
 * 오프라인 상태에서 실패한 요청을 큐에 저장하고, 온라인 복구 시 재시도
 * 
 * IndexedDB 또는 localStorage를 사용하여 영구 저장
 */

export interface QueuedRequest {
  id: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  url: string;
  data?: unknown;
  headers?: Record<string, string>;
  timestamp: number;
  retries: number;
  maxRetries?: number;
}

export interface OfflineQueueOptions {
  /**
   * 최대 재시도 횟수
   */
  maxRetries?: number;
  
  /**
   * 재시도 간격 (밀리초)
   */
  retryInterval?: number;
  
  /**
   * 큐 최대 크기
   */
  maxQueueSize?: number;
}

class OfflineQueueManager {
  private queue: QueuedRequest[] = [];
  private isProcessing = false;
  private options: Required<OfflineQueueOptions>;
  private storageKey = 'skyclust-offline-queue';

  constructor(options: OfflineQueueOptions = {}) {
    this.options = {
      maxRetries: options.maxRetries ?? 3,
      retryInterval: options.retryInterval ?? 5000,
      maxQueueSize: options.maxQueueSize ?? 100,
    };

    // localStorage에서 큐 복원
    this.loadQueue();
  }

  /**
   * localStorage에서 큐 로드
   */
  private loadQueue(): void {
    if (typeof window === 'undefined') return;

    try {
      const stored = localStorage.getItem(this.storageKey);
      if (stored) {
        this.queue = JSON.parse(stored);
      }
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.warn('Failed to load offline queue from localStorage:', error);
      }
      this.queue = [];
    }
  }

  /**
   * localStorage에 큐 저장
   */
  private saveQueue(): void {
    if (typeof window === 'undefined') return;

    try {
      localStorage.setItem(this.storageKey, JSON.stringify(this.queue));
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.warn('Failed to save offline queue to localStorage:', error);
      }
    }
  }

  /**
   * 요청을 큐에 추가
   */
  addRequest(request: Omit<QueuedRequest, 'id' | 'timestamp' | 'retries'>): string {
    // 큐 크기 제한
    if (this.queue.length >= this.options.maxQueueSize) {
      // 가장 오래된 요청 제거
      this.queue.shift();
    }

    const queuedRequest: QueuedRequest = {
      ...request,
      id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      timestamp: Date.now(),
      retries: 0,
      maxRetries: request.maxRetries ?? this.options.maxRetries,
    };

    this.queue.push(queuedRequest);
    this.saveQueue();

    return queuedRequest.id;
  }

  /**
   * 큐에서 요청 제거
   */
  removeRequest(id: string): boolean {
    const index = this.queue.findIndex(req => req.id === id);
    if (index === -1) return false;

    this.queue.splice(index, 1);
    this.saveQueue();
    return true;
  }

  /**
   * 큐 전체 비우기
   */
  clear(): void {
    this.queue = [];
    this.saveQueue();
  }

  /**
   * 큐의 모든 요청 반환
   */
  getQueue(): ReadonlyArray<QueuedRequest> {
    return [...this.queue];
  }

  /**
   * 큐 크기 반환
   */
  getSize(): number {
    return this.queue.length;
  }

  /**
   * 큐 처리 (온라인 복구 시)
   */
  async processQueue(
    requestExecutor: (request: QueuedRequest) => Promise<Response>
  ): Promise<void> {
    if (this.isProcessing) return;
    if (this.queue.length === 0) return;

    this.isProcessing = true;

    const failedRequests: QueuedRequest[] = [];

    for (const request of this.queue) {
      try {
        await requestExecutor(request);
        // 성공한 요청은 큐에서 제거
        this.removeRequest(request.id);
      } catch (_error) {
        // 재시도 횟수 증가
        request.retries++;

        if (request.retries >= (request.maxRetries ?? this.options.maxRetries)) {
          // 최대 재시도 횟수 초과 - 큐에서 제거
          if (process.env.NODE_ENV === 'development') {
            console.warn(`Request ${request.id} exceeded max retries, removing from queue`);
          }
          this.removeRequest(request.id);
        } else {
          // 재시도 가능한 요청은 유지
          failedRequests.push(request);
        }
      }

      // 요청 간 딜레이 (서버 부하 방지)
      await new Promise(resolve => setTimeout(resolve, this.options.retryInterval));
    }

    this.saveQueue();
    this.isProcessing = false;

    // 일부 요청이 실패했지만 재시도 가능하면 나중에 다시 시도
    if (failedRequests.length > 0 && navigator.onLine) {
      setTimeout(() => {
        this.processQueue(requestExecutor);
      }, this.options.retryInterval * 2);
    }
  }

  /**
   * 특정 요청 재시도
   */
  async retryRequest(
    id: string,
    requestExecutor: (request: QueuedRequest) => Promise<Response>
  ): Promise<boolean> {
    const request = this.queue.find(req => req.id === id);
    if (!request) return false;

    try {
      await requestExecutor(request);
      this.removeRequest(id);
      return true;
    } catch (_error) {
      request.retries++;
      this.saveQueue();
      return false;
    }
  }
}

// 싱글톤 인스턴스
let queueManagerInstance: OfflineQueueManager | null = null;

/**
 * Offline Queue Manager 인스턴스 가져오기
 */
export function getOfflineQueue(options?: OfflineQueueOptions): OfflineQueueManager {
  if (!queueManagerInstance) {
    queueManagerInstance = new OfflineQueueManager(options);
  }
  return queueManagerInstance;
}

/**
 * 오프라인 큐 리셋 (테스트용)
 */
export function resetOfflineQueue(): void {
  queueManagerInstance = null;
}

