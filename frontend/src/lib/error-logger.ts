/**
 * Error Logger
 * 에러 로깅 유틸리티
 * 
 * 에러를 구조화하여 로깅하고, 필요시 외부 서비스로 전송
 * 
 * Note: 순환 참조를 피하기 위해 logger를 import하지 않고 직접 console을 사용합니다.
 */

export interface ErrorLog {
  /**
   * 에러 ID
   */
  id: string;
  
  /**
   * 에러 메시지
   */
  message: string;
  
  /**
   * 에러 스택
   */
  stack?: string;
  
  /**
   * 에러 타입
   */
  type: string;
  
  /**
   * 에러 코드
   */
  code?: string;
  
  /**
   * HTTP 상태 코드
   */
  statusCode?: number;
  
  /**
   * 타임스탬프
   */
  timestamp: string;
  
  /**
   * 사용자 에이전트
   */
  userAgent?: string;
  
  /**
   * 현재 URL
   */
  url?: string;
  
  /**
   * 추가 컨텍스트
   */
  context?: Record<string, unknown>;
  
  /**
   * 컴포넌트 스택 (Error Boundary에서)
   */
  componentStack?: string;
}

/**
 * Error Logger 클래스
 */
class ErrorLogger {
  private logs: ErrorLog[] = [];
  private maxLogs = 50; // 최대 저장 로그 수
  private storageKey = 'skyclust-error-logs';

  /**
   * 에러 로그 생성
   */
  createLog(
    error: Error | unknown,
    context?: Record<string, unknown>,
    componentStack?: string
  ): ErrorLog {
    const id = `err-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    const timestamp = new Date().toISOString();

    let message = 'Unknown error';
    let stack: string | undefined;
    let type = 'Unknown';
    let code: string | undefined;
    let statusCode: number | undefined;

    if (error instanceof Error) {
      message = error.message;
      stack = error.stack;
      type = error.name;
    } else if (typeof error === 'string') {
      message = error;
      type = 'StringError';
    } else if (error && typeof error === 'object') {
      if ('message' in error) {
        message = String(error.message);
      }
      if ('name' in error) {
        type = String(error.name);
      }
      if ('stack' in error) {
        stack = String(error.stack);
      }
      if ('code' in error) {
        code = String(error.code);
      }
      if ('status' in error) {
        statusCode = Number(error.status);
      }
      if ('statusCode' in error) {
        statusCode = Number(error.statusCode);
      }
    }

    const log: ErrorLog = {
      id,
      message,
      stack,
      type,
      code,
      statusCode,
      timestamp,
      userAgent: typeof window !== 'undefined' ? navigator.userAgent : undefined,
      url: typeof window !== 'undefined' ? window.location.href : undefined,
      context,
      componentStack,
    };

    return log;
  }

  /**
   * 에러 로그 저장
   */
  log(error: Error | unknown, context?: Record<string, unknown>, componentStack?: string): string {
    const log = this.createLog(error, context, componentStack);
    
    this.logs.push(log);
    
    // 최대 개수 초과 시 오래된 로그 제거
    if (this.logs.length > this.maxLogs) {
      this.logs.shift();
    }

    // localStorage에 저장
    this.saveToStorage();

    // 에러 로깅 (개발 환경에서만 콘솔에 표시)
    // 순환 참조를 피하기 위해 직접 console 사용
    if (process.env.NODE_ENV === 'development') {
      console.error('[ERROR] Error logged', error, {
        ...context,
        errorId: log.id,
        errorType: log.type,
        errorCode: log.code,
        statusCode: log.statusCode,
      });
    }

    // 외부 로깅 서비스로 전송 (Sentry)
    this.sendToExternalService(log).catch(() => {
      // 에러 전송 실패는 조용히 무시
    });

    return log.id;
  }

  /**
   * localStorage에 저장
   */
  private saveToStorage(): void {
    if (typeof window === 'undefined') return;

    try {
      localStorage.setItem(this.storageKey, JSON.stringify(this.logs));
    } catch (error) {
      // 순환 참조를 피하기 위해 직접 console 사용
      if (process.env.NODE_ENV === 'development') {
        console.warn('[WARN] Failed to save error logs to localStorage', error);
      }
    }
  }

  /**
   * localStorage에서 로드
   */
  loadFromStorage(): ErrorLog[] {
    if (typeof window === 'undefined') return [];

    try {
      const stored = localStorage.getItem(this.storageKey);
      if (stored) {
        this.logs = JSON.parse(stored) as ErrorLog[];
        return [...this.logs];
      }
    } catch (error) {
      // 순환 참조를 피하기 위해 직접 console 사용
      if (process.env.NODE_ENV === 'development') {
        console.warn('[WARN] Failed to load error logs from localStorage', error);
      }
    }

    return [];
  }

  /**
   * 로그 가져오기
   */
  getLogs(limit?: number): ErrorLog[] {
    const logs = limit ? this.logs.slice(-limit) : [...this.logs];
    return logs;
  }

  /**
   * 특정 로그 가져오기
   */
  getLog(id: string): ErrorLog | undefined {
    return this.logs.find(log => log.id === id);
  }

  /**
   * 로그 삭제
   */
  deleteLog(id: string): boolean {
    const index = this.logs.findIndex(log => log.id === id);
    if (index === -1) return false;

    this.logs.splice(index, 1);
    this.saveToStorage();
    return true;
  }

  /**
   * 모든 로그 삭제
   */
  clearLogs(): void {
    this.logs = [];
    this.saveToStorage();
  }

  /**
   * 외부 로깅 서비스로 전송 (Sentry)
   */
  private async sendToExternalService(log: ErrorLog): Promise<void> {
    // 클라이언트에서만 Sentry 사용
    if (typeof window === 'undefined') return;

    try {

      // Dynamic import로 Sentry SDK 가져오기
      const Sentry = await import('@sentry/nextjs');

      // Error 객체 재구성
      const error = new Error(log.message);
      error.name = log.type;
      if (log.stack) {
        error.stack = log.stack;
      }

      // Sentry에 컨텍스트 추가
      Sentry.withScope((scope) => {
        if (log.context) {
          scope.setContext('errorContext', log.context);
        }

        // 추가 정보를 컨텍스트로 설정
        if (log.code) {
          scope.setContext('errorCode', { code: log.code });
        }

        if (log.statusCode) {
          scope.setContext('httpStatus', { status: log.statusCode });
        }

        if (log.url) {
          scope.setContext('url', { url: log.url });
        }

        if (log.componentStack) {
          scope.setContext('componentStack', { stack: log.componentStack });
        }

        // 추가 메타데이터
        scope.setTag('errorId', log.id);
        if (log.timestamp) {
          scope.setTag('timestamp', log.timestamp);
        }
        if (log.userAgent) {
          scope.setTag('userAgent', log.userAgent);
        }

        // 에러 전송
        Sentry.captureException(error);
      });
    } catch (error) {
      // Sentry 전송 실패는 조용히 무시 (무한 루프 방지)
      // 순환 참조를 피하기 위해 직접 console 사용
      if (process.env.NODE_ENV === 'development') {
        console.warn('[WARN] Failed to send error log to Sentry', error);
      }
    }
  }
}

// 싱글톤 인스턴스
let loggerInstance: ErrorLogger | null = null;

/**
 * Error Logger 인스턴스 가져오기
 */
export function getErrorLogger(): ErrorLogger {
  if (!loggerInstance) {
    loggerInstance = new ErrorLogger();
    // 초기화 시 localStorage에서 로드
    loggerInstance.loadFromStorage();
  }
  return loggerInstance;
}

/**
 * 에러 로깅 헬퍼 함수
 */
export function logError(
  error: Error | unknown,
  context?: Record<string, unknown>,
  componentStack?: string
): string {
  return getErrorLogger().log(error, context, componentStack);
}

