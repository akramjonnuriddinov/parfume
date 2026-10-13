<!-- OrderSuccessView.vue -->
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { DEFAULT_PHONE, useSettingsStore } from '../stores/settings'
import { orderNumber, telHref } from '../utils/format'

const { t } = useI18n()
const route = useRoute()
const settings = useSettingsStore()

const number = computed(() => orderNumber(Number(route.params.id)))
const phone = computed(() => settings.settings.phone || DEFAULT_PHONE)
</script>

<template>
    <section class="mx-auto max-w-[1240px] px-4 py-12 sm:px-6 lg:py-20">
        <div class="mx-auto max-w-xl rounded-2xl bg-hero px-6 py-12 text-center sm:px-12">
            <span class="mx-auto flex size-16 items-center justify-center rounded-full bg-white text-brand">
                <svg class="size-8" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d="m5 12 5 5L20 7" />
                </svg>
            </span>
            <h1 class="mt-6 text-4xl font-medium tracking-tight text-ink">{{ t('success.title') }}</h1>
            <p class="mt-3 text-base text-muted">
                {{ t('success.number') }}
                <span class="font-semibold text-ink">{{ number }}</span>
            </p>
            <p class="mx-auto mt-4 max-w-sm text-sm leading-6 text-slate-600">{{ t('success.text') }}</p>

            <div class="mt-8 flex flex-col justify-center gap-3 sm:flex-row">
                <RouterLink
                    to="/catalog"
                    class="inline-flex h-12 items-center justify-center rounded-lg bg-brand px-6 text-sm font-medium text-white hover:bg-brand-hover"
                >
                    {{ t('success.continue') }}
                </RouterLink>
                <a
                    :href="telHref(phone)"
                    class="inline-flex h-12 items-center justify-center rounded-lg border border-line bg-white px-6 text-sm font-medium text-ink hover:border-brand hover:text-brand"
                >
                    {{ phone }}
                </a>
            </div>
        </div>
    </section>
</template>
