<!-- AdminLayout.vue -->
<script setup lang="ts">
import AppLogo from '../../components/AppLogo.vue'
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import {
    ChatDotRound,
    Collection,
    Expand,
    Goods,
    MagicStick,
    Odometer,
    Setting,
    Tickets,
} from '@element-plus/icons-vue'
import LanguageSwitcher from '../../components/LanguageSwitcher.vue'
import { useAdminCountersStore } from '../../stores/adminCounters'
import { useAuthStore } from '../../stores/auth'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const counters = useAdminCountersStore()

const drawerOpen = ref(false)

const items = computed(() => [
    { to: '/admin', label: t('admin.nav.dashboard'), icon: Odometer, badge: 0 },
    { to: '/admin/orders', label: t('admin.nav.orders'), icon: Tickets, badge: counters.newOrders },
    { to: '/admin/products', label: t('admin.nav.products'), icon: Goods, badge: 0 },
    { to: '/admin/categories', label: t('admin.nav.categories'), icon: Collection, badge: 0 },
    { to: '/admin/messages', label: t('admin.nav.messages'), icon: ChatDotRound, badge: counters.unreadMessages },
    { to: '/admin/chats', label: t('admin.nav.chats'), icon: MagicStick, badge: 0 },
    { to: '/admin/settings', label: t('admin.nav.settings'), icon: Setting, badge: 0 },
])

const activePath = computed(() => {
    const match = [...items.value]
        .sort((a, b) => b.to.length - a.to.length)
        .find((item) => route.path === item.to || route.path.startsWith(`${item.to}/`))
    return match?.to ?? '/admin'
})

const pageTitle = computed(() => (route.meta.title ? t(route.meta.title) : ''))

onMounted(() => {
    auth.fetchMe().catch(() => undefined)
})

watch(
    () => route.fullPath,
    () => {
        drawerOpen.value = false
        counters.refresh()
    },
    { immediate: true }
)

function onCommand(command: string | number | object): void {
    if (command === 'logout') {
        auth.logout()
        router.replace('/admin/login')
    } else if (command === 'settings') {
        router.push('/admin/settings')
    }
}
</script>

<template>
    <div class="min-h-screen bg-surface">
        <!-- Sidebar (desktop) -->
        <aside
            class="fixed inset-y-0 left-0 z-30 hidden w-64 flex-col border-r border-line bg-white lg:flex"
        >
            <RouterLink
                to="/admin"
                class="flex h-[72px] shrink-0 items-center gap-2.5 border-b border-line px-6 text-ink"
            >
                <AppLogo height="h-10" />
                <span class="ml-auto rounded-md bg-surface px-1.5 py-0.5 text-[10px] font-medium tracking-wide text-muted uppercase">
                    Admin
                </span>
            </RouterLink>
            <nav
                :aria-label="t('admin.nav.label')"
                class="flex-1 space-y-1 overflow-y-auto p-3"
            >
                <RouterLink
                    v-for="item in items"
                    :key="item.to"
                    :to="item.to"
                    class="flex h-11 items-center gap-3 rounded-lg px-3 text-sm transition-colors"
                    :class="activePath === item.to ? 'bg-brand-soft font-medium text-brand' : 'text-muted hover:bg-surface hover:text-ink'"
                >
                    <el-icon :size="18"><component :is="item.icon" /></el-icon>
                    <span class="flex-1">{{ item.label }}</span>
                    <span
                        v-if="item.badge > 0"
                        class="flex h-5 min-w-5 items-center justify-center rounded-full bg-brand px-1.5 text-[11px] font-medium text-white"
                    >
                        {{ item.badge > 99 ? '99+' : item.badge }}
                    </span>
                </RouterLink>
            </nav>
            <div class="border-t border-line p-3">
                <a
                    href="/"
                    target="_blank"
                    class="flex h-10 items-center gap-2 rounded-lg px-3 text-sm text-muted hover:bg-surface hover:text-brand"
                >
                    <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                        <path d="M14 4h6v6M20 4l-9 9M18 14v5a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V7a1 1 0 0 1 1-1h5" />
                    </svg>
                    {{ t('admin.viewSite') }}
                </a>
            </div>
        </aside>

        <!-- Sidebar (mobile) -->
        <el-drawer
            v-model="drawerOpen"
            direction="ltr"
            size="272px"
            :with-header="false"
        >
            <nav class="space-y-1">
                <RouterLink
                    v-for="item in items"
                    :key="item.to"
                    :to="item.to"
                    class="flex h-11 items-center gap-3 rounded-lg px-3 text-sm"
                    :class="activePath === item.to ? 'bg-brand-soft font-medium text-brand' : 'text-muted'"
                >
                    <el-icon :size="18"><component :is="item.icon" /></el-icon>
                    <span class="flex-1">{{ item.label }}</span>
                    <span
                        v-if="item.badge > 0"
                        class="flex h-5 min-w-5 items-center justify-center rounded-full bg-brand px-1.5 text-[11px] font-medium text-white"
                    >
                        {{ item.badge }}
                    </span>
                </RouterLink>
                <a
                    href="/"
                    target="_blank"
                    class="flex h-11 items-center gap-3 rounded-lg px-3 text-sm text-muted"
                >
                    {{ t('admin.viewSite') }}
                </a>
            </nav>
        </el-drawer>

        <div class="lg:pl-64">
            <!-- Topbar -->
            <header
                class="sticky top-0 z-20 flex h-[72px] items-center gap-3 border-b border-line bg-white px-4 sm:px-6"
            >
                <button
                    type="button"
                    :aria-label="t('header.openMenu')"
                    class="flex size-10 cursor-pointer items-center justify-center rounded-lg border border-line text-ink lg:hidden"
                    @click="drawerOpen = true"
                >
                    <el-icon :size="18"><Expand /></el-icon>
                </button>

                <h1 class="truncate text-lg font-medium text-ink">{{ pageTitle }}</h1>

                <div class="ml-auto flex items-center gap-2 sm:gap-3">
                    <LanguageSwitcher />
                    <el-dropdown
                        trigger="click"
                        @command="onCommand"
                    >
                        <button
                            type="button"
                            class="flex h-11 cursor-pointer items-center gap-2 rounded-lg border border-line px-2 pr-3 text-sm text-ink hover:border-brand"
                        >
                            <span class="flex size-7 items-center justify-center rounded-md bg-brand text-xs font-semibold text-white uppercase">
                                {{ (auth.admin?.username ?? 'A').slice(0, 1) }}
                            </span>
                            <span class="hidden sm:inline">{{ auth.admin?.username ?? 'admin' }}</span>
                        </button>
                        <template #dropdown>
                            <el-dropdown-menu>
                                <el-dropdown-item command="settings">{{ t('admin.nav.settings') }}</el-dropdown-item>
                                <el-dropdown-item
                                    command="logout"
                                    divided
                                >
                                    {{ t('admin.logout') }}
                                </el-dropdown-item>
                            </el-dropdown-menu>
                        </template>
                    </el-dropdown>
                </div>
            </header>

            <main class="mx-auto max-w-[1400px] p-4 sm:p-6 lg:p-8">
                <RouterView />
            </main>
        </div>
    </div>
</template>
