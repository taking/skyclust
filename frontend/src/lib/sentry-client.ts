/**
 * Sentry Client
 * Sentry 에러 트래킹 서비스 클라이언트
 */

interface SentryClient {
  captureException: (error: Error, context?: Record<string, unknown>) => void;
  captureMessage: (message: string, level?: 'error' | 'warning' | 'info') => void;
  setUser: (user: { id?: string; email?: string; username?: string }) => void;
  setContext: (name: string, context: Record<string, unknown>) => void;
  addBreadcrumb: (breadcrumb: { message: string; level?: 'error' | 'warning' | 'info'; category?: string }) => void;
}

let sentryClient: SentryClient | null = null;

/**
 * Sentry 클라이언트 초기화
 */
export function initSentry(): void {
  if (typeof window === 'undefined') return;

  // 환경 변수에서 DSN 확인
  const dsn = process.env.NEXT_PUBLIC_SENTRY_DSN;
  
  if (!dsn) {
    if (process.env.NODE_ENV === 'development') {
      console.log('[Sentry] DSN not configured. Error tracking disabled.');
    }
    return;
  }

  // Dynamic import로 Sentry 로드 (번들 크기 최적화)
  import('@sentry/nextjs').then((Sentry) => {
    try {
      Sentry.init({
        dsn,
        environment: process.env.NODE_ENV || 'development',
        tracesSampleRate: process.env.NODE_ENV === 'production' ? 0.1 : 1.0,
        beforeSend(event, hint) {
          // 개발 환경에서는 콘솔에 로그
          if (process.env.NODE_ENV === 'development') {
            console.log('[Sentry] Error captured:', event, hint);
          }
          return event;
        },
        ignoreErrors: [
          // 네트워크 에러는 무시 (너무 많을 수 있음)
          'NetworkError',
          'Network request failed',
          'Failed to fetch',
          // 브라우저 확장 프로그램 에러 무시
          'ResizeObserver loop limit exceeded',
          'Non-Error promise rejection captured',
        ],
      });

      sentryClient = {
        captureException: (error: Error, context?: Record<string, unknown>) => {
          Sentry.withScope((scope) => {
            if (context) {
              Object.entries(context).forEach(([key, value]) => {
                scope.setContext(key, { value });
              });
            }
            Sentry.captureException(error);
          });
        },
        captureMessage: (message: string, level: 'error' | 'warning' | 'info' = 'error') => {
          Sentry.captureMessage(message, level);
        },
        setUser: (user: { id?: string; email?: string; username?: string }) => {
          Sentry.setUser(user);
        },
        setContext: (name: string, context: Record<string, unknown>) => {
          Sentry.setContext(name, context);
        },
        addBreadcrumb: (breadcrumb: { message: string; level?: 'error' | 'warning' | 'info'; category?: string }) => {
          Sentry.addBreadcrumb({
            message: breadcrumb.message,
            level: breadcrumb.level || 'info',
            category: breadcrumb.category || 'custom',
          });
        },
      };

      if (process.env.NODE_ENV === 'development') {
        console.log('[Sentry] Initialized successfully');
      }
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.error('[Sentry] Failed to initialize:', error);
      }
    }
  }).catch((error) => {
    if (process.env.NODE_ENV === 'development') {
      console.error('[Sentry] Failed to load SDK:', error);
    }
  });
}

/**
 * Sentry 클라이언트 가져오기
 */
export function getSentryClient(): SentryClient | null {
  return sentryClient;
}

/**
 * Sentry가 활성화되어 있는지 확인
 */
export function isSentryEnabled(): boolean {
  return sentryClient !== null && !!process.env.NEXT_PUBLIC_SENTRY_DSN;
}

