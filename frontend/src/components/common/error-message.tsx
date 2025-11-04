/**
 * Error Message Component
 * 에러 메시지를 표시하는 컴포넌트
 * 
 * @deprecated InlineError 컴포넌트를 사용하는 것을 권장합니다.
 * 이 컴포넌트는 하위 호환성을 위해 유지됩니다.
 */

'use client';

import { useState } from 'react';
import { AlertCircle, RefreshCw, WifiOff } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { getUserFriendlyErrorMessage, isRetryableError, isOffline, NetworkError } from '@/lib/error-handler';
import { InlineError } from './error-components';

export interface ErrorMessageProps {
  error: unknown;
  title?: string;
  onRetry?: () => void | Promise<void>;
  className?: string;
  /**
   * 닫기 함수
   */
  onDismiss?: () => void;
}

/**
 * ErrorMessage Component
 * 
 * @deprecated InlineError를 사용하세요.
 */
export function ErrorMessage({ 
  error, 
  title, 
  onRetry, 
  className,
  onDismiss,
}: ErrorMessageProps) {
  // InlineError를 사용하여 구현
  return (
    <InlineError
      error={error}
      title={title}
      onRetry={onRetry}
      onDismiss={onDismiss}
      className={className}
    />
  );
}

