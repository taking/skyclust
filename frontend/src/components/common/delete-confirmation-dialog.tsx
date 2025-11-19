'use client';

import { useState, useEffect } from 'react';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useTranslation } from '@/hooks/use-translation';

interface DeleteConfirmationDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
  title?: string;
  description?: string;
  confirmText?: string;
  cancelText?: string;
  isLoading?: boolean;
  /**
   * 리소스명 확인이 필요한 경우 리소스명을 전달
   * 이 값이 제공되면 사용자가 리소스명을 입력해야 삭제 가능
   */
  resourceName?: string;
  /**
   * 리소스명 확인 메시지
   */
  resourceNameLabel?: string;
  /**
   * 리소스명 확인 플레이스홀더
   */
  resourceNamePlaceholder?: string;
}

/**
 * 삭제 확인 모달 컴포넌트
 * 
 * @example
 * ```tsx
 * const [isOpen, setIsOpen] = useState(false);
 * 
 * <DeleteConfirmationDialog
 *   open={isOpen}
 *   onOpenChange={setIsOpen}
 *   onConfirm={() => {
 *     // 삭제 로직
 *     handleDelete();
 *     setIsOpen(false);
 *   }}
 *   title="리소스 삭제"
 *   description="이 작업은 되돌릴 수 없습니다."
 * />
 * ```
 * 
 * @example 리소스명 확인이 필요한 경우
 * ```tsx
 * <DeleteConfirmationDialog
 *   open={isOpen}
 *   onOpenChange={setIsOpen}
 *   onConfirm={handleDelete}
 *   title="워크스페이스 삭제"
 *   description="이 작업은 되돌릴 수 없습니다."
 *   resourceName={workspace.name}
 *   resourceNameLabel="워크스페이스 이름"
 * />
 * ```
 */
export function DeleteConfirmationDialog({
  open,
  onOpenChange,
  onConfirm,
  title,
  description,
  confirmText,
  cancelText,
  isLoading = false,
  resourceName,
  resourceNameLabel,
  resourceNamePlaceholder,
}: DeleteConfirmationDialogProps) {
  const { t } = useTranslation();
  const [inputValue, setInputValue] = useState('');
  const requiresNameConfirmation = !!resourceName;
  const isNameMatch = requiresNameConfirmation ? inputValue.trim() === resourceName.trim() : true;
  const canConfirm = !requiresNameConfirmation || isNameMatch;
  
  const defaultTitle = title || t('common.deleteConfirmation.title');
  const defaultDescription = description || t('common.deleteConfirmation.defaultDescription');
  const defaultConfirmText = confirmText || t('common.delete');
  const defaultCancelText = cancelText || t('common.cancel');
  const defaultResourceNameLabel = resourceNameLabel || t('common.deleteConfirmation.resourceNameLabel');

  // 모달이 닫힐 때 입력값 초기화
  useEffect(() => {
    if (!open) {
      setInputValue('');
    }
  }, [open]);

  const handleConfirm = () => {
    if (!canConfirm) return;
    onConfirm();
    setInputValue('');
  };

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      setInputValue('');
    }
    onOpenChange(newOpen);
  };

  return (
    <AlertDialog open={open} onOpenChange={handleOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{defaultTitle}</AlertDialogTitle>
          <AlertDialogDescription>{defaultDescription}</AlertDialogDescription>
        </AlertDialogHeader>
        
        {requiresNameConfirmation && (
          <div className="space-y-2 py-4">
            <Label htmlFor="resource-name-input" className="text-sm font-medium">
              {t('common.deleteConfirmation.resourceNameInputLabel', { resourceName: defaultResourceNameLabel })}
            </Label>
            <Input
              id="resource-name-input"
              type="text"
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              placeholder={resourceNamePlaceholder || resourceName}
              disabled={isLoading}
              className="w-full"
              autoFocus
              onKeyDown={(e) => {
                if (e.key === 'Enter' && canConfirm && !isLoading) {
                  handleConfirm();
                }
              }}
            />
            {inputValue && !isNameMatch && (
              <p className="text-sm text-destructive">
                {t('common.deleteConfirmation.nameMismatch')}
              </p>
            )}
          </div>
        )}

        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading}>{defaultCancelText}</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleConfirm}
            disabled={isLoading || !canConfirm}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isLoading ? t('common.deleteConfirmation.processing') : defaultConfirmText}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

