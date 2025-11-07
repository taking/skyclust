/**
 * Profile Notifications Tab Component
 * 프로필 페이지의 알림 설정 탭 컴포넌트
 */

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { Bell, Settings } from 'lucide-react';
import { UseFormReturn } from 'react-hook-form';
import { useTranslation } from '@/hooks/use-translation';

interface ProfileNotificationsTabProps {
  form: UseFormReturn<{
    emailNotifications: boolean;
    pushNotifications: boolean;
    securityAlerts: boolean;
    systemUpdates: boolean;
  }>;
  onSubmit: (data: {
    emailNotifications: boolean;
    pushNotifications: boolean;
    securityAlerts: boolean;
    systemUpdates: boolean;
  }) => void;
  isSubmitting: boolean;
}

export function ProfileNotificationsTab({
  form,
  onSubmit,
  isSubmitting,
}: ProfileNotificationsTabProps) {
  const { t } = useTranslation();

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Bell className="mr-2 h-5 w-5" />
          {t('auth.notificationPreferences')}
        </CardTitle>
        <CardDescription>{t('auth.notificationPreferencesDescription')}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <Label htmlFor="emailNotifications">{t('auth.emailNotifications')}</Label>
                <p className="text-sm text-gray-500">{t('auth.emailNotificationsDescription')}</p>
              </div>
              <input
                id="emailNotifications"
                type="checkbox"
                {...form.register('emailNotifications')}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
            </div>

            <Separator />

            <div className="flex items-center justify-between">
              <div>
                <Label htmlFor="pushNotifications">{t('auth.pushNotifications')}</Label>
                <p className="text-sm text-gray-500">{t('auth.pushNotificationsDescription')}</p>
              </div>
              <input
                id="pushNotifications"
                type="checkbox"
                {...form.register('pushNotifications')}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
            </div>

            <Separator />

            <div className="flex items-center justify-between">
              <div>
                <Label htmlFor="securityAlerts">{t('auth.securityAlerts')}</Label>
                <p className="text-sm text-gray-500">{t('auth.securityAlertsDescription')}</p>
              </div>
              <input
                id="securityAlerts"
                type="checkbox"
                {...form.register('securityAlerts')}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
            </div>

            <Separator />

            <div className="flex items-center justify-between">
              <div>
                <Label htmlFor="systemUpdates">{t('auth.systemUpdates')}</Label>
                <p className="text-sm text-gray-500">{t('auth.systemUpdatesDescription')}</p>
              </div>
              <input
                id="systemUpdates"
                type="checkbox"
                {...form.register('systemUpdates')}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
            </div>
          </div>

          <Button 
            type="submit" 
            disabled={isSubmitting}
            className="w-full md:w-auto"
          >
            <Settings className="mr-2 h-4 w-4" />
            {isSubmitting ? t('auth.saving') : t('auth.savePreferences')}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

