import { create } from 'zustand';
import { CloudProvider } from '@/lib/types';

interface ProviderState {
  selectedProvider: CloudProvider | null;
  setSelectedProvider: (provider: CloudProvider | null) => void;
}

export const useProviderStore = create<ProviderState>((set) => ({
  selectedProvider: null,
  setSelectedProvider: (provider) => set({ selectedProvider: provider }),
}));

