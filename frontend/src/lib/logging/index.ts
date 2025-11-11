/**
 * Logging Module
 * 통합 로깅 모듈
 */

export { Logger, logger, log } from './logger';
export type { LogLevel, LoggerOptions } from './logger';

// ErrorLogger 타입도 export (하위 호환성)
export type { ErrorLog } from '../error-handling';
export { getErrorLogger } from '../error-handling';


