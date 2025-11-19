/**
 * StepperSelectedValues Component
 * Stepper 폼에서 선택된 값들을 표시하는 컴포넌트
 */

'use client';

import { Badge } from '@/components/ui/badge';
import { useTranslation } from '@/hooks/use-translation';

export interface SelectedValue {
  label: string;
  value: string | null | undefined;
  placeholder?: string;
}

export interface StepperSelectedValuesProps {
  /**
   * 선택된 값들의 배열
   */
  values: SelectedValue[];
  
  /**
   * 컴포넌트 클래스명
   */
  className?: string;
}

export function StepperSelectedValues({
  values,
  className,
}: StepperSelectedValuesProps) {
  const { t } = useTranslation();

  if (!values || values.length === 0) {
    return null;
  }

  return (
    <div className={`mt-4 pt-4 border-t ${className || ''}`}>
      <div className="flex items-center gap-2 text-sm text-muted-foreground flex-wrap">
        <span className="text-xs font-medium mr-1 whitespace-nowrap">
          {t('common.selectedValues') || '선택된 값:'}
        </span>
        <div className="flex items-center gap-2 flex-wrap max-w-full">
          {values.map((item, index) => {
            const displayValue = item.value || item.placeholder || t('common.notSelected') || '미선택';
            const isEmpty = !item.value;
            
            return (
              <div key={index} className="flex items-center gap-2 flex-shrink-0">
                {index > 0 && (
                  <span className="text-muted-foreground/50 hidden sm:inline">|</span>
                )}
                <Badge 
                  variant={isEmpty ? "secondary" : "outline"} 
                  className={`text-xs ${isEmpty ? 'opacity-60' : ''} max-w-[200px] truncate`}
                  title={displayValue}
                >
                  {displayValue}
                </Badge>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

