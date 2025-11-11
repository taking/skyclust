/**
 * CreateResourceStepperLayout Component
 * 리소스 생성 페이지의 공통 Stepper 레이아웃 컴포넌트
 */

'use client';

import { ReactNode } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Stepper } from '@/components/ui/stepper';
import { StepContent } from '@/components/ui/stepper';
import { useTranslation } from '@/hooks/use-translation';
import { CreateResourcePageHeader, CreateResourcePageHeaderProps } from './create-resource-page-header';
import { StepperNavigationButtons, StepperNavigationButtonsProps } from './stepper-navigation-buttons';

export interface StepConfig {
  /**
   * Step 라벨 (i18n key 또는 직접 텍스트)
   */
  label: string;
  
  /**
   * Step 설명 (i18n key 또는 직접 텍스트)
   */
  description: string;
}

export interface CreateResourceStepperLayoutProps extends Omit<CreateResourcePageHeaderProps, 'onCancel'> {
  /**
   * Step 설정 배열
   */
  steps: StepConfig[];
  
  /**
   * 현재 Step 번호 (1부터 시작)
   */
  currentStep: number;
  
  /**
   * Step Content 렌더링 함수
   */
  renderStepContent: () => ReactNode;
  
  /**
   * Navigation 버튼 Props (currentStep, totalSteps, advancedStepNumber는 자동으로 전달됨)
   */
  navigationProps: Omit<StepperNavigationButtonsProps, 'currentStep' | 'totalSteps' | 'advancedStepNumber'>;
  
  /**
   * 취소 핸들러
   */
  onCancel: () => void;
  
  /**
   * Advanced Step 번호 (선택적)
   */
  advancedStepNumber?: number;
}

export function CreateResourceStepperLayout({
  backPath,
  title,
  description,
  descriptionParams,
  steps,
  currentStep,
  renderStepContent,
  navigationProps,
  onCancel,
  advancedStepNumber,
}: CreateResourceStepperLayoutProps) {
  const { t } = useTranslation();

  // Step 라벨과 설명 번역
  const translatedSteps = steps.map(step => ({
    label: step.label.startsWith('common.') || step.label.includes('.')
      ? t(step.label)
      : step.label.replace(/^(network|kubernetes)\./, ''),
    description: step.description.startsWith('common.') || step.description.includes('.')
      ? t(step.description)
      : step.description.replace(/^(network|kubernetes)\./, ''),
  }));

  // 현재 Step의 라벨과 설명
  const currentStepConfig = steps[currentStep - 1];
  const currentStepLabel = currentStepConfig?.label.startsWith('common.') || currentStepConfig?.label.includes('.')
    ? t(currentStepConfig.label)
    : currentStepConfig?.label.replace(/^(network|kubernetes)\./, '') || '';
  
  const currentStepDescription = currentStepConfig?.description.startsWith('common.') || currentStepConfig?.description.includes('.')
    ? t(currentStepConfig.description)
    : currentStepConfig?.description.replace(/^(network|kubernetes)\./, '') || '';

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <CreateResourcePageHeader
          backPath={backPath}
          title={title}
          description={description}
          descriptionParams={descriptionParams}
          onCancel={onCancel}
        />

        {/* Stepper */}
        <Card className="mb-6">
          <CardContent className="pt-6">
            <Stepper
              currentStep={currentStep}
              totalSteps={steps.length}
              steps={translatedSteps}
            />
          </CardContent>
        </Card>

        {/* Step Content */}
        <Card>
          <CardHeader className="pb-6">
            <CardTitle>{currentStepLabel}</CardTitle>
            {currentStepDescription && (
              <CardDescription>{currentStepDescription}</CardDescription>
            )}
          </CardHeader>
          <CardContent className="pt-0">
            <StepContent>{renderStepContent()}</StepContent>

            {/* Navigation Buttons */}
            <StepperNavigationButtons
              {...navigationProps}
              currentStep={currentStep}
              totalSteps={steps.length}
              advancedStepNumber={advancedStepNumber}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

