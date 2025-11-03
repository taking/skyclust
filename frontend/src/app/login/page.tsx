/**
 * Login Page
 * 로그인 페이지
 * 
 * use-form-with-validation 훅을 사용한 리팩토링 버전
 */

'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Form } from '@/components/ui/form';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useAuthStore } from '@/store/auth';
import { authService } from '@/services/auth';
import { LoginForm } from '@/lib/types';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { getUserFriendlyErrorMessage } from '@/lib/error-handler';
import Link from 'next/link';
import * as z from 'zod';

const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
});

export default function LoginPage() {
  const { login, initialize } = useAuthStore();
  const router = useRouter();

  // Check if already authenticated - only redirect if definitely authenticated
  useEffect(() => {
    initialize();
    
    let hasRedirected = false;
    
    const checkAuth = () => {
      if (hasRedirected) return;
      
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
      
      if (storedToken || isAuth) {
        hasRedirected = true;
        router.replace('/dashboard');
      }
    };

    const timer = setTimeout(checkAuth, 500);
    return () => clearTimeout(timer);
  }, [router, initialize]);

  const {
    form,
    handleSubmit,
    isLoading,
    error,
    getFieldError,
    getFieldValidationState,
  } = useFormWithValidation<LoginForm>({
    schema: loginSchema,
    defaultValues: {
      email: '',
      password: '',
    },
    onSubmit: async (data) => {
      const response = await authService.login(data);
      login(response);
      router.push('/dashboard');
    },
    onError: (error) => {
      // Error is handled by the hook's error state
    },
    getErrorMessage: getUserFriendlyErrorMessage,
  });

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
          <Form {...form}>
            <form onSubmit={handleSubmit} className="space-y-4">
              <EnhancedField
                name="email"
                label="Email"
                type="email"
                placeholder="Enter your email"
                required
                getFieldError={getFieldError}
                getFieldValidationState={getFieldValidationState}
              />
              <EnhancedField
                name="password"
                label="Password"
                type="password"
                placeholder="Enter your password"
                required
                getFieldError={getFieldError}
                getFieldValidationState={getFieldValidationState}
              />
              {error && (
                <div className="text-sm text-red-600 text-center" role="alert">
                  {error}
                </div>
              )}
              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading ? 'Signing in...' : 'Sign in'}
              </Button>
            </form>
          </Form>
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

