/**
 * useIsMobile Hook
 * 모바일 디바이스 여부를 확인하는 간단한 훅
 * 
 * @deprecated useResponsive 훅을 사용하는 것을 권장합니다.
 * 이 파일은 하위 호환성을 위해 유지됩니다.
 */

import { useResponsive } from './use-responsive';

/**
 * 모바일 디바이스 여부만 반환하는 간단한 훅
 * 
 * 더 많은 기능이 필요하면 useResponsive를 사용하세요.
 */
export function useIsMobile(): boolean {
  const { isMobile } = useResponsive();
  return isMobile;
}
