import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { ApiError } from '../api/http'
import {
    chatApi,
    type ChatCartAction,
    type ChatComparison,
    type ChatLead,
    type ChatOrderProposal,
    type ChatPlacedOrder,
    type ChatShownProduct,
} from '../api/chat'
import type { Locale, Localized, VariantKind } from '../api/types'
import { useCartStore } from './cart'

const STORAGE_KEY = 'mariana-chat'
const SESSION_KEY = 'mariana-chat-session'
const MAX_STORED = 40

/** A cart change made by the assistant, shown as a chip under its reply. */
export interface ChatCartChange {
    op: ChatCartAction['op']
    item?: Localized & { brand: string }
    volume?: { volume_ml: number; kind: VariantKind }
    quantity?: number
}

export interface ChatEntry {
    id: number
    role: 'user' | 'assistant'
    content: string
    products?: ChatShownProduct[]
    changes?: ChatCartChange[]
    comparison?: ChatComparison
    proposal?: ChatOrderProposal
    /** Set once the proposal on this entry was confirmed. */
    proposalConfirmed?: boolean
    order?: ChatPlacedOrder
    leads?: ChatLead[]
    quickReplies?: string[]
    failed?: boolean
}

function loadJSON<T>(key: string, fallback: T): T {
    try {
        const raw = localStorage.getItem(key)
        return raw ? (JSON.parse(raw) as T) : fallback
    } catch {
        return fallback
    }
}

function save(key: string, value: unknown): void {
    try {
        localStorage.setItem(key, JSON.stringify(value))
    } catch {
        // storage unavailable
    }
}

export const useChatStore = defineStore('chat', () => {
    const cart = useCartStore()

    const enabled = ref(false)
    const open = ref(false)
    const pending = ref(false)
    const confirming = ref(false)
    const sessionId = ref<string>(loadJSON(SESSION_KEY, ''))
    const stored = loadJSON<ChatEntry[]>(STORAGE_KEY, [])
    const entries = ref<ChatEntry[]>(Array.isArray(stored) ? stored : [])
    let nextId = entries.value.reduce((max, e) => Math.max(max, e.id), 0) + 1

    watch(entries, (value) => save(STORAGE_KEY, value.slice(-MAX_STORED)), { deep: true })
    watch(sessionId, (value) => save(SESSION_KEY, value))

    async function checkStatus(): Promise<void> {
        try {
            enabled.value = (await chatApi.status()).enabled
        } catch {
            enabled.value = false
        }
    }

    function applyActions(actions: ChatCartAction[]): ChatCartChange[] {
        const changes: ChatCartChange[] = []
        for (const action of actions) {
            switch (action.op) {
                case 'clear':
                    cart.clear()
                    changes.push({ op: 'clear' })
                    break
                case 'remove': {
                    const item = cart.find(action.variant_id)
                    if (item) {
                        cart.remove(action.variant_id)
                        changes.push({ op: 'remove', item: { ...item }, volume: { ...item } })
                    }
                    break
                }
                case 'add':
                case 'set': {
                    const variant = action.product.variants.find((v) => v.id === action.variant_id)
                    if (!variant) break
                    if (action.op === 'set' && cart.find(variant.id)) {
                        cart.setQuantity(variant.id, action.quantity)
                    } else {
                        cart.add(action.product, variant, action.quantity)
                    }
                    changes.push({
                        op: action.op,
                        item: action.product,
                        volume: variant,
                        quantity: action.quantity,
                    })
                    break
                }
            }
        }
        return changes
    }

    async function send(text: string, locale: Locale): Promise<void> {
        const content = text.trim()
        if (!content || pending.value) return

        entries.value.push({ id: nextId++, role: 'user', content })
        pending.value = true

        try {
            const reply = await chatApi.send({
                session_id: sessionId.value,
                message: content,
                cart: cart.items.map((i) => ({ variant_id: i.variant_id, quantity: i.quantity })),
                locale,
            })
            sessionId.value = reply.session_id
            const changes = applyActions(reply.cart_actions)
            if (reply.order) {
                cart.clear()
            }
            entries.value.push({
                id: nextId++,
                role: 'assistant',
                content: reply.reply,
                products: reply.products,
                changes,
                comparison: reply.comparison,
                proposal: reply.order_proposal,
                order: reply.order,
                leads: reply.leads,
                quickReplies: reply.quick_replies,
            })
            if (reply.open_cart) {
                cart.drawerOpen = true
            }
        } catch {
            entries.value.push({ id: nextId++, role: 'assistant', content: '', failed: true })
        } finally {
            pending.value = false
        }
    }

    /** Drops a failed reply and re-sends the user message before it. */
    async function retry(locale: Locale): Promise<void> {
        const last = entries.value.at(-1)
        const question = entries.value.at(-2)
        if (!last?.failed || question?.role !== 'user') return
        entries.value.splice(-2, 2)
        await send(question.content, locale)
    }

    /**
     * Places the order proposed on the given entry. Returns false when the
     * proposal is no longer valid (e.g. already placed or replaced).
     */
    async function confirmOrder(entry: ChatEntry): Promise<boolean> {
        if (!entry.proposal || confirming.value) return false
        confirming.value = true
        try {
            const order = await chatApi.confirmOrder(sessionId.value)
            entry.proposalConfirmed = true
            cart.clear()
            entries.value.push({ id: nextId++, role: 'assistant', content: '', order })
            return true
        } catch (error) {
            if (error instanceof ApiError && error.status === 409) {
                entry.proposal = undefined
                return false
            }
            throw error
        } finally {
            confirming.value = false
        }
    }

    function reset(): void {
        entries.value = []
        sessionId.value = ''
    }

    return {
        enabled,
        open,
        pending,
        confirming,
        sessionId,
        entries,
        checkStatus,
        send,
        retry,
        confirmOrder,
        reset,
    }
})
