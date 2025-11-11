/**
 * Logger
 * 모든 로깅을 통합한 단일 인터페이스
 * 
 * Logger와 ErrorLogger의 기능을 통합하여 일관된 로깅 경험 제공
 * console.log/error/warn/debug를 이 인터페이스를 통해서만 사용하도록 강제
 */

import { getErrorLogger, type ErrorLog } from '../error-handling';

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
 * 모든 로깅 기능을 통합한 클래스
 */
export class Logger {
  private static instance: Logger;
  private errorLogger = getErrorLogger();
  private options: LoggerOptions;

  private constructor(options: LoggerOptions = {}) {
    this.options = {
      minLevel: options.minLevel || (process.env.NODE_ENV === 'production' ? 'warn' : 'debug'),
      devOnly: options.devOnly ?? false,
    };
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
};

