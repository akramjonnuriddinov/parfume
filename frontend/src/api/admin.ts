import { request } from './http'
import { useAuthStore } from '../stores/auth'
import type {
    Admin,
    Category,
    CategoryInput,
    ChatSessionDetail,
    ChatSessionSummary,
    Message,
    MissingProduct,
    Order,
    OrderStatus,
    Page,
    Product,
    ProductInput,
    Settings,
    Stats,
} from './types'

function authed<T>(
    path: string,
    options: Parameters<typeof request>[1] = {}
): Promise<T> {
    return request<T>(path, { ...options, token: useAuthStore().token })
}

export const adminApi = {
    login: (username: string, password: string) =>
        request<{ token: string; admin: Admin }>('/admin/login', {
            method: 'POST',
            body: { username, password },
        }),
    me: () => authed<Admin>('/admin/me'),
    changePassword: (current_password: string, new_password: string) =>
        authed<void>('/admin/password', {
            method: 'POST',
            body: { current_password, new_password },
        }),
    stats: () => authed<Stats>('/admin/stats'),

    upload: (file: File) => {
        const form = new FormData()
        form.append('file', file)
        return authed<{ url: string }>('/admin/uploads', {
            method: 'POST',
            body: form,
        })
    },

    categories: () => authed<Category[]>('/admin/categories'),
    createCategory: (body: CategoryInput) =>
        authed<Category>('/admin/categories', { method: 'POST', body }),
    updateCategory: (id: number, body: CategoryInput) =>
        authed<Category>(`/admin/categories/${id}`, { method: 'PUT', body }),
    deleteCategory: (id: number) =>
        authed<void>(`/admin/categories/${id}`, { method: 'DELETE' }),

    products: (query: {
        q?: string
        category_id?: number | ''
        page?: number
        limit?: number
    }) => authed<Page<Product>>('/admin/products', { query }),
    product: (id: number) => authed<Product>(`/admin/products/${id}`),
    createProduct: (body: ProductInput) =>
        authed<Product>('/admin/products', { method: 'POST', body }),
    updateProduct: (id: number, body: ProductInput) =>
        authed<Product>(`/admin/products/${id}`, { method: 'PUT', body }),
    deleteProduct: (id: number) =>
        authed<void>(`/admin/products/${id}`, { method: 'DELETE' }),

    orders: (query: {
        status?: string
        q?: string
        page?: number
        limit?: number
    }) => authed<Page<Order>>('/admin/orders', { query }),
    order: (id: number) => authed<Order>(`/admin/orders/${id}`),
    updateOrder: (id: number, status: OrderStatus, admin_note: string) =>
        authed<Order>(`/admin/orders/${id}`, {
            method: 'PATCH',
            body: { status, admin_note },
        }),
    deleteOrder: (id: number) =>
        authed<void>(`/admin/orders/${id}`, { method: 'DELETE' }),

    messages: (query: { unread?: boolean; page?: number; limit?: number }) =>
        authed<Page<Message>>('/admin/messages', { query }),
    markMessage: (id: number, is_read: boolean) =>
        authed<void>(`/admin/messages/${id}`, {
            method: 'PATCH',
            body: { is_read },
        }),
    deleteMessage: (id: number) =>
        authed<void>(`/admin/messages/${id}`, { method: 'DELETE' }),

    chats: (query: {
        outcome?: '' | 'order' | 'lead'
        q?: string
        page?: number
        limit?: number
    }) => authed<Page<ChatSessionSummary>>('/admin/chats', { query }),
    chat: (id: string) => authed<ChatSessionDetail>(`/admin/chats/${id}`),
    deleteChat: (id: string) =>
        authed<void>(`/admin/chats/${id}`, { method: 'DELETE' }),
    missingProducts: (days = 0) =>
        authed<MissingProduct[]>('/admin/missing-products', { query: { days } }),
    deleteMissingProduct: (name: string) =>
        authed<void>('/admin/missing-products/delete', {
            method: 'POST',
            body: { name },
        }),

    settings: () => authed<Settings>('/admin/settings'),
    saveSettings: (body: Settings) =>
        authed<Settings>('/admin/settings', { method: 'PUT', body }),
}
