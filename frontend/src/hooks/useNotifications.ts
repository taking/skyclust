/**
 * Notification Hooks
 * 알림 관련 React Query 훅
 */

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { notificationService, Notification, NotificationPreferences, NotificationStats } from '@/services/notification';
import { toast } from 'react-hot-toast';

export const useNotifications = (
  limit: number = 20,
  offset: number = 0,
  unreadOnly: boolean = false,
  category?: string,
  priority?: string
) => {
  return useQuery({
    queryKey: ['notifications', limit, offset, unreadOnly, category, priority],
    queryFn: () => notificationService.getNotifications(limit, offset, unreadOnly, category, priority),
  });
};

export const useNotification = (notificationId: string) => {
  return useQuery({
    queryKey: ['notifications', 'detail', notificationId],
    queryFn: () => notificationService.getNotification(notificationId),
    enabled: !!notificationId,
  });
};

export const useNotificationPreferences = () => {
  return useQuery({
    queryKey: ['notifications', 'preferences'],
    queryFn: () => notificationService.getNotificationPreferences(),
  });
};

export const useNotificationStats = () => {
  return useQuery({
    queryKey: ['notifications', 'stats'],
    queryFn: () => notificationService.getNotificationStats(),
  });
};

export const useMarkAsRead = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (notificationId: string) => notificationService.markAsRead(notificationId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      queryClient.invalidateQueries({ queryKey: ['notifications', 'stats'] });
    },
    onError: (error: unknown) => {
      const message = (error as { response?: { data?: { error?: string } } })?.response?.data?.error || '알림 읽음 처리에 실패했습니다.';
      toast.error(message);
    },
  });
};

export const useMarkAllAsRead = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => notificationService.markAllAsRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      queryClient.invalidateQueries({ queryKey: ['notifications', 'stats'] });
      toast.success('모든 알림을 읽음 처리했습니다.');
    },
    onError: (error: unknown) => {
      const message = (error as { response?: { data?: { error?: string } } })?.response?.data?.error || '알림 읽음 처리에 실패했습니다.';
      toast.error(message);
    },
  });
};

export const useDeleteNotification = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (notificationId: string) => notificationService.deleteNotification(notificationId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      queryClient.invalidateQueries({ queryKey: ['notifications', 'stats'] });
      toast.success('알림을 삭제했습니다.');
    },
    onError: (error: unknown) => {
      const message = (error as { response?: { data?: { error?: string } } })?.response?.data?.error || '알림 삭제에 실패했습니다.';
      toast.error(message);
    },
  });
};

export const useDeleteNotifications = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (notificationIds: string[]) => notificationService.deleteNotifications(notificationIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      queryClient.invalidateQueries({ queryKey: ['notifications', 'stats'] });
      toast.success('선택한 알림을 삭제했습니다.');
    },
    onError: (error: unknown) => {
      const message = (error as { response?: { data?: { error?: string } } })?.response?.data?.error || '알림 삭제에 실패했습니다.';
      toast.error(message);
    },
  });
};

export const useUpdateNotificationPreferences = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (preferences: Partial<NotificationPreferences>) => 
      notificationService.updateNotificationPreferences(preferences),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications', 'preferences'] });
      toast.success('알림 설정을 업데이트했습니다.');
    },
    onError: (error: unknown) => {
      const message = (error as { response?: { data?: { error?: string } } })?.response?.data?.error || '알림 설정 업데이트에 실패했습니다.';
      toast.error(message);
    },
  });
};

export const useSendTestNotification = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: {
      type: string;
      title: string;
      message: string;
      category?: string;
      priority?: string;
    }) => notificationService.sendTestNotification(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      queryClient.invalidateQueries({ queryKey: ['notifications', 'stats'] });
      toast.success('테스트 알림을 전송했습니다.');
    },
    onError: (error: unknown) => {
      const message = (error as { response?: { data?: { error?: string } } })?.response?.data?.error || '테스트 알림 전송에 실패했습니다.';
      toast.error(message);
    },
  });
};