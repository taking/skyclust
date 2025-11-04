/**
 * Error Boundary Components
 * 통합된 에러 경계 컴포넌트
 * 
 * 다양한 에러 타입을 처리하고 사용자에게 친화적인 에러 UI를 제공합니다.
 */

'use client';

import { ErrorBoundary, ErrorBoundaryPropsWithFallback } from 'react-error-boundary';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { AlertTriangle, RefreshCw, Home, Bug, Mail } from 'lucide-react';
import { getUserFriendlyErrorMessage, NetworkError, ServerError } from '@/lib/error-handler';
import { useState } from 'react';

export interface ErrorFallbackProps {
  error: Error;
  resetErrorBoundary: () => void;
  /**
   * 에러 타입에 따른 추가 정보
   */
  errorType?: 'network' | 'server' | 'client' | 'unknown';
  /**
   * 에러 로그 ID (로깅 서비스에서 사용)
   */
  errorId?: string;
}

/**
 * Error Fallback Component
 * 에러 발생 시 표시되는 폴백 UI
 */
export function ErrorFallback({ 
  error, 
  resetErrorBoundary,
  errorType = 'unknown',
  errorId,
}: ErrorFallbackProps) {
  const [showDetails, setShowDetails] = useState(false);
  
  const friendlyMessage = getUserFriendlyErrorMessage(error);
  const isNetworkError = error instanceof NetworkError || errorType === 'network';
  const isServerError = error instanceof ServerError || errorType === 'server';

  // 에러 타입별 아이콘 및 색상
  const getErrorIcon = () => {
    if (isNetworkError) {
      return <AlertTriangle className="h-12 w-12 text-yellow-500" />;
    }
    if (isServerError) {
      return <AlertTriangle className="h-12 w-12 text-red-500" />;
    }
    return <Bug className="h-12 w-12 text-red-500" />;
  };

  // 에러 제목
  const getErrorTitle = () => {
    if (isNetworkError) {
      return 'Connection Error';
    }
    if (isServerError) {
      return 'Server Error';
    }
    return 'Something went wrong';
  };

  // 에러 설명
  const getErrorDescription = () => {
    if (isNetworkError) {
      return 'Unable to connect to the server. Please check your internet connection and try again.';
    }
    if (isServerError) {
      return 'The server encountered an error while processing your request. Please try again later or contact support if the problem persists.';
    }
    return 'An unexpected error occurred. Please try again or contact support if the problem persists.';
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
      <Card className="w-full max-w-lg">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4">
            {getErrorIcon()}
          </div>
          <CardTitle className="text-xl">{getErrorTitle()}</CardTitle>
          <CardDescription className="mt-2">
            {getErrorDescription()}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* 사용자 친화적 에러 메시지 */}
          <div className="bg-gray-50 dark:bg-gray-900 rounded-lg p-4">
            <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
              {friendlyMessage}
            </p>
          </div>

          {/* 에러 상세 정보 */}
          <details className="text-sm">
            <summary 
              className="cursor-pointer font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100"
              onClick={() => setShowDetails(!showDetails)}
            >
              {showDetails ? 'Hide' : 'Show'} technical details
            </summary>
            <div className="mt-2 p-3 bg-gray-100 dark:bg-gray-800 rounded text-xs overflow-auto max-h-64">
              <div className="space-y-2">
                <div>
                  <strong>Error Message:</strong>
                  <pre className="mt-1 text-gray-700 dark:text-gray-300">{error.message}</pre>
                </div>
                {error.stack && (
                  <div>
                    <strong>Stack Trace:</strong>
                    <pre className="mt-1 text-gray-600 dark:text-gray-400 whitespace-pre-wrap break-words">
                      {error.stack}
                    </pre>
                  </div>
                )}
                {errorId && (
                  <div>
                    <strong>Error ID:</strong>
                    <code className="ml-1 text-xs">{errorId}</code>
                  </div>
                )}
                {error.name && (
                  <div>
                    <strong>Error Type:</strong>
                    <code className="ml-1 text-xs">{error.name}</code>
                  </div>
                )}
              </div>
            </div>
          </details>

          {/* 액션 버튼 */}
          <div className="flex flex-col sm:flex-row gap-2 pt-2">
            <Button 
              onClick={resetErrorBoundary} 
              className="flex-1"
              disabled={isNetworkError}
            >
              <RefreshCw className="mr-2 h-4 w-4" />
              Try again
            </Button>
            <Button 
              variant="outline" 
              onClick={() => window.location.href = '/'}
              className="flex-1"
            >
              <Home className="mr-2 h-4 w-4" />
              Go home
            </Button>
            <Button
              variant="outline"
              onClick={() => {
                const subject = encodeURIComponent(`Error Report - ${errorId || 'Unknown'}`);
                const body = encodeURIComponent(
                  `Error Details:\n\n` +
                  `Message: ${error.message}\n` +
                  `Type: ${error.name}\n` +
                  `Error ID: ${errorId || 'N/A'}\n\n` +
                  `Stack Trace:\n${error.stack || 'N/A'}`
                );
                window.location.href = `mailto:support@skyclust.com?subject=${subject}&body=${body}`;
              }}
              className="flex-1"
            >
              <Mail className="mr-2 h-4 w-4" />
              Report
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export interface AppErrorBoundaryProps {
  children: React.ReactNode;
  /**
   * 에러 로깅 함수
   */
  onError?: (error: Error, errorInfo: React.ErrorInfo) => void;
  /**
   * 에러 발생 시 실행할 추가 로직
   */
  fallback?: React.ComponentType<{ error: Error; resetErrorBoundary: () => void }>;
  /**
   * 리셋 키 (컴포넌트를 강제로 리마운트)
   */
  resetKeys?: ErrorBoundaryPropsWithFallback['resetKeys'];
}

/**
 * App Error Boundary
 * 애플리케이션 전역 에러 경계
 * 
 * 예상치 못한 에러를 잡아서 사용자에게 친화적인 에러 UI를 표시합니다.
 */
export function AppErrorBoundary({ 
  children,
  onError,
  fallback,
  resetKeys,
}: AppErrorBoundaryProps) {
  const handleError = (error: Error, errorInfo: React.ErrorInfo) => {
    // Development 환경에서만 콘솔 로깅
    if (process.env.NODE_ENV === 'development') {
      console.error('Error caught by boundary:', error, errorInfo);
    }
    
    // 에러 타입 판단
    const errorType = error instanceof NetworkError 
      ? 'network' 
      : error instanceof ServerError 
      ? 'server'
      : 'unknown';
    
    // 에러 ID 생성 (로깅 서비스 연동 시 사용)
    const errorId = `err-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    
    // ErrorHandler를 통한 로깅 (Sentry 포함)
    const { ErrorHandler } = require('@/lib/error-handler');
    ErrorHandler.logError(error, {
      componentStack: errorInfo.componentStack,
      errorType,
      errorId,
    });
    
    // 사용자 정의 로깅 함수 호출
    if (onError) {
      onError(error, errorInfo);
    }
  };

  const DefaultErrorFallback = (props: { error: Error; resetErrorBoundary: () => void }) => {
    const { error, resetErrorBoundary } = props;
    const errorType = error instanceof NetworkError 
      ? 'network' 
      : error instanceof ServerError 
      ? 'server'
      : 'unknown';
    
    const errorId = `err-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    
    return (
      <ErrorFallback
        error={error}
        resetErrorBoundary={resetErrorBoundary}
        errorType={errorType}
        errorId={errorId}
      />
    );
  };
  DefaultErrorFallback.displayName = 'DefaultErrorFallback';

  const FallbackComponent = fallback || DefaultErrorFallback;

  return (
    <ErrorBoundary
      FallbackComponent={FallbackComponent}
      onError={handleError}
      resetKeys={resetKeys}
    >
      {children}
    </ErrorBoundary>
  );
}

/**
 * ErrorBoundaryWithFallback
 * 커스텀 폴백을 사용할 수 있는 에러 경계
 */
export interface ErrorBoundaryWithFallbackProps {
  children: React.ReactNode;
  fallback: ErrorBoundaryPropsWithFallback['fallback'];
  onError?: (error: Error, errorInfo: React.ErrorInfo) => void;
  resetKeys?: ErrorBoundaryPropsWithFallback['resetKeys'];
  onReset?: () => void;
}

export function ErrorBoundaryWithFallback({
  children,
  fallback,
  onError,
  resetKeys,
  onReset,
}: ErrorBoundaryWithFallbackProps) {
  return (
    <ErrorBoundary
      fallback={fallback}
      onError={onError}
      resetKeys={resetKeys}
      onReset={onReset}
    >
      {children}
    </ErrorBoundary>
  );
}
