import { useEffect, useCallback, useRef } from 'react';
import { useAuthStore } from '@/store/auth';
import { sseService } from '@/services/sse';
import { useToast } from '@/hooks/use-toast';
import type { SSECallbacks } from '@/lib/types/sse';

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

  // VM 에러 콜백 등록 (SSE는 시스템 에러를 onError로 처리)
  const onVMError = useCallback((callback: (data: VMError) => void) => {
    callbacksRef.current.onError = (event: Event) => {
      try {
        const errorEvent = event as unknown as { data?: string };
        if (errorEvent.data) {
          const errorData = JSON.parse(errorEvent.data) as VMError;
          callback(errorData);
        }
      } catch {
        // Error parsing failed, ignore
      }
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

  // VM 구독
  const subscribeToVM = useCallback((vmId: string) => {
    sseService.subscribeToVM(vmId);
  }, []);

  // VM 구독 해제
  const unsubscribeFromVM = useCallback((vmId: string) => {
    sseService.unsubscribeFromVM(vmId);
  }, []);

  // Provider 구독
  const subscribeToProvider = useCallback((provider: string) => {
    sseService.subscribeToEvent(`provider-${provider}`);
  }, []);

  // Provider 구독 해제
  const unsubscribeFromProvider = useCallback((provider: string) => {
    sseService.unsubscribeFromEvent(`provider-${provider}`);
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
