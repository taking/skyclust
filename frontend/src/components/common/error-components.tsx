/**
 * Error Components
 * 다양한 에러 상황을 표시하는 컴포넌트 모음
 */

'use client';

import { ReactNode } from 'react';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { 
  AlertCircle, 
  AlertTriangle, 
  RefreshCw, 
  WifiOff, 
  ServerOff,
  FileX,
  ShieldAlert,
  X,
  Home,
} from 'lucide-react';
import { getUserFriendlyErrorMessage, NetworkError, ServerError, isRetryableError } from '@/lib/error-handler';
import { cn } from '@/lib/utils';

export interface InlineErrorProps {
  /**
   * 에러 객체
   */
  error: unknown;
  
  /**
   * 제목
   */
  title?: string;
  
  /**
   * 재시도 함수
   */
  onRetry?: () => void | Promise<void>;
  
  /**
   * 닫기 함수
   */
  onDismiss?: () => void;
  
  /**
   * 재시도 중 여부
   */
  isRetrying?: boolean;
  
  /**
   * 추가 클래스명
   */
  className?: string;
  
  /**
   * 컴팩트 모드 (작은 크기)
   */
  compact?: boolean;
  
  /**
   * 아이콘 표시 여부
   */
  showIcon?: boolean;
}

/**
 * InlineError Component
 * 인라인으로 사용할 수 있는 에러 메시지 컴포넌트
 */
export function InlineError({
  error,
  title,
  onRetry,
  onDismiss,
  isRetrying = false,
  className,
  compact = false,
  showIcon = true,
}: InlineErrorProps) {
  // 1. 사용자 친화적인 에러 메시지 가져오기
  const message = getUserFriendlyErrorMessage(error);
  
  // 2. 재시도 가능 여부 확인: onRetry 함수가 있고 에러가 재시도 가능한 타입인지 확인
  const retryable = onRetry && isRetryableError(error);
  
  // 3. 에러 타입 확인
  const isNetworkError = error instanceof NetworkError;
  const isServerError = error instanceof ServerError;

  /**
   * 에러 타입에 따른 아이콘 반환
   * 네트워크 에러: WifiOff, 서버 에러: ServerOff, 기타: AlertCircle
   */
  const getIcon = () => {
    if (isNetworkError) return <WifiOff className="h-4 w-4" />;
    if (isServerError) return <ServerOff className="h-4 w-4" />;
    return <AlertCircle className="h-4 w-4" />;
  };

  /**
   * 에러 타입에 따른 Alert variant 반환
   * 네트워크 에러: default, 서버/기타 에러: destructive
   */
  const getVariant = () => {
    if (isNetworkError) return 'default';
    if (isServerError) return 'destructive';
    return 'destructive';
  };

  return (
    <Alert 
      variant={getVariant()}
      className={cn(
        compact && 'py-2',
        className
      )}
    >
      <div className="flex items-start gap-3">
        {/* 4. 아이콘 표시 (옵션이 활성화된 경우) */}
        {showIcon && (
          <div className="flex-shrink-0 mt-0.5">
            {getIcon()}
          </div>
        )}
        <div className="flex-1 min-w-0">
          {/* 5. 제목 표시 (제공된 경우) */}
          {title && (
            <AlertTitle className={compact ? 'text-sm' : ''}>
              {title}
            </AlertTitle>
          )}
          {/* 6. 에러 메시지 표시 */}
          <AlertDescription className={cn(
            'text-sm',
            compact && 'text-xs'
          )}>
            {message}
          </AlertDescription>
          {/* 7. 액션 버튼: 재시도 또는 닫기 */}
          {(retryable || onDismiss) && (
            <div className={cn(
              'flex items-center gap-2 mt-3',
              compact && 'mt-2'
            )}>
              {/* 7-1. 재시도 버튼 (재시도 가능한 경우) */}
              {retryable && (
                <Button
                  variant="outline"
                  size={compact ? 'sm' : 'default'}
                  onClick={onRetry}
                  disabled={isRetrying}
                  className="h-auto py-1.5"
                >
                  <RefreshCw className={cn(
                    'mr-2 h-3 w-3',
                    isRetrying && 'animate-spin'
                  )} />
                  {isRetrying ? 'Retrying...' : 'Retry'}
                </Button>
              )}
              {/* 7-2. 닫기 버튼 (onDismiss가 제공된 경우) */}
              {onDismiss && (
                <Button
                  variant="ghost"
                  size={compact ? 'sm' : 'default'}
                  onClick={onDismiss}
                  className="h-auto py-1.5"
                >
                  <X className="h-3 w-3" />
                </Button>
              )}
            </div>
          )}
        </div>
      </div>
    </Alert>
  );
}

/**
 * ErrorCard Component
 * 카드 형태의 에러 표시 컴포넌트
 */
export interface ErrorCardProps {
  error: unknown;
  title?: string;
  description?: string;
  onRetry?: () => void | Promise<void>;
  onDismiss?: () => void;
  isRetrying?: boolean;
  className?: string;
}

export function ErrorCard({
  error,
  title,
  description,
  onRetry,
  onDismiss,
  isRetrying = false,
  className,
}: ErrorCardProps) {
  // 1. 사용자 친화적인 에러 메시지 가져오기
  const message = getUserFriendlyErrorMessage(error);
  
  // 2. 재시도 가능 여부 확인
  const retryable = onRetry && isRetryableError(error);
  
  // 3. 에러 타입 확인
  const isNetworkError = error instanceof NetworkError;
  const isServerError = error instanceof ServerError;

  /**
   * 에러 타입에 따른 아이콘 반환 (카드용 큰 아이콘)
   * 네트워크 에러: WifiOff (노란색), 서버 에러: ServerOff (빨간색), 기타: AlertTriangle (빨간색)
   */
  const getIcon = () => {
    if (isNetworkError) return <WifiOff className="h-8 w-8 text-yellow-500" />;
    if (isServerError) return <ServerOff className="h-8 w-8 text-red-500" />;
    return <AlertTriangle className="h-8 w-8 text-red-500" />;
  };

  return (
    <Card className={className}>
      <CardHeader>
        <div className="flex items-center gap-4">
          {/* 4. 에러 아이콘 표시 */}
          {getIcon()}
          <div className="flex-1">
            {/* 5. 제목: 제공된 title이 있으면 사용, 없으면 에러 타입에 따라 기본 제목 */}
            <CardTitle>
              {title || (isNetworkError ? 'Connection Error' : 'Error')}
            </CardTitle>
            {/* 6. 설명 (제공된 경우) */}
            {description && (
              <CardDescription className="mt-1">
                {description}
              </CardDescription>
            )}
          </div>
          {/* 7. 닫기 버튼 (onDismiss가 제공된 경우) */}
          {onDismiss && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onDismiss}
              className="flex-shrink-0"
            >
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* 8. 에러 메시지 표시 (배경색이 있는 박스) */}
        <div className="bg-muted rounded-lg p-3">
          <p className="text-sm">{message}</p>
        </div>
        
        {/* 9. 재시도 버튼 (재시도 가능한 경우) */}
        {retryable && (
          <Button
            onClick={onRetry}
            disabled={isRetrying}
            className="w-full"
          >
            <RefreshCw className={cn(
              'mr-2 h-4 w-4',
              isRetrying && 'animate-spin'
            )} />
            {isRetrying ? 'Retrying...' : 'Try again'}
          </Button>
        )}
      </CardContent>
    </Card>
  );
}

/**
 * ErrorPage Component
 * 전체 페이지용 에러 컴포넌트
 */
export interface ErrorPageProps {
  error: unknown;
  title?: string;
  description?: string;
  actions?: ReactNode;
  className?: string;
}

export function ErrorPage({
  error,
  title,
  description,
  actions,
  className,
}: ErrorPageProps) {
  const message = getUserFriendlyErrorMessage(error);
  const isNetworkError = error instanceof NetworkError;
  const isServerError = error instanceof ServerError;

  const getIcon = () => {
    if (isNetworkError) return <WifiOff className="h-16 w-16 text-yellow-500" />;
    if (isServerError) return <ServerOff className="h-16 w-16 text-red-500" />;
    return <AlertTriangle className="h-16 w-16 text-red-500" />;
  };

  const getTitle = () => {
    if (title) return title;
    if (isNetworkError) return 'Connection Error';
    if (isServerError) return 'Server Error';
    return 'Something went wrong';
  };

  const getDescription = () => {
    if (description) return description;
    if (isNetworkError) {
      return 'Unable to connect to the server. Please check your internet connection.';
    }
    if (isServerError) {
      return 'The server encountered an error. Please try again later.';
    }
    return 'An unexpected error occurred. Please try again or contact support.';
  };

  return (
    <div className={cn(
      'min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 p-4',
      className
    )}>
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4">
            {getIcon()}
          </div>
          <CardTitle className="text-2xl">{getTitle()}</CardTitle>
          <CardDescription className="mt-2">
            {getDescription()}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="bg-muted rounded-lg p-4">
            <p className="text-sm font-medium">{message}</p>
          </div>
          
          {actions && (
            <div className="flex flex-col gap-2">
              {actions}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

/**
 * NotFound Component
 * 404 에러 표시 컴포넌트
 */
export interface NotFoundProps {
  resource?: string;
  onGoBack?: () => void;
  className?: string;
}

export function NotFound({
  resource = 'page',
  onGoBack,
  className,
}: NotFoundProps) {
  return (
    <ErrorPage
      error={new Error(`The ${resource} you're looking for doesn't exist.`)}
      title={`${resource.charAt(0).toUpperCase() + resource.slice(1)} Not Found`}
      description={`The ${resource} you requested could not be found. It may have been moved or deleted.`}
      actions={
        <>
          {onGoBack && (
            <Button onClick={onGoBack} variant="outline" className="w-full">
              Go back
            </Button>
          )}
          <Button 
            onClick={() => window.location.href = '/'}
            className="w-full"
          >
            <Home className="mr-2 h-4 w-4" />
            Go home
          </Button>
        </>
      }
      className={className}
    />
  );
}

/**
 * Unauthorized Component
 * 401/403 에러 표시 컴포넌트
 */
export interface UnauthorizedProps {
  onLogin?: () => void;
  className?: string;
}

export function Unauthorized({
  onLogin,
  className,
}: UnauthorizedProps) {
  return (
    <ErrorPage
      error={new ServerError('Unauthorized', 401)}
      title="Access Denied"
      description="You don't have permission to access this resource. Please log in with an authorized account."
      actions={
        <>
          {onLogin && (
            <Button onClick={onLogin} className="w-full">
              <ShieldAlert className="mr-2 h-4 w-4" />
              Log in
            </Button>
          )}
          <Button 
            onClick={() => window.location.href = '/'}
            variant="outline"
            className="w-full"
          >
            <Home className="mr-2 h-4 w-4" />
            Go home
          </Button>
        </>
      }
      className={className}
    />
  );
}

