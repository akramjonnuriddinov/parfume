<!-- ChatComparison.vue — side-by-side product comparison in a chat reply -->
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { ChatComparison } from '../../api/chat'
import type { Product } from '../../api/types'
import { useCartStore } from '../../stores/cart'
import { useInquiryStore } from '../../stores/inquiry'
import { useLocalized } from '../../utils/useLocalized'
import { usePerfume } from '../../utils/usePerfume'
import ProductImage from '../ProductImage.vue'

const props = defineProps<{ comparison: ChatComparison }>()

const { t } = useI18n()
const { tr, price } = useLocalized()
const perfume = usePerfume()
const cart = useCartStore()
const inquiry = useInquiryStore()

const rows = computed(() => {
    const ps = props.comparison.products
    const notes = (p: Product) => [p.notes_top, p.notes_heart, p.notes_base].filter(Boolean).join(' · ') || '—'
    return [
        {
            label: t('chat.compare.price'),
            values: ps.map((p) => (p.min_price !== null ? t('product.priceFrom', { price: price(p.min_price) }) : t('product.outOfStock'))),
        },
        { label: t('chat.compare.gender'), values: ps.map((p) => perfume.gender(p.gender)) },
        { label: t('chat.compare.concentration'), values: ps.map((p) => perfume.concentration(p.concentration)) },
        { label: t('chat.compare.notes'), values: ps.map(notes) },
        ...props.comparison.rows,
    ]
})

function add(product: Product): void {
    const variant = cart.defaultVariant(product)
    if (!variant) return
    cart.add(product, variant)
    ElMessage.success({ message: t('cart.added', { name: `${tr(product)}, ${perfume.volume(variant)}` }), grouping: true })
}

function inCart(product: Product): boolean {
    return product.variants.some((v) => cart.find(v.id))
}
</script>

<template>
    <div class="overflow-hidden rounded-xl border border-line bg-white">
        <div class="overflow-x-auto">
            <table class="w-full min-w-[320px] table-fixed border-collapse text-xs">
                <thead>
                    <tr>
                        <th class="w-24 border-b border-line bg-surface p-2 text-left font-medium text-muted">
                            {{ t('chat.compare.title') }}
                        </th>
                        <th
                            v-for="product in comparison.products"
                            :key="product.id"
                            class="border-b border-l border-line p-2 align-top font-normal"
                        >
                            <RouterLink
                                :to="`/products/${product.id}`"
                                class="flex flex-col items-center gap-1.5 text-center hover:text-brand"
                            >
                                <span class="size-12 overflow-hidden rounded-lg bg-surface p-1">
                                    <ProductImage :src="product.image_url" />
                                </span>
                                <span class="block text-[10px] tracking-wider text-gold uppercase">{{ product.brand }}</span>
                                <span class="line-clamp-2 text-xs leading-4 font-medium text-ink">{{ tr(product) }}</span>
                            </RouterLink>
                        </th>
                    </tr>
                </thead>
                <tbody>
                    <tr
                        v-for="row in rows"
                        :key="row.label"
                        class="even:bg-surface/50"
                    >
                        <th
                            scope="row"
                            class="border-b border-line p-2 text-left align-top font-medium text-muted"
                        >
                            {{ row.label }}
                        </th>
                        <td
                            v-for="(value, i) in row.values"
                            :key="i"
                            class="border-b border-l border-line p-2 align-top leading-4 text-ink"
                        >
                            {{ value }}
                        </td>
                    </tr>
                    <tr>
                        <td class="p-2" />
                        <td
                            v-for="product in comparison.products"
                            :key="product.id"
                            class="border-l border-line p-2 text-center"
                        >
                            <span
                                v-if="inCart(product)"
                                class="text-xs font-medium text-emerald-700"
                            >
                                {{ t('chat.inCart') }}
                            </span>
                            <button
                                v-else-if="!product.in_stock"
                                type="button"
                                class="h-8 w-full cursor-pointer rounded-lg border border-brand px-2 text-xs font-medium text-brand hover:bg-brand-soft"
                                @click="inquiry.open(product)"
                            >
                                {{ t('product.notifyMe') }}
                            </button>
                            <button
                                v-else
                                type="button"
                                :disabled="!product.in_stock"
                                class="h-8 w-full cursor-pointer rounded-lg bg-brand px-2 text-xs font-medium text-white hover:bg-brand-hover disabled:cursor-not-allowed disabled:bg-slate-300"
                                @click="add(product)"
                            >
                                {{ t('chat.addToCart') }}
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
        <p
            v-if="comparison.verdict"
            class="border-t border-line bg-brand-soft/60 px-3 py-2 text-xs leading-5 text-ink"
        >
            {{ comparison.verdict }}
        </p>
    </div>
</template>
