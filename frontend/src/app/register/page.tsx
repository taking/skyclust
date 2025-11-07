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
import { useTranslation } from '@/hooks/use-translation';
import { createValidationSchemas } from '@/lib/validations';
import { TIME } from '@/lib/constants';

export default function RegisterPage() {
  const [success, setSuccess] = useState(false);
  const router = useRouter();
  const { t } = useTranslation();
  
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
      await authService.register(data);
      setSuccess(true);
      setTimeout(() => {
        router.push('/login');
      }, TIME.DELAY.REGISTER_REDIRECT);
    },
        onError: (_error) => {
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
                {t('auth.registrationSuccessful')}
              </h2>
              <p className="text-gray-600">
                {t('auth.registrationSuccessDescription')}
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
          <CardTitle className="text-2xl font-bold text-center">{t('auth.signUpTitle')}</CardTitle>
          <CardDescription className="text-center">
            {t('auth.signUpDescription')}
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
                <div className="text-sm text-red-600 text-center" role="alert">
                  {error}
                </div>
              )}
              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading ? t('auth.signingUp') : t('auth.signUp')}
              </Button>
            </form>
          </Form>
          <div className="mt-4 text-center text-sm">
            {t('auth.alreadyHaveAccount')}{' '}
            <Link href="/login" className="text-blue-600 hover:underline">
              {t('auth.signIn')}
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

