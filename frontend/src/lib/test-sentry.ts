/**
 * Sentry 테스트 유틸리티
 * 브라우저 콘솔에서 테스트할 수 있는 함수들
 */

/**
 * Sentry 테스트 메시지 전송
 * 브라우저 콘솔에서 호출: testSentryMessage()
 */
export async function testSentryMessage(): Promise<void> {
  if (typeof window === 'undefined') {
    console.error('This function can only be called in the browser');
    return;
  }

  const Sentry = await import('@sentry/nextjs');
  
  Sentry.captureMessage('Sentry test message from console', 'info');
  console.log('[Sentry Test] Test message sent! Check your Sentry dashboard.');
}

/**
 * Sentry 테스트 에러 전송
 * 브라우저 콘솔에서 호출: testSentryError()
 */
export async function testSentryError(): Promise<void> {
  if (typeof window === 'undefined') {
    console.error('This function can only be called in the browser');
    return;
  }

  const Sentry = await import('@sentry/nextjs');
  
  const testError = new Error('Sentry test error from console');
  Sentry.withScope((scope) => {
    scope.setContext('test', { test: true, source: 'manual-test' });
    Sentry.captureException(testError);
  });
  console.log('[Sentry Test] Test error sent! Check your Sentry dashboard.');
}

// 전역에서 접근 가능하도록 window에 추가 (개발 환경에서만)
if (typeof window !== 'undefined' && process.env.NODE_ENV === 'development') {
  (window as any).testSentryMessage = testSentryMessage;
  (window as any).testSentryError = testSentryError;
  console.log('[Sentry Test] Test functions available:');
  console.log('  - testSentryMessage() - 테스트 메시지 전송');
  console.log('  - testSentryError() - 테스트 에러 전송');
}

