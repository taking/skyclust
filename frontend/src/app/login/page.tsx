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
import { useAuthHydration } from '@/hooks/use-auth-hydration';
import { authService } from '@/services/auth';
import { LoginForm } from '@/lib/types';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { getUserFriendlyErrorMessage } from '@/lib/error-handling';
import Link from 'next/link';
import { useTranslation } from '@/hooks/use-translation';
import { createValidationSchemas } from '@/lib/validation';

export default function LoginPage() {
  const { login } = useAuthStore();
  const router = useRouter();
  const { t } = useTranslation();
  const { isHydrated, isAuthenticated, isLoading: isAuthLoading } = useAuthHydration({
    hydrationDelay: 300,
    checkLegacyToken: true,
  });
  
  const { loginSchema } = createValidationSchemas(t);

  // Check if already authenticated - only redirect if definitely authenticated
  useEffect(() => {
    if (!isHydrated || isAuthLoading) {
      return;
    }

    if (isAuthenticated) {
      router.replace('/dashboard');
    }
  }, [router, isHydrated, isAuthenticated, isAuthLoading]);

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
        onError: (_error) => {
      // Error is handled by the hook's error state
    },
    getErrorMessage: getUserFriendlyErrorMessage,
  });

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold text-center">{t('auth.signInTitle')}</CardTitle>
          <CardDescription className="text-center">
            {t('auth.signInDescription')}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={handleSubmit} className="space-y-4">
              <EnhancedField
                name="email"
                label={t('auth.email')}
                type="email"
                placeholder={t('auth.emailPlaceholder')}
                required
                getFieldError={getFieldError}
                getFieldValidationState={getFieldValidationState}
              />
              <EnhancedField
                name="password"
                label={t('auth.password')}
                type="password"
                placeholder={t('auth.passwordPlaceholder')}
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
                {isLoading ? t('auth.signingIn') : t('auth.signIn')}
              </Button>
            </form>
          </Form>
          <div className="mt-4 text-center text-sm">
            {t('auth.dontHaveAccount')}{' '}
            <Link href="/register" className="text-blue-600 hover:underline">
              {t('auth.signUp')}
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

