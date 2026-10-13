import type { FormItemRule } from 'element-plus'
import { i18n } from '../i18n'

const t = (key: string) => i18n.global.t(key)

export function nameRules(): FormItemRule[] {
    return [
        { required: true, message: t('form.nameRequired'), trigger: 'blur' },
        { min: 2, max: 120, message: t('form.nameLength'), trigger: 'blur' },
    ]
}

export function phoneRules(): FormItemRule[] {
    return [
        { required: true, message: t('form.phoneRequired'), trigger: 'blur' },
        {
            validator: (_rule, value: string, callback) => {
                if (value.replace(/\D/g, '').length !== 9) {
                    callback(new Error(t('form.phoneInvalid')))
                } else {
                    callback()
                }
            },
            trigger: 'blur',
        },
    ]
}

/** "90 123 45 67" → "+998901234567" */
export function toFullPhone(local: string): string {
    return `+998${local.replace(/\D/g, '')}`
}
