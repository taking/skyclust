'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useAuthStore } from '@/store/auth';
import { authService } from '@/services/auth';
import { LoginForm } from '@/lib/types';
import { getUserFriendlyErrorMessage } from '@/lib/error-handler';
import Link from 'next/link';

const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
});

export default function LoginPage() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const { login, isAuthenticated, token, initialize } = useAuthStore();
  const router = useRouter();

  // Check if already authenticated - only redirect if definitely authenticated
  useEffect(() => {
    initialize();
    
    // Only check once on mount, don't re-trigger on state changes
    let hasRedirected = false;
    
    const checkAuth = () => {
      if (hasRedirected) return;
      
      // Only redirect if we have both token and authenticated state
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
      
      const isAuth = parsedAuth?.state?.isAuthenticated && parsedAuth?.state?.token;
      
      // Only redirect if we have clear authentication evidence
      if (storedToken || isAuth) {
        hasRedirected = true;
        router.replace('/dashboard');
      }
    };

    // Small delay to allow Zustand persist to hydrate
    const timer = setTimeout(checkAuth, 500);
    return () => clearTimeout(timer);
  }, [router, initialize]); // Remove dependencies that change frequently

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data: LoginForm) => {
    try {
      setIsLoading(true);
      setError('');
      const response = await authService.login(data);
      login(response);
      router.push('/dashboard');
    } catch (err: unknown) {
      setError(getUserFriendlyErrorMessage(err));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold text-center">Sign in</CardTitle>
          <CardDescription className="text-center">
            Enter your email and password to sign in to your account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="Enter your email"
                {...register('email')}
              />
              {errors.email && (
                <p className="text-sm text-red-600">{errors.email.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="Enter your password"
                {...register('password')}
              />
              {errors.password && (
                <p className="text-sm text-red-600">{errors.password.message}</p>
              )}
            </div>
            {error && (
              <div className="text-sm text-red-600 text-center">{error}</div>
            )}
            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? 'Signing in...' : 'Sign in'}
            </Button>
          </form>
          <div className="mt-4 text-center text-sm">
            Don&apos;t have an account?{' '}
            <Link href="/register" className="text-blue-600 hover:underline">
              Sign up
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
