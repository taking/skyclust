/**
 * Profile Security Tab Component
 * 프로필 페이지의 보안 탭 컴포넌트
 */

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Shield, Key, Eye, EyeOff } from 'lucide-react';
import { UseFormReturn } from 'react-hook-form';
import { useTranslation } from '@/hooks/use-translation';

interface ProfileSecurityTabProps {
  form: UseFormReturn<{
    currentPassword: string;
    newPassword: string;
    confirmPassword: string;
  }>;
  onSubmit: (data: { currentPassword: string; newPassword: string; confirmPassword: string }) => void;
  isSubmitting: boolean;
  showCurrentPassword: boolean;
  showNewPassword: boolean;
  showConfirmPassword: boolean;
  onToggleCurrentPassword: () => void;
  onToggleNewPassword: () => void;
  onToggleConfirmPassword: () => void;
}

export function ProfileSecurityTab({
  form,
  onSubmit,
  isSubmitting,
  showCurrentPassword,
  showNewPassword,
  showConfirmPassword,
  onToggleCurrentPassword,
  onToggleNewPassword,
  onToggleConfirmPassword,
}: ProfileSecurityTabProps) {
  const { t } = useTranslation();

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Shield className="mr-2 h-5 w-5" />
          {t('auth.changePassword')}
        </CardTitle>
        <CardDescription>{t('auth.changePasswordDescription')}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="currentPassword">{t('auth.currentPassword')}</Label>
            <div className="relative">
              <Input
                id="currentPassword"
                type={showCurrentPassword ? 'text' : 'password'}
                {...form.register('currentPassword')}
                placeholder={t('auth.currentPasswordPlaceholder')}
              />
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                onClick={onToggleCurrentPassword}
              >
                {showCurrentPassword ? (
                  <EyeOff className="h-4 w-4" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
              </Button>
            </div>
            {form.formState.errors.currentPassword && (
              <p className="text-sm text-red-600">
                {form.formState.errors.currentPassword.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="newPassword">{t('auth.newPassword')}</Label>
            <div className="relative">
              <Input
                id="newPassword"
                type={showNewPassword ? 'text' : 'password'}
                {...form.register('newPassword')}
                placeholder={t('auth.newPasswordPlaceholder')}
              />
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                onClick={onToggleNewPassword}
              >
                {showNewPassword ? (
                  <EyeOff className="h-4 w-4" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
              </Button>
            </div>
            {form.formState.errors.newPassword && (
              <p className="text-sm text-red-600">
                {form.formState.errors.newPassword.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="confirmPassword">{t('auth.confirmNewPassword')}</Label>
            <div className="relative">
              <Input
                id="confirmPassword"
                type={showConfirmPassword ? 'text' : 'password'}
                {...form.register('confirmPassword')}
                placeholder={t('auth.confirmNewPasswordPlaceholder')}
              />
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                onClick={onToggleConfirmPassword}
              >
                {showConfirmPassword ? (
                  <EyeOff className="h-4 w-4" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
              </Button>
            </div>
            {form.formState.errors.confirmPassword && (
              <p className="text-sm text-red-600">
                {form.formState.errors.confirmPassword.message}
              </p>
            )}
          </div>

          <Button 
            type="submit" 
            disabled={isSubmitting}
            className="w-full md:w-auto"
          >
            <Key className="mr-2 h-4 w-4" />
            {isSubmitting ? t('auth.changing') : t('auth.changePassword')}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

