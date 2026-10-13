<!-- GenderSection.vue — "for her / for him / unisex" entry tiles -->
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Gender } from '../../../api/types'

const { t } = useI18n()

const emit = defineEmits<{ select: [gender: Gender] }>()

const tiles = computed(() => [
    { gender: 'women' as const, image: '/images/perfumes/rose.svg', bg: 'bg-[#f8ecea]' },
    { gender: 'men' as const, image: '/images/perfumes/blue.svg', bg: 'bg-[#e9eef3]' },
    { gender: 'unisex' as const, image: '/images/perfumes/amber.svg', bg: 'bg-[#f5ede1]' },
])
</script>

<template>
    <section
        class="bg-white"
        aria-labelledby="gender-title"
    >
        <div class="mx-auto max-w-[1240px] px-4 py-10 sm:px-6 lg:py-14">
            <h2
                id="gender-title"
                class="text-2xl font-medium tracking-tight text-ink sm:text-3xl"
            >
                {{ t('home.gender.title') }}
            </h2>
            <div class="mt-8 grid gap-4 sm:grid-cols-3">
                <button
                    v-for="tile in tiles"
                    :key="tile.gender"
                    type="button"
                    class="group relative flex h-48 cursor-pointer items-end overflow-hidden rounded-2xl p-6 text-left focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand sm:h-56"
                    :class="tile.bg"
                    @click="emit('select', tile.gender)"
                >
                    <img
                        :src="tile.image"
                        alt=""
                        class="absolute -right-4 -bottom-6 w-36 transition-transform duration-300 group-hover:-translate-y-1 group-hover:scale-105 sm:w-40"
                        aria-hidden="true"
                    >
                    <span class="relative">
                        <span class="block text-2xl font-medium text-ink">{{ t(`home.gender.${tile.gender}`) }}</span>
                        <span class="mt-1 inline-flex items-center gap-1 text-sm text-brand">
                            {{ t('home.gender.cta') }}
                            <svg class="size-4 transition-transform group-hover:translate-x-0.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                <path d="M5 12h14m-6-6 6 6-6 6" />
                            </svg>
                        </span>
                    </span>
                </button>
            </div>
        </div>
    </section>
</template>
