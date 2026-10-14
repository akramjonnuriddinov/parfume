import { useI18n } from 'vue-i18n'
import type { Concentration, Gender, Tier, Variant } from '../api/types'

/** Labels shared by cards, the product page, the cart and the admin. */
export function usePerfume() {
    const { t } = useI18n()

    return {
        gender: (g: Gender) => t(`perfume.gender.${g}`),
        concentration: (c: Concentration) => t(`perfume.concentration.${c}`),
        tier: (x: Tier) => (x ? t(`perfume.tier.${x}`) : ''),
        tierInfo: (x: Tier) => (x ? t(`perfume.tierInfo.${x}`) : ''),
        /** "10 ml · raspiv" / "100 ml" */
        volume: (v: Pick<Variant, 'volume_ml' | 'kind'>) =>
            v.kind === 'decant' ? `${v.volume_ml} ml · ${t('perfume.decant')}` : `${v.volume_ml} ml`,
        notes: (raw: string) =>
            raw
                .split(',')
                .map((n) => n.trim())
                .filter(Boolean),
    }
}
