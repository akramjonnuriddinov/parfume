<!-- ChatProductList.vue — perfume cards in a chat reply, optionally grouped into sets -->
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { ChatShownProduct } from '../../api/chat'
import type { Variant } from '../../api/types'
import { useCartStore } from '../../stores/cart'
import { useLocalized } from '../../utils/useLocalized'
import ChatProductCard from './ChatProductCard.vue'

const props = defineProps<{ products: ChatShownProduct[] }>()

const { t } = useI18n()
const { price } = useLocalized()
const cart = useCartStore()

/** The volume the assistant suggested, if it can still be bought. */
function suggested(item: ChatShownProduct): Variant | undefined {
    const v = item.product.variants.find((x) => x.id === item.variant_id)
    return v?.in_stock ? v : undefined
}

const groups = computed(() => {
    const map = new Map<string, ChatShownProduct[]>()
    for (const item of props.products) {
        const key = item.group ?? ''
        map.set(key, [...(map.get(key) ?? []), item])
    }
    return [...map.entries()].map(([name, items]) => ({
        name,
        items,
        total: items.reduce((sum, i) => sum + (suggested(i)?.price ?? 0) * i.quantity, 0),
    }))
})

function missing(items: ChatShownProduct[]): ChatShownProduct[] {
    return items.filter((i) => {
        const v = suggested(i)
        return v !== undefined && !cart.find(v.id)
    })
}

function addAll(items: ChatShownProduct[]): void {
    for (const i of missing(items)) {
        cart.add(i.product, suggested(i)!, i.quantity)
    }
    ElMessage.success({ message: t('chat.addedAll'), grouping: true })
}
</script>

<template>
    <div class="space-y-3">
        <section
            v-for="group in groups"
            :key="group.name"
            class="space-y-2"
            :class="group.name ? 'rounded-2xl border border-line bg-white/70 p-2' : ''"
        >
            <header
                v-if="group.name"
                class="flex items-baseline justify-between gap-2 px-1 pt-0.5"
            >
                <h3 class="text-sm font-semibold text-ink">{{ group.name }}</h3>
                <span class="text-xs text-muted">
                    {{ t('chat.setTotal') }}: <span class="font-semibold text-ink">{{ price(group.total) }}</span>
                </span>
            </header>

            <ChatProductCard
                v-for="item in group.items"
                :key="item.product.id"
                :item="item"
            />

            <button
                v-if="missing(group.items).length > (group.name ? 0 : 1)"
                type="button"
                class="flex h-10 w-full cursor-pointer items-center justify-center gap-2 rounded-xl border border-brand bg-white text-sm font-medium text-brand hover:bg-brand-soft"
                @click="addAll(group.items)"
            >
                <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d="M3 3h2l2.5 12h11L21 7H6" />
                    <circle cx="9" cy="20" r="1" />
                    <circle cx="18" cy="20" r="1" />
                </svg>
                {{
                    group.name
                        ? t('chat.addSet', { n: missing(group.items).length })
                        : t('chat.addAll', { n: missing(group.items).length })
                }}
            </button>
        </section>
    </div>
</template>
