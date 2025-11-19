/**
 * useProviderRegionFilter Hook
 * Provider별 Region 필터링
 * 
 * 기능:
 * - 각 Provider별로 독립적인 Region 선택
 * - Provider별 Region 상태 관리
 * - Region 선택/해제 토글
 */

import { useState, useCallback, useMemo, useEffect, useRef } from 'react';
import { getRegionsByProvider } from '@/lib/regions';
import type { CloudProvider } from '@/lib/types/kubernetes';

export interface ProviderRegionSelection {
  [provider: CloudProvider]: string[];
}

interface UseProviderRegionFilterOptions {
  /**
   * 선택된 providers
   */
  providers: CloudProvider[];
  
  /**
   * 초기 선택된 regions (Provider별)
   */
  initialSelectedRegions?: ProviderRegionSelection;
  
  /**
   * Region 선택 변경 핸들러
   */
  onRegionSelectionChange?: (selectedRegions: ProviderRegionSelection) => void;
}

/**
 * Provider별 Region 필터 Hook
 * 
 * @example
 * ```tsx
 * const {
 *   selectedRegions,
 *   toggleRegion,
 *   toggleAllRegionsForProvider,
 *   clearAllRegions,
 *   regionsByProvider,
 * } = useProviderRegionFilter({
 *   providers: ['aws', 'gcp', 'azure'],
 *   onRegionSelectionChange: (regions) => {
 *     console.log('Selected regions:', regions);
 *   },
 * });
 * ```
 */
export function useProviderRegionFilter({
  providers,
  initialSelectedRegions,
  onRegionSelectionChange,
}: UseProviderRegionFilterOptions) {
  
  // Provider별 Region 선택 상태
  const [selectedRegions, setSelectedRegions] = useState<ProviderRegionSelection>(
    initialSelectedRegions || {}
  );
  
  // initialSelectedRegions가 변경되면 state 업데이트
  const prevInitialRef = useRef<string>('');
  const isInitialMountRef = useRef<boolean>(true);
  const isInternalUpdateRef = useRef<boolean>(false);
  const prevSelectedRegionsRef = useRef<ProviderRegionSelection>({});
  
  useEffect(() => {
    const currentKey = JSON.stringify(initialSelectedRegions || {});
    
    // 초기 마운트 시에는 initialSelectedRegions를 그대로 사용
    if (isInitialMountRef.current) {
      isInitialMountRef.current = false;
      const initialRegions = initialSelectedRegions || {};
      if (Object.keys(initialRegions).length > 0) {
        setSelectedRegions(initialRegions);
        prevSelectedRegionsRef.current = initialRegions;
      }
      prevInitialRef.current = currentKey;
      return;
    }
    
    // 이후 변경 시에만 업데이트 (내부 업데이트가 아닐 때만)
    if (!isInternalUpdateRef.current && prevInitialRef.current !== currentKey) {
      if (initialSelectedRegions && Object.keys(initialSelectedRegions).length > 0) {
        setSelectedRegions(initialSelectedRegions);
        prevSelectedRegionsRef.current = initialSelectedRegions;
      }
      prevInitialRef.current = currentKey;
    }
  }, [initialSelectedRegions]);
  
  // selectedRegions 변경 시 onRegionSelectionChange 호출 (렌더링 후)
  useEffect(() => {
    // 초기 마운트는 스킵
    if (isInitialMountRef.current) {
      prevSelectedRegionsRef.current = selectedRegions;
      return;
    }
    
    // 이전 값과 비교하여 실제 변경이 있었는지 확인
    const prevKey = JSON.stringify(prevSelectedRegionsRef.current);
    const currentKey = JSON.stringify(selectedRegions);
    
    if (prevKey !== currentKey) {
      // 내부 업데이트인 경우에만 콜백 호출 (렌더링 완료 후)
      if (isInternalUpdateRef.current && onRegionSelectionChange) {
        // 다음 이벤트 루프에서 호출되도록 스케줄링
        setTimeout(() => {
          onRegionSelectionChange(selectedRegions);
        }, 0);
      }
      
      // 이전 값 업데이트
      prevSelectedRegionsRef.current = selectedRegions;
    }
    
    // 플래그 리셋 (항상 리셋)
    isInternalUpdateRef.current = false;
  }, [selectedRegions, onRegionSelectionChange]);
  
  // Provider가 변경되면 선택되지 않은 Provider의 Region 제거 (초기 마운트 제외)
  const prevProvidersRef = useRef<string>('');
  const isProvidersInitialMountRef = useRef<boolean>(true);
  
  useEffect(() => {
    // 초기 마운트 시에는 스킵
    if (isProvidersInitialMountRef.current) {
      isProvidersInitialMountRef.current = false;
      prevProvidersRef.current = providers.sort().join(',');
      return;
    }
    
    setSelectedRegions(prev => {
      const updated: ProviderRegionSelection = {};
      let hasChanged = false;
      
      providers.forEach(provider => {
        updated[provider] = prev[provider] || [];
      });
      
      // 이전에 선택되었지만 현재 providers에 없는 provider의 region 제거
      Object.keys(prev).forEach(provider => {
        if (!providers.includes(provider as CloudProvider)) {
          hasChanged = true;
        }
      });
      
      // 변경이 없으면 이전 상태 유지
      if (!hasChanged && providers.every(p => prev[p] !== undefined)) {
        return prev;
      }
      
      return updated;
    });
    
    prevProvidersRef.current = providers.sort().join(',');
  }, [providers]);
  
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
        result[provider] = regions.map(r => r.value);
      }
    });
    
    return result;
  }, [providers]);
  
  // Region 토글
  const toggleRegion = useCallback((provider: CloudProvider, region: string) => {
    isInternalUpdateRef.current = true;
    setSelectedRegions(prev => {
      const providerRegions = prev[provider] || [];
      const isSelected = providerRegions.includes(region);
      
      return {
        ...prev,
        [provider]: isSelected
          ? providerRegions.filter(r => r !== region)
          : [...providerRegions, region],
      };
    });
  }, []);
  
  // Provider의 모든 Region 선택/해제
  const toggleAllRegionsForProvider = useCallback((provider: CloudProvider) => {
    isInternalUpdateRef.current = true;
    setSelectedRegions(prev => {
      const providerRegions = regionsByProvider[provider] || [];
      const currentRegions = prev[provider] || [];
      const allSelected = providerRegions.every(r => currentRegions.includes(r));
      
      return {
        ...prev,
        [provider]: allSelected ? [] : [...providerRegions],
      };
    });
  }, [regionsByProvider]);
  
  // 모든 Region 선택 해제
  const clearAllRegions = useCallback(() => {
    const emptySelection: ProviderRegionSelection = {};
    providers.forEach(provider => {
      emptySelection[provider] = [];
    });
    
    isInternalUpdateRef.current = true;
    setSelectedRegions(emptySelection);
  }, [providers]);
  
  // 선택된 Region이 있는지 확인
  const hasSelectedRegions = useMemo(() => {
    return providers.some(provider => {
      const regions = selectedRegions[provider] || [];
      return regions.length > 0;
    });
  }, [providers, selectedRegions]);
  
  // Provider별 선택된 Region 개수
  const selectedCountsByProvider = useMemo(() => {
    const counts: Record<CloudProvider, number> = {
      aws: 0,
      gcp: 0,
      azure: 0,
    };
    
    providers.forEach(provider => {
      counts[provider] = (selectedRegions[provider] || []).length;
    });
    
    return counts;
  }, [providers, selectedRegions]);
  
  return {
    selectedRegions,
    regionsByProvider,
    toggleRegion,
    toggleAllRegionsForProvider,
    clearAllRegions,
    hasSelectedRegions,
    selectedCountsByProvider,
  };
}

