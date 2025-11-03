/**
 * Notification Hooks
 * 알림 관련 React Query 훅
 */

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { notificationService } from '@/services/notification';
import type { Notification, NotificationPreferences, NotificationStats } from '@/lib/types/notification';
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
    staleTime: 30 * 1000, // 30초 - 알림은 자주 업데이트될 수 있음
    gcTime: 5 * 60 * 1000, // 5 minutes - GC 시간
    refetchInterval: 30000, // 30초마다 refetch (실시간 알림)
    refetchIntervalInBackground: false, // 백그라운드 polling 비활성화
  });
};

export const useNotification = (notificationId: string) => {
  return useQuery({
    queryKey: ['notifications', 'detail', notificationId],
    queryFn: () => notificationService.getNotification(notificationId),
    enabled: !!notificationId,
    staleTime: 5 * 60 * 1000, // 5 minutes - 알림 상세는 자주 변경되지 않음
    gcTime: 10 * 60 * 1000, // 10 minutes - GC 시간
  });
};

export const useNotificationPreferences = () => {
  return useQuery({
    queryKey: ['notifications', 'preferences'],
    queryFn: () => notificationService.getNotificationPreferences(),
    staleTime: 30 * 60 * 1000, // 30 minutes - 설정은 자주 변경되지 않음
    gcTime: 60 * 60 * 1000, // 1 hour - GC 시간
  });
};

export const useNotificationStats = () => {
  return useQuery({
    queryKey: ['notifications', 'stats'],
    queryFn: () => notificationService.getNotificationStats(),
    staleTime: 30 * 1000, // 30초 - 통계는 실시간성 필요
    gcTime: 5 * 60 * 1000, // 5 minutes - GC 시간
    refetchInterval: 30000, // 30초마다 refetch
    refetchIntervalInBackground: false, // 백그라운드 polling 비활성화
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

