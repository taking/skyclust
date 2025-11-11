/**
 * Logger
 * 모든 로깅을 통합한 단일 인터페이스
 * 
 * Logger와 ErrorLogger의 기능을 통합하여 일관된 로깅 경험 제공
 * console.log/error/warn/debug를 이 인터페이스를 통해서만 사용하도록 강제
 */

import type { ErrorLog } from '../error-handling/types';

/**
 * 로그 레벨
 */
export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

/**
 * Logger 옵션
 */
export interface LoggerOptions {
  /**
   * 최소 로그 레벨 (이 레벨 이상만 로깅)
   */
  minLevel?: LogLevel;

  /**
   * 개발 환경에서만 로깅할지 여부
   */
  devOnly?: boolean;
}

/**
 * Error Logger 내부 클래스
 * 에러 로깅 유틸리티
 */
class ErrorLogger {
  private logs: ErrorLog[] = [];
  private maxLogs = 50;
  private storageKey = 'skyclust-error-logs';

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

  log(error: Error | unknown, context?: Record<string, unknown>, componentStack?: string): string {
    const log = this.createLog(error, context, componentStack);
    
    this.logs.push(log);
    
    if (this.logs.length > this.maxLogs) {
      this.logs.shift();
    }

    this.saveToStorage();

    if (process.env.NODE_ENV === 'development') {
      console.error('[ERROR] Error logged', error, {
        ...context,
        errorId: log.id,
        errorType: log.type,
        errorCode: log.code,
        statusCode: log.statusCode,
      });
    }

    this.sendToExternalService(log).catch(() => {
      // 에러 전송 실패는 조용히 무시
    });

    return log.id;
  }

  private saveToStorage(): void {
    if (typeof window === 'undefined') return;

    try {
      localStorage.setItem(this.storageKey, JSON.stringify(this.logs));
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.warn('[WARN] Failed to save error logs to localStorage', error);
      }
    }
  }

  loadFromStorage(): ErrorLog[] {
    if (typeof window === 'undefined') return [];

    try {
      const stored = localStorage.getItem(this.storageKey);
      if (stored) {
        this.logs = JSON.parse(stored) as ErrorLog[];
        return [...this.logs];
      }
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.warn('[WARN] Failed to load error logs from localStorage', error);
      }
    }

    return [];
  }

  getLogs(limit?: number): ErrorLog[] {
    const logs = limit ? this.logs.slice(-limit) : [...this.logs];
    return logs;
  }

  getLog(id: string): ErrorLog | undefined {
    return this.logs.find(log => log.id === id);
  }

  deleteLog(id: string): boolean {
    const index = this.logs.findIndex(log => log.id === id);
    if (index === -1) return false;

    this.logs.splice(index, 1);
    this.saveToStorage();
    return true;
  }

  clearLogs(): void {
    this.logs = [];
    this.saveToStorage();
  }

  private async sendToExternalService(log: ErrorLog): Promise<void> {
    if (typeof window === 'undefined') return;

    try {
      const Sentry = await import('@sentry/nextjs');

      const error = new Error(log.message);
      error.name = log.type;
      if (log.stack) {
        error.stack = log.stack;
      }

      Sentry.withScope((scope) => {
        if (log.context) {
          scope.setContext('errorContext', log.context);
        }

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

        scope.setTag('errorId', log.id);
        if (log.timestamp) {
          scope.setTag('timestamp', log.timestamp);
        }
        if (log.userAgent) {
          scope.setTag('userAgent', log.userAgent);
        }

        Sentry.captureException(error);
      });
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.warn('[WARN] Failed to send error log to Sentry', error);
      }
    }
  }
}

/**
 * Logger 클래스
 * 모든 로깅 기능을 통합한 클래스
 */
export class Logger {
  private static instance: Logger;
  private errorLogger: ErrorLogger;
  private options: LoggerOptions;

  private constructor(options: LoggerOptions = {}) {
    this.options = {
      minLevel: options.minLevel || (process.env.NODE_ENV === 'production' ? 'warn' : 'debug'),
      devOnly: options.devOnly ?? false,
    };
    this.errorLogger = new ErrorLogger();
    // 초기화 시 localStorage에서 로드
    this.errorLogger.loadFromStorage();
  }

  /**
   * 인스턴스 가져오기
   */
  static getInstance(options?: LoggerOptions): Logger {
    if (!Logger.instance) {
      Logger.instance = new Logger(options);
    }
    return Logger.instance;
  }

  /**
   * 로그 레벨 확인
   */
  private shouldLog(level: LogLevel): boolean {
    if (this.options.devOnly && process.env.NODE_ENV === 'production') {
      return false;
    }

    const levels: LogLevel[] = ['debug', 'info', 'warn', 'error'];
    const minLevelIndex = levels.indexOf(this.options.minLevel || 'debug');
    const currentLevelIndex = levels.indexOf(level);

    return currentLevelIndex >= minLevelIndex;
  }

  /**
   * Debug 로그
   */
  debug(message: string, context?: Record<string, unknown>): void {
    if (!this.shouldLog('debug')) return;

    if (typeof window !== 'undefined' && process.env.NODE_ENV !== 'production') {
      console.debug(`[DEBUG] ${message}`, context || '');
    }
  }

  /**
   * Info 로그
   */
  info(message: string, context?: Record<string, unknown>): void {
    if (!this.shouldLog('info')) return;

    if (typeof window !== 'undefined') {
      console.info(`[INFO] ${message}`, context || '');
    }
  }

  /**
   * Warning 로그
   */
  warn(message: string, error?: unknown, context?: Record<string, unknown>): void {
    if (!this.shouldLog('warn')) return;

    if (typeof window !== 'undefined') {
      console.warn(`[WARN] ${message}`, error || '', context || '');
    }

    // 에러가 있으면 ErrorLogger에도 기록
    if (error) {
      const errorObj = error instanceof Error ? error : new Error(message);
      this.errorLogger.log(errorObj, {
        ...context,
        level: 'warn',
        originalMessage: message,
      });
    }
  }

  /**
   * Error 로그
   */
  error(message: string, error: unknown, context?: Record<string, unknown>): void {
    if (!this.shouldLog('error')) return;

    if (typeof window !== 'undefined') {
      console.error(`[ERROR] ${message}`, error, context || '');
    }

    // ErrorLogger에 기록
    const errorObj = error instanceof Error ? error : new Error(message);
    this.errorLogger.log(errorObj, {
      ...context,
      level: 'error',
      originalMessage: message,
    });
  }

  /**
   * 에러 로깅 (ErrorLogger와 통합)
   */
  logError(
    error: Error | unknown,
    context?: Record<string, unknown>,
    componentStack?: string
  ): string {
    return this.errorLogger.log(error, context, componentStack);
  }

  /**
   * 에러 로그 가져오기
   */
  getErrorLogs(limit?: number): ErrorLog[] {
    return this.errorLogger.getLogs(limit);
  }

  /**
   * 특정 에러 로그 가져오기
   */
  getErrorLog(id: string): ErrorLog | undefined {
    return this.errorLogger.getLog(id);
  }

  /**
   * 에러 로그 삭제
   */
  deleteErrorLog(id: string): boolean {
    return this.errorLogger.deleteLog(id);
  }

  /**
   * 모든 에러 로그 삭제
   */
  clearErrorLogs(): void {
    this.errorLogger.clearLogs();
  }
}


/**
 * 편의를 위한 싱글톤 인스턴스 export
 */
export const logger = Logger.getInstance();

/**
 * 통합 로깅 인터페이스
 * 모든 로깅은 이 객체를 통해서만 수행
 */
export const log = {
  debug: (message: string, context?: Record<string, unknown>) => {
    logger.debug(message, context);
  },
  info: (message: string, context?: Record<string, unknown>) => {
    logger.info(message, context);
  },
  warn: (message: string, error?: unknown, context?: Record<string, unknown>) => {
    logger.warn(message, error, context);
  },
  error: (message: string, error: unknown, context?: Record<string, unknown>) => {
    logger.error(message, error, context);
  },
  logError: (
    error: Error | unknown,
    context?: Record<string, unknown>,
    componentStack?: string
  ): string => {
    return logger.logError(error, context, componentStack);
  },
  getErrorLogs: (limit?: number) => logger.getErrorLogs(limit),
  getErrorLog: (id: string) => logger.getErrorLog(id),
  deleteErrorLog: (id: string) => logger.deleteErrorLog(id),
  clearErrorLogs: () => logger.clearErrorLogs(),
};

// 타입 export
export type { ErrorLog } from '../error-handling/types';

