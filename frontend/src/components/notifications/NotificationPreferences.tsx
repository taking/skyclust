/**
 * Notification Preferences Component
 * 알림 설정 컴포넌트
 */

'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Bell, Save, TestTube } from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Switch } from '@/components/ui/switch';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import {
  useNotificationPreferences,
  useUpdateNotificationPreferences,
  useSendTestNotification,
} from '@/hooks/use-notifications';
import type { NotificationPreferences } from '@/lib/types/notification';

const preferencesSchema = z.object({
  email_enabled: z.boolean(),
  push_enabled: z.boolean(),
  browser_enabled: z.boolean(),
  in_app_enabled: z.boolean(),
  system_notifications: z.boolean(),
  vm_notifications: z.boolean(),
  cost_notifications: z.boolean(),
  security_notifications: z.boolean(),
  low_priority_enabled: z.boolean(),
  medium_priority_enabled: z.boolean(),
  high_priority_enabled: z.boolean(),
  urgent_priority_enabled: z.boolean(),
  quiet_hours_start: z.string().optional(),
  quiet_hours_end: z.string().optional(),
  timezone: z.string(),
});

type PreferencesFormData = z.infer<typeof preferencesSchema>;

export function NotificationPreferencesComponent() {
  const { data: preferences, isLoading } = useNotificationPreferences();
  const updatePreferencesMutation = useUpdateNotificationPreferences();
  const sendTestNotificationMutation = useSendTestNotification();

  const form = useForm<PreferencesFormData>({
    resolver: zodResolver(preferencesSchema),
    defaultValues: {
      email_enabled: true,
      push_enabled: true,
      browser_enabled: true,
      in_app_enabled: true,
      system_notifications: true,
      vm_notifications: true,
      cost_notifications: true,
      security_notifications: true,
      low_priority_enabled: true,
      medium_priority_enabled: true,
      high_priority_enabled: true,
      urgent_priority_enabled: true,
      timezone: 'UTC',
    },
  });

  // 데이터 로드 시 폼 업데이트
  if (preferences && !form.formState.isDirty) {
    form.reset({
      email_enabled: preferences.email_enabled,
      push_enabled: preferences.push_enabled,
      browser_enabled: preferences.browser_enabled,
      in_app_enabled: preferences.in_app_enabled,
      system_notifications: preferences.system_notifications,
      vm_notifications: preferences.vm_notifications,
      cost_notifications: preferences.cost_notifications,
      security_notifications: preferences.security_notifications,
      low_priority_enabled: preferences.low_priority_enabled,
      medium_priority_enabled: preferences.medium_priority_enabled,
      high_priority_enabled: preferences.high_priority_enabled,
      urgent_priority_enabled: preferences.urgent_priority_enabled,
      quiet_hours_start: preferences.quiet_hours_start || '',
      quiet_hours_end: preferences.quiet_hours_end || '',
      timezone: preferences.timezone,
    });
  }

  const onSubmit = async (data: PreferencesFormData) => {
    await updatePreferencesMutation.mutateAsync(data);
  };

  const handleSendTestNotification = async () => {
    await sendTestNotificationMutation.mutateAsync({
      type: 'info',
      title: '테스트 알림',
      message: '알림 설정이 올바르게 작동하는지 확인하는 테스트 알림입니다.',
      category: 'system',
      priority: 'medium',
    });
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>알림 설정</CardTitle>
          <CardDescription>알림 수신 방식을 설정하세요.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="flex items-center justify-between">
                <div className="space-y-1">
                  <div className="h-4 w-32 bg-gray-200 rounded animate-pulse" />
                  <div className="h-3 w-48 bg-gray-100 rounded animate-pulse" />
                </div>
                <div className="h-6 w-12 bg-gray-200 rounded animate-pulse" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Bell className="h-5 w-5" />
            알림 설정
          </CardTitle>
          <CardDescription>
            알림 수신 방식을 설정하세요.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              {/* 전송 방식 설정 */}
              <div className="space-y-4">
                <h3 className="text-lg font-semibold">전송 방식</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <FormField
                    control={form.control}
                    name="email_enabled"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                        <div className="space-y-0.5">
                          <FormLabel className="text-base">이메일 알림</FormLabel>
                          <FormDescription>
                            이메일로 알림을 받습니다.
                          </FormDescription>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="push_enabled"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                        <div className="space-y-0.5">
                          <FormLabel className="text-base">푸시 알림</FormLabel>
                          <FormDescription>
                            모바일 앱으로 푸시 알림을 받습니다.
                          </FormDescription>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="browser_enabled"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                        <div className="space-y-0.5">
                          <FormLabel className="text-base">브라우저 알림</FormLabel>
                          <FormDescription>
                            브라우저에서 알림을 받습니다.
                          </FormDescription>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="in_app_enabled"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                        <div className="space-y-0.5">
                          <FormLabel className="text-base">앱 내 알림</FormLabel>
                          <FormDescription>
                            웹 앱 내에서 알림을 받습니다.
                          </FormDescription>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                </div>
              </div>

              <Separator />

              {/* 카테고리별 설정 */}
              <div className="space-y-4">
                <h3 className="text-lg font-semibold">카테고리별 알림</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <FormField
                    control={form.control}
                    name="system_notifications"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                        <div className="space-y-0.5">
                          <FormLabel className="text-base">시스템 알림</FormLabel>
                          <FormDescription>
                            시스템 상태 및 업데이트 알림
                          </FormDescription>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="vm_notifications"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                        <div className="space-y-0.5">
                          <FormLabel className="text-base">VM 알림</FormLabel>
                          <FormDescription>
                            가상머신 상태 변경 알림
                          </FormDescription>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="cost_notifications"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                        <div className="space-y-0.5">
                          <FormLabel className="text-base">비용 알림</FormLabel>
                          <FormDescription>
                            비용 및 예산 관련 알림
                          </FormDescription>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="security_notifications"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                        <div className="space-y-0.5">
                          <FormLabel className="text-base">보안 알림</FormLabel>
                          <FormDescription>
                            보안 관련 중요 알림
                          </FormDescription>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                </div>
              </div>

              <Separator />

              {/* 우선순위별 설정 */}
              <div className="space-y-4">
                <h3 className="text-lg font-semibold">우선순위별 알림</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <FormField
                    control={form.control}
                    name="urgent_priority_enabled"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                        <div className="space-y-0.5">
                          <FormLabel className="text-base text-red-600">긴급</FormLabel>
                          <FormDescription>
                            즉시 확인이 필요한 알림
                          </FormDescription>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="high_priority_enabled"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                        <div className="space-y-0.5">
                          <FormLabel className="text-base text-orange-600">높음</FormLabel>
                          <FormDescription>
                            중요하지만 긴급하지 않은 알림
                          </FormDescription>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="medium_priority_enabled"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                        <div className="space-y-0.5">
                          <FormLabel className="text-base text-yellow-600">보통</FormLabel>
                          <FormDescription>
                            일반적인 정보 알림
                          </FormDescription>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="low_priority_enabled"
                    render={({ field }) => (
                      <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                        <div className="space-y-0.5">
                          <FormLabel className="text-base text-green-600">낮음</FormLabel>
                          <FormDescription>
                            참고용 정보 알림
                          </FormDescription>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </FormItem>
                    )}
                  />
                </div>
              </div>

              <Separator />

              {/* 시간 설정 */}
              <div className="space-y-4">
                <h3 className="text-lg font-semibold">시간 설정</h3>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <FormField
                    control={form.control}
                    name="quiet_hours_start"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>조용한 시간 시작</FormLabel>
                        <FormControl>
                          <Input
                            type="time"
                            {...field}
                            placeholder="HH:MM"
                          />
                        </FormControl>
                        <FormDescription>
                          이 시간 이후에는 알림을 받지 않습니다.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="quiet_hours_end"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>조용한 시간 종료</FormLabel>
                        <FormControl>
                          <Input
                            type="time"
                            {...field}
                            placeholder="HH:MM"
                          />
                        </FormControl>
                        <FormDescription>
                          이 시간부터 다시 알림을 받습니다.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="timezone"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>시간대</FormLabel>
                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="시간대 선택" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            <SelectItem value="UTC">UTC</SelectItem>
                            <SelectItem value="Asia/Seoul">Asia/Seoul (KST)</SelectItem>
                            <SelectItem value="America/New_York">America/New_York (EST)</SelectItem>
                            <SelectItem value="Europe/London">Europe/London (GMT)</SelectItem>
                            <SelectItem value="Asia/Tokyo">Asia/Tokyo (JST)</SelectItem>
                          </SelectContent>
                        </Select>
                        <FormDescription>
                          알림 시간 계산 기준
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
              </div>

              <div className="flex items-center justify-between pt-6">
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleSendTestNotification}
                  disabled={sendTestNotificationMutation.isPending}
                >
                  <TestTube className="mr-2 h-4 w-4" />
                  테스트 알림 전송
                </Button>

                <Button
                  type="submit"
                  disabled={updatePreferencesMutation.isPending}
                >
                  <Save className="mr-2 h-4 w-4" />
                  {updatePreferencesMutation.isPending ? '저장 중...' : '설정 저장'}
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  );
}
