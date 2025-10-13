import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import { QueryProvider } from '@/components/providers/query-provider';
import { ThemeProvider } from '@/components/providers/theme-provider';
import { ToastProvider } from '@/components/ui/toast';
import { AppErrorBoundary } from '@/components/error-boundary';

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
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <AppErrorBoundary>
            <QueryProvider>
              {children}
              <ToastProvider />
            </QueryProvider>
          </AppErrorBoundary>
        </ThemeProvider>
      </body>
    </html>
  );
}