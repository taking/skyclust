'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/auth';
import { Spinner } from '@/components/ui/spinner';

export default function HomePage() {
  const { isAuthenticated, token, initialize } = useAuthStore();
  const router = useRouter();
  const [isChecking, setIsChecking] = useState(true);

  useEffect(() => {
    // Initialize auth store (syncs with localStorage)
    initialize();

    // Wait for Zustand persist to hydrate, then check auth
    const checkAuth = () => {
      // Check localStorage directly to avoid Zustand hydration timing issues
      const storedToken = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
      const authStorage = typeof window !== 'undefined' 
        ? localStorage.getItem('auth-storage') 
        : null;
      
      let parsedAuth: { state?: { isAuthenticated?: boolean; token?: string } } = {};
      try {
        if (authStorage) {
          parsedAuth = JSON.parse(authStorage);
        }
      } catch {
        // Ignore parse errors
      }
      
      const isAuth = storedToken || (parsedAuth?.state?.isAuthenticated && parsedAuth?.state?.token);

      if (isAuth) {
        router.replace('/dashboard');
      } else {
        router.replace('/login');
      }
      setIsChecking(false);
    };

    // Small delay to allow Zustand persist to hydrate
    const timer = setTimeout(checkAuth, 500);
    
    return () => clearTimeout(timer);
  }, [router, initialize]); // Remove frequently changing dependencies

  if (isChecking) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center space-y-4">
          <h1 className="text-4xl font-bold text-gray-900">SkyClust</h1>
          <Spinner size="lg" label="Loading..." />
        </div>
      </div>
    );
  }

  return null;
}