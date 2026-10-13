<!-- ProductsSection.vue -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Product } from '../../../api/types'
import ProductCard from '../../../components/ProductCard.vue'
import ProductCardSkeleton from '../../../components/ProductCardSkeleton.vue'

interface Props {
    products: Product[]
    loading?: boolean
}

withDefaults(defineProps<Props>(), { loading: false })

const emit = defineEmits<{
    catalog: []
}>()

const { t } = useI18n()
</script>

<template>
    <section
        class="bg-white"
        aria-labelledby="products-title"
    >
        <div class="mx-auto max-w-[1240px] px-4 py-12 sm:px-6 lg:py-16">
            <div
                class="mb-8 flex items-end justify-between gap-5 sm:mb-10"
            >
                <h2
                    id="products-title"
                    class="text-2xl font-medium tracking-tight text-ink sm:text-3xl"
                >
                    {{ t('home.products.title') }}
                </h2>

                <button
                    type="button"
                    class="inline-flex shrink-0 cursor-pointer items-center text-sm text-brand underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand"
                    @click="emit('catalog')"
                >
                    {{ t('home.products.all') }}
                </button>
            </div>

            <div
                class="grid grid-cols-2 gap-x-4 gap-y-10 sm:gap-x-6 lg:grid-cols-4"
            >
                <template v-if="loading">
                    <ProductCardSkeleton
                        v-for="n in 8"
                        :key="n"
                    />
                </template>
                <ProductCard
                    v-for="product in products"
                    v-else
                    :key="product.id"
                    :product="product"
                />
            </div>
        </div>
    </section>
</template>
