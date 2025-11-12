import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import { QueryProvider } from '@/components/providers/query-provider';
import { ThemeProvider } from '@/components/providers/theme-provider';
import { ToastProvider } from '@/components/ui/toast';
import { AppErrorBoundary } from '@/components/error-boundary';
import { OfflineBanner } from '@/components/common/offline-banner';
import { SentryProvider } from '@/components/providers/sentry-provider';
import { I18nProvider } from '@/components/providers/i18n-provider';
// 개발 환경에서 Sentry 테스트 함수 로드
if (process.env.NODE_ENV === 'development') {
  import('@/lib/test');
}

// 개발 환경에서 개발 도구 관련 경고 필터링
if (typeof window !== 'undefined' && process.env.NODE_ENV === 'development') {
  const originalError = console.error;
  const originalWarn = console.warn;
  
  // console.error 필터링
  console.error = (...args: unknown[]) => {
    // HMR WebSocket 에러는 무시 (정상적인 개발 모드 동작)
    const message = args[0]?.toString() || '';
    if (
      message.includes('WebSocket connection to') &&
      message.includes('_next/webpack-hmr')
    ) {
      return;
    }
    originalError.apply(console, args);
  };
  
  // console.warn 필터링 (소스맵 경고 등)
  console.warn = (...args: unknown[]) => {
    const message = args[0]?.toString() || '';
    // 소스맵 관련 경고는 무시 (정상적인 개발 모드 동작)
    if (
      message.includes('소스 맵에 유효하지 않은') ||
      message.includes('sourcesContent') ||
      message.includes('source map') ||
      message.includes('_next/static/chunks')
    ) {
      return;
    }
    originalWarn.apply(console, args);
  };
}

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'SkyClust - Multi-Cloud Management Platform',
  description: 'Unified multi-cloud management platform for VMs, infrastructure, and resources',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ko" suppressHydrationWarning>
      <body className={inter.className}>
        <SentryProvider>
          <ThemeProvider
            attribute="class"
            defaultTheme="system"
            enableSystem
            disableTransitionOnChange
          >
            <I18nProvider>
              <AppErrorBoundary>
                <QueryProvider>
                  <OfflineBanner position="top" autoHide showRefreshButton />
                  {children}
                  <ToastProvider />
                </QueryProvider>
              </AppErrorBoundary>
            </I18nProvider>
          </ThemeProvider>
        </SentryProvider>
      </body>
    </html>
  );
}