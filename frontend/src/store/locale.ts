/**
 * Locale Store
 * 언어 설정을 관리하는 Zustand 스토어
 */

import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { defaultLocale, type Locale } from '@/i18n/config';

interface LocaleState {
  locale: Locale;
  setLocale: (locale: Locale) => void;
}

export const useLocaleStore = create<LocaleState>()(
  persist(
    (set) => ({
      locale: defaultLocale,
      setLocale: (locale) => set({ locale }),
    }),
    {
      name: 'locale-storage',
    }
  )
);

