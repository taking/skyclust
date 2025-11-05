/**
 * i18n Instance
 * 다국어 인스턴스
 */

import { createTranslator } from 'next-intl';
import { defaultLocale } from './config';
import koMessages from './messages/ko.json';
import enMessages from './messages/en.json';

const messages = {
  ko: koMessages,
  en: enMessages,
};

export type Messages = typeof koMessages;

export function getTranslations(locale: string = defaultLocale) {
  return createTranslator({
    locale,
    messages: messages[locale as keyof typeof messages] || messages.ko,
  });
}

export function getMessages(locale: string = defaultLocale) {
  return messages[locale as keyof typeof messages] || messages.ko;
}

