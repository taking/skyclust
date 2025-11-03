/**
 * Skeleton Component
 * 로딩 상태를 표시하는 기본 Skeleton 컴포넌트
 * 
 * 다양한 크기와 스타일을 지원하는 범용 컴포넌트
 */

import { cn } from "@/lib/utils"

export interface SkeletonProps extends React.ComponentProps<"div"> {
  /**
   * Skeleton의 크기 variant
   * - 'default': 기본 크기
   * - 'sm': 작은 크기
   * - 'lg': 큰 크기
   * - 'xl': 매우 큰 크기
   */
  size?: 'default' | 'sm' | 'lg' | 'xl';
  
  /**
   * Skeleton의 모양 variant
   * - 'rect': 사각형 (기본)
   * - 'circle': 원형
   * - 'text': 텍스트 라인
   */
  variant?: 'rect' | 'circle' | 'text';
  
  /**
   * 애니메이션 효과
   * - true: 펄스 애니메이션 활성화 (기본)
   * - false: 애니메이션 비활성화
   */
  animate?: boolean;
}

function Skeleton({ 
  className,
  size = 'default',
  variant = 'rect',
  animate = true,
  ...props 
}: SkeletonProps) {
  const sizeClasses = {
    sm: 'h-2',
    default: 'h-4',
    lg: 'h-6',
    xl: 'h-8',
  };

  const variantClasses = {
    rect: 'rounded-md',
    circle: 'rounded-full',
    text: 'rounded',
  };

  return (
    <div
      data-slot="skeleton"
      className={cn(
        "bg-accent",
        sizeClasses[size],
        variantClasses[variant],
        animate && "animate-pulse",
        className
      )}
      {...props}
    />
  )
}

export { Skeleton }
