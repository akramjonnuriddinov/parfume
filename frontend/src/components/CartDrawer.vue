<!-- CartDrawer.vue -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useCartStore } from '../stores/cart'
import { useLocalized } from '../utils/useLocalized'
import { usePerfume } from '../utils/usePerfume'
import ProductImage from './ProductImage.vue'
import QuantityStepper from './QuantityStepper.vue'

const { t } = useI18n()
const { tr, price } = useLocalized()
const perfume = usePerfume()
const cart = useCartStore()
const router = useRouter()

function go(path: string): void {
    cart.drawerOpen = false
    router.push(path)
}
</script>

<template>
    <el-drawer
        v-model="cart.drawerOpen"
        :title="t('cart.title')"
        direction="rtl"
        size="min(420px, 100vw)"
    >
        <div
            v-if="cart.items.length === 0"
            class="flex h-full flex-col items-center justify-center text-center"
        >
            <span class="flex size-14 items-center justify-center rounded-full bg-brand-soft text-brand">
                <svg class="size-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d="M6 7h12l-1 13H7L6 7Z" />
                    <path d="M9 7V5a3 3 0 0 1 6 0v2" />
                </svg>
            </span>
            <p class="mt-4 text-base font-medium text-ink">{{ t('cart.empty') }}</p>
            <p class="mt-1 max-w-64 text-sm text-muted">{{ t('cart.emptyHint') }}</p>
            <el-button
                type="primary"
                class="mt-6! h-11! rounded-lg! px-6!"
                @click="go('/catalog')"
            >
                {{ t('cart.goCatalog') }}
            </el-button>
        </div>

        <div
            v-else
            class="flex h-full flex-col"
        >
            <ul class="-mt-2 flex-1 divide-y divide-line overflow-y-auto">
                <li
                    v-for="item in cart.items"
                    :key="item.variant_id"
                    class="flex gap-3 py-4"
                >
                    <button
                        type="button"
                        class="size-18 shrink-0 cursor-pointer overflow-hidden rounded-lg bg-surface p-2"
                        :aria-label="tr(item)"
                        @click="go(`/products/${item.product_id}`)"
                    >
                        <ProductImage :src="item.image_url" />
                    </button>

                    <div class="flex min-w-0 flex-1 flex-col">
                        <div class="flex items-start justify-between gap-2">
                            <button
                                type="button"
                                class="min-w-0 cursor-pointer text-left"
                                @click="go(`/products/${item.product_id}`)"
                            >
                                <span class="block truncate text-[11px] font-medium tracking-wider text-gold uppercase">{{ item.brand }}</span>
                                <span class="line-clamp-2 text-sm leading-5 font-medium text-ink hover:text-brand">{{ tr(item) }}</span>
                                <span class="block text-xs text-muted">{{ perfume.volume(item) }}</span>
                            </button>
                            <button
                                type="button"
                                :aria-label="t('cart.remove')"
                                class="shrink-0 cursor-pointer rounded p-1 text-slate-400 hover:text-red-500"
                                @click="cart.remove(item.variant_id)"
                            >
                                <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true">
                                    <path d="M6 6l12 12M18 6 6 18" />
                                </svg>
                            </button>
                        </div>
                        <div class="mt-auto flex items-center justify-between gap-2 pt-2">
                            <QuantityStepper
                                size="sm"
                                :model-value="item.quantity"
                                @update:model-value="cart.setQuantity(item.variant_id, $event)"
                                @remove="cart.remove(item.variant_id)"
                            />
                            <span class="text-sm font-semibold text-ink">
                                {{ price(item.price * item.quantity) }}
                            </span>
                        </div>
                    </div>
                </li>
            </ul>

            <div class="border-t border-line pt-4">
                <div class="flex items-center justify-between">
                    <span class="text-sm text-muted">{{ t('cart.total') }}</span>
                    <span class="text-lg font-semibold text-ink">{{ price(cart.total) }}</span>
                </div>
                <el-button
                    type="primary"
                    class="mt-4! h-12! w-full! rounded-lg! text-sm! font-medium!"
                    @click="go('/cart')"
                >
                    {{ t('cart.checkout') }}
                </el-button>
            </div>
        </div>
    </el-drawer>
</template>
