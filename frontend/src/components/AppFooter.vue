<!-- AppFooter.vue -->
<script setup lang="ts">
import AppLogo from './AppLogo.vue'
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useCategoriesStore } from '../stores/categories'
import { DEFAULT_PHONE, useSettingsStore } from '../stores/settings'
import { telHref } from '../utils/format'
import { useLocalized } from '../utils/useLocalized'

const { t } = useI18n()
const { tr, setting } = useLocalized()
const categories = useCategoriesStore()
const settingsStore = useSettingsStore()

onMounted(() => categories.load())

const currentYear = new Date().getFullYear()

const s = computed(() => settingsStore.settings)
const phone = computed(() => s.value.phone || DEFAULT_PHONE)
const address = computed(() => setting(s.value, 'address'))

const informationLinks = computed(() => [
    { label: t('nav.about'), to: '/about' },
    { label: t('nav.delivery'), to: '/delivery' },
    { label: t('nav.contact'), to: '/contact' },
])
</script>

<template>
    <footer class="border-t border-line bg-surface">
        <div class="mx-auto max-w-[1240px] px-4 sm:px-6">
            <div
                class="grid grid-cols-2 gap-x-6 gap-y-10 py-12 lg:grid-cols-[1.4fr_1fr_1fr_1.2fr] lg:gap-12 lg:py-14"
            >
                <!-- Brand -->
                <div class="col-span-2 lg:col-span-1">
                    <RouterLink
                        to="/"
                        class="inline-flex items-center gap-2.5 rounded text-ink focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand"
                        :aria-label="t('header.homeLabel')"
                    >
                        <AppLogo height="h-14" />
                    </RouterLink>

                    <p class="mt-4 max-w-64 text-sm leading-6 text-muted">
                        {{ t('footer.about') }}
                    </p>
                </div>

                <!-- Katalog -->
                <nav aria-labelledby="footer-catalog-title">
                    <h2
                        id="footer-catalog-title"
                        class="text-sm font-medium text-ink"
                    >
                        {{ t('nav.catalog') }}
                    </h2>

                    <ul class="mt-4 space-y-3">
                        <li
                            v-for="category in categories.items"
                            :key="category.id"
                        >
                            <RouterLink
                                :to="{ path: '/catalog', query: { category: category.slug } }"
                                class="rounded text-sm leading-6 text-muted transition-colors hover:text-brand focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand"
                            >
                                {{ tr(category) }}
                            </RouterLink>
                        </li>
                    </ul>
                </nav>

                <!-- Ma’lumotlar -->
                <nav aria-labelledby="footer-info-title">
                    <h2
                        id="footer-info-title"
                        class="text-sm font-medium text-ink"
                    >
                        {{ t('footer.information') }}
                    </h2>

                    <ul class="mt-4 space-y-3">
                        <li
                            v-for="link in informationLinks"
                            :key="link.to"
                        >
                            <RouterLink
                                :to="link.to"
                                class="rounded text-sm leading-6 text-muted transition-colors hover:text-brand focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand"
                            >
                                {{ link.label }}
                            </RouterLink>
                        </li>
                    </ul>
                </nav>

                <!-- Aloqa -->
                <div class="col-span-2 lg:col-span-1">
                    <h2 class="text-sm font-medium text-ink">
                        {{ t('nav.contact') }}
                    </h2>

                    <address class="mt-4 flex flex-col items-start gap-3 not-italic">
                        <a
                            :href="telHref(phone)"
                            class="rounded text-base font-medium text-ink transition-colors hover:text-brand focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand"
                        >
                            {{ phone }}
                        </a>

                        <a
                            v-if="s.telegram_url"
                            :href="s.telegram_url"
                            target="_blank"
                            rel="noopener noreferrer"
                            class="rounded text-sm text-brand underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand"
                        >
                            {{ t('footer.telegram') }}
                        </a>

                        <p
                            v-if="address"
                            class="max-w-64 text-sm leading-6 text-muted"
                        >
                            {{ address }}
                        </p>
                    </address>
                </div>
            </div>

            <div
                class="flex flex-col gap-3 border-t border-line py-5 text-xs leading-5 text-muted sm:flex-row sm:items-center sm:justify-between"
            >
                <p>
                    © {{ currentYear }} Mariana Parfumes
                </p>

                <RouterLink
                    to="/privacy"
                    class="w-fit rounded underline-offset-4 hover:text-brand hover:underline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand"
                >
                    {{ t('privacy.title') }}
                </RouterLink>
            </div>
        </div>
    </footer>
</template>
