/**
 * Notification Actions Hook
 * 알림 액션 핸들러 통합 관리 훅
 * 
 * 알림 관련 모든 액션(읽음 처리, 삭제 등)을 통합 관리합니다.
 * 
 * @example
 * ```tsx
 * const {
 *   handleMarkAsRead,
 *   handleMarkAllAsRead,
 *   handleDeleteNotification,
 *   handleDeleteSelected,
 *   isMarkingAsRead,
 *   isMarkingAllAsRead,
 *   isDeleting,
 *   isDeletingSelected,
 * } = useNotificationActions({
 *   selectedNotifications,
 *   onSelectionClear: () => clearSelection(),
 * });
 * ```
 */

import { useCallback } from 'react';
import {
  useMarkAsRead,
  useMarkAllAsRead,
  useDeleteNotification,
  useDeleteNotifications,
} from '@/hooks/use-notifications';

export interface UseNotificationActionsOptions {
  /** 선택된 알림 ID 목록 (벌크 삭제용) */
  selectedNotifications?: string[];
  /** 선택 초기화 콜백 (벌크 삭제 후 호출) */
  onSelectionClear?: () => void;
}

export interface UseNotificationActionsReturn {
  /** 개별 알림 읽음 처리 핸들러 */
  handleMarkAsRead: (notificationId: string) => Promise<void>;
  /** 모든 알림 읽음 처리 핸들러 */
  handleMarkAllAsRead: () => Promise<void>;
  /** 개별 알림 삭제 핸들러 */
  handleDeleteNotification: (notificationId: string) => Promise<void>;
  /** 선택된 알림 일괄 삭제 핸들러 */
  handleDeleteSelected: () => Promise<void>;
  /** 개별 알림 읽음 처리 중 여부 */
  isMarkingAsRead: boolean;
  /** 모든 알림 읽음 처리 중 여부 */
  isMarkingAllAsRead: boolean;
  /** 개별 알림 삭제 중 여부 */
  isDeleting: boolean;
  /** 선택된 알림 일괄 삭제 중 여부 */
  isDeletingSelected: boolean;
}

/**
 * 알림 액션 핸들러 통합 관리 훅
 */
export function useNotificationActions(
  options: UseNotificationActionsOptions = {}
): UseNotificationActionsReturn {
  const { selectedNotifications = [], onSelectionClear } = options;

  const markAsReadMutation = useMarkAsRead();
  const markAllAsReadMutation = useMarkAllAsRead();
  const deleteNotificationMutation = useDeleteNotification();
  const deleteNotificationsMutation = useDeleteNotifications();

  const handleMarkAsRead = useCallback(
    async (notificationId: string) => {
      await markAsReadMutation.mutateAsync(notificationId);
    },
    [markAsReadMutation]
  );

  const handleMarkAllAsRead = useCallback(async () => {
    await markAllAsReadMutation.mutateAsync(undefined);
  }, [markAllAsReadMutation]);

  const handleDeleteNotification = useCallback(
    async (notificationId: string) => {
      await deleteNotificationMutation.mutateAsync(notificationId);
    },
    [deleteNotificationMutation]
  );

  const handleDeleteSelected = useCallback(async () => {
    if (selectedNotifications.length > 0) {
      await deleteNotificationsMutation.mutateAsync(selectedNotifications);
      onSelectionClear?.();
    }
  }, [selectedNotifications, deleteNotificationsMutation, onSelectionClear]);

  return {
    handleMarkAsRead,
    handleMarkAllAsRead,
    handleDeleteNotification,
    handleDeleteSelected,
    isMarkingAsRead: markAsReadMutation.isPending,
    isMarkingAllAsRead: markAllAsReadMutation.isPending,
    isDeleting: deleteNotificationMutation.isPending,
    isDeletingSelected: deleteNotificationsMutation.isPending,
  };
}

