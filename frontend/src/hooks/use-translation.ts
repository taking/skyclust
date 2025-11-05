/**
 * useTranslation Hook
 * 다국어 번역 훅
 */

'use client';

import { useTranslations as useNextIntlTranslations } from 'next-intl';
import { useLocaleStore } from '@/store/locale';

export function useTranslation(namespace?: string) {
  const { locale, setLocale } = useLocaleStore();
  const t = useNextIntlTranslations(namespace);

  return {
    t,
    locale,
    setLocale,
  };
}

