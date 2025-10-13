import { useEffect, useCallback, useRef } from 'react';
import { useAuthStore } from '@/store/auth';
import { webSocketService } from '@/services/websocket';
import { useToast } from '@/hooks/useToast';

interface VMStatusUpdate {
  vmId: string;
  status: string;
  timestamp: number;
}

interface VMResourceUpdate {
  vmId: string;
  cpu: number;
  memory: number;
  disk: number;
  timestamp: number;
}

interface VMError {
  vmId: string;
  error: string;
  timestamp: number;
}

interface ProviderStatusUpdate {
  provider: string;
  status: string;
  timestamp: number;
}

interface ProviderInstanceUpdate {
  provider: string;
  instances: unknown[];
  timestamp: number;
}

interface SystemNotification {
  type: 'info' | 'warning' | 'error';
  message: string;
  timestamp: number;
}

interface SystemAlert {
  level: 'low' | 'medium' | 'high' | 'critical';
  message: string;
  timestamp: number;
}

export function useRealtimeMonitoring() {
  const { token } = useAuthStore();
  const { success, error } = useToast();
  const callbacksRef = useRef<{
    onVMStatusUpdate?: (data: VMStatusUpdate) => void;
    onVMResourceUpdate?: (data: VMResourceUpdate) => void;
    onVMError?: (data: VMError) => void;
    onProviderStatusUpdate?: (data: ProviderStatusUpdate) => void;
    onProviderInstanceUpdate?: (data: ProviderInstanceUpdate) => void;
    onSystemNotification?: (data: SystemNotification) => void;
    onSystemAlert?: (data: SystemAlert) => void;
  }>({});

  // WebSocket 연결
  useEffect(() => {
    if (token) {
      webSocketService.connect(token);
    }

    return () => {
      webSocketService.disconnect();
    };
  }, [token]);

  // VM 상태 업데이트 콜백 등록
  const onVMStatusUpdate = useCallback((callback: (data: VMStatusUpdate) => void) => {
    callbacksRef.current.onVMStatusUpdate = callback;
    webSocketService.onVMStatusUpdate(callback);
  }, []);

  // VM 리소스 업데이트 콜백 등록
  const onVMResourceUpdate = useCallback((callback: (data: VMResourceUpdate) => void) => {
    callbacksRef.current.onVMResourceUpdate = callback;
    webSocketService.onVMResourceUpdate(callback);
  }, []);

  // VM 에러 콜백 등록
  const onVMError = useCallback((callback: (data: VMError) => void) => {
    callbacksRef.current.onVMError = callback;
    webSocketService.onVMError(callback);
  }, []);

  // Provider 상태 업데이트 콜백 등록
  const onProviderStatusUpdate = useCallback((callback: (data: ProviderStatusUpdate) => void) => {
    callbacksRef.current.onProviderStatusUpdate = callback;
    webSocketService.onProviderStatusUpdate(callback);
  }, []);

  // Provider 인스턴스 업데이트 콜백 등록
  const onProviderInstanceUpdate = useCallback((callback: (data: ProviderInstanceUpdate) => void) => {
    callbacksRef.current.onProviderInstanceUpdate = callback;
    webSocketService.onProviderInstanceUpdate(callback);
  }, []);

  // 시스템 알림 콜백 등록
  const onSystemNotification = useCallback((callback: (data: SystemNotification) => void) => {
    callbacksRef.current.onSystemNotification = callback;
    webSocketService.onSystemNotification(callback);
  }, []);

  // 시스템 알림 콜백 등록
  const onSystemAlert = useCallback((callback: (data: SystemAlert) => void) => {
    callbacksRef.current.onSystemAlert = callback;
    webSocketService.onSystemAlert(callback);
  }, []);

  // VM 구독
  const subscribeToVM = useCallback((vmId: string) => {
    webSocketService.subscribeToVM(vmId);
  }, []);

  // VM 구독 해제
  const unsubscribeFromVM = useCallback((vmId: string) => {
    webSocketService.unsubscribeFromVM(vmId);
  }, []);

  // Provider 구독
  const subscribeToProvider = useCallback((provider: string) => {
    webSocketService.subscribeToProvider(provider);
  }, []);

  // Provider 구독 해제
  const unsubscribeFromProvider = useCallback((provider: string) => {
    webSocketService.unsubscribeFromProvider(provider);
  }, []);

  // 연결 상태 확인
  const isConnected = useCallback(() => {
    return webSocketService.isConnected();
  }, []);

  // 기본 알림 설정
  useEffect(() => {
    const handleSystemNotification = (data: SystemNotification) => {
      if (data.type === 'error') {
        error(data.message);
      } else {
        success(data.message);
      }
    };

    const handleSystemAlert = (data: SystemAlert) => {
      const alertMessage = `[${data.level.toUpperCase()}] ${data.message}`;
      if (data.level === 'critical' || data.level === 'high') {
        error(alertMessage);
      } else {
        success(alertMessage);
      }
    };

    onSystemNotification(handleSystemNotification);
    onSystemAlert(handleSystemAlert);

    return () => {
      webSocketService.offSystemNotification(handleSystemNotification);
      webSocketService.offSystemAlert(handleSystemAlert);
    };
  }, [success, error, onSystemNotification, onSystemAlert]);

  // 컴포넌트 언마운트 시 이벤트 리스너 정리
  useEffect(() => {
    const callbacks = callbacksRef.current;
    return () => {
      if (callbacks.onVMStatusUpdate) {
        webSocketService.offVMStatusUpdate(callbacks.onVMStatusUpdate);
      }
      if (callbacks.onVMResourceUpdate) {
        webSocketService.offVMResourceUpdate(callbacks.onVMResourceUpdate);
      }
      if (callbacks.onVMError) {
        webSocketService.offVMError(callbacks.onVMError);
      }
      if (callbacks.onProviderStatusUpdate) {
        webSocketService.offProviderStatusUpdate(callbacks.onProviderStatusUpdate);
      }
      if (callbacks.onProviderInstanceUpdate) {
        webSocketService.offProviderInstanceUpdate(callbacks.onProviderInstanceUpdate);
      }
    };
  }, []);

  return {
    onVMStatusUpdate,
    onVMResourceUpdate,
    onVMError,
    onProviderStatusUpdate,
    onProviderInstanceUpdate,
    onSystemNotification,
    onSystemAlert,
    subscribeToVM,
    unsubscribeFromVM,
    subscribeToProvider,
    unsubscribeFromProvider,
    isConnected,
  };
}
