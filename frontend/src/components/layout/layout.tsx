'use client';

import { Header } from './header';
import { Sidebar } from './sidebar';
import { useAuthStore } from '@/store/auth';
import { useRouter, usePathname } from 'next/navigation';
import { useEffect, useState } from 'react';
import { SkipLink } from '@/components/accessibility/skip-link';
import { LiveRegion } from '@/components/accessibility/live-region';
import { GlobalKeyboardShortcuts } from '@/components/common/global-keyboard-shortcuts';
import { OfflineBanner } from '@/components/common/offline-banner';
import { KeyboardShortcutsHelp } from '@/components/common/keyboard-shortcuts-help';
import { KeyboardShortcut } from '@/hooks/use-keyboard-shortcuts';
import { HelpCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface LayoutProps {
  children: React.ReactNode;
}

export function Layout({ children }: LayoutProps) {
  const { isAuthenticated, initialize } = useAuthStore();
  const router = useRouter();
  const pathname = usePathname();
  const [isHydrated, setIsHydrated] = useState(false);

  // Initialize auth store and wait for hydration
  useEffect(() => {
    initialize();
    
    // Wait for Zustand persist to rehydrate
    const checkHydration = () => {
      if (typeof window !== 'undefined') {
        // Check if auth-storage exists in localStorage (primary source)
        const authStorage = localStorage.getItem('auth-storage');
        // Fallback to legacy token for backward compatibility
        const legacyToken = localStorage.getItem('token');
        
        // If either exists, we need to wait a bit for Zustand to rehydrate
        if (authStorage || legacyToken) {
          // Give Zustand time to rehydrate
          setTimeout(() => {
            setIsHydrated(true);
          }, 100);
        } else {
          // No auth data, safe to check immediately
          setIsHydrated(true);
        }
      } else {
        setIsHydrated(true);
      }
    };

    checkHydration();
  }, [initialize]);

  // Only redirect if we're not on login/register pages and hydration is complete
  useEffect(() => {
    if (!isHydrated) return;
    
    const publicPaths = ['/login', '/register'];
    const isPublicPath = publicPaths.includes(pathname);
    
    if (!isAuthenticated && !isPublicPath) {
      router.replace('/login');
    }
  }, [isAuthenticated, isHydrated, router, pathname]);

  // Don't render anything until hydration is complete
  if (!isHydrated) {
    return null;
  }

  // Don't render layout on login/register pages
  const publicPaths = ['/login', '/register'];
  if (publicPaths.includes(pathname)) {
    return <>{children}</>;
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="flex h-screen bg-background">
      <GlobalKeyboardShortcuts />
      <SkipLink href="#main-content">Skip to main content</SkipLink>
      <OfflineBanner position="top" autoHide showRefreshButton />
      <Sidebar />
      <div className="flex flex-1 flex-col overflow-hidden">
        <Header />
        <main 
          id="main-content"
          className="flex-1 overflow-y-auto bg-muted/30 p-3 sm:p-4 md:p-6"
          role="main"
          tabIndex={-1}
        >
          {children}
        </main>
      </div>
      <LiveRegion message="" />
      {/* Floating Help Button - Safari compatible */}
      <div 
        className="fixed bottom-6 right-6 z-50" 
        style={{ 
          position: 'fixed', 
          bottom: '1.5rem', 
          right: '1.5rem', 
          zIndex: 50,
          WebkitTransform: 'translateZ(0)', // Safari hardware acceleration
          transform: 'translateZ(0)',
          willChange: 'transform', // Optimize for Safari
        }}
      >
        <KeyboardShortcutsHelp
          shortcuts={(typeof window !== 'undefined' && (window as Window & { __keyboardShortcuts?: KeyboardShortcut[] }).__keyboardShortcuts) || []}
          trigger={
            <Button
              variant="default"
              size="icon"
              className="h-12 w-12 rounded-full shadow-lg hover:shadow-xl transition-shadow"
              aria-label="Show keyboard shortcuts"
            >
              <HelpCircle className="h-6 w-6" />
            </Button>
          }
        />
      </div>
    </div>
  );
}
