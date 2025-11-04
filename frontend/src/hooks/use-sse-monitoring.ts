import { useEffect, useCallback, useRef } from 'react';
import { useAuthStore } from '@/store/auth';
import { sseService, SSECallbacks } from '@/services/sse';
import { useToast } from '@/hooks/use-toast';

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

export function useSSEMonitoring() {
  const { token } = useAuthStore();
  const { success, error } = useToast();
  const callbacksRef = useRef<SSECallbacks>({});

  // SSE 연결
  useEffect(() => {
    if (token) {
      sseService.connect(token, callbacksRef.current);
    }

    return () => {
      sseService.disconnect();
    };
  }, [token]);

  // 자동 구독 설정
  useEffect(() => {
    if (sseService.isConnected()) {
      // 기본 이벤트 구독
      sseService.subscribeToEvent('vm-status');
      sseService.subscribeToEvent('vm-resource');
      sseService.subscribeToEvent('provider-status');
      sseService.subscribeToEvent('provider-instance');
      sseService.subscribeToEvent('system-notification');
      sseService.subscribeToEvent('system-alert');
    }
  }, [token]);

  // VM 상태 업데이트 콜백 등록
  const onVMStatusUpdate = useCallback((callback: (data: VMStatusUpdate) => void) => {
    callbacksRef.current.onVMStatusUpdate = (data: unknown) => {
      callback(data as VMStatusUpdate);
    };
  }, []);

  // VM 리소스 업데이트 콜백 등록
  const onVMResourceUpdate = useCallback((callback: (data: VMResourceUpdate) => void) => {
    callbacksRef.current.onVMResourceUpdate = (data: unknown) => {
      callback(data as VMResourceUpdate);
    };
  }, []);

  // Provider 상태 업데이트 콜백 등록
  const onProviderStatusUpdate = useCallback((callback: (data: ProviderStatusUpdate) => void) => {
    callbacksRef.current.onProviderStatusUpdate = (data: unknown) => {
      callback(data as ProviderStatusUpdate);
    };
  }, []);

  // Provider 인스턴스 업데이트 콜백 등록
  const onProviderInstanceUpdate = useCallback((callback: (data: ProviderInstanceUpdate) => void) => {
    callbacksRef.current.onProviderInstanceUpdate = (data: unknown) => {
      callback(data as ProviderInstanceUpdate);
    };
  }, []);

  // 시스템 알림 콜백 등록
  const onSystemNotification = useCallback((callback: (data: SystemNotification) => void) => {
    callbacksRef.current.onSystemNotification = (data: unknown) => {
      callback(data as SystemNotification);
    };
  }, []);

  // 시스템 알림 콜백 등록
  const onSystemAlert = useCallback((callback: (data: SystemAlert) => void) => {
    callbacksRef.current.onSystemAlert = (data: unknown) => {
      callback(data as SystemAlert);
    };
  }, []);

  // 연결 상태 콜백 등록
  const onConnected = useCallback((callback: (data: unknown) => void) => {
    callbacksRef.current.onConnected = callback;
  }, []);

  // 에러 콜백 등록
  const onError = useCallback((callback: (error: Event) => void) => {
    callbacksRef.current.onError = callback;
  }, []);

  // 연결 상태 확인
  const isConnected = useCallback(() => {
    return sseService.isConnected();
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
  }, [success, error, onSystemNotification, onSystemAlert]);

  // VM 구독 관리
  const subscribeToVM = useCallback((vmId: string) => {
    sseService.subscribeToVM(vmId);
  }, []);

  const unsubscribeFromVM = useCallback((vmId: string) => {
    sseService.unsubscribeFromVM(vmId);
  }, []);

  // Provider 구독 관리
  const subscribeToProvider = useCallback((provider: string) => {
    sseService.subscribeToProvider(provider);
  }, []);

  const unsubscribeFromProvider = useCallback((provider: string) => {
    sseService.unsubscribeFromProvider(provider);
  }, []);

  // 이벤트 구독 관리
  const subscribeToEvent = useCallback((eventType: string) => {
    sseService.subscribeToEvent(eventType);
  }, []);

  const unsubscribeFromEvent = useCallback((eventType: string) => {
    sseService.unsubscribeFromEvent(eventType);
  }, []);

  // 구독 상태 조회
  const getSubscriptions = useCallback(() => {
    return sseService.getSubscriptions();
  }, []);

  return {
    onVMStatusUpdate,
    onVMResourceUpdate,
    onProviderStatusUpdate,
    onProviderInstanceUpdate,
    onSystemNotification,
    onSystemAlert,
    onConnected,
    onError,
    isConnected,
    subscribeToVM,
    unsubscribeFromVM,
    subscribeToProvider,
    unsubscribeFromProvider,
    subscribeToEvent,
    unsubscribeFromEvent,
    getSubscriptions,
  };
}

