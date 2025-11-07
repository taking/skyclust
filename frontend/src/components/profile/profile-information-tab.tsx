/**
 * Profile Information Tab Component
 * 프로필 페이지의 개인정보 탭 컴포넌트
 */

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Save } from 'lucide-react';
import { UseFormReturn } from 'react-hook-form';
import { useTranslation } from '@/hooks/use-translation';

interface ProfileInformationTabProps {
  form: UseFormReturn<{
    username: string;
    email: string;
  }>;
  onSubmit: (data: { username: string; email: string }) => void;
  isSubmitting: boolean;
}

export function ProfileInformationTab({
  form,
  onSubmit,
  isSubmitting,
}: ProfileInformationTabProps) {
  const { t } = useTranslation();

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('auth.personalInformation')}</CardTitle>
        <CardDescription>{t('auth.personalInformationDescription')}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="username">{t('auth.username')}</Label>
              <Input
                id="username"
                {...form.register('username')}
                placeholder={t('auth.usernamePlaceholder')}
              />
              {form.formState.errors.username && (
                <p className="text-sm text-red-600">
                  {form.formState.errors.username.message}
                </p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">{t('auth.email')}</Label>
              <Input
                id="email"
                type="email"
                {...form.register('email')}
                placeholder={t('auth.emailPlaceholder')}
              />
              {form.formState.errors.email && (
                <p className="text-sm text-red-600">
                  {form.formState.errors.email.message}
                </p>
              )}
            </div>
          </div>
          <Button 
            type="submit" 
            disabled={isSubmitting}
            className="w-full md:w-auto"
          >
            <Save className="mr-2 h-4 w-4" />
            {isSubmitting ? t('auth.saving') : t('auth.saveChanges')}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

