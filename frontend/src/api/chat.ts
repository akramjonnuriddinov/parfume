import { request } from './http'
import type { Locale, Product, Variant } from './types'

/** A cart line sent to the assistant: always a specific volume. */
export interface ChatCartRef {
    variant_id: number
    quantity: number
}

export interface ChatShownProduct {
    product: Product
    /** Volume the assistant suggests; 0 when none is in stock. */
    variant_id: number
    quantity: number
    note?: string
    group?: string
}

export type ChatCartAction =
    | { op: 'add' | 'set'; variant_id: number; quantity: number; product: Product }
    | { op: 'remove'; variant_id: number }
    | { op: 'clear' }

export interface ChatComparison {
    products: Product[]
    rows: { label: string; values: string[] }[]
    verdict: string
}

export interface ChatOrderProposal {
    full_name: string
    phone: string
    comment: string
    items: { product: Product; variant: Variant; quantity: number; subtotal: number }[]
    total: number
}

export interface ChatPlacedOrder {
    id: number
    total: number
}

export interface ChatLead {
    kind: 'callback' | 'quote'
    id: number
}

export interface ChatRequest {
    session_id: string
    message: string
    cart: ChatCartRef[]
    locale: Locale
}

export interface ChatReply {
    session_id: string
    reply: string
    products: ChatShownProduct[]
    cart_actions: ChatCartAction[]
    open_cart: boolean
    comparison?: ChatComparison
    order_proposal?: ChatOrderProposal
    order?: ChatPlacedOrder
    quick_replies: string[]
    leads: ChatLead[]
}

export const chatApi = {
    status: () => request<{ enabled: boolean }>('/chat'),
    send: (payload: ChatRequest) => request<ChatReply>('/chat', { method: 'POST', body: payload }),
    confirmOrder: (sessionId: string) =>
        request<ChatPlacedOrder>('/chat/order', {
            method: 'POST',
            body: { session_id: sessionId },
        }),
}
