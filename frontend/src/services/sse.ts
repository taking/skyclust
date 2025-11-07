/**
 * Server-Sent Events (SSE) Service
 * NATS 기반 실시간 데이터 수신
 */

import type { SSECallbacks, SSEErrorInfo } from '@/lib/types/sse';
import { API_CONFIG } from '@/lib/api-config';
import { API_ENDPOINTS } from '@/lib/api-endpoints';
import { CONNECTION, STORAGE_KEYS } from '@/lib/constants';
import { getErrorLogger } from '@/lib/error-logger';
import { parseSSEMessage } from '@/lib/sse-compression';
import { logger } from '@/lib/logger';

class SSEService {
  private eventSource: EventSource | null = null;
  private callbacks: SSECallbacks = {};
  private reconnectAttempts = 0;
  private maxReconnectAttempts = CONNECTION.SSE.MAX_RECONNECT_ATTEMPTS;
  private baseReconnectDelay = CONNECTION.SSE.BASE_RECONNECT_DELAY;
  private maxReconnectDelay = CONNECTION.SSE.MAX_RECONNECT_DELAY;
  private isConnecting = false;
  private clientId: string | null = null;
  private subscribedEvents = new Set<string>();
  private subscribedVMs = new Set<string>();

  connect(token: string, callbacks: SSECallbacks = {}): void {
    if (this.eventSource?.readyState === EventSource.OPEN) {
      return;
    }

    if (this.isConnecting) {
      return;
    }

    this.isConnecting = true;
    this.callbacks = callbacks;

    const endpoint = `${API_ENDPOINTS.sse.connect()}?token=${token}`;
    const url = `${API_CONFIG.BASE_URL}${API_CONFIG.API_PREFIX}/${API_CONFIG.VERSION}${endpoint}`;
    this.eventSource = new EventSource(url);

    this.setupEventListeners();
  }

  disconnect(): void {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
    this.isConnecting = false;
    this.reconnectAttempts = 0;
  }

  /**
   * SSE 메시지를 파싱합니다.
   * Backend에서 전송하는 메시지 구조:
   * - event: <eventType>
   * - compressed: true (압축된 경우, 선택적)
   * - data: <JSON 문자열 또는 Base64 인코딩된 압축 데이터>
   */
  private parseSSEEvent(event: MessageEvent): { data: unknown; isCompressed: boolean } {
    try {
      // SSE 메시지 파싱 (event.data는 이미 파싱된 문자열)
      // Backend에서 전송하는 형식: "event: <type>\ncompressed: true\ndata: <data>\n\n"
      // 또는 "event: <type>\ndata: <data>\n\n"
      
      // EventSource가 자동으로 파싱하므로 event.data에 이미 JSON 문자열이 있음
      // 하지만 압축 플래그는 별도로 확인해야 함
      
      // 임시로 압축 여부를 확인하기 위해 데이터를 분석
      // 실제로는 커스텀 SSE 파서를 사용하거나 서버에서 별도 헤더로 전송해야 함
      // 현재는 Backend에서 compressed 플래그를 별도 필드로 전송하지 않으므로
      // 데이터 형식으로 판단 (Base64 문자열 패턴 확인)
      
      const rawData = event.data;
      let isCompressed = false;
      
      // Base64 인코딩 패턴 확인 (간단한 휴리스틱)
      // 실제로는 Backend에서 명시적으로 플래그를 전송해야 함
      // 여기서는 데이터가 매우 길고 Base64 패턴인 경우 압축으로 간주
      if (typeof rawData === 'string' && rawData.length > 100) {
        // Base64 패턴 확인 (대략적인 휴리스틱)
        const base64Pattern = /^[A-Za-z0-9+/=]+$/;
        if (base64Pattern.test(rawData.trim()) && rawData.length % 4 === 0) {
          isCompressed = true;
        }
      }
      
      const parsedData = parseSSEMessage(rawData, isCompressed);
      
      return { data: parsedData, isCompressed };
    } catch (error) {
      getErrorLogger().log(error, { service: 'SSE', action: 'parseEvent' });
      // 파싱 실패 시 원본 데이터 반환 시도
      try {
        return { data: JSON.parse(event.data), isCompressed: false };
      } catch {
        throw error;
      }
    }
  }

  /**
   * SSE 메시지에서 실제 이벤트 데이터를 추출합니다.
   * Backend에서 전송하는 구조: { id, event, data: { data: <actualData> }, timestamp }
   */
  private extractEventData(parsedData: unknown): unknown {
    if (typeof parsedData === 'object' && parsedData !== null) {
      const message = parsedData as Record<string, unknown>;
      // Backend 구조: message.data.data에 실제 데이터가 있음
      if (message.data && typeof message.data === 'object' && message.data !== null) {
        const dataWrapper = message.data as Record<string, unknown>;
        if ('data' in dataWrapper) {
          return dataWrapper.data;
        }
      }
      // Fallback: data 필드 자체가 데이터인 경우
      if ('data' in message) {
        return message.data;
      }
    }
    return parsedData;
  }

  private setupEventListeners(): void {
    if (!this.eventSource) return;

    // 연결 성공
    this.eventSource.addEventListener('connected', (event) => {
      logger.debug('SSE connected', { event });
      try {
        const { data: parsedData } = this.parseSSEEvent(event);
        const data = this.extractEventData(parsedData);
        const eventData = data as { client_id?: string };
        this.clientId = eventData.client_id || null;
        this.reconnectAttempts = 0;
        this.isConnecting = false;
        this.callbacks.onConnected?.(data);
      } catch (error) {
        getErrorLogger().log(error, { service: 'SSE', action: 'connected' });
        // Fallback: 기존 방식으로 파싱 시도
        try {
          const data = JSON.parse(event.data);
          this.clientId = data.client_id;
          this.reconnectAttempts = 0;
          this.isConnecting = false;
          this.callbacks.onConnected?.(data);
        } catch (fallbackError) {
          getErrorLogger().log(fallbackError, { service: 'SSE', action: 'connected-fallback' });
        }
      }
    });

    // 공통 이벤트 처리 헬퍼
    const handleEvent = (eventType: string, callback?: (data: unknown) => void) => {
      return (event: MessageEvent) => {
        try {
          const { data: parsedData } = this.parseSSEEvent(event);
          const data = this.extractEventData(parsedData);
          callback?.(data);
        } catch (error) {
          getErrorLogger().log(error, { service: 'SSE', action: eventType });
          // Fallback: 기존 방식으로 파싱 시도
          try {
            const data = JSON.parse(event.data);
            callback?.(data);
          } catch (fallbackError) {
            getErrorLogger().log(fallbackError, { service: 'SSE', action: `${eventType}-fallback` });
          }
        }
      };
    };

    // VM 상태 업데이트
    this.eventSource.addEventListener('vm-status', handleEvent('vm-status', this.callbacks.onVMStatusUpdate));

    // VM 리소스 업데이트
    this.eventSource.addEventListener('vm-resource', handleEvent('vm-resource', this.callbacks.onVMResourceUpdate));

    // Provider 상태 업데이트
    this.eventSource.addEventListener('provider-status', handleEvent('provider-status', this.callbacks.onProviderStatusUpdate));

    // Provider 인스턴스 업데이트
    this.eventSource.addEventListener('provider-instance', handleEvent('provider-instance', this.callbacks.onProviderInstanceUpdate));

    // 시스템 알림
    this.eventSource.addEventListener('system-notification', handleEvent('system-notification', this.callbacks.onSystemNotification));

    // 시스템 알림
    this.eventSource.addEventListener('system-alert', handleEvent('system-alert', this.callbacks.onSystemAlert));

    // Kubernetes 클러스터 이벤트
    this.eventSource.addEventListener('kubernetes-cluster-created', handleEvent('kubernetes-cluster-created', this.callbacks.onKubernetesClusterCreated));
    this.eventSource.addEventListener('kubernetes-cluster-updated', handleEvent('kubernetes-cluster-updated', this.callbacks.onKubernetesClusterUpdated));
    this.eventSource.addEventListener('kubernetes-cluster-deleted', handleEvent('kubernetes-cluster-deleted', this.callbacks.onKubernetesClusterDeleted));
    this.eventSource.addEventListener('kubernetes-cluster-list', handleEvent('kubernetes-cluster-list', this.callbacks.onKubernetesClusterList));

    // Kubernetes Node Pool 이벤트
    this.eventSource.addEventListener('kubernetes-node-pool-created', handleEvent('kubernetes-node-pool-created', this.callbacks.onKubernetesNodePoolCreated));
    this.eventSource.addEventListener('kubernetes-node-pool-updated', handleEvent('kubernetes-node-pool-updated', this.callbacks.onKubernetesNodePoolUpdated));
    this.eventSource.addEventListener('kubernetes-node-pool-deleted', handleEvent('kubernetes-node-pool-deleted', this.callbacks.onKubernetesNodePoolDeleted));

    // Kubernetes Node 이벤트
    this.eventSource.addEventListener('kubernetes-node-created', handleEvent('kubernetes-node-created', this.callbacks.onKubernetesNodeCreated));
    this.eventSource.addEventListener('kubernetes-node-updated', handleEvent('kubernetes-node-updated', this.callbacks.onKubernetesNodeUpdated));
    this.eventSource.addEventListener('kubernetes-node-deleted', handleEvent('kubernetes-node-deleted', this.callbacks.onKubernetesNodeDeleted));

    // Network VPC 이벤트
    this.eventSource.addEventListener('network-vpc-created', handleEvent('network-vpc-created', this.callbacks.onNetworkVPCCreated));
    this.eventSource.addEventListener('network-vpc-updated', handleEvent('network-vpc-updated', this.callbacks.onNetworkVPCUpdated));
    this.eventSource.addEventListener('network-vpc-deleted', handleEvent('network-vpc-deleted', this.callbacks.onNetworkVPCDeleted));
    this.eventSource.addEventListener('network-vpc-list', handleEvent('network-vpc-list', this.callbacks.onNetworkVPCList));

    // Network Subnet 이벤트
    this.eventSource.addEventListener('network-subnet-created', handleEvent('network-subnet-created', this.callbacks.onNetworkSubnetCreated));
    this.eventSource.addEventListener('network-subnet-updated', handleEvent('network-subnet-updated', this.callbacks.onNetworkSubnetUpdated));
    this.eventSource.addEventListener('network-subnet-deleted', handleEvent('network-subnet-deleted', this.callbacks.onNetworkSubnetDeleted));
    this.eventSource.addEventListener('network-subnet-list', handleEvent('network-subnet-list', this.callbacks.onNetworkSubnetList));

    // Network Security Group 이벤트
    this.eventSource.addEventListener('network-security-group-created', handleEvent('network-security-group-created', this.callbacks.onNetworkSecurityGroupCreated));
    this.eventSource.addEventListener('network-security-group-updated', handleEvent('network-security-group-updated', this.callbacks.onNetworkSecurityGroupUpdated));
    this.eventSource.addEventListener('network-security-group-deleted', handleEvent('network-security-group-deleted', this.callbacks.onNetworkSecurityGroupDeleted));
    this.eventSource.addEventListener('network-security-group-list', handleEvent('network-security-group-list', this.callbacks.onNetworkSecurityGroupList));

    // VM 이벤트 (추가)
    this.eventSource.addEventListener('vm-created', handleEvent('vm-created', this.callbacks.onVMCreated));
    this.eventSource.addEventListener('vm-updated', handleEvent('vm-updated', this.callbacks.onVMUpdated));
    this.eventSource.addEventListener('vm-deleted', handleEvent('vm-deleted', this.callbacks.onVMDeleted));
    this.eventSource.addEventListener('vm-list', handleEvent('vm-list', this.callbacks.onVMList));

    // 에러 처리
    this.eventSource.addEventListener('error', (_event) => {
      // Event 객체는 직렬화할 수 없으므로 필요한 정보만 추출
      const readyState = this.eventSource?.readyState;
      const url = this.eventSource?.url;
      
      // readyState에 따른 메시지 생성
      let errorMessage = 'SSE connection error';
      if (readyState === EventSource.CONNECTING) {
        errorMessage = 'SSE connection failed during initialization';
      } else if (readyState === EventSource.CLOSED) {
        errorMessage = 'SSE connection closed unexpectedly';
      } else if (readyState === EventSource.OPEN) {
        errorMessage = 'SSE connection error occurred';
      }
      
      const errorInfo: SSEErrorInfo & { message?: string } = {
        type: 'SSE',
        readyState: readyState,
        url: url,
        timestamp: new Date().toISOString(),
        message: errorMessage,
      };
      
      const readyStateText = readyState === EventSource.CONNECTING ? 'CONNECTING' : 
                            readyState === EventSource.OPEN ? 'OPEN' : 
                            readyState === EventSource.CLOSED ? 'CLOSED' : 'UNKNOWN';
      
      // Error 객체로 변환하여 ErrorLogger가 메시지를 인식할 수 있도록 함
      const error = new Error(errorMessage);
      
      logger.error(errorMessage, error, {
        service: 'SSE',
        readyState,
        readyStateText,
        url,
        timestamp: errorInfo.timestamp,
      });
      (error as Error & { readyState?: number; url?: string }).readyState = readyState;
      (error as Error & { readyState?: number; url?: string }).url = url;
      error.name = 'SSEError';
      
      getErrorLogger().log(error, { type: 'SSE', readyState, url });
      this.isConnecting = false;
      this.callbacks.onError?.(errorInfo);
      this.handleReconnect();
    });

    // 연결 해제
    this.eventSource.addEventListener('close', () => {
      logger.info('SSE connection closed');
      this.isConnecting = false;
      this.handleReconnect();
    });
  }

  private handleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      const error = new Error('Max reconnection attempts reached');
      logger.error('Max reconnection attempts reached', error, { 
        service: 'SSE', 
        attempts: this.reconnectAttempts 
      });
      return;
    }

    this.reconnectAttempts++;
    
    // Exponential backoff with jitter and max delay
    const exponentialDelay = this.baseReconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    const jitter = Math.random() * 1000; // 0-1000ms jitter
    const delay = Math.min(exponentialDelay + jitter, this.maxReconnectDelay);

    logger.debug(`Reconnecting SSE in ${Math.round(delay)}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`, {
      delay: Math.round(delay),
      attempt: this.reconnectAttempts,
      maxAttempts: this.maxReconnectAttempts,
    });

    setTimeout(() => {
      if (this.eventSource?.readyState === EventSource.CLOSED) {
        // 토큰을 다시 가져와서 재연결 (auth-storage에서 가져오기)
        let token: string | null = null;
        try {
          const authStorage = localStorage.getItem(STORAGE_KEYS.AUTH_STORAGE);
          if (authStorage) {
            const parsed = JSON.parse(authStorage);
            token = parsed?.state?.token || null;
          }
          // Fallback to legacy token for backward compatibility
          if (!token) {
            token = localStorage.getItem('token');
          }
        } catch {
          token = localStorage.getItem('token');
        }
        if (token) {
          this.connect(token, this.callbacks);
        }
      }
    }, delay);
  }

  isConnected(): boolean {
    return this.eventSource?.readyState === EventSource.OPEN;
  }

  getReadyState(): number | null {
    return this.eventSource?.readyState ?? null;
  }

  // 구독 관리 메서드들
  subscribeToEvent(eventType: string): void {
    this.subscribedEvents.add(eventType);
    this.sendSubscriptionUpdate();
  }

  unsubscribeFromEvent(eventType: string): void {
    this.subscribedEvents.delete(eventType);
    this.sendSubscriptionUpdate();
  }

  subscribeToVM(vmId: string): void {
    this.subscribedVMs.add(vmId);
    this.sendSubscriptionUpdate();
  }

  unsubscribeFromVM(vmId: string): void {
    this.subscribedVMs.delete(vmId);
    this.sendSubscriptionUpdate();
  }

  private sendSubscriptionUpdate(): void {
    if (!this.clientId || !this.isConnected()) {
      return;
    }

    // 구독 정보를 서버에 전송 (실제 구현에서는 별도 API 엔드포인트 사용)
    const subscriptionData = {
      clientId: this.clientId,
      events: Array.from(this.subscribedEvents),
      vms: Array.from(this.subscribedVMs),
    };

    logger.debug('Subscription update', { subscriptionData });
    // TODO: 서버에 구독 정보 전송
  }

  getSubscriptions() {
    return {
      events: Array.from(this.subscribedEvents),
      vms: Array.from(this.subscribedVMs),
    };
  }
}

export const sseService = new SSEService();
