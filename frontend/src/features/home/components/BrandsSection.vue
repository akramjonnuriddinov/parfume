<!-- BrandsSection.vue — brands in stock, each opens its catalog -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Brand } from '../../../api/types'

defineProps<{ brands: Brand[] }>()
const emit = defineEmits<{ select: [brand: string] }>()

const { t } = useI18n()
</script>

<template>
    <section
        v-if="brands.length"
        class="bg-white"
        aria-labelledby="brands-title"
    >
        <div class="mx-auto max-w-[1240px] px-4 py-10 sm:px-6 lg:py-14">
            <h2
                id="brands-title"
                class="text-2xl font-medium tracking-tight text-ink sm:text-3xl"
            >
                {{ t('home.brands.title') }}
            </h2>
            <div class="mt-8 flex flex-wrap gap-3">
                <button
                    v-for="b in brands"
                    :key="b.name"
                    type="button"
                    class="cursor-pointer rounded-full border border-line px-5 py-2.5 text-lg font-medium text-ink transition-colors hover:border-brand hover:bg-brand-soft hover:text-brand"
                    @click="emit('select', b.name)"
                >
                    {{ b.name }}
                </button>
            </div>
        </div>
    </section>
</template>
