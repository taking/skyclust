/**
 * Credential Preference Utilities
 * 자격증명 선호도 및 사용 통계 관리 유틸리티
 * 
 * Backend API가 준비되기 전까지 localStorage를 사용하여
 * 최근 사용한 자격증명, Provider별 기본 자격증명 등을 추적
 */

import { logger } from './logger';
import type { Credential, CloudProvider } from './types';

export interface CredentialUsage {
  credentialId: string;
  provider: CloudProvider;
  lastUsedAt: string;
  usageCount: number;
  resourceTypes: string[]; // 'kubernetes', 'network', 'compute'
}

export interface CredentialPreference {
  workspaceId: string;
  defaultCredentials: Partial<Record<CloudProvider, string | null>>; // Provider별 기본 자격증명
  recentlyUsed: CredentialUsage[];
}

const PREFERENCE_STORAGE_KEY = 'credential-preferences';
const MAX_RECENTLY_USED = 10; // 최근 사용한 자격증명 최대 개수

/**
 * 자격증명 사용 추적
 */
export function trackCredentialUsage(
  workspaceId: string,
  credentialId: string,
  provider: CloudProvider,
  resourceType: 'kubernetes' | 'network' | 'compute'
): void {
  try {
    const preferences = loadPreferences(workspaceId);
    
    // 기존 사용 기록 찾기
    const existingIndex = preferences.recentlyUsed.findIndex(
      (usage) => usage.credentialId === credentialId
    );
    
    if (existingIndex >= 0) {
      // 기존 기록 업데이트
      const existing = preferences.recentlyUsed[existingIndex];
      existing.lastUsedAt = new Date().toISOString();
      existing.usageCount += 1;
      if (!existing.resourceTypes.includes(resourceType)) {
        existing.resourceTypes.push(resourceType);
      }
      // 최근 사용 목록의 맨 앞으로 이동
      preferences.recentlyUsed.splice(existingIndex, 1);
      preferences.recentlyUsed.unshift(existing);
    } else {
      // 새 기록 추가
      const newUsage: CredentialUsage = {
        credentialId,
        provider,
        lastUsedAt: new Date().toISOString(),
        usageCount: 1,
        resourceTypes: [resourceType],
      };
      preferences.recentlyUsed.unshift(newUsage);
      
      // 최대 개수 제한
      if (preferences.recentlyUsed.length > MAX_RECENTLY_USED) {
        preferences.recentlyUsed = preferences.recentlyUsed.slice(0, MAX_RECENTLY_USED);
      }
    }
    
    savePreferences(workspaceId, preferences);
  } catch (error) {
    logger.warn('Failed to track credential usage', error, { workspaceId, credentialId, provider, resourceType });
  }
}

/**
 * Provider별 기본 자격증명 설정
 */
export function setDefaultCredential(
  workspaceId: string,
  provider: CloudProvider,
  credentialId: string
): void {
  try {
    const preferences = loadPreferences(workspaceId);
    preferences.defaultCredentials[provider] = credentialId;
    savePreferences(workspaceId, preferences);
  } catch (error) {
    logger.warn('Failed to set default credential', error, { workspaceId, provider, credentialId });
  }
}

/**
 * Provider별 기본 자격증명 가져오기
 */
export function getDefaultCredential(
  workspaceId: string,
  provider: CloudProvider
): string | null {
  try {
    const preferences = loadPreferences(workspaceId);
    return preferences.defaultCredentials[provider] || null;
  } catch (error) {
    logger.warn('Failed to get default credential', error, { workspaceId, provider });
    return null;
  }
}

/**
 * 최근 사용한 자격증명 가져오기
 */
export function getRecentlyUsedCredentials(
  workspaceId: string,
  provider?: CloudProvider,
  limit: number = 5
): CredentialUsage[] {
  try {
    const preferences = loadPreferences(workspaceId);
    let recentlyUsed = preferences.recentlyUsed;
    
    // Provider 필터링
    if (provider) {
      recentlyUsed = recentlyUsed.filter((usage) => usage.provider === provider);
    }
    
    // 최신순 정렬 및 제한
    return recentlyUsed
      .sort((a, b) => new Date(b.lastUsedAt).getTime() - new Date(a.lastUsedAt).getTime())
      .slice(0, limit);
  } catch (error) {
    logger.warn('Failed to get recently used credentials', error, { workspaceId, provider, limit });
    return [];
  }
}

/**
 * 자격증명 추천 (우선순위 기반)
 */
export function getRecommendedCredential(
  workspaceId: string,
  credentials: Credential[],
  provider?: CloudProvider,
  _resourceType?: 'kubernetes' | 'network' | 'compute'
): string | null {
  if (credentials.length === 0) {
    return null;
  }
  
  // Provider 필터링
  let filteredCredentials = credentials;
  if (provider) {
    filteredCredentials = credentials.filter((c) => c.provider === provider);
  }
  
  if (filteredCredentials.length === 0) {
    return null;
  }
  
  // 1. 자격증명이 1개면 자동 선택
  if (filteredCredentials.length === 1) {
    return filteredCredentials[0].id;
  }
  
  // 2. Provider별 기본 자격증명 확인
  if (provider) {
    const defaultId = getDefaultCredential(workspaceId, provider);
    if (defaultId && filteredCredentials.some((c) => c.id === defaultId)) {
      return defaultId;
    }
  }
  
  // 3. 최근 사용한 자격증명 확인
  const recentlyUsed = getRecentlyUsedCredentials(workspaceId, provider, 1);
  if (recentlyUsed.length > 0) {
    const recentId = recentlyUsed[0].credentialId;
    if (filteredCredentials.some((c) => c.id === recentId)) {
      return recentId;
    }
  }
  
  // 4. 첫 번째 자격증명 반환
  return filteredCredentials[0].id;
}

/**
 * 선호도 데이터 로드
 */
function loadPreferences(workspaceId: string): CredentialPreference {
  try {
    const stored = localStorage.getItem(`${PREFERENCE_STORAGE_KEY}-${workspaceId}`);
    if (stored) {
      const parsed = JSON.parse(stored);
      // 기본값 병합
      return {
        workspaceId,
        defaultCredentials: {
          aws: parsed.defaultCredentials?.aws || null,
          gcp: parsed.defaultCredentials?.gcp || null,
          azure: parsed.defaultCredentials?.azure || null,
          ncp: parsed.defaultCredentials?.ncp || null,
        },
        recentlyUsed: parsed.recentlyUsed || [],
      };
    }
  } catch (error) {
    logger.warn('Failed to load credential preferences', error, { workspaceId });
  }
  
  // 기본값 반환
  return {
    workspaceId,
    defaultCredentials: {
      aws: null,
      gcp: null,
      azure: null,
      ncp: null,
    },
    recentlyUsed: [],
  };
}

/**
 * 선호도 데이터 저장
 */
function savePreferences(workspaceId: string, preferences: CredentialPreference): void {
  try {
    localStorage.setItem(
      `${PREFERENCE_STORAGE_KEY}-${workspaceId}`,
      JSON.stringify(preferences)
    );
  } catch (error) {
    logger.warn('Failed to save credential preferences', error, { workspaceId });
  }
}

/**
 * 워크스페이스별 선호도 초기화
 */
export function clearPreferences(workspaceId: string): void {
  try {
    localStorage.removeItem(`${PREFERENCE_STORAGE_KEY}-${workspaceId}`);
  } catch (error) {
    logger.warn('Failed to clear credential preferences', error, { workspaceId });
  }
}

