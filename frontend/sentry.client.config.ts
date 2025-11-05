// This file configures the initialization of Sentry on the client.
// The config you add here will be used whenever a users loads a page in their browser.
// https://docs.sentry.io/platforms/javascript/guides/nextjs/

import * as Sentry from "@sentry/nextjs";

Sentry.init({
  dsn: process.env.NEXT_PUBLIC_SENTRY_DSN || "http://9dd7b83debf801f1246f338022665a98@localhost:9000/2",

  // Adjust this value in production, or use tracesSampler for greater control
  tracesSampleRate: process.env.NODE_ENV === 'production' ? 0.1 : 1.0,

  // Setting this option to true will print useful information to the console while you're setting up Sentry.
  debug: process.env.NODE_ENV === 'development',

  // Uncomment the line below to enable Spotlight (https://spotlightjs.com)
  // spotlight: process.env.NODE_ENV === 'development',

  // Session Replay 기능을 사용하려면 아래 주석을 해제하세요
  // replaysOnErrorSampleRate: 1.0,
  // replaysSessionSampleRate: 0.1,
  // integrations: [
  //   Sentry.replayIntegration({
  //     maskAllText: true,
  //     blockAllMedia: true,
  //   }),
  // ],

  beforeSend(event, hint) {
    // 개발 환경에서는 콘솔에 로그
    if (process.env.NODE_ENV === 'development') {
      console.log('[Sentry Client] Error captured:', event, hint);
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

