import { createI18n } from "vue-i18n";

import {default as en} from './locales/en.json'
import {default as zh_Hans} from './locales/zh-Hans.json'

const loadLocaleMessages = () => {
  return {
    en: en,
    'zh-Hans': zh_Hans,
  }
}

export const i18n = createI18n({
  legacy: false,
  allowComposition: true,
  locale: 'en',
  fallbackLocale: 'en',
  messages: loadLocaleMessages(),
})
