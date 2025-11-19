/**
 * Date Formatting Utilities
 * 날짜 포맷팅 유틸리티 함수
 * 
 * 다국어 지원 및 Invalid Date 에러 방지를 포함한 날짜 포맷팅 함수들
 */

import { format } from 'date-fns';
import { ko, enUS } from 'date-fns/locale';

export type Locale = 'ko' | 'en';

/**
 * Locale에 따른 date-fns locale 객체 반환
 */
export function getDateFnsLocale(locale: Locale = 'ko') {
  return locale === 'ko' ? ko : enUS;
}

/**
 * 날짜 문자열을 Date 객체로 안전하게 변환
 * Invalid Date인 경우 null 반환
 */
export function parseDate(dateString: string | null | undefined): Date | null {
  if (!dateString) return null;
  
  const date = new Date(dateString);
  
  // Invalid Date 체크
  if (isNaN(date.getTime())) {
    return null;
  }
  
  return date;
}

/**
 * 날짜를 로케일별로 포맷팅
 * Invalid Date인 경우 fallback 반환
 */
export function formatDate(
  dateString: string | null | undefined,
  formatStr: string = 'yyyy-MM-dd',
  locale: Locale = 'ko',
  fallback: string = '-'
): string {
  const date = parseDate(dateString);
  if (!date) return fallback;
  
  try {
    return format(date, formatStr, {
      locale: getDateFnsLocale(locale),
    });
  } catch {
    return fallback;
  }
}

/**
 * 날짜를 로케일별 날짜 형식으로 포맷팅 (yyyy-MM-dd)
 */
export function formatDateOnly(
  dateString: string | null | undefined,
  locale: Locale = 'ko',
  fallback: string = '-'
): string {
  return formatDate(dateString, 'yyyy-MM-dd', locale, fallback);
}

/**
 * 날짜를 로케일별 날짜+시간 형식으로 포맷팅
 */
export function formatDateTime(
  dateString: string | null | undefined,
  locale: Locale = 'ko',
  fallback: string = '-'
): string {
  const formatStr = locale === 'ko' ? 'yyyy-MM-dd HH:mm:ss' : 'yyyy-MM-dd hh:mm:ss a';
  return formatDate(dateString, formatStr, locale, fallback);
}

/**
 * 날짜를 로케일별 짧은 날짜 형식으로 포맷팅
 */
export function formatShortDate(
  dateString: string | null | undefined,
  locale: Locale = 'ko',
  fallback: string = '-'
): string {
  const formatStr = locale === 'ko' ? 'yyyy.MM.dd' : 'MM/dd/yyyy';
  return formatDate(dateString, formatStr, locale, fallback);
}

/**
 * 날짜를 로케일별 시간 형식으로 포맷팅
 */
export function formatTime(
  dateString: string | null | undefined,
  locale: Locale = 'ko',
  fallback: string = '-'
): string {
  const formatStr = locale === 'ko' ? 'HH:mm:ss' : 'hh:mm:ss a';
  return formatDate(dateString, formatStr, locale, fallback);
}

/**
 * 날짜를 로케일별로 toLocaleDateString 사용
 * Invalid Date인 경우 fallback 반환
 */
export function toLocaleDateString(
  dateString: string | null | undefined,
  locale: Locale = 'ko',
  options?: Intl.DateTimeFormatOptions,
  fallback: string = '-'
): string {
  const date = parseDate(dateString);
  if (!date) return fallback;
  
  try {
    const localeString = locale === 'ko' ? 'ko-KR' : 'en-US';
    return date.toLocaleDateString(localeString, options);
  } catch {
    return fallback;
  }
}

/**
 * 날짜를 로케일별로 toLocaleTimeString 사용
 * Invalid Date인 경우 fallback 반환
 */
export function toLocaleTimeString(
  dateString: string | null | undefined,
  locale: Locale = 'ko',
  options?: Intl.DateTimeFormatOptions,
  fallback: string = '-'
): string {
  const date = parseDate(dateString);
  if (!date) return fallback;
  
  try {
    const localeString = locale === 'ko' ? 'ko-KR' : 'en-US';
    return date.toLocaleTimeString(localeString, options);
  } catch {
    return fallback;
  }
}

/**
 * 날짜를 로케일별로 toLocaleString 사용
 * Invalid Date인 경우 fallback 반환
 */
export function toLocaleString(
  dateString: string | null | undefined,
  locale: Locale = 'ko',
  options?: Intl.DateTimeFormatOptions,
  fallback: string = '-'
): string {
  const date = parseDate(dateString);
  if (!date) return fallback;
  
  try {
    const localeString = locale === 'ko' ? 'ko-KR' : 'en-US';
    return date.toLocaleString(localeString, options);
  } catch {
    return fallback;
  }
}

