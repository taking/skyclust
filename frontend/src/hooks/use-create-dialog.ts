/**
 * Create Dialog Hook
 * 
 * Create dialog를 열기 위한 전역 이벤트 리스너를 관리하는 hook
 * 키보드 단축키나 다른 컴포넌트에서 dialog를 열 수 있도록 통합
 * 
 * @example
 * ```tsx
 * const [isCreateDialogOpen, setIsCreateDialogOpen] = useCreateDialog(EVENTS.CREATE_DIALOG.VM);
 * ```
 */

import { useState, useEffect } from 'react';
import { EVENTS } from '@/lib/constants';

/**
 * Create dialog를 관리하는 hook
 * 
 * @param eventName - CustomEvent 이름 (EVENTS.CREATE_DIALOG.* 사용 권장)
 * @returns [isOpen, setIsOpen] - dialog 열림 상태와 setter 함수
 */
export function useCreateDialog(eventName: typeof EVENTS.CREATE_DIALOG[keyof typeof EVENTS.CREATE_DIALOG] | string): [boolean, (open: boolean) => void] {
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    // Event listener for opening dialog
    const handleOpenDialog = () => {
      setIsOpen(true);
    };

    // Register event listener
    window.addEventListener(eventName, handleOpenDialog);

    // Cleanup: remove event listener on unmount
    return () => {
      window.removeEventListener(eventName, handleOpenDialog);
    };
  }, [eventName]);

  return [isOpen, setIsOpen];
}

