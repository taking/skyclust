/**
 * Notification Selection Hook
 * 알림 선택 상태 관리 훅
 * 
 * 알림 목록에서 여러 알림을 선택하고 관리하는 로직을 제공합니다.
 * 
 * @example
 * ```tsx
 * const {
 *   selectedNotifications,
 *   handleSelectNotification,
 *   handleSelectAll,
 *   clearSelection,
 * } = useNotificationSelection(notifications);
 * ```
 */

import { useState, useCallback } from 'react';
import type { Notification } from '@/lib/types/notification';

export interface UseNotificationSelectionReturn {
  /** 선택된 알림 ID 목록 */
  selectedNotifications: string[];
  /** 개별 알림 선택/해제 핸들러 */
  handleSelectNotification: (notificationId: string, checked: boolean) => void;
  /** 전체 선택/해제 핸들러 */
  handleSelectAll: (checked: boolean) => void;
  /** 선택 초기화 */
  clearSelection: () => void;
  /** 모든 알림 선택 */
  selectAll: () => void;
}

export interface UseNotificationSelectionOptions {
  /** 현재 표시된 알림 목록 */
  notifications?: Notification[];
}

/**
 * 알림 선택 상태 관리 훅
 */
export function useNotificationSelection(
  options: UseNotificationSelectionOptions = {}
): UseNotificationSelectionReturn {
  const { notifications = [] } = options;
  const [selectedNotifications, setSelectedNotifications] = useState<string[]>([]);

  const handleSelectNotification = useCallback((notificationId: string, checked: boolean) => {
    if (checked) {
      setSelectedNotifications(prev => [...prev, notificationId]);
    } else {
      setSelectedNotifications(prev => prev.filter(id => id !== notificationId));
    }
  }, []);

  const handleSelectAll = useCallback((checked: boolean) => {
    if (checked && notifications.length > 0) {
      setSelectedNotifications(notifications.map(n => n.id));
    } else {
      setSelectedNotifications([]);
    }
  }, [notifications]);

  const clearSelection = useCallback(() => {
    setSelectedNotifications([]);
  }, []);

  const selectAll = useCallback(() => {
    if (notifications.length > 0) {
      setSelectedNotifications(notifications.map(n => n.id));
    }
  }, [notifications]);

  return {
    selectedNotifications,
    handleSelectNotification,
    handleSelectAll,
    clearSelection,
    selectAll,
  };
}

