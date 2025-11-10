/**
 * Stepper Component
 * Step 방식 폼을 위한 스테퍼 UI 컴포넌트
 */

'use client';

import * as React from 'react';
import { cn } from '@/lib/utils';
import { Check } from 'lucide-react';

interface StepperProps {
  currentStep: number;
  totalSteps: number;
  steps: Array<{
    label: string;
    description?: string;
  }>;
  className?: string;
}

export function Stepper({ currentStep, totalSteps: _totalSteps, steps, className }: StepperProps) {
  return (
    <div className={cn('w-full', className)}>
      <div className="flex items-center justify-between">
        {steps.map((step, index) => {
          const stepNumber = index + 1;
          const isActive = stepNumber === currentStep;
          const isCompleted = stepNumber < currentStep;
          const isPending = stepNumber > currentStep;

          return (
            <React.Fragment key={stepNumber}>
              <div className="flex flex-col items-center flex-1">
                <div className="flex items-start w-full">
                  {/* Step Circle */}
                  <div
                    className={cn(
                      'flex items-center justify-center w-10 h-10 rounded-full border-2 transition-colors flex-shrink-0',
                      isCompleted && 'bg-primary border-primary text-primary-foreground',
                      isActive && 'bg-primary border-primary text-primary-foreground',
                      isPending && 'bg-background border-muted-foreground/30 text-muted-foreground'
                    )}
                  >
                    {isCompleted ? (
                      <Check className="h-5 w-5" />
                    ) : (
                      <span className="text-sm font-semibold">{stepNumber}</span>
                    )}
                  </div>
                  
                  {/* Step Label */}
                  <div className="ml-4 flex-1 min-w-0">
                    <div
                      className={cn(
                        'text-sm font-medium leading-tight',
                        isActive && 'text-foreground',
                        isCompleted && 'text-foreground',
                        isPending && 'text-muted-foreground'
                      )}
                    >
                      {step.label}
                    </div>
                    {step.description && (
                      <div
                        className={cn(
                          'text-xs mt-1 leading-tight',
                          isActive && 'text-muted-foreground',
                          isCompleted && 'text-muted-foreground',
                          isPending && 'text-muted-foreground/60'
                        )}
                      >
                        {step.description}
                      </div>
                    )}
                  </div>
                </div>
              </div>

              {/* Connector Line */}
              {index < steps.length - 1 && (
                <div
                  className={cn(
                    'h-0.5 flex-1 mx-4',
                    isCompleted ? 'bg-primary' : 'bg-muted-foreground/30'
                  )}
                />
              )}
            </React.Fragment>
          );
        })}
      </div>
    </div>
  );
}

interface StepContentProps {
  children: React.ReactNode;
  className?: string;
}

export function StepContent({ children, className }: StepContentProps) {
  return (
    <div className={cn('', className)}>
      {children}
    </div>
  );
}

