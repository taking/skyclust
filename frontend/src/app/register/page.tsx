/**
 * Register Page
 * 회원가입 페이지
 * 
 * use-form-with-validation 훅을 사용한 리팩토링 버전
 */

'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Form } from '@/components/ui/form';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { authService } from '@/services/auth';
import { RegisterForm } from '@/lib/types';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import Link from 'next/link';
import * as z from 'zod';

const registerSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters'),
  email: z.string().email('Invalid email address'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
});

export default function RegisterPage() {
  const [success, setSuccess] = useState(false);
  const router = useRouter();

  const {
    form,
    handleSubmit,
    isLoading,
    error,
    reset,
    getFieldError,
    getFieldValidationState,
  } = useFormWithValidation<RegisterForm>({
    schema: registerSchema,
    defaultValues: {
      name: '',
      email: '',
      password: '',
    },
    onSubmit: async (data) => {
      await authService.register(data);
      setSuccess(true);
      setTimeout(() => {
        router.push('/login');
      }, 2000);
    },
    onError: (error) => {
      // Error is handled by the hook's error state
    },
    getErrorMessage: (err) => {
      if (err && typeof err === 'object' && 'response' in err) {
        const response = err.response as { data?: { error?: { message?: string } } };
        return response?.data?.error?.message || 'Registration failed';
      }
      return 'Registration failed';
    },
    resetOnSuccess: true,
  });

  if (success) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6">
            <div className="text-center">
              <div className="text-green-600 text-6xl mb-4">✓</div>
              <h2 className="text-2xl font-bold text-gray-900 mb-2">
                Registration Successful!
              </h2>
              <p className="text-gray-600">
                Your account has been created. Redirecting to login...
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold text-center">Sign up</CardTitle>
          <CardDescription className="text-center">
            Create a new account to get started
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={handleSubmit} className="space-y-4">
              <EnhancedField
                name="name"
                label="Full Name"
                type="text"
                placeholder="Enter your full name"
                required
                getFieldError={getFieldError}
                getFieldValidationState={getFieldValidationState}
              />
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
                {isLoading ? 'Creating account...' : 'Sign up'}
              </Button>
            </form>
          </Form>
          <div className="mt-4 text-center text-sm">
            Already have an account?{' '}
            <Link href="/login" className="text-blue-600 hover:underline">
              Sign in
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

