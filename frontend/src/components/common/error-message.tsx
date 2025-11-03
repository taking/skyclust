'use client';

import { AlertCircle, RefreshCw, WifiOff } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { getUserFriendlyErrorMessage, isRetryableError, isOffline, NetworkError } from '@/lib/error-handler';
import { useState } from 'react';

interface ErrorMessageProps {
  error: unknown;
  title?: string;
  onRetry?: () => void;
  className?: string;
}

export function ErrorMessage({ error, title, onRetry, className }: ErrorMessageProps) {
  const [isRetrying, setIsRetrying] = useState(false);

  const message = getUserFriendlyErrorMessage(error);
  const retryable = onRetry && isRetryableError(error);
  const offline = isOffline();

  const handleRetry = async () => {
    if (!onRetry) return;
    
    setIsRetrying(true);
    try {
      await onRetry();
    } finally {
      setIsRetrying(false);
    }
  };

  return (
    <Alert variant="destructive" className={className}>
      <div className="flex items-start gap-3">
        {offline || error instanceof NetworkError ? (
          <WifiOff className="h-5 w-5" />
        ) : (
          <AlertCircle className="h-5 w-5" />
        )}
        <div className="flex-1">
          {title && <AlertTitle>{title}</AlertTitle>}
          <AlertDescription>
            {message}
            {offline && (
              <span className="block mt-2">
                Please check your internet connection.
              </span>
            )}
          </AlertDescription>
          {retryable && (
            <div className="mt-4">
              <Button
                variant="outline"
                size="sm"
                onClick={handleRetry}
                disabled={isRetrying || offline}
              >
                <RefreshCw className={`mr-2 h-4 w-4 ${isRetrying ? 'animate-spin' : ''}`} />
                {isRetrying ? 'Retrying...' : 'Retry'}
              </Button>
            </div>
          )}
        </div>
      </div>
    </Alert>
  );
}

