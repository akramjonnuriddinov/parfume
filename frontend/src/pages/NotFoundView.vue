<!-- NotFoundView.vue -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
    <section class="mx-auto max-w-[1240px] px-4 py-20 text-center sm:px-6 lg:py-28">
        <p class="text-sm font-medium text-brand">404</p>
        <h1 class="mt-3 text-4xl font-medium tracking-tight text-ink sm:text-4xl">{{ t('notFound.title') }}</h1>
        <p class="mt-3 text-base text-muted">{{ t('notFound.text') }}</p>
        <RouterLink
            to="/"
            class="mt-8 inline-flex h-12 items-center rounded-lg bg-brand px-6 text-sm font-medium text-white hover:bg-brand-hover"
        >
            {{ t('notFound.home') }}
        </RouterLink>
    </section>
</template>
