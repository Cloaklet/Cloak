import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { createI18n } from 'vue-i18n'
import {default as en} from './locales/en.json'
import {default as zh_Hans} from './locales/zh-Hans.json'

const loadLocaleMessages = () => {
  return {
    en: en,
    'zh-Hans': zh_Hans,
  }
}

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  messages: loadLocaleMessages(),
})
const app = createApp(App)

app.use(createPinia())
app.use(i18n)

app.mount('#app-root')
