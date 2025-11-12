/**
 * Credential Context Store
 * Credential과 Region 선택 상태를 전역으로 관리하는 스토어
 * 
 * Zustand persist를 사용하여 새로고침 시에도 상태 유지
 */

import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { devtools } from 'zustand/middleware';
import type { StateCreator } from 'zustand';

/**
 * 개발 환경 확인
 */
const isDevelopment = process.env.NODE_ENV === 'development';

interface CredentialContextState {
  selectedCredentialId: string | null;
  selectedRegion: string | null;
  selectedResourceGroup: string | null; // Azure-specific: Resource Group selection
  setSelectedCredential: (credentialId: string | null) => void;
  setSelectedRegion: (region: string | null) => void;
  setSelectedResourceGroup: (resourceGroup: string | null) => void;
  clearSelection: () => void;
}

const storeCreator: StateCreator<CredentialContextState> = (set) => ({
  // 초기 상태
  selectedCredentialId: null,
  selectedRegion: null,
  selectedResourceGroup: null,
  
  /**
   * 선택된 자격 증명 ID 설정
   * @param credentialId - 자격 증명 ID (null이면 선택 해제)
   */
  setSelectedCredential: (credentialId) => 
    set({ selectedCredentialId: credentialId }),
  
  /**
   * 선택된 리전 설정
   * @param region - 리전 (null이면 선택 해제)
   */
  setSelectedRegion: (region) => 
    set({ selectedRegion: region }),
  
  /**
   * 선택된 Resource Group 설정 (Azure 전용)
   * @param resourceGroup - Resource Group (null이면 선택 해제)
   */
  setSelectedResourceGroup: (resourceGroup) => 
    set({ selectedResourceGroup: resourceGroup }),
  
  /**
   * 선택 초기화
   * 자격 증명, 리전, Resource Group 선택을 모두 초기화합니다.
   */
  clearSelection: () => 
    set({ selectedCredentialId: null, selectedRegion: null, selectedResourceGroup: null }),
});

const persistedStore = persist(
  storeCreator,
  {
    name: 'credential-context-storage',
    storage: createJSONStorage(() => localStorage),
    partialize: (state) => ({
      selectedCredentialId: state.selectedCredentialId,
      selectedRegion: state.selectedRegion,
      selectedResourceGroup: state.selectedResourceGroup,
    }),
  }
);

// 개발 환경에서만 devtools 적용
export const useCredentialContextStore = (isDevelopment 
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ? create<CredentialContextState>()(devtools(persistedStore as any, { name: 'CredentialContextStore' }) as any)
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  : create<CredentialContextState>()(persistedStore as any)
// eslint-disable-next-line @typescript-eslint/no-explicit-any
) as any;

