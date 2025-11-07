/**
 * Cluster Tag Dialog Hook
 * 클러스터 태그 다이얼로그 상태 관리 훅
 * 
 * @example
 * ```tsx
 * const {
 *   isOpen,
 *   tagKey,
 *   tagValue,
 *   openDialog,
 *   closeDialog,
 *   setTagKey,
 *   setTagValue,
 *   reset,
 * } = useClusterTagDialog();
 * ```
 */

import { useState, useCallback } from 'react';

export interface UseClusterTagDialogReturn {
  /** 다이얼로그 열림 상태 */
  isOpen: boolean;
  /** 태그 키 */
  tagKey: string;
  /** 태그 값 */
  tagValue: string;
  /** 다이얼로그 열기 */
  openDialog: () => void;
  /** 다이얼로그 닫기 */
  closeDialog: () => void;
  /** 태그 키 설정 */
  setTagKey: (key: string) => void;
  /** 태그 값 설정 */
  setTagValue: (value: string) => void;
  /** 상태 초기화 */
  reset: () => void;
}

/**
 * 클러스터 태그 다이얼로그 상태 관리 훅
 */
export function useClusterTagDialog(): UseClusterTagDialogReturn {
  const [isOpen, setIsOpen] = useState(false);
  const [tagKey, setTagKey] = useState('');
  const [tagValue, setTagValue] = useState('');

  const openDialog = useCallback(() => {
    setIsOpen(true);
  }, []);

  const closeDialog = useCallback(() => {
    setIsOpen(false);
    setTagKey('');
    setTagValue('');
  }, []);

  const reset = useCallback(() => {
    setIsOpen(false);
    setTagKey('');
    setTagValue('');
  }, []);

  return {
    isOpen,
    tagKey,
    tagValue,
    openDialog,
    closeDialog,
    setTagKey,
    setTagValue,
    reset,
  };
}

