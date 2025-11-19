/**
 * useCredentialRegionSync Hook
 * Credential 및 Region 선택 상태를 URL과 동기화
 * 
 * 전략:
 * - URL을 Single Source of Truth로 사용
 * - Store 변경 시 URL 업데이트 (debounced)
 * - URL 변경 시 Store 업데이트
 * - LocalStorage는 fallback으로만 사용
 */

'use client';

import { useEffect, useRef, useMemo, useCallback } from 'react';
import { useRouter, usePathname, useSearchParams } from 'next/navigation';
import { useCredentialContextStore } from '@/store/credential-context';
import {
  encodeCredentials,
  decodeCredentials,
  encodeProviderRegions,
  decodeProviderRegions,
  isUrlTooLong,
  MAX_URL_LENGTH,
} from '@/lib/utils/url-state';
import type { ProviderRegionSelection } from '@/hooks/use-provider-region-filter';
import { log } from '@/lib/logging';

const URL_SYNC_DEBOUNCE_MS = 300;
const STORAGE_KEY = 'credential-region-state';

/**
 * Credentials 배열을 정규화하여 비교 가능한 문자열로 변환
 */
function normalizeCredentials(ids: string[]): string {
  return [...ids].sort().join(',');
}

/**
 * Provider regions 객체를 정규화하여 비교 가능한 문자열로 변환
 */
function normalizeProviderRegions(regions: ProviderRegionSelection): string {
  const normalized = {
    aws: [...((regions as Record<string, string[]>)['aws'] || [])].sort(),
    gcp: [...((regions as Record<string, string[]>)['gcp'] || [])].sort(),
    azure: [...((regions as Record<string, string[]>)['azure'] || [])].sort(),
  };
  return JSON.stringify(normalized);
}

/**
 * URL 동기화 Hook
 * 
 * @example
 * ```tsx
 * function ClustersPage() {
 *   useCredentialRegionSync();
 *   // Store에서 값을 읽으면 자동으로 URL과 동기화됨
 * }
 * ```
 */
export function useCredentialRegionSync() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  
  const {
    selectedCredentialIds,
    selectedRegion,
    providerSelectedRegions,
    setSelectedCredentials,
    setSelectedRegion,
    setProviderSelectedRegions,
  } = useCredentialContextStore();
  
  // Debounce를 위한 ref
  const urlUpdateTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const isInitialMountRef = useRef(true);
  const prevUrlStateRef = useRef<string>('');
  
  // URL에서 상태 읽기 (memoized)
  const urlState = useMemo(() => {
    const urlCredentials = decodeCredentials(searchParams.get('credentials'));
    const urlRegion = searchParams.get('region');
    const urlProviderRegions = decodeProviderRegions(searchParams);
    
    return {
      credentials: urlCredentials,
      region: urlRegion,
      providerRegions: urlProviderRegions,
    };
  }, [searchParams]);
  
  // URL → Store 동기화 (읽기) - 초기 마운트 및 URL 변경 시
  useEffect(() => {
    const { credentials, region, providerRegions } = urlState;
    
    // 현재 URL 상태를 정규화하여 비교
    const currentUrlStateKey = JSON.stringify({
      credentials: normalizeCredentials(credentials),
      region: region || '',
      providerRegions: normalizeProviderRegions(providerRegions),
    });
    
    // 초기 마운트 시 URL에서 읽기
    if (isInitialMountRef.current) {
      // URL을 Single Source of Truth로 사용
      // URL에 값이 있으면 우선 사용
      if (credentials.length > 0) {
        setSelectedCredentials(credentials);
      } else {
        // URL에 credentials가 없으면 Store도 초기화 (URL 우선 원칙)
        if (selectedCredentialIds.length > 0) {
          setSelectedCredentials([]);
        }
      }
      
      if (region) {
        setSelectedRegion(region);
      } else {
        // URL에 region이 없으면 Store도 초기화 (URL 우선 원칙)
        // localStorage에 저장된 값이 URL에 반영되지 않도록 함
        if (selectedRegion) {
          setSelectedRegion(null);
        }
      }
      
      // Provider별 Region이 URL에 있으면 사용
      const hasProviderRegions = Object.values(providerRegions).some(r => r.length > 0);
      if (hasProviderRegions) {
        setProviderSelectedRegions(providerRegions);
      } else {
        // URL에 provider regions가 없으면 Store도 초기화 (URL 우선 원칙)
        const hasCurrentRegions = Object.values(providerSelectedRegions).some(r => r.length > 0);
        if (hasCurrentRegions) {
          setProviderSelectedRegions({
            aws: [],
            gcp: [],
            azure: [],
          });
        }
      }
      
      // LocalStorage fallback은 제거 (URL 우선 원칙)
      // URL이 Single Source of Truth이므로 localStorage는 사용하지 않음
      
      prevUrlStateRef.current = currentUrlStateKey;
      isInitialMountRef.current = false;
      return;
    }
    
    // URL 변경 감지 시 Store 업데이트 (초기 마운트 이후)
    if (prevUrlStateRef.current === currentUrlStateKey) {
      return;
    }
    
    prevUrlStateRef.current = currentUrlStateKey;
    
    // Credentials 비교 및 업데이트 (정규화된 비교)
    if (credentials.length > 0) {
      const currentCredentialsKey = normalizeCredentials(selectedCredentialIds);
      const urlCredentialsKey = normalizeCredentials(credentials);
      if (currentCredentialsKey !== urlCredentialsKey) {
        setSelectedCredentials(credentials);
      }
    } else if (selectedCredentialIds.length > 0) {
      // URL에 credentials가 없으면 Store도 초기화 (URL 우선 원칙)
      setSelectedCredentials([]);
    }
    
    // Region 비교 및 업데이트
    if (region && region !== selectedRegion) {
      setSelectedRegion(region);
    } else if (!region && selectedRegion) {
      // URL에 region이 없으면 Store도 초기화 (URL 우선 원칙)
      setSelectedRegion(null);
    }
    
    // Provider별 Region 비교 및 업데이트 (정규화된 깊은 비교)
    const hasProviderRegions = Object.values(providerRegions).some(r => r.length > 0);
    if (hasProviderRegions) {
      const currentRegionsKey = normalizeProviderRegions(providerSelectedRegions);
      const urlRegionsKey = normalizeProviderRegions(providerRegions);
      if (currentRegionsKey !== urlRegionsKey) {
        setProviderSelectedRegions(providerRegions);
      }
    } else {
      // URL에 provider regions가 없으면 Store도 초기화 (URL 우선 원칙)
      const hasCurrentRegions = Object.values(providerSelectedRegions).some(r => r.length > 0);
      if (hasCurrentRegions) {
        setProviderSelectedRegions({
          aws: [],
          gcp: [],
          azure: [],
        });
      }
    }
  }, [urlState, selectedCredentialIds, selectedRegion, providerSelectedRegions, setSelectedCredentials, setSelectedRegion, setProviderSelectedRegions]);
  
  
  // Store → URL 동기화 (쓰기) - Debounced
  const syncToUrl = useCallback(() => {
    if (urlUpdateTimeoutRef.current) {
      clearTimeout(urlUpdateTimeoutRef.current);
    }
    
    urlUpdateTimeoutRef.current = setTimeout(() => {
      const params = new URLSearchParams(searchParams.toString());
      
      // Multi-credential 선택
      if (selectedCredentialIds.length > 1) {
        params.set('credentials', encodeCredentials(selectedCredentialIds));
      } else if (selectedCredentialIds.length === 1) {
        params.delete('credentials');
      } else {
        params.delete('credentials');
      }
      
      // Region 선택
      if (selectedRegion) {
        params.set('region', selectedRegion);
      } else {
        params.delete('region');
      }
      
      // Provider별 Region 선택
      const providerRegionsParams = encodeProviderRegions(providerSelectedRegions);
      Object.entries(providerRegionsParams).forEach(([key, value]) => {
        params.set(key, value);
      });
      // 선택되지 않은 provider의 region 파라미터 제거
      ['aws_regions', 'gcp_regions', 'azure_regions'].forEach(key => {
        if (!providerRegionsParams[key]) {
          params.delete(key);
        }
      });
      
      // URL 길이 확인
      const newUrl = `${pathname}?${params.toString()}`;
      
      if (isUrlTooLong(newUrl)) {
        // URL이 너무 길면 localStorage에 저장하고 간단한 플래그만 URL에 저장
        try {
          localStorage.setItem(STORAGE_KEY, JSON.stringify({
            credentials: selectedCredentialIds,
            region: selectedRegion,
            providerRegions: providerSelectedRegions,
          }));
          params.set('use_storage', 'true');
          params.delete('credentials');
          Object.keys(providerRegionsParams).forEach(key => {
            params.delete(key);
          });
        } catch (error) {
          log.error('Failed to store credential-region state', error, {
            service: 'useCredentialRegionSync',
            action: 'storeState',
          });
        }
      } else {
        // URL이 정상 길이면 localStorage에서 제거
        params.delete('use_storage');
        try {
          localStorage.removeItem(STORAGE_KEY);
        } catch (error) {
          log.error('Failed to remove stored credential-region state', error, {
            service: 'useCredentialRegionSync',
            action: 'removeStoredState',
          });
        }
      }
      
      const finalUrl = `${pathname}?${params.toString()}`;
      router.replace(finalUrl, { scroll: false });
    }, URL_SYNC_DEBOUNCE_MS);
  }, [selectedCredentialIds, selectedRegion, providerSelectedRegions, pathname, searchParams, router]);
  
  // Store 변경 시 URL 동기화
  useEffect(() => {
    if (isInitialMountRef.current) {
      return;
    }
    
    syncToUrl();
    
    return () => {
      if (urlUpdateTimeoutRef.current) {
        clearTimeout(urlUpdateTimeoutRef.current);
      }
    };
  }, [selectedCredentialIds, selectedRegion, providerSelectedRegions, syncToUrl]);
  
}

