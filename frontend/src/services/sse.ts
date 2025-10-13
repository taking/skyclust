/**
 * Server-Sent Events (SSE) Service
 * NATS 기반 실시간 데이터 수신
 */

export interface SSEMessage {
  type: string;
  data: unknown;
  timestamp: number;
}

export interface SSECallbacks {
  onVMStatusUpdate?: (data: unknown) => void;
  onVMResourceUpdate?: (data: unknown) => void;
  onProviderStatusUpdate?: (data: unknown) => void;
  onProviderInstanceUpdate?: (data: unknown) => void;
  onSystemNotification?: (data: unknown) => void;
  onSystemAlert?: (data: unknown) => void;
  onConnected?: (data: unknown) => void;
  onError?: (error: Event) => void;
}

class SSEService {
  private eventSource: EventSource | null = null;
  private callbacks: SSECallbacks = {};
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private isConnecting = false;
  private clientId: string | null = null;
  private subscribedEvents = new Set<string>();
  private subscribedVMs = new Set<string>();
  private subscribedProviders = new Set<string>();

  connect(token: string, callbacks: SSECallbacks = {}): void {
    if (this.eventSource?.readyState === EventSource.OPEN) {
      return;
    }

    if (this.isConnecting) {
      return;
    }

    this.isConnecting = true;
    this.callbacks = callbacks;

    const url = `${process.env.NEXT_PUBLIC_API_URL}/api/events?token=${token}`;
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

  private setupEventListeners(): void {
    if (!this.eventSource) return;

    // 연결 성공
    this.eventSource.addEventListener('connected', (event) => {
      console.log('SSE connected:', event);
      const data = JSON.parse(event.data);
      this.clientId = data.client_id;
      this.reconnectAttempts = 0;
      this.isConnecting = false;
      this.callbacks.onConnected?.(data);
    });

    // VM 상태 업데이트
    this.eventSource.addEventListener('vm-status', (event) => {
      const data = JSON.parse(event.data);
      this.callbacks.onVMStatusUpdate?.(data);
    });

    // VM 리소스 업데이트
    this.eventSource.addEventListener('vm-resource', (event) => {
      const data = JSON.parse(event.data);
      this.callbacks.onVMResourceUpdate?.(data);
    });

    // Provider 상태 업데이트
    this.eventSource.addEventListener('provider-status', (event) => {
      const data = JSON.parse(event.data);
      this.callbacks.onProviderStatusUpdate?.(data);
    });

    // Provider 인스턴스 업데이트
    this.eventSource.addEventListener('provider-instance', (event) => {
      const data = JSON.parse(event.data);
      this.callbacks.onProviderInstanceUpdate?.(data);
    });

    // 시스템 알림
    this.eventSource.addEventListener('system-notification', (event) => {
      const data = JSON.parse(event.data);
      this.callbacks.onSystemNotification?.(data);
    });

    // 시스템 알림
    this.eventSource.addEventListener('system-alert', (event) => {
      const data = JSON.parse(event.data);
      this.callbacks.onSystemAlert?.(data);
    });

    // 에러 처리
    this.eventSource.addEventListener('error', (event) => {
      console.error('SSE error:', event);
      this.isConnecting = false;
      this.callbacks.onError?.(event);
      this.handleReconnect();
    });

    // 연결 해제
    this.eventSource.addEventListener('close', () => {
      console.log('SSE connection closed');
      this.isConnecting = false;
      this.handleReconnect();
    });
  }

  private handleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    console.log(`Reconnecting SSE in ${delay}ms (attempt ${this.reconnectAttempts})`);

    setTimeout(() => {
      if (this.eventSource?.readyState === EventSource.CLOSED) {
        // 토큰을 다시 가져와서 재연결
        const token = localStorage.getItem('token');
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

  subscribeToProvider(provider: string): void {
    this.subscribedProviders.add(provider);
    this.sendSubscriptionUpdate();
  }

  unsubscribeFromProvider(provider: string): void {
    this.subscribedProviders.delete(provider);
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
      providers: Array.from(this.subscribedProviders),
    };

    console.log('Subscription update:', subscriptionData);
    // TODO: 서버에 구독 정보 전송
  }

  getSubscriptions() {
    return {
      events: Array.from(this.subscribedEvents),
      vms: Array.from(this.subscribedVMs),
      providers: Array.from(this.subscribedProviders),
    };
  }
}

export const sseService = new SSEService();
