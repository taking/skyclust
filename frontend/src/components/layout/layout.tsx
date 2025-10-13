'use client';

import { Header } from './header';
import { Sidebar } from './sidebar';
import { useAuthStore } from '@/store/auth';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import { SkipLink } from '@/components/accessibility/skip-link';
import { LiveRegion } from '@/components/accessibility/live-region';

interface LayoutProps {
  children: React.ReactNode;
}

export function Layout({ children }: LayoutProps) {
  const { isAuthenticated } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, router]);

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="flex h-screen bg-background">
      <SkipLink href="#main-content">Skip to main content</SkipLink>
      <Sidebar />
      <div className="flex flex-1 flex-col overflow-hidden">
        <Header />
        <main 
          id="main-content"
          className="flex-1 overflow-y-auto bg-muted/30 p-4 sm:p-6"
          role="main"
          tabIndex={-1}
        >
          {children}
        </main>
      </div>
      <LiveRegion message="" />
    </div>
  );
}
