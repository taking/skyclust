// This file configures the initialization of Sentry on the client.
// The config you add here will be used whenever a users loads a page in their browser.
// https://docs.sentry.io/platforms/javascript/guides/nextjs/

import * as Sentry from "@sentry/nextjs";

Sentry.init({
  dsn: process.env.NEXT_PUBLIC_SENTRY_DSN || "http://9dd7b83debf801f1246f338022665a98@localhost:9000/2",

  // Tracing 설정
  // Set tracesSampleRate to 1.0 to capture 100% of transactions for performance monitoring.
  // We recommend adjusting this value in production
  tracesSampleRate: process.env.NODE_ENV === 'production' ? 0.1 : 1.0,

  // Integrations
  integrations: [
    Sentry.browserTracingIntegration(),
    Sentry.browserProfilingIntegration(),
    Sentry.replayIntegration({
      maskAllText: true,
      blockAllMedia: true,
    }),
  ],

  // Set 'tracePropagationTargets' to control for which URLs distributed tracing should be enabled
  tracePropagationTargets: [
    "localhost",
    /^https:\/\/.*\.io\/api/,
    process.env.NEXT_PUBLIC_API_URL ? new RegExp(`^${process.env.NEXT_PUBLIC_API_URL}`) : /^https:\/\/.*/,
  ],

  // Browser Profiling 설정
  // Set profilesSampleRate to 1.0 to profile every transaction.
  // Since profilesSampleRate is relative to tracesSampleRate,
  // the final profiling rate can be computed as tracesSampleRate * profilesSampleRate
  profilesSampleRate: process.env.NODE_ENV === 'production' ? 0.1 : 1.0,

  // Session Replay 설정
  // This sets the sample rate at 10%. You may want to change it to 100% while in development and then sample at a lower rate in production.
  replaysSessionSampleRate: process.env.NODE_ENV === 'production' ? 0.1 : 1.0,
  // If you're not already sampling the entire session, change the sample rate to 100% when sampling sessions where errors occur.
  replaysOnErrorSampleRate: 1.0,

  // Setting this option to true will print useful information to the console while you're setting up Sentry.
  debug: process.env.NODE_ENV === 'development',

  // Uncomment the line below to enable Spotlight (https://spotlightjs.com)
  // spotlight: process.env.NODE_ENV === 'development',

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

  // Enable logs to be sent to Sentry
  enableLogs: true,

  // Enable sending user PII (Personally Identifiable Information)
  // https://docs.sentry.io/platforms/javascript/guides/nextjs/configuration/options/#sendDefaultPii
  sendDefaultPii: true,
});

// Next.js 라우터 전환 추적을 위한 export
export const onRouterTransitionStart = Sentry.captureRouterTransitionStart;

