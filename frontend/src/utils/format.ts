import type { Locale } from '../api/types'

const currency: Record<Locale, string> = {
    uz: 'so‘m',
    ru: 'сум',
    en: 'UZS',
}

const numberFormats: Record<Locale, Intl.NumberFormat> = {
    uz: new Intl.NumberFormat('ru-RU'),
    ru: new Intl.NumberFormat('ru-RU'),
    en: new Intl.NumberFormat('en-US'),
}

const dateLocales: Record<Locale, string> = {
    uz: 'ru-RU',
    ru: 'ru-RU',
    en: 'en-GB',
}

export function formatPrice(value: number, locale: Locale): string {
    return `${numberFormats[locale].format(value)} ${currency[locale]}`
}

export function formatDate(iso: string, locale: Locale, withTime = true): string {
    return new Intl.DateTimeFormat(dateLocales[locale], {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        ...(withTime ? { hour: '2-digit', minute: '2-digit' } : {}),
    }).format(new Date(iso))
}

/** +998901234567 → +998 90 123 45 67 */
export function formatPhone(phone: string): string {
    const m = phone.match(/^\+?998(\d{2})(\d{3})(\d{2})(\d{2})$/)
    return m ? `+998 ${m[1]} ${m[2]} ${m[3]} ${m[4]}` : phone
}

/** Formats up to 9 local digits as "90 123 45 67" while typing. */
export function formatLocalPhoneInput(value: string): string {
    const d = value.replace(/\D/g, '').slice(0, 9)
    return [d.slice(0, 2), d.slice(2, 5), d.slice(5, 7), d.slice(7, 9)]
        .filter(Boolean)
        .join(' ')
}

export function telHref(phone: string): string {
    return `tel:${phone.replace(/[^\d+]/g, '')}`
}

export function orderNumber(id: number): string {
    return `#${String(id).padStart(5, '0')}`
}

type LocalizedField = 'name' | 'description'

export function localized(
    obj: object | null | undefined,
    field: LocalizedField,
    locale: Locale
): string {
    if (!obj) return ''
    const record = obj as Record<string, unknown>
    const value = record[`${field}_${locale}`]
    return typeof value === 'string' && value
        ? value
        : String(record[`${field}_uz`] ?? '')
}
