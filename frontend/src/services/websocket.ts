import { io, Socket } from 'socket.io-client';

class WebSocketService {
  private socket: Socket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  connect(token: string): void {
    if (this.socket?.connected) {
      return;
    }

    this.socket = io(process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8081', {
      auth: {
        token,
      },
      transports: ['websocket', 'polling'],
    });

    this.setupEventListeners();
  }

  disconnect(): void {
    if (this.socket) {
      this.socket.disconnect();
      this.socket = null;
    }
  }

  private setupEventListeners(): void {
    if (!this.socket) return;

    this.socket.on('connect', () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
    });

    this.socket.on('disconnect', (reason) => {
      console.log('WebSocket disconnected:', reason);
      this.handleReconnect();
    });

    this.socket.on('connect_error', (error) => {
      console.error('WebSocket connection error:', error);
      this.handleReconnect();
    });
  }

  private handleReconnect(): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
      
      setTimeout(() => {
        if (this.socket && !this.socket.connected) {
          this.socket.connect();
        }
      }, delay);
    }
  }

  // VM 관련 이벤트
  onVMStatusUpdate(callback: (data: { vmId: string; status: string; timestamp: number }) => void): void {
    this.socket?.on('vm:status_update', callback);
  }

  onVMResourceUpdate(callback: (data: { vmId: string; cpu: number; memory: number; disk: number; timestamp: number }) => void): void {
    this.socket?.on('vm:resource_update', callback);
  }

  onVMError(callback: (data: { vmId: string; error: string; timestamp: number }) => void): void {
    this.socket?.on('vm:error', callback);
  }

  // Provider 관련 이벤트
  onProviderStatusUpdate(callback: (data: { provider: string; status: string; timestamp: number }) => void): void {
    this.socket?.on('provider:status_update', callback);
  }

  onProviderInstanceUpdate(callback: (data: { provider: string; instances: unknown[]; timestamp: number }) => void): void {
    this.socket?.on('provider:instance_update', callback);
  }

  // 시스템 알림
  onSystemNotification(callback: (data: { type: 'info' | 'warning' | 'error'; message: string; timestamp: number }) => void): void {
    this.socket?.on('system:notification', callback);
  }

  onSystemAlert(callback: (data: { level: 'low' | 'medium' | 'high' | 'critical'; message: string; timestamp: number }) => void): void {
    this.socket?.on('system:alert', callback);
  }

  // 구독 관리
  subscribeToVM(vmId: string): void {
    this.socket?.emit('subscribe:vm', { vmId });
  }

  unsubscribeFromVM(vmId: string): void {
    this.socket?.emit('unsubscribe:vm', { vmId });
  }

  subscribeToProvider(provider: string): void {
    this.socket?.emit('subscribe:provider', { provider });
  }

  unsubscribeFromProvider(provider: string): void {
    this.socket?.emit('unsubscribe:provider', { provider });
  }

  // 이벤트 리스너 제거
  offVMStatusUpdate(callback?: (data: { vmId: string; status: string; timestamp: number }) => void): void {
    this.socket?.off('vm:status_update', callback);
  }

  offVMResourceUpdate(callback?: (data: { vmId: string; cpu: number; memory: number; disk: number; timestamp: number }) => void): void {
    this.socket?.off('vm:resource_update', callback);
  }

  offVMError(callback?: (data: { vmId: string; error: string; timestamp: number }) => void): void {
    this.socket?.off('vm:error', callback);
  }

  offProviderStatusUpdate(callback?: (data: { provider: string; status: string; timestamp: number }) => void): void {
    this.socket?.off('provider:status_update', callback);
  }

  offProviderInstanceUpdate(callback?: (data: { provider: string; instances: unknown[]; timestamp: number }) => void): void {
    this.socket?.off('provider:instance_update', callback);
  }

  offSystemNotification(callback?: (data: { type: 'info' | 'warning' | 'error'; message: string; timestamp: number }) => void): void {
    this.socket?.off('system:notification', callback);
  }

  offSystemAlert(callback?: (data: { level: 'low' | 'medium' | 'high' | 'critical'; message: string; timestamp: number }) => void): void {
    this.socket?.off('system:alert', callback);
  }

  // 연결 상태 확인
  isConnected(): boolean {
    return this.socket?.connected || false;
  }

  getSocket(): Socket | null {
    return this.socket;
  }
}

export const webSocketService = new WebSocketService();
