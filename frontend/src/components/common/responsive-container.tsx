/**
 * Responsive Container Component
 * 반응형 컨테이너 컴포넌트
 * 
 * 화면 크기에 따라 다른 레이아웃을 제공합니다.
 */

'use client';

import { ReactNode } from 'react';
import { useResponsive } from '@/hooks/use-responsive';
import { cn } from '@/lib/utils';

export interface ResponsiveContainerProps {
  /**
   * 모바일에서 표시할 컨텐츠
   */
  mobile?: ReactNode;
  
  /**
   * 태블릿에서 표시할 컨텐츠
   */
  tablet?: ReactNode;
  
  /**
   * 데스크톱에서 표시할 컨텐츠
   */
  desktop?: ReactNode;
  
  /**
   * 기본 컨텐츠 (모든 화면 크기에서 표시)
   */
  children?: ReactNode;
  
  /**
   * 추가 클래스명
   */
  className?: string;
  
  /**
   * 모바일에서 숨길지 여부
   */
  hideOnMobile?: boolean;
  
  /**
   * 데스크톱에서 숨길지 여부
   */
  hideOnDesktop?: boolean;
}

/**
 * ResponsiveContainer Component
 * 
 * 화면 크기에 따라 다른 컨텐츠를 표시하는 컴포넌트
 * 
 * @example
 * ```tsx
 * <ResponsiveContainer
 *   mobile={<MobileView />}
 *   desktop={<DesktopView />}
 * />
 * ```
 */
export function ResponsiveContainer({
  mobile,
  tablet,
  desktop,
  children,
  className,
  hideOnMobile = false,
  hideOnDesktop = false,
}: ResponsiveContainerProps) {
  const { isMobile, isTablet, isDesktop } = useResponsive();

  // 숨김 설정
  if (hideOnMobile && isMobile) return null;
  if (hideOnDesktop && isDesktop) return null;

  // 우선순위: mobile/tablet/desktop > children
  if (isMobile && mobile !== undefined) {
    return <div className={className}>{mobile}</div>;
  }
  
  if (isTablet && tablet !== undefined) {
    return <div className={className}>{tablet}</div>;
  }
  
  if (isDesktop && desktop !== undefined) {
    return <div className={className}>{desktop}</div>;
  }

  return <div className={className}>{children}</div>;
}

/**
 * MobileOnly Component
 * 모바일에서만 표시되는 컴포넌트
 */
export function MobileOnly({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <ResponsiveContainer hideOnDesktop className={className}>
      {children}
    </ResponsiveContainer>
  );
}

/**
 * DesktopOnly Component
 * 데스크톱에서만 표시되는 컴포넌트
 */
export function DesktopOnly({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <ResponsiveContainer hideOnMobile className={className}>
      {children}
    </ResponsiveContainer>
  );
}

/**
 * TabletOnly Component
 * 태블릿에서만 표시되는 컴포넌트
 */
export function TabletOnly({ children, className }: { children: ReactNode; className?: string }) {
  const { isTablet } = useResponsive();
  if (!isTablet) return null;
  return <div className={className}>{children}</div>;
}

