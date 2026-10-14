import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import DefaultLayout from '../layouts/DefaultLayout.vue'
import { i18n } from '../i18n'
import { useAuthStore } from '../stores/auth'

declare module 'vue-router' {
    interface RouteMeta {
        title?: string
        requiresAuth?: boolean
    }
}

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        component: DefaultLayout,
        children: [
            {
                name: 'home',
                path: '',
                component: () => import('../pages/HomeView.vue'),
            },
            {
                name: 'catalog',
                path: 'catalog',
                component: () => import('../pages/CatalogView.vue'),
                meta: { title: 'nav.catalog' },
            },
            {
                name: 'product',
                path: 'products/:id(\\d+)',
                component: () => import('../pages/ProductView.vue'),
            },
            {
                name: 'cart',
                path: 'cart',
                component: () => import('../pages/CartView.vue'),
                meta: { title: 'cart.title' },
            },
            {
                name: 'order-success',
                path: 'order/success/:id(\\d+)',
                component: () => import('../pages/OrderSuccessView.vue'),
                meta: { title: 'success.title' },
            },
            {
                name: 'about',
                path: 'about',
                component: () => import('../pages/AboutView.vue'),
                meta: { title: 'nav.about' },
            },
            {
                name: 'delivery',
                path: 'delivery',
                component: () => import('../pages/DeliveryView.vue'),
                meta: { title: 'nav.delivery' },
            },
            {
                name: 'contact',
                path: 'contact',
                component: () => import('../pages/ContactView.vue'),
                meta: { title: 'nav.contact' },
            },
            {
                name: 'privacy',
                path: 'privacy',
                component: () => import('../pages/PrivacyView.vue'),
                meta: { title: 'privacy.title' },
            },
        ],
    },
    {
        name: 'admin-login',
        path: '/admin/login',
        component: () => import('../admin/pages/LoginView.vue'),
        meta: { title: 'admin.login.title' },
    },
    {
        path: '/admin',
        component: () => import('../admin/layouts/AdminLayout.vue'),
        meta: { requiresAuth: true },
        children: [
            {
                name: 'admin-dashboard',
                path: '',
                component: () => import('../admin/pages/DashboardView.vue'),
                meta: { title: 'admin.nav.dashboard', requiresAuth: true },
            },
            {
                name: 'admin-orders',
                path: 'orders',
                component: () => import('../admin/pages/OrdersView.vue'),
                meta: { title: 'admin.nav.orders', requiresAuth: true },
            },
            {
                name: 'admin-order',
                path: 'orders/:id(\\d+)',
                component: () => import('../admin/pages/OrderDetailView.vue'),
                meta: { title: 'admin.nav.orders', requiresAuth: true },
            },
            {
                name: 'admin-products',
                path: 'products',
                component: () => import('../admin/pages/ProductsView.vue'),
                meta: { title: 'admin.nav.products', requiresAuth: true },
            },
            {
                name: 'admin-product-new',
                path: 'products/new',
                component: () => import('../admin/pages/ProductFormView.vue'),
                meta: { title: 'admin.products.new', requiresAuth: true },
            },
            {
                name: 'admin-product-edit',
                path: 'products/:id(\\d+)',
                component: () => import('../admin/pages/ProductFormView.vue'),
                meta: { title: 'admin.products.edit', requiresAuth: true },
            },
            {
                name: 'admin-categories',
                path: 'categories',
                component: () => import('../admin/pages/CategoriesView.vue'),
                meta: { title: 'admin.nav.categories', requiresAuth: true },
            },
            {
                name: 'admin-messages',
                path: 'messages',
                component: () => import('../admin/pages/MessagesView.vue'),
                meta: { title: 'admin.nav.messages', requiresAuth: true },
            },
            {
                name: 'admin-chats',
                path: 'chats',
                component: () => import('../admin/pages/ChatsView.vue'),
                meta: { title: 'admin.nav.chats', requiresAuth: true },
            },
            {
                name: 'admin-chat',
                path: 'chats/:id([0-9a-f]{32})',
                component: () => import('../admin/pages/ChatDetailView.vue'),
                meta: { title: 'admin.nav.chats', requiresAuth: true },
            },
            {
                name: 'admin-settings',
                path: 'settings',
                component: () => import('../admin/pages/SettingsView.vue'),
                meta: { title: 'admin.nav.settings', requiresAuth: true },
            },
        ],
    },
    {
        path: '/:pathMatch(.*)*',
        component: DefaultLayout,
        children: [
            {
                name: 'not-found',
                path: '',
                component: () => import('../pages/NotFoundView.vue'),
                meta: { title: 'notFound.title' },
            },
        ],
    },
]

export const router = createRouter({
    history: createWebHistory(),
    routes,
    scrollBehavior(to, from, saved) {
        if (saved) return saved
        if (to.path === from.path) return false
        return { top: 0 }
    },
})

router.beforeEach((to) => {
    const auth = useAuthStore()
    if (to.meta.requiresAuth && !auth.isAuthenticated) {
        return { name: 'admin-login', query: { redirect: to.fullPath } }
    }
    if (to.name === 'admin-login' && auth.isAuthenticated) {
        return { name: 'admin-dashboard' }
    }
})

// A stale/missing lazy chunk (new deploy, dev re-bundle) would leave a blank
// page — reload the target route once instead.
router.onError((error: Error, to) => {
    const isChunkError = /dynamically imported module|Importing a module script failed/i.test(error?.message ?? '')
    if (!isChunkError) return
    const key = `chunk-reload:${to.fullPath}`
    if (sessionStorage.getItem(key)) return
    sessionStorage.setItem(key, '1')
    window.location.assign(to.fullPath)
})

router.afterEach((to) => sessionStorage.removeItem(`chunk-reload:${to.fullPath}`))

export function applyDocumentTitle(): void {
    const route = router.currentRoute.value
    const t = i18n.global.t
    const base = 'Mariana Parfumes'
    document.title = route.meta.title ? `${t(route.meta.title)} — ${base}` : `${base} — ${t('meta.tagline')}`
}

router.afterEach(() => applyDocumentTitle())
