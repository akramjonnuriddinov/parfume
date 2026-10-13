import { useI18n } from 'vue-i18n'
import type { Locale } from '../api/types'
import { formatDate, formatPrice, localized } from './format'

export function useLocalized() {
    const { locale } = useI18n()
    const current = () => locale.value as Locale

    return {
        tr: (obj: object | null | undefined, field: 'name' | 'description' = 'name') =>
            localized(obj, field, current()),
        price: (value: number) => formatPrice(value, current()),
        date: (iso: string, withTime = true) => formatDate(iso, current(), withTime),
        setting: (settings: Record<string, string>, key: string) =>
            settings[`${key}_${current()}`] || settings[`${key}_uz`] || '',
    }
}
