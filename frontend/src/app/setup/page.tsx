/**
 * Setup Page
 * 초기 설정 페이지 - 관리자 계정 생성
 * 
 * 시스템에 사용자가 없을 때 관리자 계정을 생성하는 페이지입니다.
 */

'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Form } from '@/components/ui/form';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { authService } from '@/services/auth';
import { RegisterForm } from '@/lib/types';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { useTranslation } from '@/hooks/use-translation';
import { createValidationSchemas } from '@/lib/validations';
import { useSystemInitialized } from '@/hooks/use-system-initialized';
import { useAuthStore } from '@/store/auth';
import { Spinner } from '@/components/ui/loading-states';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { CheckCircle2, AlertCircle } from 'lucide-react';

export default function SetupPage() {
  const [success, setSuccess] = useState(false);
  const router = useRouter();
  const { t } = useTranslation();
  const { login } = useAuthStore();
  
  // 시스템 초기화 상태 확인
  const { data: initStatus, isLoading: isLoadingStatus, error: statusError } = useSystemInitialized();

  // 이미 초기화된 경우 로그인 페이지로 리다이렉트
  useEffect(() => {
    if (initStatus?.initialized) {
      router.replace('/login');
    }
  }, [initStatus, router]);

  const { registerSchema } = createValidationSchemas(t);

  const {
    form,
    handleSubmit,
    isLoading,
    error,
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
      try {
        // 관리자 계정 생성 (첫 사용자는 자동으로 Admin 역할 부여)
        const response = await authService.register(data);
        
        // 자동 로그인
        login({
          token: response.token,
          user: response.user,
          expiresAt: response.expires_at,
        });

        setSuccess(true);
        
        // 성공 후 대시보드로 리다이렉트
        setTimeout(() => {
          router.push('/dashboard');
        }, 2000);
      } catch (err) {
        // 에러는 hook에서 처리됨
        throw err;
      }
    },
    onError: (_error) => {
      // Error is handled by the hook's error state
    },
    getErrorMessage: (err) => {
      if (err && typeof err === 'object' && 'response' in err) {
        const response = err.response as { data?: { error?: { message?: string } } };
        return response?.data?.error?.message || '관리자 계정 생성에 실패했습니다.';
      }
      return '관리자 계정 생성에 실패했습니다.';
    },
    resetOnSuccess: false,
  });

  // 로딩 중
  if (isLoadingStatus) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center space-y-4">
          <Spinner size="lg" label={t('setup.checkingStatus')} />
        </div>
      </div>
    );
  }

  // 에러 발생
  if (statusError) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6">
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                {t('setup.statusCheckError')}
              </AlertDescription>
            </Alert>
          </CardContent>
        </Card>
      </div>
    );
  }

  // 이미 초기화된 경우 (리다이렉트 대기 중)
  if (initStatus?.initialized) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center space-y-4">
          <Spinner size="lg" label={t('setup.redirecting')} />
        </div>
      </div>
    );
  }

  // 성공 화면
  if (success) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6">
            <div className="text-center">
              <div className="text-green-600 text-6xl mb-4 flex justify-center">
                <CheckCircle2 className="h-16 w-16" />
              </div>
              <h2 className="text-2xl font-bold text-gray-900 mb-2">
                {t('setup.successTitle')}
              </h2>
              <p className="text-gray-600 mb-4">
                {t('setup.successDescription')}
              </p>
              <p className="text-sm text-gray-500">
                {t('setup.redirectingToDashboard')}
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // 관리자 계정 생성 폼
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold text-center">
            {t('setup.title')}
          </CardTitle>
          <CardDescription className="text-center">
            {t('setup.description')}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={handleSubmit} className="space-y-4">
              <EnhancedField
                name="name"
                label={t('auth.fullName')}
                type="text"
                placeholder={t('auth.namePlaceholder')}
                required
                getFieldError={getFieldError}
                getFieldValidationState={getFieldValidationState}
              />
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
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}
              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading ? t('setup.creating') : t('setup.createAdmin')}
              </Button>
            </form>
          </Form>
          <div className="mt-4 text-center text-sm text-gray-500">
            {t('setup.note')}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

