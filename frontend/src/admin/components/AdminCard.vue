<!-- AdminCard.vue -->
<script setup lang="ts">
withDefaults(defineProps<{ title?: string; padded?: boolean }>(), { padded: true })
</script>

<template>
    <section class="rounded-xl border border-line bg-white">
        <header
            v-if="title || $slots.actions"
            class="flex flex-wrap items-center justify-between gap-3 border-b border-line px-5 py-4"
        >
            <h2 class="text-base font-medium text-ink">{{ title }}</h2>
            <slot name="actions" />
        </header>
        <div :class="padded ? 'p-5' : ''">
            <slot />
        </div>
    </section>
</template>
