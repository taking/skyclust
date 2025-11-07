/**
 * Logger
 * 통합 로깅 유틸리티
 * 
 * 모든 로깅을 통일된 인터페이스로 제공합니다.
 * debug, info, warn, error 레벨을 지원하며,
 * ErrorLogger를 내부적으로 사용하여 에러 로깅과 Sentry 연동을 처리합니다.
 * 
 * @example
 * ```tsx
 * import { logger } from '@/lib/logger';
 * 
 * logger.debug('Debug message', { context: 'test' });
 * logger.info('Info message');
 * logger.warn('Warning message', error);
 * logger.error('Error message', error, { operation: 'createVM' });
 * ```
 */

import { getErrorLogger, type ErrorLog } from './error-logger';

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
 * Logger 클래스
 * 통합 로깅 기능 제공
 */
class Logger {
  private errorLogger = getErrorLogger();
  private options: LoggerOptions;

  constructor(options: LoggerOptions = {}) {
    this.options = {
      minLevel: options.minLevel || (process.env.NODE_ENV === 'production' ? 'warn' : 'debug'),
      devOnly: options.devOnly ?? false,
    };
  }

  /**
   * 로그 레벨 우선순위
   */
  private getLevelPriority(level: LogLevel): number {
    const priorities: Record<LogLevel, number> = {
      debug: 0,
      info: 1,
      warn: 2,
      error: 3,
    };
    return priorities[level];
  }

  /**
   * 로그 레벨 확인
   */
  private shouldLog(level: LogLevel): boolean {
    const minPriority = this.getLevelPriority(this.options.minLevel || 'debug');
    const currentPriority = this.getLevelPriority(level);
    
    if (currentPriority < minPriority) {
      return false;
    }

    if (this.options.devOnly && process.env.NODE_ENV === 'production') {
      return false;
    }

    return true;
  }

  /**
   * 개발 환경에서만 콘솔에 출력
   */
  private logToConsole(level: LogLevel, message: string, data?: unknown): void {
    if (process.env.NODE_ENV !== 'development') {
      return;
    }

    const prefix = `[${level.toUpperCase()}]`;
    const timestamp = new Date().toISOString();
    const logMessage = `${prefix} ${timestamp} - ${message}`;

    switch (level) {
      case 'debug':
        if (data) {
          console.debug(logMessage, data);
        } else {
          console.debug(logMessage);
        }
        break;
      case 'info':
        if (data) {
          console.info(logMessage, data);
        } else {
          console.info(logMessage);
        }
        break;
      case 'warn':
        if (data) {
          console.warn(logMessage, data);
        } else {
          console.warn(logMessage);
        }
        break;
      case 'error':
        if (data) {
          console.error(logMessage, data);
        } else {
          console.error(logMessage);
        }
        break;
    }
  }

  /**
   * Debug 로그
   * 개발 환경에서만 표시되는 상세 정보
   */
  debug(message: string, context?: Record<string, unknown>): void {
    if (!this.shouldLog('debug')) {
      return;
    }

    this.logToConsole('debug', message, context);
    
    // Debug 레벨은 ErrorLogger에 저장하지 않음 (너무 많은 로그 방지)
  }

  /**
   * Info 로그
   * 일반적인 정보성 메시지
   */
  info(message: string, context?: Record<string, unknown>): void {
    if (!this.shouldLog('info')) {
      return;
    }

    this.logToConsole('info', message, context);
    
    // Info 레벨은 ErrorLogger에 저장하지 않음
  }

  /**
   * Warning 로그
   * 경고 메시지 (에러는 아니지만 주의 필요)
   */
  warn(message: string, error?: unknown, context?: Record<string, unknown>): void {
    if (!this.shouldLog('warn')) {
      return;
    }

    this.logToConsole('warn', message, error || context);

    // Warning이 에러 객체와 함께 오면 ErrorLogger로 전송
    if (error) {
      const errorObj = error instanceof Error 
        ? error 
        : new Error(message);
      
      this.errorLogger.log(errorObj, {
        ...context,
        level: 'warn',
        originalMessage: message,
      });
    }
  }

  /**
   * Error 로그
   * 에러 메시지 (에러 로깅 및 Sentry 전송)
   */
  error(message: string, error: unknown, context?: Record<string, unknown>): void {
    if (!this.shouldLog('error')) {
      return;
    }

    this.logToConsole('error', message, error);

    // ErrorLogger로 전송 (Sentry 연동 포함)
    const errorObj = error instanceof Error 
      ? error 
      : new Error(message);
    
    this.errorLogger.log(errorObj, {
      ...context,
      level: 'error',
      originalMessage: message,
    });
  }

  /**
   * 에러 로그 (ErrorLogger 직접 사용)
   * ErrorLogger의 모든 기능을 사용할 때
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

// 싱글톤 인스턴스
let loggerInstance: Logger | null = null;

/**
 * Logger 인스턴스 가져오기
 */
export function getLogger(): Logger {
  if (!loggerInstance) {
    loggerInstance = new Logger();
  }
  return loggerInstance;
}

/**
 * 기본 Logger 인스턴스 (편의를 위한 export)
 */
export const logger = getLogger();

