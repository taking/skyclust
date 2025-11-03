/**
 * Server-Sent Events (SSE) 관련 타입 정의
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

