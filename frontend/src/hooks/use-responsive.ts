/**
 * useResponsive Hook
 * 반응형 디자인을 위한 포괄적인 훅
 * 
 * 다양한 브레이크포인트를 감지하고 디바이스 타입을 제공합니다.
 */

import { useState, useEffect } from 'react';

/**
 * Tailwind CSS 기본 breakpoints (기준: 0-640px 미만)
 * - sm: 640px 이상
 * - md: 768px 이상
 * - lg: 1024px 이상
 * - xl: 1280px 이상
 * - 2xl: 1536px 이상
 */
export const BREAKPOINTS = {
  sm: 640,
  md: 768,
  lg: 1024,
  xl: 1280,
  '2xl': 1536,
} as const;

export type BreakpointKey = keyof typeof BREAKPOINTS;

export interface UseResponsiveReturn {
  /**
   * 현재 화면 너비
   */
  width: number;
  
  /**
   * 현재 화면 높이
   */
  height: number;
  
  /**
   * 모바일 디바이스 여부 (< 768px)
   */
  isMobile: boolean;
  
  /**
   * 태블릿 디바이스 여부 (768px ~ 1024px)
   */
  isTablet: boolean;
  
  /**
   * 데스크톱 디바이스 여부 (>= 1024px)
   */
  isDesktop: boolean;
  
  /**
   * 큰 화면 여부 (>= 1280px)
   */
  isLargeScreen: boolean;
  
  /**
   * 매우 큰 화면 여부 (>= 1536px)
   */
  isXLargeScreen: boolean;
  
  /**
   * 특정 breakpoint 이상인지 확인
   */
  isBreakpoint: (key: BreakpointKey) => boolean;
  
  /**
   * 특정 breakpoint 미만인지 확인
   */
  isBelowBreakpoint: (key: BreakpointKey) => boolean;
  
  /**
   * 현재 활성화된 breakpoint 키
   */
  activeBreakpoint: BreakpointKey | null;
  
  /**
   * 터치 가능한 디바이스 여부
   */
  isTouchDevice: boolean;
  
  /**
   * 다크 모드 지원 여부
   */
  prefersDarkMode: boolean;
  
  /**
   * 낮은 모션 선호도 여부
   */
  prefersReducedMotion: boolean;
}

/**
 * useResponsive Hook
 * 
 * @example
 * ```tsx
 * const { isMobile, isDesktop, width } = useResponsive();
 * 
 * if (isMobile) {
 *   return <MobileComponent />;
 * }
 * return <DesktopComponent />;
 * ```
 */
export function useResponsive(): UseResponsiveReturn {
  const [dimensions, setDimensions] = useState<{ width: number; height: number }>(() => {
    if (typeof window === 'undefined') {
      return { width: 0, height: 0 };
    }
    return {
      width: window.innerWidth,
      height: window.innerHeight,
    };
  });

  const [prefersDarkMode, setPrefersDarkMode] = useState(false);
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false);
  const [isTouchDevice, setIsTouchDevice] = useState(false);

  useEffect(() => {
    // 초기 터치 디바이스 감지
    setIsTouchDevice(
      'ontouchstart' in window ||
      navigator.maxTouchPoints > 0 ||
      // @ts-expect-error - 일부 브라우저 지원
      navigator.msMaxTouchPoints > 0
    );

    // 다크 모드 감지
    const darkModeQuery = window.matchMedia('(prefers-color-scheme: dark)');
    setPrefersDarkMode(darkModeQuery.matches);
    
    const handleDarkModeChange = (e: MediaQueryListEvent) => {
      setPrefersDarkMode(e.matches);
    };
    darkModeQuery.addEventListener('change', handleDarkModeChange);

    // 낮은 모션 선호도 감지
    const motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    setPrefersReducedMotion(motionQuery.matches);
    
    const handleMotionChange = (e: MediaQueryListEvent) => {
      setPrefersReducedMotion(e.matches);
    };
    motionQuery.addEventListener('change', handleMotionChange);

    // 화면 크기 감지
    const handleResize = () => {
      setDimensions({
        width: window.innerWidth,
        height: window.innerHeight,
      });
    };

    handleResize(); // 초기 크기 설정
    window.addEventListener('resize', handleResize);
    
    // Orientation change 감지
    window.addEventListener('orientationchange', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      window.removeEventListener('orientationchange', handleResize);
      darkModeQuery.removeEventListener('change', handleDarkModeChange);
      motionQuery.removeEventListener('change', handleMotionChange);
    };
  }, []);

  const { width, height } = dimensions;

  // Breakpoint 확인 함수들
  const isBreakpoint = (key: BreakpointKey): boolean => {
    return width >= BREAKPOINTS[key];
  };

  const isBelowBreakpoint = (key: BreakpointKey): boolean => {
    return width < BREAKPOINTS[key];
  };

  // 현재 활성화된 breakpoint 찾기 (큰 것부터 확인)
  const activeBreakpoint: BreakpointKey | null = (() => {
    const keys: BreakpointKey[] = ['2xl', 'xl', 'lg', 'md', 'sm'];
    for (const key of keys) {
      if (width >= BREAKPOINTS[key]) {
        return key;
      }
    }
    return null;
  })();

  // 디바이스 타입 판단
  const isMobile = width < BREAKPOINTS.md; // < 768px
  const isTablet = width >= BREAKPOINTS.md && width < BREAKPOINTS.lg; // 768px ~ 1024px
  const isDesktop = width >= BREAKPOINTS.lg; // >= 1024px
  const isLargeScreen = width >= BREAKPOINTS.xl; // >= 1280px
  const isXLargeScreen = width >= BREAKPOINTS['2xl']; // >= 1536px

  return {
    width,
    height,
    isMobile,
    isTablet,
    isDesktop,
    isLargeScreen,
    isXLargeScreen,
    isBreakpoint,
    isBelowBreakpoint,
    activeBreakpoint,
    isTouchDevice,
    prefersDarkMode,
    prefersReducedMotion,
  };
}

/**
 * useIsMobile Hook
 * 모바일 디바이스 여부만 확인하는 간단한 훅
 * 
 * @deprecated useResponsive 훅을 사용하세요
 */
export function useIsMobile(): boolean {
  const { isMobile } = useResponsive();
  return isMobile;
}

/**
 * useBreakpoint Hook
 * 특정 breakpoint 상태를 확인하는 훅
 * 
 * @example
 * ```tsx
 * const isDesktop = useBreakpoint('lg');
 * ```
 */
export function useBreakpoint(breakpoint: BreakpointKey): boolean {
  const { isBreakpoint } = useResponsive();
  return isBreakpoint(breakpoint);
}

