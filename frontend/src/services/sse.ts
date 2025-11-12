/**
 * Server-Sent Events (SSE) Service
 * NATS 기반 실시간 데이터 수신
 */

import type { SSECallbacks, SSEErrorInfo } from '@/lib/types/sse';
import { API_CONFIG, API_ENDPOINTS, api, getApiUrl } from '@/lib/api';
import { CONNECTION, STORAGE_KEYS } from '@/lib/constants';
import { logger } from '@/lib/logging/logger';
import { parseSSEMessage } from '@/lib/sse';
import { log } from '@/lib/logging';

interface SubscriptionFilters {
  providers?: string[];
  credential_ids?: string[];
  regions?: string[];
}

interface SubscriptionInfo {
  eventType: string;
  filters?: SubscriptionFilters;
  timestamp: Date;
}

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
  private currentToken: string | null = null; // 현재 연결된 토큰 추적
  private lastEventId: string | null = null; // 마지막 수신한 이벤트 ID
  private retryInterval: number = 3000; // 재연결 간격 (밀리초, 기본 3초)
  // 구독 추적: subscriptionKey -> SubscriptionInfo
  private activeSubscriptions = new Map<string, SubscriptionInfo>();

  connect(token: string, callbacks: SSECallbacks = {}): void {
    // 토큰이 없으면 연결하지 않음
    if (!token || token.trim() === '') {
      log.warn('SSE connect called with empty token, skipping connection');
      return;
    }

    // 이미 같은 토큰으로 연결되어 있으면 재연결하지 않음
    if (this.eventSource?.readyState === EventSource.OPEN && this.currentToken === token) {
      log.debug('SSE already connected with same token, skipping reconnection', {
        token: token.substring(0, 20) + '...',
      });
      // 콜백만 업데이트 (연결은 유지)
      this.updateCallbacks(callbacks);
      return;
    }

    // 연결 중이면 대기
    if (this.isConnecting) {
      log.debug('SSE connection already in progress, skipping duplicate connection attempt');
      return;
    }

    // 기존 연결이 있으면 먼저 정리
    if (this.eventSource) {
      log.debug('Closing existing SSE connection before reconnecting');
      this.disconnect();
    }

    this.isConnecting = true;
    this.currentToken = token;
    this.callbacks = callbacks;

    // localStorage에서 마지막 이벤트 ID 로드
    this.loadLastEventId();

    // Token과 Last-Event-ID를 URL 인코딩하여 쿼리 파라미터로 전달
    const encodedToken = encodeURIComponent(token);
    const params = new URLSearchParams({
      token: encodedToken,
    });
    
    // Last-Event-ID가 있으면 쿼리 파라미터로 추가
    if (this.lastEventId) {
      params.set('last_event_id', this.lastEventId);
    }

    const endpoint = `${API_ENDPOINTS.sse.connect()}?${params.toString()}`;
    const url = `${API_CONFIG.BASE_URL}${API_CONFIG.API_PREFIX}/${API_CONFIG.VERSION}/${endpoint}`;
    
    log.debug('Connecting to SSE', {
      url: url.replace(token, '***').replace(this.lastEventId || '', '***'),
      endpoint,
      lastEventId: this.lastEventId ? '***' : null,
    });

    try {
      this.eventSource = new EventSource(url);
      this.setupEventListeners();
    } catch (error) {
      log.error('Failed to create EventSource', error, {
        service: 'SSE',
        action: 'connect',
        url: url.replace(token, '***').replace(this.lastEventId || '', '***'),
      });
      this.isConnecting = false;
      this.currentToken = null;
      throw error;
    }
  }

  disconnect(): void {
    if (this.eventSource) {
      log.debug('Disconnecting SSE', {
        readyState: this.eventSource.readyState,
        url: this.eventSource.url?.replace(this.currentToken || '', '***'),
      });
      this.eventSource.close();
      this.eventSource = null;
    }
    this.isConnecting = false;
    this.reconnectAttempts = 0;
    this.currentToken = null;
    this.clientId = null;
    // 구독 정보 초기화
    this.subscribedEvents.clear();
    this.activeSubscriptions.clear();
    this.subscribedVMs.clear();
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
      logger.logError(error, { service: 'SSE', action: 'parseEvent' });
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
      log.info('SSE connected successfully', {
        hasClientId: !!this.clientId,
        reconnectAttempts: this.reconnectAttempts,
      });
      try {
        const { data: parsedData } = this.parseSSEEvent(event);
        const data = this.extractEventData(parsedData);
        const eventData = data as { client_id?: string; connection_id?: string };
        // connection_id 또는 client_id 중 하나를 사용
        this.clientId = eventData.connection_id || eventData.client_id || null;
        this.reconnectAttempts = 0;
        this.isConnecting = false;
        log.debug('SSE connection established', {
          clientId: this.clientId,
          data: eventData,
        });
        this.callbacks.onConnected?.(data);
      } catch (error) {
        logger.logError(error, { service: 'SSE', action: 'connected' });
        // Fallback: 기존 방식으로 파싱 시도
        try {
          const data = JSON.parse(event.data);
          this.clientId = data.connection_id || data.client_id || null;
          this.reconnectAttempts = 0;
          this.isConnecting = false;
          log.debug('SSE connection established (fallback)', {
            clientId: this.clientId,
          });
          this.callbacks.onConnected?.(data);
        } catch (fallbackError) {
          logger.logError(fallbackError, { service: 'SSE', action: 'connected-fallback' });
          this.isConnecting = false;
        }
      }
    });

    // Retry 이벤트 처리
    this.eventSource.addEventListener('message', (event: MessageEvent) => {
      // Retry 이벤트 처리 (서버에서 전송한 retry 간격)
      if (event.data && event.data.startsWith('retry:')) {
        const retryValue = parseInt(event.data.replace('retry:', '').trim(), 10);
        if (!isNaN(retryValue) && retryValue > 0) {
          this.retryInterval = retryValue;
          log.debug('SSE retry interval updated', { retryInterval: this.retryInterval });
        }
      }

      // 이벤트 ID 추적 및 저장
      if (event.lastEventId) {
        this.lastEventId = event.lastEventId;
        this.saveLastEventId(event.lastEventId);
      }
    });

    // 공통 이벤트 처리 헬퍼
    const handleEvent = (eventType: string, callback?: (data: unknown) => void) => {
      return (event: MessageEvent) => {
        // 이벤트 ID 추적 및 저장
        if (event.lastEventId) {
          this.lastEventId = event.lastEventId;
          this.saveLastEventId(event.lastEventId);
        }

        try {
          const { data: parsedData } = this.parseSSEEvent(event);
          const data = this.extractEventData(parsedData);
          callback?.(data);
        } catch (error) {
          logger.logError(error, { service: 'SSE', action: eventType });
          // Fallback: 기존 방식으로 파싱 시도
          try {
            const data = JSON.parse(event.data);
            callback?.(data);
          } catch (fallbackError) {
            logger.logError(fallbackError, { service: 'SSE', action: `${eventType}-fallback` });
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

    // Azure Resource Group 이벤트
    this.eventSource.addEventListener('azure-resource-group-created', handleEvent('azure-resource-group-created', this.callbacks.onAzureResourceGroupCreated));
    this.eventSource.addEventListener('azure-resource-group-updated', handleEvent('azure-resource-group-updated', this.callbacks.onAzureResourceGroupUpdated));
    this.eventSource.addEventListener('azure-resource-group-deleted', handleEvent('azure-resource-group-deleted', this.callbacks.onAzureResourceGroupDeleted));
    this.eventSource.addEventListener('azure-resource-group-list', handleEvent('azure-resource-group-list', this.callbacks.onAzureResourceGroupList));

    // Dashboard Summary 이벤트
    this.eventSource.addEventListener('dashboard-summary-updated', handleEvent('dashboard-summary-updated', this.callbacks.onDashboardSummaryUpdated));

    // 에러 처리
    this.eventSource.addEventListener('error', (_event) => {
      // Event 객체는 직렬화할 수 없으므로 필요한 정보만 추출
      const readyState = this.eventSource?.readyState;
      const url = this.eventSource?.url;
      
      // URL에서 토큰 부분을 마스킹하여 로그에 출력
      const maskedUrl = url ? url.replace(/token=[^&]*/, 'token=***') : undefined;
      
      // readyState에 따른 메시지 생성
      let errorMessage = 'SSE connection error';
      let shouldReconnect = true;
      
      if (readyState === EventSource.CONNECTING) {
        errorMessage = 'SSE connection failed during initialization';
        // CONNECTING 상태에서 에러는 인증 실패일 가능성이 높음
        log.warn('SSE connection failed during initialization - possible authentication issue', {
          readyState,
          url: maskedUrl,
          hasToken: !!this.currentToken,
        });
      } else if (readyState === EventSource.CLOSED) {
        errorMessage = 'SSE connection closed unexpectedly';
        // CLOSED 상태는 정상적인 종료일 수도 있으므로 재연결 시도
      } else if (readyState === EventSource.OPEN) {
        errorMessage = 'SSE connection error occurred';
        // OPEN 상태에서 에러는 네트워크 문제일 가능성이 높음
      }
      
      const errorInfo: SSEErrorInfo & { message?: string } = {
        type: 'SSE',
        readyState: readyState,
        url: maskedUrl,
        timestamp: new Date().toISOString(),
        message: errorMessage,
      };
      
      const readyStateText = readyState === EventSource.CONNECTING ? 'CONNECTING' : 
                            readyState === EventSource.OPEN ? 'OPEN' : 
                            readyState === EventSource.CLOSED ? 'CLOSED' : 'UNKNOWN';
      
      // Error 객체로 변환하여 ErrorLogger가 메시지를 인식할 수 있도록 함
      const error = new Error(errorMessage);
      
      log.error(errorMessage, error, {
        service: 'SSE',
        readyState,
        readyStateText,
        url: maskedUrl,
        timestamp: errorInfo.timestamp,
        reconnectAttempts: this.reconnectAttempts,
        maxReconnectAttempts: this.maxReconnectAttempts,
      });
      
      (error as Error & { readyState?: number; url?: string }).readyState = readyState;
      (error as Error & { readyState?: number; url?: string }).url = maskedUrl;
      error.name = 'SSEError';
      
      logger.logError(error, { 
        type: 'SSE', 
        readyState, 
        url: maskedUrl,
        reconnectAttempts: this.reconnectAttempts,
      });
      
      this.isConnecting = false;
      this.callbacks.onError?.(errorInfo);
      
      // 재연결 시도 (최대 시도 횟수 내에서만)
      if (shouldReconnect) {
        this.handleReconnect();
      }
    });

    // 연결 해제
    this.eventSource.addEventListener('close', () => {
      log.info('SSE connection closed');
      this.isConnecting = false;
      this.handleReconnect();
    });
  }

  private handleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      const error = new Error('Max reconnection attempts reached');
      log.error('Max reconnection attempts reached', error, { 
        service: 'SSE', 
        attempts: this.reconnectAttempts 
      });
      return;
    }

    this.reconnectAttempts++;
    
    // 서버에서 전송한 retryInterval을 우선 사용, 없으면 exponential backoff
    let delay: number;
    if (this.retryInterval > 0) {
      // 서버에서 지정한 재연결 간격 사용 (jitter 추가)
      const jitter = Math.random() * 500; // 0-500ms jitter
      delay = this.retryInterval + jitter;
    } else {
      // Exponential backoff with jitter and max delay
      const exponentialDelay = this.baseReconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
      const jitter = Math.random() * 1000; // 0-1000ms jitter
      delay = Math.min(exponentialDelay + jitter, this.maxReconnectDelay);
    }

    log.debug(`Reconnecting SSE in ${Math.round(delay)}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`, {
      delay: Math.round(delay),
      attempt: this.reconnectAttempts,
      maxAttempts: this.maxReconnectAttempts,
      retryInterval: this.retryInterval,
    });

    setTimeout(() => {
      if (this.eventSource?.readyState === EventSource.CLOSED || !this.eventSource) {
        // 토큰을 다시 가져와서 재연결 (auth-storage에서 가져오기)
        let token: string | null = this.currentToken; // 먼저 현재 토큰 사용
        
        // 현재 토큰이 없으면 스토리지에서 가져오기
        if (!token) {
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

  getClientId(): string | null {
    return this.clientId;
  }

  // 구독 키 생성 (eventType + filters의 해시)
  private getSubscriptionKey(eventType: string, filters?: SubscriptionFilters): string {
    if (!filters || Object.keys(filters).length === 0) {
      return eventType;
    }
    
    // 필터를 정렬하여 일관된 키 생성
    const sortedFilters = JSON.stringify({
      providers: filters.providers?.sort(),
      credential_ids: filters.credential_ids?.sort(),
      regions: filters.regions?.sort(),
    });
    
    return `${eventType}:${sortedFilters}`;
  }

  // 구독 관리 메서드들
  async subscribeToEvent(eventType: string, filters?: SubscriptionFilters): Promise<void> {
    // 연결 상태 확인
    if (!this.isConnected()) {
      log.warn('Cannot subscribe: SSE not connected', {
        eventType,
        connected: this.isConnected(),
        readyState: this.eventSource?.readyState,
      });
      return;
    }

    // clientId가 없으면 잠시 대기 후 재시도 (최대 3초)
    if (!this.clientId) {
      log.debug('Waiting for clientId to be set...', { eventType });
      // 최대 3초 대기 (300ms 간격으로 10번 시도)
      for (let i = 0; i < 10; i++) {
        await new Promise(resolve => setTimeout(resolve, 300));
        if (this.clientId) {
          log.debug('ClientId set, proceeding with subscription', { eventType, clientId: this.clientId });
          break;
        }
      }
      // 여전히 clientId가 없으면 실패
      if (!this.clientId) {
        log.error('Failed to get clientId after waiting, cannot subscribe', { eventType });
        return;
      }
    }

    const subscriptionKey = this.getSubscriptionKey(eventType, filters);
    
    // 이미 동일한 구독이 활성화되어 있는지 확인
    if (this.activeSubscriptions.has(subscriptionKey)) {
      const existing = this.activeSubscriptions.get(subscriptionKey)!;
      // 중복이지만 subscribedEvents Set에는 eventType이 포함되어야 함
      this.subscribedEvents.add(eventType);
      log.debug('Subscription already active, skipping', {
        eventType,
        filters,
        existingTimestamp: existing.timestamp,
      });
      return;
    }

    try {
      await api.post(getApiUrl(API_ENDPOINTS.sse.subscribe()), {
        event_type: eventType,
        filters: filters || {},
      });

      // 구독 정보 저장
      // subscribedEvents Set에는 eventType만 저장 (중복 방지)
      this.subscribedEvents.add(eventType);
      // activeSubscriptions Map에는 필터 정보와 함께 저장
      this.activeSubscriptions.set(subscriptionKey, {
        eventType,
        filters,
        timestamp: new Date(),
      });
      
      log.debug('Subscribed to event', { eventType, filters, subscriptionKey });
    } catch (error) {
      logger.logError(error, {
        service: 'SSE',
        action: 'subscribeToEvent',
        eventType,
        filters,
      });
      throw error;
    }
  }

  async unsubscribeFromEvent(eventType: string, filters?: SubscriptionFilters): Promise<void> {
    if (!this.isConnected() || !this.clientId) {
      log.warn('Cannot unsubscribe: SSE not connected', {
        eventType,
        connected: this.isConnected(),
        clientId: this.clientId,
      });
      return;
    }

    const subscriptionKey = this.getSubscriptionKey(eventType, filters);
    
    // 구독이 활성화되어 있지 않으면 스킵
    if (!this.activeSubscriptions.has(subscriptionKey)) {
      log.debug('Subscription not found, skipping unsubscribe', {
        eventType,
        filters,
        subscriptionKey,
      });
      return;
    }

    try {
      await api.post(getApiUrl(API_ENDPOINTS.sse.unsubscribe()), {
        event_type: eventType,
        filters: filters || {},
      });

      // 구독 정보 제거
      this.activeSubscriptions.delete(subscriptionKey);
      
      // 해당 eventType의 다른 구독이 없으면 Set에서도 제거
      const hasOtherSubscriptions = Array.from(this.activeSubscriptions.values())
        .some(sub => sub.eventType === eventType);
      if (!hasOtherSubscriptions) {
        this.subscribedEvents.delete(eventType);
      }
      
      log.debug('Unsubscribed from event', { eventType, filters, subscriptionKey });
    } catch (error) {
      logger.logError(error, {
        service: 'SSE',
        action: 'unsubscribeFromEvent',
        eventType,
        filters,
      });
      throw error;
    }
  }

  // 특정 eventType의 모든 구독 해제
  async unsubscribeFromAllEventType(eventType: string): Promise<void> {
    const subscriptionsToRemove: string[] = [];
    
    // 해당 eventType의 모든 구독 찾기
    for (const [key, info] of this.activeSubscriptions.entries()) {
      if (info.eventType === eventType) {
        subscriptionsToRemove.push(key);
      }
    }

    // 모든 구독 해제
    for (const key of subscriptionsToRemove) {
      const info = this.activeSubscriptions.get(key)!;
      await this.unsubscribeFromEvent(eventType, info.filters);
    }
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
    // 구독 정보는 이제 subscribeToEvent/unsubscribeFromEvent에서 직접 API 호출
    // 이 메서드는 더 이상 필요하지 않지만, 호환성을 위해 유지
    log.debug('Subscription state updated', {
      events: Array.from(this.subscribedEvents),
      vms: Array.from(this.subscribedVMs),
    });
  }

  getSubscriptions() {
    return {
      events: Array.from(this.subscribedEvents),
      vms: Array.from(this.subscribedVMs),
    };
  }

  /**
   * 현재 구독 중인 이벤트 타입 Set을 반환합니다.
   * @returns 구독 중인 이벤트 타입 Set
   */
  getSubscribedEvents(): Set<string> {
    return new Set(this.subscribedEvents);
  }

  /**
   * 활성 구독 정보를 조회합니다.
   * @returns 활성 구독 정보 Map
   */
  getActiveSubscriptions(): Map<string, SubscriptionInfo> {
    return new Map(this.activeSubscriptions);
  }

  /**
   * 특정 eventType의 활성 구독 개수를 조회합니다.
   * @param eventType - 이벤트 타입
   * @returns 활성 구독 개수
   */
  getActiveSubscriptionCount(eventType: string): number {
    return Array.from(this.activeSubscriptions.values())
      .filter(info => info.eventType === eventType)
      .length;
  }

  /**
   * 여러 이벤트를 한 번에 구독합니다.
   * @param eventTypes - 구독할 이벤트 타입 배열
   * @param filters - 필터 옵션
   */
  async subscribeToEvents(
    eventTypes: string[],
    filters?: {
      providers?: string[];
      credential_ids?: string[];
      regions?: string[];
    }
  ): Promise<void> {
    const promises = eventTypes.map((eventType) =>
      this.subscribeToEvent(eventType, filters)
    );
    await Promise.allSettled(promises);
  }

  /**
   * 여러 이벤트를 한 번에 구독 해제합니다.
   * @param eventTypes - 구독 해제할 이벤트 타입 배열
   */
  async unsubscribeFromEvents(eventTypes: string[]): Promise<void> {
    const promises = eventTypes.map((eventType) =>
      this.unsubscribeFromEvent(eventType)
    );
    await Promise.allSettled(promises);
  }

  /**
   * 필요한 이벤트만 구독하고 불필요한 구독을 해제합니다.
   * @param requiredEvents - 필요한 이벤트 타입 Set
   * @param filters - 필터 옵션
   */
  async syncSubscriptions(
    requiredEvents: Set<string>,
    filters?: {
      providers?: string[];
      credential_ids?: string[];
      regions?: string[];
    }
  ): Promise<void> {
    // SSE 연결 확인
    if (!this.isConnected()) {
      log.warn('Cannot sync subscriptions: SSE not connected', {
        requiredEvents: Array.from(requiredEvents),
        filters,
      });
      throw new Error('SSE not connected');
    }

    const currentEvents = this.getSubscribedEvents();
    
    // 시스템 이벤트는 항상 유지 (system-notification, system-alert)
    const systemEvents = new Set(['system-notification', 'system-alert']);
    
    // 구독할 이벤트: requiredEvents에 있지만 현재 구독되지 않은 이벤트
    const toSubscribe = Array.from(requiredEvents).filter(
      (event) => !currentEvents.has(event) && !systemEvents.has(event)
    );
    
    // 구독 해제할 이벤트: 현재 구독 중이지만 requiredEvents에 없고, 시스템 이벤트가 아닌 이벤트
    const toUnsubscribe = Array.from(currentEvents).filter(
      (event) => !requiredEvents.has(event) && !systemEvents.has(event)
    );

    if (toSubscribe.length > 0) {
      log.debug('Subscribing to new events', {
        events: toSubscribe,
        filters,
      });
      await this.subscribeToEvents(toSubscribe, filters);
    }

    if (toUnsubscribe.length > 0) {
      log.debug('Unsubscribing from unused events', {
        events: toUnsubscribe,
      });
      await this.unsubscribeFromEvents(toUnsubscribe);
    }
  }

  /**
   * 현재 등록된 콜백을 반환합니다.
   * @returns 현재 콜백 객체
   */
  getCallbacks(): SSECallbacks {
    return { ...this.callbacks };
  }

  /**
   * 콜백을 업데이트합니다. 기존 콜백과 병합됩니다.
   * @param newCallbacks - 추가할 콜백
   */
  updateCallbacks(newCallbacks: SSECallbacks): void {
    this.callbacks = { ...this.callbacks, ...newCallbacks };
    log.debug('SSE callbacks updated', {
      callbackCount: Object.keys(this.callbacks).length,
    });
  }

  /**
   * 마지막 이벤트 ID를 localStorage에서 로드합니다.
   */
  private loadLastEventId(): void {
    try {
      if (typeof window !== 'undefined') {
        const stored = localStorage.getItem(STORAGE_KEYS.SSE_LAST_EVENT_ID);
        if (stored) {
          this.lastEventId = stored;
          log.debug('Last event ID loaded from storage', {
            lastEventId: this.lastEventId.substring(0, 20) + '...',
          });
        }
      }
    } catch (error) {
      logger.logError(error, { service: 'SSE', action: 'loadLastEventId' });
    }
  }

  /**
   * 마지막 이벤트 ID를 localStorage에 저장합니다.
   * @param eventId - 저장할 이벤트 ID
   */
  private saveLastEventId(eventId: string): void {
    try {
      if (typeof window !== 'undefined' && eventId) {
        localStorage.setItem(STORAGE_KEYS.SSE_LAST_EVENT_ID, eventId);
        log.debug('Last event ID saved to storage', {
          eventId: eventId.substring(0, 20) + '...',
        });
      }
    } catch (error) {
      logger.logError(error, { service: 'SSE', action: 'saveLastEventId' });
    }
  }

  /**
   * 토큰 갱신 시 재연결합니다.
   * @param newToken - 새로운 토큰
   */
  async refreshToken(newToken: string): Promise<void> {
    const savedLastEventId = this.lastEventId;
    
    log.info('Refreshing SSE token', {
      hasLastEventId: !!savedLastEventId,
    });

    // 기존 연결 종료
    this.disconnect();

    // 새 토큰으로 재연결 (Last-Event-ID 포함)
    this.connect(newToken, this.callbacks);

    // Last-Event-ID는 connect() 내부에서 자동으로 로드됨
  }

  /**
   * 현재 재연결 간격을 반환합니다.
   * @returns 재연결 간격 (밀리초)
   */
  getRetryInterval(): number {
    return this.retryInterval;
  }
}

export const sseService = new SSEService();
