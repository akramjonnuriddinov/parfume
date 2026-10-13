<!-- ChatWidget.vue — floating AI shopping assistant -->
<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { Locale } from '../../api/types'
import { useCartStore } from '../../stores/cart'
import { useChatStore, type ChatCartChange, type ChatEntry } from '../../stores/chat'
import { orderNumber } from '../../utils/format'
import { useLocalized } from '../../utils/useLocalized'
import { usePerfume } from '../../utils/usePerfume'
import ChatComparison from './ChatComparison.vue'
import ChatOrderCard from './ChatOrderCard.vue'
import ChatProductList from './ChatProductList.vue'

const { t, locale } = useI18n()
const { tr, price } = useLocalized()
const perfume = usePerfume()
const chat = useChatStore()
const cart = useCartStore()

const input = ref('')
const scroller = ref<HTMLElement | null>(null)
const textarea = ref<HTMLTextAreaElement | null>(null)

const suggestions = computed(() => [
    t('chat.suggestions.gift'),
    t('chat.suggestions.similar'),
    t('chat.suggestions.office'),
    t('chat.suggestions.decants'),
])

onMounted(() => chat.checkStatus())

async function scrollToBottom(): Promise<void> {
    await nextTick()
    scroller.value?.scrollTo({ top: scroller.value.scrollHeight, behavior: 'smooth' })
}

watch(() => [chat.entries.length, chat.pending], scrollToBottom)

const quickReplies = computed(() => {
    const last = chat.entries.at(-1)
    return !chat.pending && last?.role === 'assistant' ? (last.quickReplies ?? []) : []
})
watch(
    () => chat.open,
    async (open) => {
        if (!open) return
        await scrollToBottom()
        textarea.value?.focus()
    }
)

function submit(text = input.value): void {
    if (!text.trim() || chat.pending) return
    input.value = ''
    chat.send(text, locale.value as Locale)
}

function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Enter' && !event.shiftKey && !event.isComposing) {
        event.preventDefault()
        submit()
    }
}

/** Only the latest proposal can be confirmed, and not after an order was placed. */
const activeProposalId = computed(() => {
    for (let i = chat.entries.length - 1; i >= 0; i--) {
        const entry = chat.entries[i]
        if (entry.order) return null
        if (entry.proposal) return entry.proposalConfirmed ? null : entry.id
    }
    return null
})

async function confirmOrder(entry: ChatEntry): Promise<void> {
    try {
        const ok = await chat.confirmOrder(entry)
        if (!ok) ElMessage.warning(t('chat.order.expired'))
    } catch {
        ElMessage.error(t('chat.order.failed'))
    }
}

function changeLabel(change: ChatCartChange): string {
    const name = change.item
        ? `${change.item.brand} ${tr(change.item)}${change.volume ? `, ${perfume.volume(change.volume)}` : ''}`.trim()
        : ''
    switch (change.op) {
        case 'add':
            return t('chat.changeAdd', { name, n: change.quantity })
        case 'set':
            return t('chat.changeSet', { name, n: change.quantity })
        case 'remove':
            return t('chat.changeRemove', { name })
        default:
            return t('chat.changeClear')
    }
}

function isLast(entry: ChatEntry): boolean {
    return chat.entries.at(-1)?.id === entry.id
}
</script>

<template>
    <template v-if="chat.enabled">
        <button
            v-show="!chat.open"
            type="button"
            :aria-label="t('chat.open')"
            class="fixed right-4 bottom-4 z-50 flex h-14 cursor-pointer items-center gap-2 rounded-full bg-brand pr-5 pl-4 text-white shadow-lg shadow-brand/30 transition-colors hover:bg-brand-hover focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand sm:right-6 sm:bottom-6"
            @click="chat.open = true"
        >
            <svg class="size-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="M21 12a8 8 0 0 1-11.6 7.1L4 20l1-4.6A8 8 0 1 1 21 12Z" />
                <path d="M12 8.5v.01M9 11.5h.01M15 11.5h.01" stroke-width="2.4" />
            </svg>
            <span class="text-sm font-medium">{{ t('chat.launcher') }}</span>
        </button>

        <Transition
            enter-from-class="opacity-0 translate-y-4"
            enter-active-class="transition duration-200 ease-out"
            leave-to-class="opacity-0 translate-y-4"
            leave-active-class="transition duration-150 ease-in"
        >
            <section
                v-if="chat.open"
                role="dialog"
                :aria-label="t('chat.title')"
                class="fixed inset-0 z-50 flex flex-col bg-white sm:inset-auto sm:right-6 sm:bottom-6 sm:h-[min(680px,calc(100dvh-3rem))] sm:w-[420px] sm:overflow-hidden sm:rounded-2xl sm:border sm:border-line sm:shadow-2xl"
                @keydown.esc="chat.open = false"
            >
                <header class="flex items-center gap-3 border-b border-line px-4 py-3">
                    <span class="flex size-10 shrink-0 items-center justify-center rounded-full bg-brand-soft text-brand">
                        <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <path d="M12 3v3M8 21h8M12 18v3M5 9h14a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2Z" />
                            <path d="M9 13.5h.01M15 13.5h.01" stroke-width="2.6" />
                        </svg>
                    </span>
                    <div class="min-w-0 flex-1">
                        <h2 class="truncate text-base font-semibold text-ink">{{ t('chat.title') }}</h2>
                        <p class="truncate text-xs text-muted">{{ t('chat.subtitle') }}</p>
                    </div>
                    <button
                        v-if="chat.entries.length > 0"
                        type="button"
                        :title="t('chat.reset')"
                        :aria-label="t('chat.reset')"
                        :disabled="chat.pending"
                        class="cursor-pointer rounded-lg p-2 text-muted hover:bg-surface hover:text-ink disabled:opacity-40"
                        @click="chat.reset()"
                    >
                        <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <path d="M3 12a9 9 0 1 0 3-6.7L3 8" />
                            <path d="M3 3v5h5" />
                        </svg>
                    </button>
                    <button
                        type="button"
                        :aria-label="t('common.close')"
                        class="cursor-pointer rounded-lg p-2 text-muted hover:bg-surface hover:text-ink"
                        @click="chat.open = false"
                    >
                        <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true">
                            <path d="M6 6l12 12M18 6 6 18" />
                        </svg>
                    </button>
                </header>

                <div
                    ref="scroller"
                    class="flex-1 space-y-4 overflow-y-auto overscroll-contain bg-surface/60 px-4 py-4"
                    aria-live="polite"
                >
                    <div class="max-w-[90%] rounded-2xl rounded-tl-md bg-white px-4 py-3 text-sm leading-6 text-ink shadow-xs">
                        {{ t('chat.greeting') }}
                    </div>

                    <div
                        v-if="chat.entries.length === 0"
                        class="space-y-2"
                    >
                        <p class="text-xs font-medium text-muted">{{ t('chat.suggestionsTitle') }}</p>
                        <button
                            v-for="s in suggestions"
                            :key="s"
                            type="button"
                            class="block w-full cursor-pointer rounded-xl border border-line bg-white px-3.5 py-2.5 text-left text-sm text-ink transition-colors hover:border-brand hover:text-brand"
                            @click="submit(s)"
                        >
                            {{ s }}
                        </button>
                    </div>

                    <template
                        v-for="entry in chat.entries"
                        :key="entry.id"
                    >
                        <div
                            v-if="entry.role === 'user'"
                            class="ml-auto w-fit max-w-[85%] rounded-2xl rounded-tr-md bg-brand px-4 py-2.5 text-sm leading-6 whitespace-pre-line text-white"
                        >
                            {{ entry.content }}
                        </div>

                        <div
                            v-else-if="entry.failed"
                            class="max-w-[90%] rounded-2xl rounded-tl-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
                        >
                            {{ t('chat.failed') }}
                            <button
                                v-if="isLast(entry)"
                                type="button"
                                class="ml-1 cursor-pointer font-medium underline"
                                :disabled="chat.pending"
                                @click="chat.retry(locale as Locale)"
                            >
                                {{ t('chat.retry') }}
                            </button>
                        </div>

                        <div
                            v-else
                            class="space-y-2"
                        >
                            <div
                                v-if="entry.content"
                                class="max-w-[90%] rounded-2xl rounded-tl-md bg-white px-4 py-3 text-sm leading-6 whitespace-pre-line text-ink shadow-xs"
                            >
                                {{ entry.content }}
                            </div>

                            <ChatProductList
                                v-if="entry.products?.length"
                                :products="entry.products"
                            />

                            <ChatComparison
                                v-if="entry.comparison"
                                :comparison="entry.comparison"
                            />

                            <ChatOrderCard
                                v-if="entry.proposal"
                                :proposal="entry.proposal"
                                :confirmed="!!entry.proposalConfirmed"
                                :active="activeProposalId === entry.id"
                                :loading="chat.confirming"
                                @confirm="confirmOrder(entry)"
                            />

                            <div
                                v-if="entry.order"
                                class="flex gap-3 rounded-xl border border-emerald-200 bg-emerald-50 p-3"
                            >
                                <span class="flex size-9 shrink-0 items-center justify-center rounded-full bg-emerald-600 text-white">
                                    <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                        <path d="m5 12 5 5 9-10" />
                                    </svg>
                                </span>
                                <div class="text-sm leading-5">
                                    <p class="font-semibold text-emerald-900">
                                        {{ t('chat.order.placed', { number: orderNumber(entry.order.id) }) }}
                                    </p>
                                    <p class="text-emerald-800">
                                        {{ t('cart.total') }}: {{ price(entry.order.total) }}
                                    </p>
                                    <p class="mt-1 text-xs text-emerald-700">{{ t('chat.order.placedHint') }}</p>
                                </div>
                            </div>

                            <div
                                v-if="entry.leads?.length"
                                class="flex flex-wrap gap-1.5"
                            >
                                <span
                                    v-for="lead in entry.leads"
                                    :key="lead.id"
                                    class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2.5 py-1 text-xs text-emerald-700"
                                >
                                    <svg class="size-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                        <path d="M22 16.9v3a2 2 0 0 1-2.2 2 19.8 19.8 0 0 1-8.6-3.1 19.5 19.5 0 0 1-6-6A19.8 19.8 0 0 1 2.1 4.2 2 2 0 0 1 4.1 2h3a2 2 0 0 1 2 1.7c.1.9.4 1.8.7 2.7a2 2 0 0 1-.5 2.1L8 9.8a16 16 0 0 0 6 6l1.3-1.3a2 2 0 0 1 2.1-.4c.9.3 1.8.6 2.7.7a2 2 0 0 1 1.7 2Z" />
                                    </svg>
                                    {{ lead.kind === 'quote' ? t('chat.leadQuote') : t('chat.leadCallback') }}
                                </span>
                            </div>

                            <div
                                v-if="entry.changes?.length"
                                class="flex flex-wrap items-center gap-1.5"
                            >
                                <span
                                    v-for="(change, i) in entry.changes"
                                    :key="i"
                                    class="inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-xs"
                                    :class="change.op === 'remove' || change.op === 'clear' ? 'bg-slate-100 text-slate-600' : 'bg-emerald-50 text-emerald-700'"
                                >
                                    <svg class="size-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                        <path v-if="change.op === 'remove' || change.op === 'clear'" d="M5 12h14" />
                                        <path v-else d="m5 12 5 5 9-10" />
                                    </svg>
                                    {{ changeLabel(change) }}
                                </span>
                                <button
                                    v-if="cart.count > 0"
                                    type="button"
                                    class="cursor-pointer px-1 text-xs font-medium text-brand underline-offset-2 hover:underline"
                                    @click="cart.drawerOpen = true"
                                >
                                    {{ t('chat.openCart') }}
                                </button>
                            </div>
                        </div>
                    </template>

                    <div
                        v-if="quickReplies.length"
                        class="flex flex-wrap justify-end gap-2"
                    >
                        <button
                            v-for="option in quickReplies"
                            :key="option"
                            type="button"
                            class="cursor-pointer rounded-full border border-brand/40 bg-white px-3.5 py-1.5 text-sm text-brand transition-colors hover:border-brand hover:bg-brand-soft"
                            @click="submit(option)"
                        >
                            {{ option }}
                        </button>
                    </div>

                    <div
                        v-if="chat.pending"
                        class="flex w-fit items-center gap-1 rounded-2xl rounded-tl-md bg-white px-4 py-3.5 shadow-xs"
                        :aria-label="t('chat.thinking')"
                    >
                        <span
                            v-for="i in 3"
                            :key="i"
                            class="size-1.5 animate-bounce rounded-full bg-brand/60"
                            :style="{ animationDelay: `${(i - 1) * 150}ms` }"
                        />
                    </div>
                </div>

                <form
                    class="border-t border-line bg-white px-3 pt-3 pb-[max(0.75rem,env(safe-area-inset-bottom))]"
                    @submit.prevent="submit()"
                >
                    <div class="flex items-end gap-2 rounded-xl border border-line px-3 py-1.5 focus-within:border-brand">
                        <textarea
                            ref="textarea"
                            v-model="input"
                            rows="1"
                            maxlength="2000"
                            :placeholder="t('chat.placeholder')"
                            :aria-label="t('chat.placeholder')"
                            class="field-sizing-content max-h-32 min-h-9 flex-1 resize-none bg-transparent py-1.5 text-sm leading-6 text-ink outline-none placeholder:text-slate-400"
                            @keydown="onKeydown"
                        />
                        <button
                            type="submit"
                            :aria-label="t('chat.send')"
                            :disabled="!input.trim() || chat.pending"
                            class="mb-0.5 flex size-9 shrink-0 cursor-pointer items-center justify-center rounded-lg bg-brand text-white hover:bg-brand-hover disabled:cursor-not-allowed disabled:bg-slate-200 disabled:text-slate-400"
                        >
                            <svg class="size-4.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                <path d="M5 12h14M13 6l6 6-6 6" />
                            </svg>
                        </button>
                    </div>
                    <p class="mt-2 text-center text-[11px] leading-4 text-slate-400">
                        {{ t('chat.disclaimer') }}
                    </p>
                </form>
            </section>
        </Transition>
    </template>
</template>
