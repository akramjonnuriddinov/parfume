<!-- ChatProductCard.vue — compact perfume card in AI chat replies, with the suggested volume preselected -->
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { ChatShownProduct } from '../../api/chat'
import type { Variant } from '../../api/types'
import { useCartStore } from '../../stores/cart'
import { useInquiryStore } from '../../stores/inquiry'
import { useLocalized } from '../../utils/useLocalized'
import { usePerfume } from '../../utils/usePerfume'
import ProductImage from '../ProductImage.vue'
import QuantityStepper from '../QuantityStepper.vue'

const props = defineProps<{ item: ChatShownProduct }>()

const { t } = useI18n()
const { tr, price } = useLocalized()
const perfume = usePerfume()
const cart = useCartStore()
const inquiry = useInquiryStore()

const product = computed(() => props.item.product)
const name = computed(() => tr(product.value))
const link = computed(() => `/products/${product.value.id}`)

const selected = ref<Variant | undefined>()
watch(
    () => props.item,
    (item) => {
        selected.value =
            item.product.variants.find((v) => v.id === item.variant_id) ?? cart.defaultVariant(item.product)
    },
    { immediate: true }
)

const inCart = computed(() => (selected.value ? cart.find(selected.value.id) : undefined))

function addToCart(): void {
    if (!selected.value) return
    cart.add(product.value, selected.value, props.item.quantity)
    ElMessage.success({
        message: t('cart.added', { name: `${name.value}, ${perfume.volume(selected.value)}` }),
        grouping: true,
    })
}
</script>

<template>
    <article class="flex gap-3 rounded-xl border border-line bg-white p-2.5">
        <RouterLink
            :to="link"
            :aria-label="t('product.viewDetails', { name })"
            class="size-18 shrink-0 overflow-hidden rounded-lg bg-surface p-1.5"
        >
            <ProductImage :src="product.image_url" />
        </RouterLink>

        <div class="flex min-w-0 flex-1 flex-col">
            <p class="truncate text-[10px] font-medium tracking-wider text-gold uppercase">{{ product.brand }}</p>
            <RouterLink
                :to="link"
                class="line-clamp-1 text-sm leading-5 font-medium text-ink hover:text-brand"
            >
                {{ name }}
            </RouterLink>
            <p
                v-if="item.note"
                class="mt-0.5 line-clamp-2 text-xs leading-4 text-muted"
            >
                {{ item.note }}
            </p>

            <div
                v-if="product.variants.length > 1"
                class="mt-2 flex flex-wrap gap-1"
            >
                <button
                    v-for="v in product.variants"
                    :key="v.id"
                    type="button"
                    :disabled="!v.in_stock"
                    class="cursor-pointer rounded-md border px-1.5 py-0.5 text-[11px] transition-colors disabled:cursor-not-allowed disabled:line-through disabled:opacity-50"
                    :class="selected?.id === v.id ? 'border-brand bg-brand-soft font-medium text-brand' : 'border-line text-ink hover:border-brand'"
                    @click="selected = v"
                >
                    {{ v.volume_ml }} ml<template v-if="v.kind === 'decant'"> · {{ t('perfume.decant') }}</template>
                </button>
            </div>

            <div class="mt-auto flex flex-wrap items-center justify-between gap-2 pt-2">
                <p class="text-sm font-semibold text-ink">
                    {{ selected ? price(selected.price) : t('product.outOfStock') }}
                    <span
                        v-if="item.quantity > 1"
                        class="ml-1 text-xs font-normal text-muted"
                    >
                        {{ t('chat.suggested', { n: item.quantity }) }}
                    </span>
                </p>

                <QuantityStepper
                    v-if="inCart && selected"
                    size="sm"
                    :model-value="inCart.quantity"
                    @update:model-value="cart.setQuantity(selected.id, $event)"
                    @remove="cart.remove(selected.id)"
                />
                <button
                    v-else-if="!selected?.in_stock"
                    type="button"
                    class="h-9 cursor-pointer rounded-lg border border-brand px-3 text-xs font-medium text-brand hover:bg-brand-soft"
                    @click="inquiry.open(product)"
                >
                    {{ t('product.notifyMe') }}
                </button>
                <button
                    v-else
                    type="button"
                    :aria-label="`${name} — ${t('product.addToCart')}`"
                    class="h-9 cursor-pointer rounded-lg bg-brand px-3 text-xs font-medium text-white hover:bg-brand-hover"
                    @click="addToCart"
                >
                    {{ t('chat.addToCart') }}
                </button>
            </div>
        </div>
    </article>
</template>
