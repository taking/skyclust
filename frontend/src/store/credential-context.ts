/**
 * Credential Context Store
 * Multi-provider 지원을 위한 Credential 상태 관리
 * 
 * 개선 사항:
 * - Multi-credential 선택 지원 (배열)
 * - Provider별 credential 그룹화
 * - Region 필터링 지원
 */

import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { devtools } from 'zustand/middleware';
import type { StateCreator } from 'zustand';
import type { CloudProvider } from '@/lib/types/kubernetes';
import type { ProviderRegionSelection } from '@/hooks/use-provider-region-filter';

const isDevelopment = process.env.NODE_ENV === 'development';

interface CredentialContextState {
  selectedCredentialId: string | null;
  
  selectedCredentialIds: string[];
  isMultiSelectMode: boolean;
  
  selectedRegion: string | null;
  selectedResourceGroup: string | null;
  
  // Provider별 Region 선택 (localStorage에 저장)
  providerSelectedRegions: ProviderRegionSelection;
  
  setSelectedCredential: (credentialId: string | null) => void;
  
  setSelectedCredentials: (credentialIds: string[]) => void;
  addCredential: (credentialId: string) => void;
  removeCredential: (credentialId: string) => void;
  toggleCredential: (credentialId: string) => void;
  setMultiSelectMode: (enabled: boolean) => void;
  
  setSelectedRegion: (region: string | null) => void;
  setSelectedResourceGroup: (resourceGroup: string | null) => void;
  
  // Provider별 Region 선택 관리
  setProviderSelectedRegions: (regions: ProviderRegionSelection) => void;
  
  clearSelection: () => void;
  getSelectedCredentialsByProvider: (credentials: Array<{ id: string; provider: CloudProvider }>) => Record<CloudProvider, string[]>;
}

const storeCreator: StateCreator<CredentialContextState> = (set, get) => ({
  selectedCredentialId: null,
  selectedCredentialIds: [],
  isMultiSelectMode: false,
  selectedRegion: null,
  selectedResourceGroup: null,
  providerSelectedRegions: {
    aws: [],
    gcp: [],
    azure: [],
  },
  
  setSelectedCredential: (credentialId) => 
    set({ 
      selectedCredentialId: credentialId,
      selectedCredentialIds: credentialId ? [credentialId] : [],
      isMultiSelectMode: false,
    }),
  
  setSelectedCredentials: (credentialIds) =>
    set({
      selectedCredentialIds: credentialIds,
      selectedCredentialId: credentialIds.length === 1 ? credentialIds[0] : null,
      isMultiSelectMode: credentialIds.length > 1,
    }),
  
  addCredential: (credentialId) =>
    set((state) => {
      if (state.selectedCredentialIds.includes(credentialId)) {
        return state;
      }
      return {
        selectedCredentialIds: [...state.selectedCredentialIds, credentialId],
        isMultiSelectMode: true,
        selectedCredentialId: state.selectedCredentialIds.length === 0 ? credentialId : null,
      };
    }),
  
  removeCredential: (credentialId) =>
    set((state) => {
      const updated = state.selectedCredentialIds.filter(id => id !== credentialId);
      return {
        selectedCredentialIds: updated,
        isMultiSelectMode: updated.length > 1,
        selectedCredentialId: updated.length === 1 ? updated[0] : null,
      };
    }),
  
  toggleCredential: (credentialId) => {
    const { selectedCredentialIds } = get();
    if (selectedCredentialIds.includes(credentialId)) {
      get().removeCredential(credentialId);
    } else {
      get().addCredential(credentialId);
    }
  },
  
  setMultiSelectMode: (enabled) =>
    set({ isMultiSelectMode: enabled }),
  
  setSelectedRegion: (region) => 
    set({ selectedRegion: region }),
  
  setSelectedResourceGroup: (resourceGroup) => 
    set({ selectedResourceGroup: resourceGroup }),
  
  setProviderSelectedRegions: (regions) =>
    set({ providerSelectedRegions: regions }),
  
  clearSelection: () => 
    set({ 
      selectedCredentialId: null,
      selectedCredentialIds: [],
      isMultiSelectMode: false,
      selectedRegion: null,
      selectedResourceGroup: null,
      providerSelectedRegions: {
        aws: [],
        gcp: [],
        azure: [],
      },
    }),
  
  getSelectedCredentialsByProvider: (credentials) => {
    const { selectedCredentialIds } = get();
    const result: Record<CloudProvider, string[]> = {
      aws: [],
      gcp: [],
      azure: [],
    };
    
    selectedCredentialIds.forEach(credentialId => {
      const credential = credentials.find(c => c.id === credentialId);
      if (credential && credential.provider in result) {
        result[credential.provider as CloudProvider].push(credentialId);
      }
    });
    
    return result;
  },
});

const persistedStore = persist(
  storeCreator,
  {
    name: 'credential-context-storage',
    storage: createJSONStorage(() => localStorage),
    partialize: (state) => ({
      selectedCredentialId: state.selectedCredentialId,
      selectedCredentialIds: state.selectedCredentialIds,
      selectedRegion: state.selectedRegion,
      selectedResourceGroup: state.selectedResourceGroup,
      providerSelectedRegions: state.providerSelectedRegions,
    }),
  }
);

export const useCredentialContextStore = create<CredentialContextState>()(
  isDevelopment 
    ? (devtools(persistedStore as StateCreator<CredentialContextState>, { name: 'CredentialContextStore' }) as StateCreator<CredentialContextState>)
    : persistedStore
);

