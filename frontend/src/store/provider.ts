/**
 * Provider Store
 * 클라우드 프로바이더 선택 상태 관리 스토어
 * 
 * Zustand Devtools 지원 (개발 환경에서만 활성화)
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type { StateCreator } from 'zustand';
import { CloudProvider } from '@/lib/types';

/**
 * 개발 환경 확인
 */
const isDevelopment = process.env.NODE_ENV === 'development';

interface ProviderState {
  selectedProvider: CloudProvider | null;
  setSelectedProvider: (provider: CloudProvider | null) => void;
}

const storeCreator: StateCreator<ProviderState> = (set) => ({
  selectedProvider: null,
  setSelectedProvider: (provider) => set({ selectedProvider: provider }),
});

// 개발 환경에서만 devtools 적용
export const useProviderStore = create<ProviderState>()(
  isDevelopment 
    ? (devtools(storeCreator, { name: 'ProviderStore' }) as StateCreator<ProviderState>)
    : storeCreator
);
