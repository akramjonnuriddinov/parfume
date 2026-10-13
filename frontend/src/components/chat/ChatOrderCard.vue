<!-- ChatOrderCard.vue — order summary the assistant proposed, with a confirm button -->
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ChatOrderProposal } from '../../api/chat'
import { useCartStore } from '../../stores/cart'
import { formatPhone } from '../../utils/format'
import { useLocalized } from '../../utils/useLocalized'
import { usePerfume } from '../../utils/usePerfume'

const props = defineProps<{
    proposal: ChatOrderProposal
    confirmed: boolean
    /** Only the latest proposal in the conversation can be confirmed. */
    active: boolean
    loading: boolean
}>()

const emit = defineEmits<{ confirm: [] }>()

const { t } = useI18n()
const { tr, price } = useLocalized()
const perfume = usePerfume()
const cart = useCartStore()

/** The proposal was built from the cart; it is stale once the cart differs. */
const matchesCart = computed(
    () =>
        props.proposal.items.length === cart.items.length &&
        props.proposal.items.every(
            (i) => cart.find(i.variant.id)?.quantity === i.quantity
        )
)
</script>

<template>
    <article class="overflow-hidden rounded-xl border border-brand/30 bg-white">
        <header class="flex items-center gap-2 border-b border-line bg-brand-soft px-3 py-2">
            <svg class="size-4 text-brand" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="M9 5H7a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2" />
                <rect x="9" y="3" width="6" height="4" rx="1" />
                <path d="M9 12h6M9 16h4" />
            </svg>
            <h3 class="text-sm font-semibold text-ink">{{ t('chat.order.title') }}</h3>
        </header>

        <ul class="divide-y divide-line px-3">
            <li
                v-for="item in proposal.items"
                :key="item.variant.id"
                class="flex items-start justify-between gap-3 py-2 text-sm"
            >
                <span class="min-w-0 text-ink">
                    <span class="line-clamp-2">{{ item.product.brand }} {{ tr(item.product) }}</span>
                    <span class="text-xs text-muted">{{ perfume.volume(item.variant) }} × {{ item.quantity }}</span>
                </span>
                <span class="shrink-0 font-medium text-ink">{{ price(item.subtotal) }}</span>
            </li>
        </ul>

        <div class="flex items-center justify-between border-t border-line px-3 py-2">
            <span class="text-sm text-muted">{{ t('cart.total') }}</span>
            <span class="text-base font-semibold text-ink">{{ price(proposal.total) }}</span>
        </div>

        <dl class="space-y-1 border-t border-line px-3 py-2 text-xs">
            <div class="flex gap-2">
                <dt class="w-20 shrink-0 text-muted">{{ t('chat.order.name') }}</dt>
                <dd class="text-ink">{{ proposal.full_name }}</dd>
            </div>
            <div class="flex gap-2">
                <dt class="w-20 shrink-0 text-muted">{{ t('chat.order.phone') }}</dt>
                <dd class="text-ink">{{ formatPhone(proposal.phone) }}</dd>
            </div>
            <div
                v-if="proposal.comment"
                class="flex gap-2"
            >
                <dt class="w-20 shrink-0 text-muted">{{ t('chat.order.comment') }}</dt>
                <dd class="text-ink">{{ proposal.comment }}</dd>
            </div>
        </dl>

        <div class="border-t border-line p-3">
            <p
                v-if="confirmed"
                class="flex items-center justify-center gap-1.5 text-sm font-medium text-emerald-700"
            >
                <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d="m5 12 5 5 9-10" />
                </svg>
                {{ t('chat.order.confirmed') }}
            </p>
            <p
                v-else-if="!active"
                class="text-center text-xs text-muted"
            >
                {{ t('chat.order.outdated') }}
            </p>
            <template v-else>
                <p
                    v-if="!matchesCart"
                    class="mb-2 text-center text-xs text-amber-700"
                >
                    {{ t('chat.order.cartChanged') }}
                </p>
                <button
                    type="button"
                    :disabled="!matchesCart || loading"
                    class="flex h-11 w-full cursor-pointer items-center justify-center gap-2 rounded-lg bg-brand text-sm font-medium text-white hover:bg-brand-hover disabled:cursor-not-allowed disabled:bg-slate-300"
                    @click="emit('confirm')"
                >
                    <span
                        v-if="loading"
                        class="size-4 animate-spin rounded-full border-2 border-white/40 border-t-white"
                        aria-hidden="true"
                    />
                    {{ t('chat.order.confirm') }}
                </button>
                <p class="mt-2 text-center text-[11px] leading-4 text-muted">
                    {{ t('chat.order.hint') }}
                </p>
            </template>
        </div>
    </article>
</template>
