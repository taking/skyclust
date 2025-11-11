/**
 * StepperNavigationButtons Component
 * Stepper 폼의 공통 Navigation 버튼 컴포넌트
 */

'use client';

import { Button } from '@/components/ui/button';
import { useTranslation } from '@/hooks/use-translation';

export interface StepperNavigationButtonsProps {
  /**
   * 현재 Step 번호 (1부터 시작)
   */
  currentStep: number;
  
  /**
   * 총 Step 수
   */
  totalSteps: number;
  
  /**
   * Advanced Step 번호 (Skip 가능한 Step)
   */
  advancedStepNumber?: number;
  
  /**
   * 로딩 상태
   */
  isLoading?: boolean;
  
  /**
   * 이전 버튼 클릭 핸들러
   */
  onPrevious: () => void;
  
  /**
   * 다음 버튼 클릭 핸들러
   */
  onNext: () => void;
  
  /**
   * 취소 버튼 클릭 핸들러
   */
  onCancel: () => void;
  
  /**
   * Advanced Step 건너뛰기 핸들러
   */
  onSkipAdvanced?: () => void;
  
  /**
   * 마지막 Step의 제출 버튼 클릭 핸들러
   */
  onSubmit?: () => void;
  
  /**
   * 제출 버튼 텍스트 (i18n key 또는 직접 텍스트)
   */
  submitButtonText?: string;
  
  /**
   * 제출 중 버튼 텍스트 (i18n key 또는 직접 텍스트)
   */
  submittingButtonText?: string;
}

export function StepperNavigationButtons({
  currentStep,
  totalSteps,
  advancedStepNumber,
  isLoading = false,
  onPrevious,
  onNext,
  onCancel,
  onSkipAdvanced,
  onSubmit,
  submitButtonText,
  submittingButtonText,
}: StepperNavigationButtonsProps) {
  const { t } = useTranslation();
  
  const isFirstStep = currentStep === 1;
  const isLastStep = currentStep === totalSteps;
  const isAdvancedStep = advancedStepNumber !== undefined && currentStep === advancedStepNumber;
  
  const defaultSubmitText = submitButtonText || t('common.create');
  const defaultSubmittingText = submittingButtonText || t('actions.creating');
  
  const submitText = defaultSubmitText.startsWith('common.') || defaultSubmitText.includes('.')
    ? t(defaultSubmitText)
    : defaultSubmitText;
  
  const submittingText = defaultSubmittingText.startsWith('actions.') || defaultSubmittingText.includes('.')
    ? t(defaultSubmittingText)
    : defaultSubmittingText;

  return (
    <div className="flex justify-between mt-8 pt-6 border-t">
      <Button
        type="button"
        variant="outline"
        onClick={isFirstStep ? onCancel : onPrevious}
        disabled={isLoading}
      >
        {isFirstStep ? t('common.cancel') : t('common.back')}
      </Button>
      <div className="flex gap-2">
        {isAdvancedStep && onSkipAdvanced && (
          <Button 
            variant="outline" 
            onClick={onSkipAdvanced}
            disabled={isLoading}
          >
            {t('network.skipAdvancedOptions')}
          </Button>
        )}
        {!isLastStep ? (
          <Button
            type="button"
            onClick={onNext}
            disabled={isLoading}
          >
            {t('common.next')}
          </Button>
        ) : (
          onSubmit && (
            <Button
              type="button"
              onClick={onSubmit}
              disabled={isLoading}
            >
              {isLoading ? submittingText : submitText}
            </Button>
          )
        )}
      </div>
    </div>
  );
}


