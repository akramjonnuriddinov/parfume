<!-- CategoriesSection.vue -->
<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useCategoriesStore } from '../../../stores/categories'
import { useLocalized } from '../../../utils/useLocalized'

const { t } = useI18n()
const { tr } = useLocalized()
const categories = useCategoriesStore()

onMounted(() => categories.load())

const emit = defineEmits<{
    select: [categorySlug: string]
    catalog: []
}>()
</script>

<template>
    <section
        class="bg-white"
        aria-labelledby="categories-title"
    >
        <div class="mx-auto max-w-[1240px] px-4 py-12 sm:px-6 lg:py-16">
            <div class="grid gap-8 lg:grid-cols-[1fr_2fr] lg:gap-16">
                <div>
                    <h2
                        id="categories-title"
                        class="text-2xl leading-tight font-medium tracking-tight text-ink sm:text-3xl"
                    >
                        {{ t('home.categories.title') }}
                    </h2>

                    <button
                        type="button"
                        class="mt-5 inline-flex cursor-pointer items-center gap-1 text-sm text-brand underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand"
                        @click="emit('catalog')"
                    >
                        {{ t('home.categories.all') }}
                    </button>
                </div>

                <div class="grid gap-x-10 sm:grid-cols-2">
                    <template v-if="!categories.loaded">
                        <div
                            v-for="n in 4"
                            :key="n"
                            class="border-t border-slate-200 py-6"
                            aria-hidden="true"
                        >
                            <div class="h-5 w-1/2 animate-pulse rounded bg-surface" />
                            <div class="mt-3 h-4 w-3/4 animate-pulse rounded bg-surface" />
                        </div>
                    </template>

                    <button
                        v-for="category in categories.items"
                        :key="category.id"
                        type="button"
                        class="group flex cursor-pointer items-center justify-between gap-5 border-t border-slate-200 py-6 text-left focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand"
                        @click="emit('select', category.slug)"
                    >
                        <div>
                            <h3
                                class="text-lg font-medium text-ink transition-colors group-hover:text-brand"
                            >
                                {{ tr(category) }}
                            </h3>

                            <p class="mt-2 text-sm leading-6 text-slate-500">
                                {{ tr(category, 'description') }}
                            </p>
                        </div>

                        <svg
                            class="size-5 shrink-0 text-slate-400 transition-colors group-hover:text-brand"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="1.5"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            aria-hidden="true"
                        >
                            <path d="m9 5 7 7-7 7" />
                        </svg>
                    </button>
                </div>
            </div>
        </div>
    </section>
</template>
