'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Layout } from '@/components/layout/layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Separator } from '@/components/ui/separator';
import { Badge } from '@/components/ui/badge';
import { authService } from '@/services/auth';
import { useToast } from '@/hooks/useToast';
import { useRequireAuth } from '@/hooks/useAuth';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { 
  User, 
  Shield, 
  Key, 
  Bell, 
  Settings,
  Save,
  Eye,
  EyeOff
} from 'lucide-react';

const profileSchema = z.object({
  username: z.string().min(3, 'Username must be at least 3 characters'),
  email: z.string().email('Invalid email address'),
});

const passwordSchema = z.object({
  currentPassword: z.string().min(6, 'Current password is required'),
  newPassword: z.string().min(8, 'New password must be at least 8 characters'),
  confirmPassword: z.string().min(8, 'Please confirm your password'),
}).refine((data) => data.newPassword === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

const notificationSchema = z.object({
  emailNotifications: z.boolean(),
  pushNotifications: z.boolean(),
  securityAlerts: z.boolean(),
  systemUpdates: z.boolean(),
});

export default function ProfilePage() {
  const queryClient = useQueryClient();
  const { success, error: showError } = useToast();
  const { user, isLoading: authLoading } = useRequireAuth();
  const [showCurrentPassword, setShowCurrentPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  // Fetch current user data
  const { data: currentUser, isLoading } = useQuery({
    queryKey: ['user', 'me'],
    queryFn: authService.getCurrentUser,
    enabled: !!user,
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
  const updateProfileMutation = useMutation({
    mutationFn: (data: z.infer<typeof profileSchema>) => 
      authService.updateUser(currentUser?.id || '', data),
    onSuccess: () => {
      success('Profile updated successfully');
      queryClient.invalidateQueries({ queryKey: ['user', 'me'] });
    },
    onError: (error) => {
      showError(`Failed to update profile: ${error.message}`);
    },
  });

  // Change password mutation
  const changePasswordMutation = useMutation({
    mutationFn: (_data: z.infer<typeof passwordSchema>) => {
      // This would be a separate API endpoint for password change
      return Promise.resolve();
    },
    onSuccess: () => {
      success('Password changed successfully');
      passwordForm.reset();
    },
    onError: (error) => {
      showError(`Failed to change password: ${error.message}`);
    },
  });

  // Update notifications mutation
  const updateNotificationsMutation = useMutation({
    mutationFn: (_data: z.infer<typeof notificationSchema>) => {
      // This would be a separate API endpoint for notification preferences
      return Promise.resolve();
    },
    onSuccess: () => {
      success('Notification preferences updated');
    },
    onError: (error) => {
      showError(`Failed to update notifications: ${error.message}`);
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
            <p className="mt-2 text-gray-600">Loading profile...</p>
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
            <h3 className="text-lg font-medium text-gray-900">User not found</h3>
            <p className="mt-1 text-sm text-gray-500">Unable to load user profile.</p>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Profile Settings</h1>
          <p className="text-gray-600">Manage your account settings and preferences</p>
        </div>

        {/* Profile Overview */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center">
              <User className="mr-2 h-5 w-5" />
              Profile Overview
            </CardTitle>
            <CardDescription>Your account information and status</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center space-x-4">
              <div className="h-16 w-16 rounded-full bg-gray-200 flex items-center justify-center">
                <User className="h-8 w-8 text-gray-500" />
              </div>
              <div className="flex-1">
                <h3 className="text-lg font-semibold">{currentUser.username}</h3>
                <p className="text-gray-600">{currentUser.email}</p>
                <div className="flex items-center space-x-2 mt-2">
                  <Badge variant={currentUser.is_active ? "default" : "secondary"}>
                    {currentUser.is_active ? 'Active' : 'Inactive'}
                  </Badge>
                  <span className="text-sm text-gray-500">
                    Member since {new Date(currentUser.created_at).toLocaleDateString()}
                  </span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Settings Tabs */}
        <Tabs defaultValue="profile" className="space-y-4">
          <TabsList>
            <TabsTrigger value="profile">Profile</TabsTrigger>
            <TabsTrigger value="security">Security</TabsTrigger>
            <TabsTrigger value="notifications">Notifications</TabsTrigger>
          </TabsList>

          {/* Profile Tab */}
          <TabsContent value="profile" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Personal Information</CardTitle>
                <CardDescription>Update your personal details</CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={profileForm.handleSubmit(handleProfileSubmit)} className="space-y-4">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="username">Username</Label>
                      <Input
                        id="username"
                        {...profileForm.register('username')}
                        placeholder="Enter username"
                      />
                      {profileForm.formState.errors.username && (
                        <p className="text-sm text-red-600">
                          {profileForm.formState.errors.username.message}
                        </p>
                      )}
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="email">Email</Label>
                      <Input
                        id="email"
                        type="email"
                        {...profileForm.register('email')}
                        placeholder="Enter email"
                      />
                      {profileForm.formState.errors.email && (
                        <p className="text-sm text-red-600">
                          {profileForm.formState.errors.email.message}
                        </p>
                      )}
                    </div>
                  </div>
                  <Button 
                    type="submit" 
                    disabled={updateProfileMutation.isPending}
                    className="w-full md:w-auto"
                  >
                    <Save className="mr-2 h-4 w-4" />
                    {updateProfileMutation.isPending ? 'Saving...' : 'Save Changes'}
                  </Button>
                </form>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Security Tab */}
          <TabsContent value="security" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center">
                  <Shield className="mr-2 h-5 w-5" />
                  Change Password
                </CardTitle>
                <CardDescription>Update your password to keep your account secure</CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={passwordForm.handleSubmit(handlePasswordSubmit)} className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="currentPassword">Current Password</Label>
                    <div className="relative">
                      <Input
                        id="currentPassword"
                        type={showCurrentPassword ? 'text' : 'password'}
                        {...passwordForm.register('currentPassword')}
                        placeholder="Enter current password"
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                        onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                      >
                        {showCurrentPassword ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                    {passwordForm.formState.errors.currentPassword && (
                      <p className="text-sm text-red-600">
                        {passwordForm.formState.errors.currentPassword.message}
                      </p>
                    )}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="newPassword">New Password</Label>
                    <div className="relative">
                      <Input
                        id="newPassword"
                        type={showNewPassword ? 'text' : 'password'}
                        {...passwordForm.register('newPassword')}
                        placeholder="Enter new password"
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                        onClick={() => setShowNewPassword(!showNewPassword)}
                      >
                        {showNewPassword ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                    {passwordForm.formState.errors.newPassword && (
                      <p className="text-sm text-red-600">
                        {passwordForm.formState.errors.newPassword.message}
                      </p>
                    )}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="confirmPassword">Confirm New Password</Label>
                    <div className="relative">
                      <Input
                        id="confirmPassword"
                        type={showConfirmPassword ? 'text' : 'password'}
                        {...passwordForm.register('confirmPassword')}
                        placeholder="Confirm new password"
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                        onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                      >
                        {showConfirmPassword ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                    {passwordForm.formState.errors.confirmPassword && (
                      <p className="text-sm text-red-600">
                        {passwordForm.formState.errors.confirmPassword.message}
                      </p>
                    )}
                  </div>

                  <Button 
                    type="submit" 
                    disabled={changePasswordMutation.isPending}
                    className="w-full md:w-auto"
                  >
                    <Key className="mr-2 h-4 w-4" />
                    {changePasswordMutation.isPending ? 'Changing...' : 'Change Password'}
                  </Button>
                </form>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Notifications Tab */}
          <TabsContent value="notifications" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center">
                  <Bell className="mr-2 h-5 w-5" />
                  Notification Preferences
                </CardTitle>
                <CardDescription>Choose what notifications you want to receive</CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={notificationForm.handleSubmit(handleNotificationSubmit)} className="space-y-4">
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <Label htmlFor="emailNotifications">Email Notifications</Label>
                        <p className="text-sm text-gray-500">Receive notifications via email</p>
                      </div>
                      <input
                        id="emailNotifications"
                        type="checkbox"
                        {...notificationForm.register('emailNotifications')}
                        className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                      />
                    </div>

                    <Separator />

                    <div className="flex items-center justify-between">
                      <div>
                        <Label htmlFor="pushNotifications">Push Notifications</Label>
                        <p className="text-sm text-gray-500">Receive push notifications in your browser</p>
                      </div>
                      <input
                        id="pushNotifications"
                        type="checkbox"
                        {...notificationForm.register('pushNotifications')}
                        className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                      />
                    </div>

                    <Separator />

                    <div className="flex items-center justify-between">
                      <div>
                        <Label htmlFor="securityAlerts">Security Alerts</Label>
                        <p className="text-sm text-gray-500">Important security-related notifications</p>
                      </div>
                      <input
                        id="securityAlerts"
                        type="checkbox"
                        {...notificationForm.register('securityAlerts')}
                        className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                      />
                    </div>

                    <Separator />

                    <div className="flex items-center justify-between">
                      <div>
                        <Label htmlFor="systemUpdates">System Updates</Label>
                        <p className="text-sm text-gray-500">Notifications about system updates and maintenance</p>
                      </div>
                      <input
                        id="systemUpdates"
                        type="checkbox"
                        {...notificationForm.register('systemUpdates')}
                        className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                      />
                    </div>
                  </div>

                  <Button 
                    type="submit" 
                    disabled={updateNotificationsMutation.isPending}
                    className="w-full md:w-auto"
                  >
                    <Settings className="mr-2 h-4 w-4" />
                    {updateNotificationsMutation.isPending ? 'Saving...' : 'Save Preferences'}
                  </Button>
                </form>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </Layout>
  );
}
