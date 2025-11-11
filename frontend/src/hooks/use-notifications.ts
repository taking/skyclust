/**
 * Notification Hooks
 * 알림 관련 React Query 훅
 */

import { useQuery } from '@tanstack/react-query';
import { notificationService } from '@/services/notification';
import type { NotificationPreferences } from '@/lib/types/notification';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import { useStandardMutation } from './use-standard-mutation';

export const useNotifications = (
  limit: number = 20,
  offset: number = 0,
  unreadOnly: boolean = false,
  category?: string,
  priority?: string
) => {
  return useQuery({
    queryKey: queryKeys.notifications.list(limit, offset, unreadOnly, category, priority),
    queryFn: () => notificationService.getNotifications(limit, offset, unreadOnly, category, priority),
    staleTime: CACHE_TIMES.REALTIME, // 30초 - 알림은 자주 업데이트될 수 있음
    gcTime: GC_TIMES.SHORT, // 5 minutes - GC 시간
    // refetchInterval 제거: SSE system-notification 이벤트로 자동 업데이트
  });
};

export const useNotification = (notificationId: string) => {
  return useQuery({
    queryKey: queryKeys.notifications.detail(notificationId),
    queryFn: () => notificationService.getNotification(notificationId),
    enabled: !!notificationId,
    staleTime: CACHE_TIMES.RESOURCE, // 5 minutes - 알림 상세는 자주 변경되지 않음
    gcTime: GC_TIMES.MEDIUM, // 10 minutes - GC 시간
  });
};

export const useNotificationPreferences = () => {
  return useQuery({
    queryKey: queryKeys.notifications.preferences(),
    queryFn: () => notificationService.getNotificationPreferences(),
    staleTime: CACHE_TIMES.STATIC, // 30 minutes - 설정은 자주 변경되지 않음
    gcTime: GC_TIMES.LONG, // 30 minutes - GC 시간 (1시간 대신 30분으로 조정)
  });
};

export const useNotificationStats = () => {
  return useQuery({
    queryKey: queryKeys.notifications.stats(),
    queryFn: () => notificationService.getNotificationStats(),
    staleTime: CACHE_TIMES.REALTIME, // 30초 - 통계는 실시간성 필요
    gcTime: GC_TIMES.SHORT, // 5 minutes - GC 시간
    // refetchInterval 제거: SSE system-notification 이벤트로 자동 업데이트
  });
};

export const useMarkAsRead = () => {
  return useStandardMutation({
    mutationFn: (notificationId: string) => notificationService.markAsRead(notificationId),
    invalidateQueries: [
      queryKeys.notifications.all,
      queryKeys.notifications.stats(),
    ],
    successMessage: '알림을 읽음 처리했습니다.',
    errorContext: { operation: 'markAsRead', resource: 'Notification' },
  });
};

export const useMarkAllAsRead = () => {
  return useStandardMutation({
    mutationFn: () => notificationService.markAllAsRead(),
    invalidateQueries: [
      queryKeys.notifications.all,
      queryKeys.notifications.stats(),
    ],
    successMessage: '모든 알림을 읽음 처리했습니다.',
    errorContext: { operation: 'markAllAsRead', resource: 'Notification' },
  });
};

export const useDeleteNotification = () => {
  return useStandardMutation({
    mutationFn: (notificationId: string) => notificationService.deleteNotification(notificationId),
    invalidateQueries: [
      queryKeys.notifications.all,
      queryKeys.notifications.stats(),
    ],
    successMessage: '알림을 삭제했습니다.',
    errorContext: { operation: 'deleteNotification', resource: 'Notification' },
  });
};

export const useDeleteNotifications = () => {
  return useStandardMutation({
    mutationFn: (notificationIds: string[]) => notificationService.deleteNotifications(notificationIds),
    invalidateQueries: [
      queryKeys.notifications.all,
      queryKeys.notifications.stats(),
    ],
    successMessage: '선택한 알림을 삭제했습니다.',
    errorContext: { operation: 'deleteNotifications', resource: 'Notification' },
  });
};

export const useUpdateNotificationPreferences = () => {
  return useStandardMutation({
    mutationFn: (preferences: Partial<NotificationPreferences>) => 
      notificationService.updateNotificationPreferences(preferences),
    invalidateQueries: [
      queryKeys.notifications.preferences(),
    ],
    successMessage: '알림 설정을 업데이트했습니다.',
    errorContext: { operation: 'updateNotificationPreferences', resource: 'Notification' },
  });
};

export const useSendTestNotification = () => {
  return useStandardMutation({
    mutationFn: (data: {
      type: string;
      title: string;
      message: string;
      category?: string;
      priority?: string;
    }) => notificationService.sendTestNotification(data),
    invalidateQueries: [
      queryKeys.notifications.all,
      queryKeys.notifications.stats(),
    ],
    successMessage: '테스트 알림을 전송했습니다.',
    errorContext: { operation: 'sendTestNotification', resource: 'Notification' },
  });
};

