<!-- PageIntro.vue -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps<{
    title?: string
    text?: string
    crumbs?: { label: string; to?: string }[]
}>()

const { t } = useI18n()

type Crumb = { label: string; to?: string }
</script>

<template>
    <div class="mx-auto max-w-[1240px] px-4 pt-6 sm:px-6 lg:pt-8">
        <nav
            :aria-label="t('common.breadcrumbs')"
            class="flex flex-wrap items-center gap-1.5 text-xs text-muted"
        >
            <RouterLink
                to="/"
                class="hover:text-brand"
            >
                {{ t('nav.home') }}
            </RouterLink>
            <template
                v-for="crumb in crumbs ?? ([{ label: title ?? '' }] as Crumb[])"
                :key="crumb.label"
            >
                <span aria-hidden="true">/</span>
                <RouterLink
                    v-if="crumb.to"
                    :to="crumb.to"
                    class="hover:text-brand"
                >
                    {{ crumb.label }}
                </RouterLink>
                <span
                    v-else
                    class="text-ink"
                >
                    {{ crumb.label }}
                </span>
            </template>
        </nav>

        <h1
            v-if="title"
            class="mt-4 text-4xl leading-tight font-medium tracking-tight text-ink sm:text-5xl"
        >
            {{ title }}
        </h1>
        <p
            v-if="text"
            class="mt-3 max-w-2xl text-base leading-7 text-muted"
        >
            {{ text }}
        </p>
        <slot />
    </div>
</template>
