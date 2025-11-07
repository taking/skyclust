/**
 * useTranslation Hook
 * 
 * 다국어 번역을 위한 React 훅입니다.
 * next-intl의 useTranslations를 래핑하여 locale 상태 관리 기능을 추가했습니다.
 * 
 * @param namespace - 번역 네임스페이스 (선택사항)
 * @returns 번역 함수, 현재 locale, locale 변경 함수
 * 
 * @example
 * ```tsx
 * // 기본 사용
 * const { t } = useTranslation();
 * <div>{t('common.save')}</div>
 * 
 * // 네임스페이스 사용
 * const { t } = useTranslation('kubernetes');
 * <div>{t('cluster.name')}</div>
 * 
 * // Locale 변경
 * const { locale, setLocale } = useTranslation();
 * <button onClick={() => setLocale('ko')}>한국어</button>
 * ```
 */

'use client';

import { useTranslations as useNextIntlTranslations } from 'next-intl';
import { useLocaleStore } from '@/store/locale';

export function useTranslation(namespace?: string) {
  // 1. Locale 스토어에서 현재 locale과 변경 함수 가져오기
  const { locale, setLocale } = useLocaleStore();
  
  // 2. next-intl의 useTranslations 훅 사용 (네임스페이스 지원)
  const t = useNextIntlTranslations(namespace);

  // 3. 번역 함수, locale, locale 변경 함수 반환
  return {
    t,
    locale,
    setLocale,
  };
}

