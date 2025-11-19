/**
 * useRegionFilter Hook
 * Multi-provider Region 필터링
 * 
 * 기능:
 * - 여러 provider의 region을 union으로 결합
 * - Region 선택 관리
 * - Provider별 region 목록 제공
 */

import { useMemo } from 'react';
import { getRegionsByProvider } from '@/lib/regions';
import type { CloudProvider } from '@/lib/types/kubernetes';

interface RegionOption {
  value: string;
  label: string;
}

interface UseRegionFilterOptions {
  /**
   * 선택된 providers
   */
  providers: CloudProvider[];
  
  /**
   * 현재 선택된 region
   */
  selectedRegion?: string | null;
  
  /**
   * Region 선택 핸들러
   */
  onRegionChange?: (region: string | null) => void;
}

/**
 * Multi-provider Region 필터 Hook
 * 
 * @example
 * ```tsx
 * const {
 *   availableRegions,
 *   regionsByProvider,
 *   selectedRegion,
 *   setSelectedRegion,
 * } = useRegionFilter({
 *   providers: ['aws', 'gcp', 'azure'],
 *   selectedRegion: 'ap-northeast-3',
 * });
 * ```
 */
export function useRegionFilter({
  providers,
  selectedRegion,
  onRegionChange,
}: UseRegionFilterOptions) {
  
  // Provider별 region 목록
  const regionsByProvider = useMemo(() => {
    const result: Record<CloudProvider, string[]> = {
      aws: [],
      gcp: [],
      azure: [],
    };
    
    providers.forEach(provider => {
      if (provider in result) {
        const regions = getRegionsByProvider(provider);
        result[provider] = regions.map((r: RegionOption) => r.value);
      }
    });
    
    return result;
  }, [providers]);
  
  // 모든 provider의 region을 union으로 결합 (중복 제거, 정렬)
  const availableRegions = useMemo(() => {
    const allRegions = new Set<string>();
    
    providers.forEach(provider => {
      const regions = regionsByProvider[provider] || [];
      regions.forEach(region => allRegions.add(region));
    });
    
    return Array.from(allRegions).sort();
  }, [providers, regionsByProvider]);
  
  // Region 선택 핸들러
  const setSelectedRegion = (region: string | null) => {
    onRegionChange?.(region);
  };
  
  // 선택된 region이 유효한지 확인
  const isRegionValid = useMemo(() => {
    if (!selectedRegion) return true; // 선택 안 함 = 유효
    return availableRegions.includes(selectedRegion);
  }, [selectedRegion, availableRegions]);
  
  return {
    availableRegions,
    regionsByProvider,
    selectedRegion: selectedRegion || null,
    setSelectedRegion,
    isRegionValid,
    hasRegions: availableRegions.length > 0,
  };
}

