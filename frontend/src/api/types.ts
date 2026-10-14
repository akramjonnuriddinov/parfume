export type Locale = 'uz' | 'ru' | 'en'

export interface Localized {
    name_uz: string
    name_ru: string
    name_en: string
}

export interface CategoryRef extends Localized {
    id: number
    slug: string
}

export interface Category extends CategoryRef {
    description_uz: string
    description_ru: string
    description_en: string
    sort_order: number
    product_count: number
    created_at: string
    updated_at: string
}

export type CategoryInput = Omit<
    Category,
    'id' | 'product_count' | 'created_at' | 'updated_at'
>

export type Gender = 'women' | 'men' | 'unisex'
export const GENDERS: Gender[] = ['women', 'men', 'unisex']
/** Storefront filters: unisex perfumes also show up under women and men. */
export const SHOP_GENDERS: Gender[] = ['women', 'men', 'unisex']

/** Quality tier the shop sells the scent in; '' = not set. */
export type Tier = '' | 'premium' | 'standart' | 'luxcopy' | 'tester'
export const TIERS: Exclude<Tier, ''>[] = ['premium', 'standart', 'luxcopy', 'tester']

export type Concentration = 'edp' | 'edt' | 'parfum' | 'extrait' | 'cologne' | 'oil'
export const CONCENTRATIONS: Concentration[] = ['edp', 'edt', 'parfum', 'extrait', 'cologne', 'oil']

/** A full bottle or a decant ("raspiv") poured from one. */
export type VariantKind = 'bottle' | 'decant'

export interface Variant {
    id: number
    volume_ml: number
    kind: VariantKind
    price: number
    in_stock: boolean
    sort_order: number
}

export interface Product extends Localized {
    id: number
    category_id: number | null
    category: CategoryRef | null
    description_uz: string
    description_ru: string
    description_en: string
    brand: string
    gender: Gender
    concentration: Concentration
    tier: Tier
    notes_top: string
    notes_heart: string
    notes_base: string
    image_url: string
    is_active: boolean
    is_featured: boolean
    sort_order: number
    variants: Variant[]
    /** Cheapest in-stock volume; null when nothing is in stock. */
    min_price: number | null
    in_stock: boolean
    created_at: string
    updated_at: string
}

export type VariantInput = Omit<Variant, 'id'> & { id?: number }

export type ProductInput = Omit<
    Product,
    'id' | 'category' | 'created_at' | 'updated_at' | 'variants' | 'min_price' | 'in_stock'
> & { variants: VariantInput[] }

export interface Brand {
    name: string
    count: number
}

export interface Page<T> {
    items: T[]
    total: number
    page: number
    limit: number
}

export type OrderSource = 'site' | 'ai_chat'

export type OrderStatus = 'new' | 'processing' | 'completed' | 'cancelled'

export const ORDER_STATUSES: OrderStatus[] = [
    'new',
    'processing',
    'completed',
    'cancelled',
]

export interface OrderItem {
    id: number
    product_id: number | null
    variant_id: number | null
    product_name: string
    product_brand: string
    product_image: string
    volume_ml: number
    kind: VariantKind
    price: number
    quantity: number
    subtotal: number
}

export interface Order {
    id: number
    full_name: string
    phone: string
    comment: string
    status: OrderStatus
    total: number
    admin_note: string
    locale: Locale
    source: OrderSource
    items_count: number
    items?: OrderItem[]
    created_at: string
    updated_at: string
}

export interface OrderPayload {
    full_name: string
    phone: string
    comment: string
    locale: Locale
    items: { variant_id: number; quantity: number }[]
    chat_session_id?: string
}

export interface Message {
    id: number
    full_name: string
    phone: string
    message: string
    product_id: number | null
    product_name: string
    is_read: boolean
    source: MessageSource
    chat_session_id: string | null
    created_at: string
}

export type MessageSource = 'site' | 'ai_callback' | 'ai_quote'

export interface MessagePayload {
    full_name: string
    phone: string
    message: string
    product_id?: number | null
}

export type Settings = Record<string, string>

export interface Stats {
    orders_total: number
    orders_by_status: Record<OrderStatus, number>
    revenue_total: number
    orders_today: number
    products_total: number
    products_active: number
    categories_total: number
    messages_unread: number
    recent_orders: Order[]
}

export interface Admin {
    id: number
    username: string
}

export interface ChatSessionSummary {
    id: string
    locale: Locale
    first_message: string
    messages_count: number
    leads_count: number
    order_id: number | null
    order_status: OrderStatus | null
    order_total: number | null
    created_at: string
    updated_at: string
}

export interface ChatTranscriptMeta {
    products?: {
        product_id: number
        variant_id?: number
        name: string
        volume?: string
        quantity: number
        group?: string
    }[]
    notes?: string[]
    order_id?: number
}

export interface ChatTranscriptMessage {
    id: number
    role: 'user' | 'assistant'
    content: string
    meta: ChatTranscriptMeta
    created_at: string
}

export interface ChatSessionDetail extends ChatSessionSummary {
    messages: ChatTranscriptMessage[]
    leads: Message[]
}

export interface MissingProduct {
    name: string
    count: number
    last_asked: string
    details: string[]
    session_ids: string[]
}
