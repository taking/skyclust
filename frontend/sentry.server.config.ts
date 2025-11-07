// This file configures the initialization of Sentry on the server.
// The config you add here will be used whenever the server handles a request.
// https://docs.sentry.io/platforms/javascript/guides/nextjs/

import * as Sentry from "@sentry/nextjs";
import { nodeProfilingIntegration } from "@sentry/profiling-node";

Sentry.init({
  dsn: process.env.NEXT_PUBLIC_SENTRY_DSN || "http://9dd7b83debf801f1246f338022665a98@localhost:9000/2",

  // Tracing 설정
  // Set tracesSampleRate to 1.0 to capture 100% of the transactions
  // Tracing must be enabled for profiling to work
  tracesSampleRate: process.env.NODE_ENV === 'production' ? 0.1 : 1.0,

  // Integrations
  integrations: [
    nodeProfilingIntegration(),
  ],

  // Node.js Profiling 설정
  // Set sampling rate for profiling - this is evaluated only once per SDK.init call
  profileSessionSampleRate: process.env.NODE_ENV === 'production' ? 0.1 : 1.0,
  // Trace lifecycle automatically enables profiling during active traces
  profileLifecycle: 'trace',

  // Setting this option to true will print useful information to the console while you're setting up Sentry.
  debug: process.env.NODE_ENV === 'development',
});

// Profiling happens automatically after setting it up with `Sentry.init()`.
// All spans (unless those discarded by sampling) will have profiling data attached to them.

