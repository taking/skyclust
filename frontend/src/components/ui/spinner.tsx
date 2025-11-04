/**
 * Spinner Component
 * 로딩 스피너 컴포넌트
 * 
 * 다양한 크기와 스타일을 지원하는 스피너
 */

'use client';

import { cn } from '@/lib/utils';

export interface SpinnerProps {
  /**
   * Spinner의 크기
   * - 'sm': 작은 크기
   * - 'default': 기본 크기
   * - 'lg': 큰 크기
   * - 'xl': 매우 큰 크기
   */
  size?: 'sm' | 'default' | 'lg' | 'xl';
  
  /**
   * 색상 variant
   * - 'default': 기본 색상
   * - 'primary': 기본 색상
   * - 'muted': 회색 색상
   */
  variant?: 'default' | 'primary' | 'muted';
  
  /**
   * 레이블 텍스트
   */
  label?: string;
  
  /**
   * 추가 클래스명
   */
  className?: string;
}

const sizeClasses = {
  sm: 'h-4 w-4 border-2',
  default: 'h-8 w-8 border-2',
  lg: 'h-12 w-12 border-[3px]',
  xl: 'h-16 w-16 border-4',
};

const variantClasses = {
  default: 'border-gray-900 border-t-transparent',
  primary: 'border-primary border-t-transparent',
  muted: 'border-gray-400 border-t-transparent',
};

export function Spinner({
  size = 'default',
  variant = 'default',
  label,
  className,
}: SpinnerProps) {
  return (
    <div className={cn('flex flex-col items-center justify-center', className)}>
      <div
        className={cn(
          'animate-spin rounded-full',
          sizeClasses[size],
          variantClasses[variant]
        )}
        role="status"
        aria-label={label || 'Loading'}
      >
        <span className="sr-only">{label || 'Loading'}</span>
      </div>
      {label && (
        <p className="mt-2 text-sm text-gray-600">{label}</p>
      )}
    </div>
  );
}

/**
 * Inline Spinner Component
 * 인라인으로 사용할 수 있는 작은 스피너
 */
export function InlineSpinner({ className }: { className?: string }) {
  return (
    <Spinner 
      size="sm" 
      variant="muted" 
      className={cn('inline-block', className)}
    />
  );
}

