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

    // Verify token is still valid by checking localStorage
    const storedToken = localStorage.getItem('token');
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

