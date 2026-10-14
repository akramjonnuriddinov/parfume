<!-- AppHeader.vue -->
<script setup lang="ts">
import AppLogo from './AppLogo.vue'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useCartStore } from '../stores/cart'
import { DEFAULT_PHONE, useSettingsStore } from '../stores/settings'
import { telHref } from '../utils/format'
import LanguageSwitcher from './LanguageSwitcher.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const cart = useCartStore()
const settingsStore = useSettingsStore()

const phone = computed(() => settingsStore.settings.phone || DEFAULT_PHONE)

const searchQuery = ref<string>('')
const menuOpen = ref<boolean>(false)

watch(
    () => route.query.q,
    (q) => {
        searchQuery.value = typeof q === 'string' ? q : ''
    },
    { immediate: true }
)

const links = computed(() => [
    { label: t('nav.catalog'), to: '/catalog' },
    { label: t('nav.women'), to: '/catalog?gender=women' },
    { label: t('nav.men'), to: '/catalog?gender=men' },
    { label: t('nav.unisex'), to: '/catalog?gender=unisex' },
    { label: t('nav.about'), to: '/about' },
    { label: t('nav.delivery'), to: '/delivery' },
    { label: t('nav.contact'), to: '/contact' },
])

function submitSearch(): void {
    const query = searchQuery.value.trim()
    router.push({ path: '/catalog', query: query ? { q: query } : {} })
}
</script>

<template>
    <header class="sticky top-0 z-40 border-b border-line bg-white">
        <div
            class="mx-auto flex max-w-[1240px] flex-wrap items-center gap-3 px-4 py-4 sm:gap-4 sm:px-6 lg:flex-nowrap lg:gap-8 lg:py-5"
        >
            <!-- Logo -->
            <RouterLink
                to="/"
                class="mr-auto inline-flex shrink-0 items-center gap-2.5 rounded-md text-ink focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand lg:mr-0"
                :aria-label="t('header.homeLabel')"
            >
                <AppLogo height="h-11 sm:h-14" />
            </RouterLink>

            <!-- Qidiruv -->
            <form
                role="search"
                class="order-last w-full lg:order-none lg:mr-auto lg:max-w-[480px] lg:flex-1"
                @submit.prevent="submitSearch"
            >
                <div
                    class="flex h-11 items-center gap-2 rounded-lg border border-line bg-surface pr-1 pl-3.5 transition-colors focus-within:border-brand"
                >
                    <svg
                        class="size-5 shrink-0 text-muted"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="1.7"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        aria-hidden="true"
                    >
                        <circle cx="10.5" cy="10.5" r="6.5" />
                        <path d="m16 16 4 4" />
                    </svg>

                    <input
                        v-model="searchQuery"
                        type="search"
                        name="q"
                        :placeholder="t('header.searchPlaceholder')"
                        :aria-label="t('header.searchPlaceholder')"
                        class="min-w-0 flex-1 bg-transparent text-sm text-ink outline-none placeholder:text-muted"
                    />

                    <button
                        type="submit"
                        class="h-9 shrink-0 cursor-pointer rounded-md bg-brand px-4 text-sm font-medium text-white transition-colors hover:bg-brand-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand"
                    >
                        {{ t('header.search') }}
                    </button>
                </div>
            </form>

            <!-- Telefon -->
            <a
                :href="telHref(phone)"
                class="hidden shrink-0 rounded text-right focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand lg:block"
            >
                <span class="block text-xs text-muted">
                    {{ t('header.consultation') }}
                </span>

                <span
                    class="mt-1 block text-sm font-medium text-ink transition-colors hover:text-brand"
                >
                    {{ phone }}
                </span>
            </a>

            <!-- Til -->
            <LanguageSwitcher class="hidden! sm:flex!" />

            <!-- Savat -->
            <button
                type="button"
                :aria-label="t('header.openCart', { count: cart.count })"
                class="relative flex size-11 shrink-0 cursor-pointer items-center justify-center rounded-lg border border-line text-ink transition-colors hover:border-brand hover:text-brand focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand"
                @click="cart.drawerOpen = true"
            >
                <svg
                    class="size-5"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.7"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    aria-hidden="true"
                >
                    <path d="M3 3h2l2.5 12h11L21 7H6" />
                    <circle cx="9" cy="20" r="1" />
                    <circle cx="18" cy="20" r="1" />
                </svg>

                <span
                    v-if="cart.count > 0"
                    class="absolute -top-2 -right-2 flex h-5 min-w-5 items-center justify-center rounded-full bg-brand px-1 text-[10px] font-medium text-white ring-2 ring-white"
                    aria-hidden="true"
                >
                    {{ cart.count > 99 ? '99+' : cart.count }}
                </span>
            </button>

            <!-- Mobil menyu -->
            <button
                type="button"
                :aria-label="t('header.openMenu')"
                aria-haspopup="dialog"
                :aria-expanded="menuOpen"
                class="flex size-11 shrink-0 cursor-pointer items-center justify-center rounded-lg border border-line text-ink transition-colors hover:text-brand focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand lg:hidden"
                @click="menuOpen = true"
            >
                <svg
                    class="size-5"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.7"
                    stroke-linecap="round"
                    aria-hidden="true"
                >
                    <path d="M4 6h16M4 12h16M4 18h16" />
                </svg>
            </button>
        </div>

        <!-- Desktop navigatsiya -->
        <div class="hidden border-t border-line lg:block">
            <nav
                :aria-label="t('header.mainMenu')"
                class="mx-auto flex max-w-[1240px] items-center gap-8 px-6"
            >
                <RouterLink
                    v-for="link in links"
                    :key="link.to"
                    :to="link.to"
                    class="py-3.5 text-sm text-muted transition-colors hover:text-brand focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand"
                    :class="route.fullPath === link.to ? 'text-brand!' : ''"
                >
                    {{ link.label }}
                </RouterLink>
            </nav>
        </div>

        <!-- Mobil navigatsiya -->
        <el-drawer
            v-model="menuOpen"
            :title="t('header.menu')"
            direction="rtl"
            size="min(340px, 100vw)"
        >
            <div class="flex h-full flex-col">
                <nav
                    :aria-label="t('header.mobileMenu')"
                    class="flex flex-col"
                >
                    <RouterLink
                        v-for="link in links"
                        :key="link.to"
                        :to="link.to"
                        class="border-b border-line py-4 text-base text-ink transition-colors hover:text-brand focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand"
                        :class="route.fullPath === link.to ? 'text-brand!' : ''"
                        @click="menuOpen = false"
                    >
                        {{ link.label }}
                    </RouterLink>
                </nav>

                <div class="mt-6">
                    <LanguageSwitcher />
                </div>

                <div class="mt-auto pt-10">
                    <p class="text-xs text-muted">
                        {{ t('header.consultation') }}
                    </p>

                    <a
                        :href="telHref(phone)"
                        class="mt-2 inline-block text-base font-medium text-brand"
                    >
                        {{ phone }}
                    </a>
                </div>
            </div>
        </el-drawer>
    </header>
</template>
