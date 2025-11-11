/**
 * useStepperForm Hook
 * Stepper 방식 폼의 공통 로직을 추상화한 커스텀 훅
 * 
 * 리소스 생성 페이지(VPC, Subnet, Security Group, Cluster 등)에서
 * 반복되는 Stepper 로직을 통합하여 일관성과 유지보수성을 향상시킵니다.
 * 
 * @example
 * ```tsx
 * const {
 *   currentStep,
 *   form,
 *   formData,
 *   updateFormData,
 *   handleNext,
 *   handlePrevious,
 *   handleSkipAdvanced,
 *   canGoNext,
 *   canGoPrevious,
 * } = useStepperForm({
 *   totalSteps: 3,
 *   schema: createVPCSchema,
 *   defaultValues: { name: '', region: '' },
 *   stepValidation: {
 *     1: ['name', 'region'],
 *     2: ['vpc_id'],
 *   },
 * });
 * ```
 */

'use client';

import { useState, useEffect, useCallback, useMemo } from 'react';
import { useForm, UseFormReturn, FieldValues, Path } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useRouter } from 'next/navigation';
import { log } from '@/lib/logging';

export interface StepValidationConfig<T extends FieldValues> {
  /**
   * Step 번호 (1부터 시작)
   */
  step: number;
  
  /**
   * 해당 Step에서 검증할 필드 목록
   */
  fields: (keyof T)[];
  
  /**
   * 커스텀 검증 함수 (선택적)
   * true를 반환하면 검증 통과, false 또는 에러 메시지를 반환하면 검증 실패
   */
  customValidation?: (formData: T) => boolean | string;
}

export interface UseStepperFormOptions<T extends FieldValues> {
  /**
   * 총 Step 수
   */
  totalSteps: number;
  
  /**
   * Zod 스키마
   */
  schema: z.ZodSchema<T>;
  
  /**
   * 기본값
   */
  defaultValues: Partial<T>;
  
  /**
   * Step별 검증 설정
   * Step 번호를 키로 하고, 검증할 필드 목록을 값으로 하는 객체
   */
  stepValidation?: Record<number, (keyof T)[] | StepValidationConfig<T>>;
  
  /**
   * Advanced Step 번호 (Skip 가능한 Step)
   * 지정된 Step은 Skip 버튼이 표시됩니다.
   */
  advancedStepNumber?: number;
  
  /**
   * Form 옵션
   */
  formOptions?: {
    mode?: 'onChange' | 'onBlur' | 'onSubmit' | 'onTouched' | 'all';
  };
  
  /**
   * 초기 Step 번호
   */
  initialStep?: number;
  
  /**
   * Form 데이터 동기화 콜백
   * 외부 상태와 동기화가 필요한 경우 사용
   */
  onFormDataChange?: (data: Partial<T>) => void;
}

export interface UseStepperFormReturn<T extends FieldValues> {
  /**
   * 현재 Step 번호 (1부터 시작)
   */
  currentStep: number;
  
  /**
   * React Hook Form 인스턴스
   */
  form: UseFormReturn<T>;
  
  /**
   * Form 데이터 (상태로 관리)
   */
  formData: Partial<T>;
  
  /**
   * Form 데이터 업데이트 함수
   */
  updateFormData: (data: Partial<T>) => void;
  
  /**
   * 다음 Step으로 이동
   */
  handleNext: () => Promise<void>;
  
  /**
   * 이전 Step으로 이동
   */
  handlePrevious: () => void;
  
  /**
   * Advanced Step 건너뛰기
   */
  handleSkipAdvanced: () => void;
  
  /**
   * Step 이동 (직접 지정)
   */
  setCurrentStep: (step: number) => void;
  
  /**
   * 다음 Step으로 이동 가능한지 여부
   */
  canGoNext: boolean;
  
  /**
   * 이전 Step으로 이동 가능한지 여부
   */
  canGoPrevious: boolean;
  
  /**
   * Advanced Step인지 여부
   */
  isAdvancedStep: boolean;
  
  /**
   * 마지막 Step인지 여부
   */
  isLastStep: boolean;
  
  /**
   * 첫 번째 Step인지 여부
   */
  isFirstStep: boolean;
  
  /**
   * Form 리셋
   */
  reset: (values?: Partial<T>) => void;
}

/**
 * Stepper Form 훅
 */
export function useStepperForm<T extends FieldValues>(
  options: UseStepperFormOptions<T>
): UseStepperFormReturn<T> {
  const {
    totalSteps,
    schema,
    defaultValues,
    stepValidation = {},
    advancedStepNumber,
    formOptions,
    initialStep = 1,
    onFormDataChange,
  } = options;

  const [currentStep, setCurrentStep] = useState(initialStep);
  const [formData, setFormData] = useState<Partial<T>>(defaultValues);

  // React Hook Form 초기화
  const form = useForm<T>({
    resolver: zodResolver(schema as any), // Type compatibility workaround
    defaultValues: defaultValues as any,
    mode: formOptions?.mode || 'onChange',
    ...formOptions,
  });

  // Form 데이터 동기화
  useEffect(() => {
    form.reset(formData as T);
  }, []); // 초기 마운트 시에만 실행

  // Form 데이터 업데이트 함수
  const updateFormData = useCallback(
    (data: Partial<T>) => {
      setFormData(prev => {
        const updated = { ...prev, ...data };
        onFormDataChange?.(updated);
        return updated;
      });
      
      // React Hook Form에도 동기화
      Object.entries(data).forEach(([key, value]) => {
        form.setValue(key as Path<T>, value as any);
      });
    },
    [form, onFormDataChange]
  );

  // Step별 검증 함수
  const validateStep = useCallback(
    async (step: number): Promise<boolean> => {
      const validation = stepValidation[step];
      
      if (!validation) {
        // 검증 설정이 없으면 통과
        return true;
      }

      // 배열 형태인 경우 (간단한 필드 목록)
      if (Array.isArray(validation)) {
        const fields = validation as (keyof T)[];
        const isValid = await form.trigger(fields as Path<T>[]);
        return isValid;
      }

      // StepValidationConfig 형태인 경우
      const config = validation as StepValidationConfig<T>;
      const fields = config.fields;
      
      // 필드 검증
      const fieldsValid = await form.trigger(fields as Path<T>[]);
      if (!fieldsValid) {
        return false;
      }

      // 커스텀 검증
      if (config.customValidation) {
        const formValues = form.getValues();
        const customResult = config.customValidation(formValues);
        if (typeof customResult === 'string') {
          log.warn('Custom validation failed', { step, message: customResult });
          return false;
        }
        return customResult;
      }

      return true;
    },
    [form, stepValidation]
  );

  // 다음 Step으로 이동
  const handleNext = useCallback(async () => {
    // 현재 Step 검증
    const isValid = await validateStep(currentStep);
    if (!isValid) {
      log.debug('Step validation failed', { step: currentStep });
      return;
    }

    // Form 데이터 동기화
    const formValues = form.getValues();
    updateFormData(formValues);

    // 다음 Step으로 이동
    if (currentStep < totalSteps) {
      setCurrentStep(prev => prev + 1);
    }
  }, [currentStep, totalSteps, validateStep, form, updateFormData]);

  // 이전 Step으로 이동
  const handlePrevious = useCallback(() => {
    if (currentStep > 1) {
      setCurrentStep(prev => prev - 1);
    }
  }, [currentStep]);

  // Advanced Step 건너뛰기
  const handleSkipAdvanced = useCallback(() => {
    if (advancedStepNumber && currentStep === advancedStepNumber) {
      // Review Step (마지막 Step)으로 이동
      setCurrentStep(totalSteps);
    }
  }, [advancedStepNumber, currentStep, totalSteps]);

  // 계산된 값들
  const canGoNext = useMemo(() => currentStep < totalSteps, [currentStep, totalSteps]);
  const canGoPrevious = useMemo(() => currentStep > 1, [currentStep]);
  const isAdvancedStep = useMemo(
    () => advancedStepNumber !== undefined && currentStep === advancedStepNumber,
    [advancedStepNumber, currentStep]
  );
  const isLastStep = useMemo(() => currentStep === totalSteps, [currentStep, totalSteps]);
  const isFirstStep = useMemo(() => currentStep === 1, [currentStep]);

  // Form 리셋
  const reset = useCallback(
    (values?: Partial<T>) => {
      const resetValues = values || defaultValues;
      setFormData(resetValues);
      form.reset(resetValues as T);
    },
    [form, defaultValues]
  );

  return {
    currentStep,
    form,
    formData,
    updateFormData,
    handleNext,
    handlePrevious,
    handleSkipAdvanced,
    setCurrentStep,
    canGoNext,
    canGoPrevious,
    isAdvancedStep,
    isLastStep,
    isFirstStep,
    reset,
  };
}


