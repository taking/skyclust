/**
 * Notification Filters Hook
 * 알림 필터링 상태 관리 훅
 * 
 * 알림 목록의 필터링 상태(unreadOnly, category, priority)를 관리합니다.
 * 
 * @example
 * ```tsx
 * const {
 *   unreadOnly,
 *   category,
 *   priority,
 *   setUnreadOnly,
 *   setCategory,
 *   setPriority,
 *   clearFilters,
 * } = useNotificationFilters();
 * ```
 */

import { useState, useCallback } from 'react';

export interface UseNotificationFiltersReturn {
  /** 읽지 않은 알림만 표시 여부 */
  unreadOnly: boolean;
  /** 선택된 카테고리 */
  category: string;
  /** 선택된 우선순위 */
  priority: string;
  /** 읽지 않은 알림만 표시 여부 설정 */
  setUnreadOnly: (value: boolean) => void;
  /** 카테고리 설정 */
  setCategory: (value: string) => void;
  /** 우선순위 설정 */
  setPriority: (value: string) => void;
  /** 모든 필터 초기화 */
  clearFilters: () => void;
}

/**
 * 알림 필터링 상태 관리 훅
 */
export function useNotificationFilters(): UseNotificationFiltersReturn {
  const [unreadOnly, setUnreadOnlyState] = useState(false);
  const [category, setCategoryState] = useState<string>('');
  const [priority, setPriorityState] = useState<string>('');

  const setUnreadOnly = useCallback((value: boolean) => {
    setUnreadOnlyState(value);
  }, []);

  const setCategory = useCallback((value: string) => {
    setCategoryState(value);
  }, []);

  const setPriority = useCallback((value: string) => {
    setPriorityState(value);
  }, []);

  const clearFilters = useCallback(() => {
    setUnreadOnlyState(false);
    setCategoryState('');
    setPriorityState('');
  }, []);

  return {
    unreadOnly,
    category,
    priority,
    setUnreadOnly,
    setCategory,
    setPriority,
    clearFilters,
  };
}

