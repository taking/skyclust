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
  setSelectedCredential: (credentialId: string | null) => void;
  setSelectedRegion: (region: string | null) => void;
  clearSelection: () => void;
}

const storeCreator: StateCreator<CredentialContextState> = (set) => ({
  selectedCredentialId: null,
  selectedRegion: null,
  setSelectedCredential: (credentialId) => 
    set({ selectedCredentialId: credentialId }),
  setSelectedRegion: (region) => 
    set({ selectedRegion: region }),
  clearSelection: () => 
    set({ selectedCredentialId: null, selectedRegion: null }),
});

const persistedStore = persist(
  storeCreator,
  {
    name: 'credential-context-storage',
    storage: createJSONStorage(() => localStorage),
    partialize: (state) => ({
      selectedCredentialId: state.selectedCredentialId,
      selectedRegion: state.selectedRegion,
    }),
  }
);

// 개발 환경에서만 devtools 적용
export const useCredentialContextStore = create<CredentialContextState>()(
  isDevelopment 
    ? (devtools(persistedStore, { name: 'CredentialContextStore' }) as StateCreator<CredentialContextState>)
    : persistedStore
);

