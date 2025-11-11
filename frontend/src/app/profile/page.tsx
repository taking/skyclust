'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Layout } from '@/components/layout/layout';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { authService } from '@/services/auth';
import { useToast } from '@/hooks/use-toast';
import { useRequireAuth } from '@/hooks/use-auth';
import { WorkspaceRequired } from '@/components/common/workspace-required';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import { useTranslation } from '@/hooks/use-translation';
import { createValidationSchemas } from '@/lib/validation';
import { ProfileOverviewCard } from '@/components/profile/profile-overview-card';
import { ProfileInformationTab } from '@/components/profile/profile-information-tab';
import { ProfileSecurityTab } from '@/components/profile/profile-security-tab';
import { ProfileNotificationsTab } from '@/components/profile/profile-notifications-tab';
import { useStandardMutation } from '@/hooks/use-standard-mutation';

export default function ProfilePage() {
  const { t } = useTranslation();
  
  const { profileSchema, passwordSchema, notificationSchema } = createValidationSchemas(t);
  const { error: showError } = useToast();
  const { user, isLoading: authLoading } = useRequireAuth();
  const [showCurrentPassword, setShowCurrentPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  // Fetch current user data (사용자 정보는 자주 변경되지 않음)
  const { data: currentUser, isLoading } = useQuery({
    queryKey: queryKeys.user.me(),
    queryFn: () => authService.getCurrentUser(),
    enabled: !!user,
    staleTime: CACHE_TIMES.RESOURCE, // 5분 - 사용자 정보는 비교적 안정적
    gcTime: GC_TIMES.MEDIUM, // 10분 - GC 시간 (15분 대신 10분으로 조정)
  });

  // Profile form
  const profileForm = useForm<z.infer<typeof profileSchema>>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      username: currentUser?.username || '',
      email: currentUser?.email || '',
    },
  });

  // Password form
  const passwordForm = useForm<z.infer<typeof passwordSchema>>({
    resolver: zodResolver(passwordSchema),
  });

  // Notification form
  const notificationForm = useForm<z.infer<typeof notificationSchema>>({
    resolver: zodResolver(notificationSchema),
    defaultValues: {
      emailNotifications: true,
      pushNotifications: true,
      securityAlerts: true,
      systemUpdates: false,
    },
  });

  // Update profile mutation
  const updateProfileMutation = useStandardMutation({
    mutationFn: (data: z.infer<typeof profileSchema>) => 
      authService.updateUser(currentUser?.id || '', data),
    invalidateQueries: [queryKeys.user.me()],
    successMessage: t('auth.profileUpdatedSuccessfully'),
    errorContext: { operation: 'updateProfile', resource: 'User' },
    onError: (error) => {
      showError(t('auth.failedToUpdateProfile', { error: error instanceof Error ? error.message : String(error) }));
    },
  });

  // Change password mutation
  const changePasswordMutation = useStandardMutation({
    mutationFn: (_data: z.infer<typeof passwordSchema>) => {
      // This would be a separate API endpoint for password change
      return Promise.resolve();
    },
    successMessage: t('auth.passwordChangedSuccessfully'),
    errorContext: { operation: 'changePassword', resource: 'User' },
    onSuccess: () => {
      passwordForm.reset();
    },
    onError: (error) => {
      showError(t('auth.failedToChangePassword', { error: error instanceof Error ? error.message : String(error) }));
    },
  });

  // Update notifications mutation
  const updateNotificationsMutation = useStandardMutation({
    mutationFn: (_data: z.infer<typeof notificationSchema>) => {
      // This would be a separate API endpoint for notification preferences
      return Promise.resolve();
    },
    successMessage: t('auth.notificationPreferencesUpdated'),
    errorContext: { operation: 'updateNotifications', resource: 'User' },
    onError: (error) => {
      showError(t('auth.failedToUpdateNotifications', { error: error instanceof Error ? error.message : String(error) }));
    },
  });

  const handleProfileSubmit = (data: z.infer<typeof profileSchema>) => {
    updateProfileMutation.mutate(data);
  };

  const handlePasswordSubmit = (data: z.infer<typeof passwordSchema>) => {
    changePasswordMutation.mutate(data);
  };

  const handleNotificationSubmit = (data: z.infer<typeof notificationSchema>) => {
    updateNotificationsMutation.mutate(data);
  };

  if (authLoading || isLoading) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
            <p className="mt-2 text-gray-600">{t('auth.loadingProfile')}</p>
          </div>
        </div>
      </Layout>
    );
  }

  if (!currentUser) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <h3 className="text-lg font-medium text-gray-900">{t('auth.userNotFound')}</h3>
            <p className="mt-1 text-sm text-gray-500">{t('auth.unableToLoadProfile')}</p>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <WorkspaceRequired>
      <Layout>
        <div className="space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{t('auth.profileSettings')}</h1>
          <p className="text-gray-600">{t('auth.profileSettingsDescription')}</p>
        </div>

        {/* Profile Overview */}
        <ProfileOverviewCard user={currentUser} />

        {/* Settings Tabs */}
        <Tabs defaultValue="profile" className="space-y-4">
          <TabsList>
            <TabsTrigger value="profile">{t('auth.profileTab')}</TabsTrigger>
            <TabsTrigger value="security">{t('auth.securityTab')}</TabsTrigger>
            <TabsTrigger value="notifications">{t('auth.notificationsTab')}</TabsTrigger>
          </TabsList>

          {/* Profile Tab */}
          <TabsContent value="profile" className="space-y-4">
            <ProfileInformationTab
              form={profileForm}
              onSubmit={handleProfileSubmit}
              isSubmitting={updateProfileMutation.isPending}
            />
          </TabsContent>

          {/* Security Tab */}
          <TabsContent value="security" className="space-y-4">
            <ProfileSecurityTab
              form={passwordForm}
              onSubmit={handlePasswordSubmit}
              isSubmitting={changePasswordMutation.isPending}
              showCurrentPassword={showCurrentPassword}
              showNewPassword={showNewPassword}
              showConfirmPassword={showConfirmPassword}
              onToggleCurrentPassword={() => setShowCurrentPassword(!showCurrentPassword)}
              onToggleNewPassword={() => setShowNewPassword(!showNewPassword)}
              onToggleConfirmPassword={() => setShowConfirmPassword(!showConfirmPassword)}
            />
          </TabsContent>

          {/* Notifications Tab */}
          <TabsContent value="notifications" className="space-y-4">
            <ProfileNotificationsTab
              form={notificationForm}
              onSubmit={handleNotificationSubmit}
              isSubmitting={updateNotificationsMutation.isPending}
            />
          </TabsContent>
        </Tabs>
      </div>
    </Layout>
    </WorkspaceRequired>
  );
}
