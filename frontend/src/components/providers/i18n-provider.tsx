/**
 * i18n Provider
 * 다국어 제공자 컴포넌트
 */

'use client';

import { NextIntlClientProvider } from 'next-intl';
import { useLocaleStore } from '@/store/locale';
import { getMessages } from '@/i18n/i18n';
import { defaultLocale } from '@/i18n/config';
import { ReactNode, useEffect, useState } from 'react';

interface I18nProviderProps {
  children: ReactNode;
}

export function I18nProvider({ children }: I18nProviderProps) {
  const { locale } = useLocaleStore();
  const [isHydrated, setIsHydrated] = useState(false);
  const [currentLocale, setCurrentLocale] = useState<typeof defaultLocale>(defaultLocale);

  // Wait for Zustand persist to rehydrate
  useEffect(() => {
    // Zustand persist가 localStorage에서 데이터를 읽어오는 시간을 기다림
    const timer = setTimeout(() => {
      setIsHydrated(true);
      setCurrentLocale(locale);
    }, 0);

    return () => clearTimeout(timer);
  }, []);

  // locale이 변경되면 업데이트
  useEffect(() => {
    if (isHydrated) {
      setCurrentLocale(locale);
    }
  }, [locale, isHydrated]);

  // 항상 현재 locale의 messages를 가져옴 (기본값으로 fallback)
  const messages = getMessages(currentLocale);

  // NextIntlClientProvider를 항상 렌더링하여 useTranslations hook이 항상 컨텍스트를 사용할 수 있도록 함
  // isHydrated가 false여도 기본 locale과 messages로 렌더링하여 에러 방지
  return (
    <NextIntlClientProvider locale={currentLocale} messages={messages}>
      {children}
    </NextIntlClientProvider>
  );
}

