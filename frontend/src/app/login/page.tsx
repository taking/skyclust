/**
 * Login Page
 * 로그인 페이지
 * 
 * use-form-with-validation 훅을 사용한 리팩토링 버전
 */

'use client';

import { useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Form } from '@/components/ui/form';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useAuthStore } from '@/store/auth';
import { useAuthHydration } from '@/hooks/use-auth-hydration';
import { authService } from '@/services/auth';
import { LoginForm } from '@/lib/types';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { ErrorHandler } from '@/lib/error-handling';
import Link from 'next/link';
import { useTranslation } from '@/hooks/use-translation';
import { createValidationSchemas } from '@/lib/validation';

export default function LoginPage() {
  const { login } = useAuthStore();
  const router = useRouter();
  const searchParams = useSearchParams();
  const { t } = useTranslation();
  const { isHydrated, isAuthenticated, isLoading: isAuthLoading } = useAuthHydration({
    hydrationDelay: 300,
    checkLegacyToken: true,
  });
  
  const { loginSchema } = createValidationSchemas(t);

  // Check if already authenticated - redirect to returnUrl, lastPath, or dashboard
  useEffect(() => {
    if (!isHydrated || isAuthLoading) {
      return;
    }

    if (isAuthenticated) {
      // 1. returnUrl이 있으면 해당 페이지로 리다이렉트
      const returnUrl = searchParams.get('returnUrl');
      if (returnUrl) {
        try {
          const decodedUrl = decodeURIComponent(returnUrl);
          // 세션 스토리지의 리다이렉트 플래그 제거
          sessionStorage.removeItem('authRedirected');
          router.replace(decodedUrl);
        } catch {
          // URL 디코딩 실패 시 dashboard로 리다이렉트
          sessionStorage.removeItem('authRedirected');
          router.replace('/dashboard');
        }
        return;
      }

      // 2. returnUrl이 없으면 세션 스토리지에서 마지막 경로 확인
      if (typeof window !== 'undefined') {
        const lastPath = sessionStorage.getItem('lastPath');
        if (lastPath && lastPath !== '/login' && lastPath !== '/register' && lastPath !== '/setup') {
          sessionStorage.removeItem('authRedirected');
          router.replace(lastPath);
          return;
        }
      }

      // 3. 마지막 경로도 없으면 workspace가 있으면 해당 workspace의 dashboard로, 없으면 /workspaces로 리다이렉트
      sessionStorage.removeItem('authRedirected');
      if (typeof window !== 'undefined') {
        import('@/store/workspace').then(({ useWorkspaceStore }) => {
          const { currentWorkspace } = useWorkspaceStore.getState();
          
          if (currentWorkspace?.id) {
            import('@/lib/routing/helpers').then(({ buildManagementPath }) => {
              router.replace(buildManagementPath(currentWorkspace.id, 'dashboard'));
            });
          } else {
            router.replace('/workspaces');
          }
        });
      } else {
        router.replace('/dashboard');
      }
    }
  }, [router, searchParams, isHydrated, isAuthenticated, isAuthLoading]);

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
      
      // 1. returnUrl이 있으면 해당 페이지로 리다이렉트
      const returnUrl = searchParams.get('returnUrl');
      if (returnUrl) {
        try {
          const decodedUrl = decodeURIComponent(returnUrl);
          // 세션 스토리지의 리다이렉트 플래그 제거
          sessionStorage.removeItem('authRedirected');
          router.push(decodedUrl);
        } catch {
          // URL 디코딩 실패 시 workspace가 있으면 해당 workspace의 dashboard로, 없으면 /workspaces로 리다이렉트
          sessionStorage.removeItem('authRedirected');
          if (typeof window !== 'undefined') {
            import('@/store/workspace').then(({ useWorkspaceStore }) => {
              const { currentWorkspace } = useWorkspaceStore.getState();
              
              if (currentWorkspace?.id) {
                import('@/lib/routing/helpers').then(({ buildManagementPath }) => {
                  router.push(buildManagementPath(currentWorkspace.id, 'dashboard'));
                });
              } else {
                router.push('/workspaces');
              }
            });
          } else {
            router.push('/dashboard');
          }
        }
        return;
      }

      // 2. returnUrl이 없으면 세션 스토리지에서 마지막 경로 확인
      if (typeof window !== 'undefined') {
        const lastPath = sessionStorage.getItem('lastPath');
        if (lastPath && lastPath !== '/login' && lastPath !== '/register' && lastPath !== '/setup') {
          sessionStorage.removeItem('authRedirected');
          router.push(lastPath);
          return;
        }
      }

      // 3. 마지막 경로도 없으면 workspace가 있으면 해당 workspace의 dashboard로, 없으면 /workspaces로 리다이렉트
      sessionStorage.removeItem('authRedirected');
      if (typeof window !== 'undefined') {
        import('@/store/workspace').then(({ useWorkspaceStore }) => {
          const { currentWorkspace } = useWorkspaceStore.getState();
          
          if (currentWorkspace?.id) {
            import('@/lib/routing/helpers').then(({ buildManagementPath }) => {
              router.push(buildManagementPath(currentWorkspace.id, 'dashboard'));
            });
          } else {
            router.push('/workspaces');
          }
        });
      } else {
        router.push('/dashboard');
      }
    },
        onError: (_error) => {
      // Error is handled by the hook's error state
    },
    getErrorMessage: ErrorHandler.getUserFriendlyMessage,
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

