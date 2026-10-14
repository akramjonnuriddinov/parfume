import { createI18n } from 'vue-i18n'
import type { Locale } from '../api/types'
import uz from './locales/uz'
import ru from './locales/ru'
import en from './locales/en'

const STORAGE_KEY = 'mariana-locale'

export const LOCALES: { code: Locale; label: string; short: string }[] = [
    { code: 'uz', label: 'O‘zbekcha', short: 'UZ' },
    { code: 'ru', label: 'Русский', short: 'RU' },
    { code: 'en', label: 'English', short: 'EN' },
]

function initialLocale(): Locale {
    try {
        const saved = localStorage.getItem(STORAGE_KEY)
        if (saved === 'uz' || saved === 'ru' || saved === 'en') return saved
    } catch {
        // storage unavailable
    }
    return 'uz'
}

export const i18n = createI18n({
    legacy: false,
    locale: initialLocale(),
    fallbackLocale: 'uz',
    messages: { uz, ru, en },
})

export function setLocale(locale: Locale): void {
    i18n.global.locale.value = locale
    document.documentElement.lang = locale
    try {
        localStorage.setItem(STORAGE_KEY, locale)
    } catch {
        // storage unavailable
    }
}

document.documentElement.lang = i18n.global.locale.value
