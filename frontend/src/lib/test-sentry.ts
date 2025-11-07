/**
 * Sentry 테스트 유틸리티
 * 브라우저 콘솔에서 테스트할 수 있는 함수들
 */

import { logger } from './logger';

/**
 * Sentry 테스트 메시지 전송
 * 브라우저 콘솔에서 호출: testSentryMessage()
 */
export async function testSentryMessage(): Promise<void> {
  if (typeof window === 'undefined') {
    logger.error('This function can only be called in the browser', new Error('Browser only function'));
    return;
  }

  const Sentry = await import('@sentry/nextjs');
  
  Sentry.captureMessage('Sentry test message from console', 'info');
  logger.info('[Sentry Test] Test message sent! Check your Sentry dashboard.');
}

/**
 * Sentry 테스트 에러 전송
 * 브라우저 콘솔에서 호출: testSentryError()
 */
export async function testSentryError(): Promise<void> {
  if (typeof window === 'undefined') {
    logger.error('This function can only be called in the browser', new Error('Browser only function'));
    return;
  }

  const Sentry = await import('@sentry/nextjs');
  
  const testError = new Error('Sentry test error from console');
  Sentry.withScope((scope) => {
    scope.setContext('test', { test: true, source: 'manual-test' });
    Sentry.captureException(testError);
  });
  logger.info('[Sentry Test] Test error sent! Check your Sentry dashboard.');
}

// 전역에서 접근 가능하도록 window에 추가 (개발 환경에서만)
if (typeof window !== 'undefined' && process.env.NODE_ENV === 'development') {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  (window as any).testSentryMessage = testSentryMessage;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  (window as any).testSentryError = testSentryError;
  logger.info('[Sentry Test] Test functions available:', {
    functions: ['testSentryMessage()', 'testSentryError()'],
  });
}

