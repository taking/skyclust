/**
 * SSE Error Recovery Hook
 * 
 * SSE 연결 실패 시 자동 재연결 및 사용자 알림을 제공하는 훅
 */

'use client';

import { useEffect, useRef } from 'react';
import { sseService } from '@/services/sse';
import { useSSEStatus } from './use-sse-status';
import { useToast } from '@/hooks/use-toast';
import { useTranslation } from '@/hooks/use-translation';
import { log } from '@/lib/logging';
import { useAuthStore } from '@/store/auth';
import { STORAGE_KEYS } from '@/lib/constants/values';

export interface UseSSEErrorRecoveryOptions {
  /**
   * 자동 재연결 활성화 여부
   */
  autoReconnect?: boolean;
  
  /**
   * 사용자 알림 표시 여부
   */
  showNotifications?: boolean;
  
  /**
   * 최대 재연결 시도 횟수
   */
  maxReconnectAttempts?: number;
}

/**
 * useSSEErrorRecovery Hook
 * 
 * SSE 연결 실패 시 자동 재연결 및 사용자 알림을 제공합니다.
 * 
 * @example
 * ```tsx
 * useSSEErrorRecovery({
 *   autoReconnect: true,
 *   showNotifications: true,
 * });
 * ```
 */
export function useSSEErrorRecovery({
  autoReconnect = true,
  showNotifications = true,
  maxReconnectAttempts = 5,
}: UseSSEErrorRecoveryOptions = {}): void {
  const { status: sseStatus } = useSSEStatus();
  const { error: showErrorToast, success: showSuccessToast } = useToast();
  const { t } = useTranslation();
  const { token } = useAuthStore();
  const reconnectAttemptsRef = useRef(0);
  const lastErrorTimeRef = useRef<number | null>(null);
  const wasConnectedRef = useRef(false);

  useEffect(() => {
    wasConnectedRef.current = sseStatus.isConnected;
  }, [sseStatus.isConnected]);

  useEffect(() => {
    if (!autoReconnect || !token) {
      return;
    }

    const callbacks = sseService.getCallbacks();
    
    const originalOnError = callbacks.onError;
    const originalOnConnected = callbacks.onConnected;

    const errorRecoveryCallbacks = {
      onError: (error: unknown) => {
        if (originalOnError) {
          originalOnError(error);
        }

        const now = Date.now();
        
        if (lastErrorTimeRef.current && now - lastErrorTimeRef.current < 5000) {
          return;
        }
        
        lastErrorTimeRef.current = now;
        reconnectAttemptsRef.current++;

        if (showNotifications) {
          const errorMessage = t('sse.connectionError') || 'SSE connection error. Attempting to reconnect...';
          showErrorToast(errorMessage);
        }

        log.warn('[SSE Error Recovery] Connection error detected', error, {
          reconnectAttempt: reconnectAttemptsRef.current,
          maxAttempts: maxReconnectAttempts,
        });

        if (reconnectAttemptsRef.current <= maxReconnectAttempts) {
          setTimeout(() => {
            if (!sseService.isConnected() && token) {
              log.info('[SSE Error Recovery] Attempting to reconnect', {
                attempt: reconnectAttemptsRef.current,
                maxAttempts: maxReconnectAttempts,
              });
              
              sseService.connect(token, callbacks).catch((reconnectError) => {
                log.error('[SSE Error Recovery] Reconnection failed', reconnectError, {
                  attempt: reconnectAttemptsRef.current,
                });
              });
            }
          }, Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current - 1), 30000));
        } else {
          if (showNotifications) {
            const maxAttemptsMessage = t('sse.maxReconnectAttemptsReached') || 
              'Maximum reconnection attempts reached. Please refresh the page.';
            showErrorToast(maxAttemptsMessage);
          }
          
          log.error('[SSE Error Recovery] Maximum reconnection attempts reached', {
            attempts: reconnectAttemptsRef.current,
            maxAttempts: maxReconnectAttempts,
          });
        }
      },
      
      onConnected: (data: unknown) => {
        if (originalOnConnected) {
          originalOnConnected(data);
        }

        if (reconnectAttemptsRef.current > 0) {
          reconnectAttemptsRef.current = 0;
          lastErrorTimeRef.current = null;
          
          if (showNotifications && wasConnectedRef.current) {
            const successMessage = t('sse.reconnected') || 'SSE connection restored.';
            showSuccessToast(successMessage);
          }
          
          log.info('[SSE Error Recovery] Connection restored', {
            previousAttempts: reconnectAttemptsRef.current,
          });
        }
      },
    };

    sseService.updateCallbacks(errorRecoveryCallbacks);

    return () => {
      reconnectAttemptsRef.current = 0;
      lastErrorTimeRef.current = null;
    };
  }, [autoReconnect, token, showNotifications, maxReconnectAttempts, t, showErrorToast, showSuccessToast, sseStatus.isConnected]);
}

