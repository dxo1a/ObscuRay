import { createI18n } from "vue-i18n";
import ru from './locales/ru.json' with { type: "json" }
import en from './locales/en.json' with { type: "json" }

export const i18n = createI18n({
    legacy: false,
    locale: navigator.language.startsWith('ru') ? 'ru' : 'en',
    fallbackLocale: 'en',
    messages: {
        ru,
        en,
    },
})