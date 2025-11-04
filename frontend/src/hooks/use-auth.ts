import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/auth';

export const useAuth = (redirectTo: string = '/login') => {
  const router = useRouter();
  const { isAuthenticated, token, user } = useAuthStore();

  useEffect(() => {
    // Check if user is authenticated
    if (!isAuthenticated || !token || !user) {
      router.push(redirectTo);
      return;
    }

    // Verify token is still valid by checking auth-storage
    let storedToken: string | null = null;
    try {
      const authStorage = typeof window !== 'undefined' ? localStorage.getItem('auth-storage') : null;
      if (authStorage) {
        const parsed = JSON.parse(authStorage);
        storedToken = parsed?.state?.token || null;
      }
      // Fallback to legacy token for backward compatibility
      if (!storedToken && typeof window !== 'undefined') {
        storedToken = localStorage.getItem('token');
      }
    } catch {
      storedToken = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
    }
    
    if (!storedToken || storedToken !== token) {
      useAuthStore.getState().logout();
      router.push(redirectTo);
      return;
    }
  }, [isAuthenticated, token, user, router, redirectTo]);

  return { isAuthenticated, user, token };
};

export const useRequireAuth = (redirectTo: string = '/login') => {
  const { isAuthenticated, user, token } = useAuth(redirectTo);
  
  return {
    isAuthenticated,
    user,
    token,
    isLoading: !isAuthenticated && !user && !token,
  };
};

