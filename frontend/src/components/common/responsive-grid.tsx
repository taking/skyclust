/**
 * Responsive Grid Component
 * 반응형 그리드 컴포넌트
 * 
 * 화면 크기에 따라 열 개수를 자동으로 조정합니다.
 */

'use client';

import { ReactNode } from 'react';
import { useResponsive } from '@/hooks/use-responsive';
import { cn } from '@/lib/utils';

export interface ResponsiveGridProps {
  /**
   * 그리드 아이템들
   */
  children: ReactNode;
  
  /**
   * 모바일 열 개수 (기본: 1)
   */
  mobileCols?: number;
  
  /**
   * 태블릿 열 개수 (기본: 2)
   */
  tabletCols?: number;
  
  /**
   * 데스크톱 열 개수 (기본: 3)
   */
  desktopCols?: number;
  
  /**
   * 큰 화면 열 개수 (기본: desktopCols)
   */
  largeCols?: number;
  
  /**
   * 간격 (기본: 4)
   */
  gap?: number;
  
  /**
   * 추가 클래스명
   */
  className?: string;
}

/**
 * ResponsiveGrid Component
 * 
 * 화면 크기에 따라 열 개수를 자동으로 조정하는 그리드
 * 
 * @example
 * ```tsx
 * <ResponsiveGrid mobileCols={1} tabletCols={2} desktopCols={3}>
 *   <Card>Item 1</Card>
 *   <Card>Item 2</Card>
 *   <Card>Item 3</Card>
 * </ResponsiveGrid>
 * ```
 */
export function ResponsiveGrid({
  children,
  mobileCols = 1,
  tabletCols = 2,
  desktopCols = 3,
  largeCols,
  gap = 4,
  className,
}: ResponsiveGridProps) {
  const { isMobile, isTablet, isLargeScreen } = useResponsive();


  // Tailwind의 동적 클래스 생성은 제한적이므로 명시적으로 처리
  const gridColsClasses = {
    1: 'grid-cols-1',
    2: 'grid-cols-2',
    3: 'grid-cols-3',
    4: 'grid-cols-4',
    5: 'grid-cols-5',
    6: 'grid-cols-6',
  } as const;

  const gapClasses = {
    1: 'gap-1',
    2: 'gap-2',
    3: 'gap-3',
    4: 'gap-4',
    5: 'gap-5',
    6: 'gap-6',
    8: 'gap-8',
  } as const;

  const mobileColsClass = gridColsClasses[mobileCols as keyof typeof gridColsClasses] || 'grid-cols-1';
  const tabletColsClass = gridColsClasses[tabletCols as keyof typeof gridColsClasses] || 'grid-cols-2';
  const desktopColsClass = gridColsClasses[desktopCols as keyof typeof gridColsClasses] || 'grid-cols-3';
  const largeColsClass = largeCols ? (gridColsClasses[largeCols as keyof typeof gridColsClasses] || 'grid-cols-4') : null;
  const gapClass = gapClasses[gap as keyof typeof gapClasses] || 'gap-4';

  return (
    <div
      className={cn(
        'grid',
        mobileColsClass,
        `md:${tabletColsClass}`,
        `lg:${desktopColsClass}`,
        largeColsClass && `xl:${largeColsClass}`,
        gapClass,
        className
      )}
    >
      {children}
    </div>
  );
}

/**
 * ResponsiveStack Component
 * 모바일에서는 세로 배치, 데스크톱에서는 가로 배치
 */
export interface ResponsiveStackProps {
  children: ReactNode;
  direction?: 'row' | 'column';
  gap?: number;
  className?: string;
  align?: 'start' | 'center' | 'end' | 'stretch';
  justify?: 'start' | 'center' | 'end' | 'between' | 'around';
}

export function ResponsiveStack({
  children,
  direction = 'column',
  gap = 4,
  className,
  align = 'stretch',
  justify = 'start',
}: ResponsiveStackProps) {
  return (
    <div
      className={cn(
        'flex',
        `flex-${direction}`,
        `md:flex-row`,
        `gap-${gap}`,
        `items-${align}`,
        `justify-${justify}`,
        className
      )}
    >
      {children}
    </div>
  );
}

