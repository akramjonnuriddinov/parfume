import { request } from './http'
import type {
    Brand,
    Category,
    MessagePayload,
    OrderPayload,
    Page,
    Product,
    Settings,
} from './types'

export interface ProductQuery {
    category?: string
    gender?: string
    brand?: string
    tier?: string
    q?: string
    featured?: boolean
    sort?: string
    ids?: string
    page?: number
    limit?: number
}

export const publicApi = {
    categories: () => request<Category[]>('/categories'),
    brands: () => request<Brand[]>('/brands'),
    products: (query: ProductQuery = {}) =>
        request<Page<Product>>('/products', { query: { ...query } }),
    product: (id: number | string) => request<Product>(`/products/${id}`),
    settings: () => request<Settings>('/settings'),
    createOrder: (payload: OrderPayload) =>
        request<{ id: number; total: number; created_at: string }>('/orders', {
            method: 'POST',
            body: payload,
        }),
    createMessage: (payload: MessagePayload) =>
        request<{ id: number }>('/messages', {
            method: 'POST',
            body: payload,
        }),
}
