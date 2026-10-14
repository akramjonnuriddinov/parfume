<!-- ProductCard.vue -->
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { Product, Variant } from '../api/types'
import { useCartStore } from '../stores/cart'
import { useInquiryStore } from '../stores/inquiry'
import { useLocalized } from '../utils/useLocalized'
import { usePerfume } from '../utils/usePerfume'
import ProductImage from './ProductImage.vue'
import QuantityStepper from './QuantityStepper.vue'

const props = defineProps<{ product: Product }>()

const { t } = useI18n()
const { tr, price } = useLocalized()
const perfume = usePerfume()
const cart = useCartStore()
const inquiry = useInquiryStore()

const name = computed(() => tr(props.product))
const link = computed(() => `/products/${props.product.id}`)

// Cards show at most three volumes; the product page lists them all.
const shownVariants = computed(() => props.product.variants.slice(0, 3))
const selected = ref<Variant | undefined>()
watch(
    () => props.product,
    (p) => (selected.value = cart.defaultVariant(p) ?? p.variants[0]),
    { immediate: true }
)

const inCart = computed(() => (selected.value ? cart.find(selected.value.id) : undefined))

function addToCart(): void {
    if (!selected.value) return
    cart.add(props.product, selected.value)
    ElMessage.success({
        message: t('cart.added', { name: `${name.value}, ${perfume.volume(selected.value)}` }),
        grouping: true,
    })
}
</script>

<template>
    <article class="flex min-w-0 flex-col overflow-hidden rounded-xl border border-line bg-hero">
        <RouterLink
            :to="link"
            :aria-label="t('product.viewDetails', { name })"
            class="relative flex aspect-square items-center justify-center overflow-hidden rounded-xl bg-surface p-5 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand sm:p-8"
        >
            <ProductImage
                :src="product.image_url"
                class="transition-transform duration-300 hover:scale-105"
            />

            <span
                class="absolute top-3 left-3 rounded-md bg-white/90 px-2 py-1 text-[11px] font-medium text-muted">
                {{ perfume.gender(product.gender) }}
            </span>
            <span
                v-if="!product.in_stock"
                class="absolute top-3 right-3 rounded-md bg-white px-2 py-1 text-[11px] font-medium text-muted"
            >
                {{ t('product.outOfStock') }}
            </span>
        </RouterLink>

        <div class="flex flex-1 flex-col border-t border-line bg-white p-2 pt-4 sm:p-3 sm:pt-4">
            <p class="truncate text-[11px] font-medium tracking-[0.12em] text-gold uppercase">
                {{ product.brand || ' ' }}
            </p>

            <h3 class="mt-1">
                <RouterLink
                    :to="link"
                    class="line-clamp-2 text-left text-sm leading-6 font-medium text-ink hover:text-brand focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand sm:text-base"
                >
                    {{ name }}
                </RouterLink>
            </h3>
            <p class="flex flex-wrap items-center gap-x-1.5 gap-y-1 text-xs text-muted">
                <span
                    v-if="product.tier"
                    class="rounded border border-gold/40 bg-gold-soft px-1.5 py-px text-[10px] font-semibold tracking-wide text-ink uppercase"
                >{{ perfume.tier(product.tier) }}</span>
                {{ perfume.concentration(product.concentration) }}
            </p>

            <div class="mt-auto pt-3">
                <div
                    v-if="shownVariants.length > 1"
                    class="mb-3 flex flex-wrap gap-1.5"
                    role="radiogroup"
                    :aria-label="t('product.volume')"
                >
                    <button
                        v-for="v in shownVariants"
                        :key="v.id"
                        type="button"
                        role="radio"
                        :aria-checked="selected?.id === v.id"
                        :disabled="!v.in_stock"
                        class="cursor-pointer rounded-md border px-2 py-1 text-xs transition-colors disabled:cursor-not-allowed disabled:line-through disabled:opacity-50"
                        :class="selected?.id === v.id ? 'border-brand bg-brand-soft font-medium text-brand' : 'border-line text-ink hover:border-brand'"
                        @click="selected = v"
                    >
                        {{ v.volume_ml }} ml
                    </button>
                </div>

                <p class="text-sm font-semibold text-ink sm:text-base">
                    <template v-if="selected">{{ price(selected.price) }}</template>
                    <span
                        v-if="selected?.kind === 'decant'"
                        class="ml-1 text-xs font-normal text-muted"
                    >· {{ t('perfume.decant') }}</span>
                </p>

                <QuantityStepper
                    v-if="inCart && selected"
                    class="mt-3"
                    :model-value="inCart.quantity"
                    @update:model-value="cart.setQuantity(selected.id, $event)"
                    @remove="cart.remove(selected.id)"
                />

                <el-button
                    v-else-if="!product.in_stock"
                    class="mt-3! ml-0! h-11! w-full! rounded-lg! border-brand! px-3! text-sm! font-medium! text-brand! hover:bg-brand-soft!"
                    plain
                    @click="inquiry.open(product)"
                >
                    {{ t('product.notifyMe') }}
                </el-button>

                <el-button
                    v-else
                    class="mt-3! ml-0! h-11! w-full! rounded-lg! border-brand! bg-brand! px-3! text-sm! font-medium! text-white! hover:border-brand-hover! hover:bg-brand-hover!"
                    :disabled="!selected?.in_stock"
                    :aria-label="`${name} — ${t('product.addToCart')}`"
                    @click="addToCart"
                >
                    {{ selected?.in_stock ? t('product.addToCart') : t('product.outOfStock') }}
                </el-button>
            </div>
        </div>
    </article>
</template>
