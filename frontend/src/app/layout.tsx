import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import { QueryProvider } from '@/components/providers/query-provider';
import { ThemeProvider } from '@/components/providers/theme-provider';
import { ToastProvider } from '@/components/ui/toast';
import { AppErrorBoundary } from '@/components/error-boundary';
import { OfflineBanner } from '@/components/common/offline-banner';
import { SentryProvider } from '@/components/providers/sentry-provider';

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
    <html lang="en" suppressHydrationWarning>
      <body className={inter.className}>
        <SentryProvider>
          <ThemeProvider
            attribute="class"
            defaultTheme="system"
            enableSystem
            disableTransitionOnChange
          >
            <AppErrorBoundary>
              <QueryProvider>
                <OfflineBanner position="top" autoHide showRefreshButton />
                {children}
                <ToastProvider />
              </QueryProvider>
            </AppErrorBoundary>
          </ThemeProvider>
        </SentryProvider>
      </body>
    </html>
  );
}