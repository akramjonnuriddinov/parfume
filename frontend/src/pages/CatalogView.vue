<!-- CatalogView.vue -->
<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { publicApi } from '../api/public'
import { SHOP_GENDERS, TIERS, type Brand, type Product } from '../api/types'
import PageIntro from '../components/PageIntro.vue'
import ProductCard from '../components/ProductCard.vue'
import ProductCardSkeleton from '../components/ProductCardSkeleton.vue'
import { useCategoriesStore } from '../stores/categories'
import { useLocalized } from '../utils/useLocalized'
import { usePerfume } from '../utils/usePerfume'

const PAGE_SIZE = 12

const { t } = useI18n()
const { tr } = useLocalized()
const perfume = usePerfume()
const route = useRoute()
const router = useRouter()
const categories = useCategoriesStore()

const products = ref<Product[]>([])
const brands = ref<Brand[]>([])
const total = ref(0)
const loading = ref(true)
const failed = ref(false)

const category = computed(() => (typeof route.query.category === 'string' ? route.query.category : ''))
const q = computed(() => (typeof route.query.q === 'string' ? route.query.q : ''))
const gender = computed(() => (typeof route.query.gender === 'string' ? route.query.gender : ''))
const brand = computed(() => (typeof route.query.brand === 'string' ? route.query.brand : ''))
const tier = computed(() => (typeof route.query.tier === 'string' ? route.query.tier : ''))
const sort = computed(() => (typeof route.query.sort === 'string' ? route.query.sort : ''))
const page = computed(() => Math.max(1, Number(route.query.page) || 1))

const activeCategory = computed(() => categories.items.find((c) => c.slug === category.value))
const title = computed(() => {
    if (activeCategory.value) return tr(activeCategory.value)
    if (brand.value) return brand.value
    if (gender.value === 'women' || gender.value === 'men') return t(`catalog.forGender.${gender.value}`)
    return t('catalog.title')
})

const sortOptions = computed(() => [
    { value: '', label: t('catalog.sort.default') },
    { value: 'newest', label: t('catalog.sort.newest') },
    { value: 'price_asc', label: t('catalog.sort.priceAsc') },
    { value: 'price_desc', label: t('catalog.sort.priceDesc') },
])

function updateQuery(patch: Record<string, string | number | undefined>): void {
    const query: Record<string, string> = {}
    const merged = {
        category: category.value,
        gender: gender.value,
        brand: brand.value,
        tier: tier.value,
        q: q.value,
        sort: sort.value,
        page: undefined,
        ...patch,
    }
    for (const [key, value] of Object.entries(merged)) {
        if (value !== undefined && value !== '' && !(key === 'page' && Number(value) === 1)) {
            query[key] = String(value)
        }
    }
    router.push({ path: '/catalog', query })
}

async function load(): Promise<void> {
    loading.value = true
    failed.value = false
    try {
        const result = await publicApi.products({
            category: category.value || undefined,
            gender: gender.value || undefined,
            brand: brand.value || undefined,
            tier: tier.value || undefined,
            q: q.value || undefined,
            sort: sort.value || undefined,
            page: page.value,
            limit: PAGE_SIZE,
        })
        products.value = result.items
        total.value = result.total
    } catch {
        failed.value = true
        products.value = []
        total.value = 0
    } finally {
        loading.value = false
    }
}

onMounted(async () => {
    categories.load()
    brands.value = await publicApi.brands().catch(() => [])
})
watch(() => route.fullPath, load, { immediate: true })
</script>

<template>
    <div>
        <PageIntro
            :title="title"
            :crumbs="activeCategory ? [{ label: t('catalog.title'), to: '/catalog' }, { label: title }] : undefined"
            :text="activeCategory ? tr(activeCategory, 'description') : undefined"
        />

        <div class="mx-auto max-w-[1240px] px-4 py-8 sm:px-6 lg:py-10">
            <!-- Kategoriyalar -->
            <div
                class="-mx-4 flex gap-2 overflow-x-auto px-4 pb-1 sm:mx-0 sm:flex-wrap sm:px-0"
                role="tablist"
                :aria-label="t('catalog.categories')"
            >
                <button
                    type="button"
                    role="tab"
                    :aria-selected="!category"
                    class="shrink-0 cursor-pointer rounded-lg border px-4 py-2 text-sm transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand"
                    :class="!category ? 'border-brand bg-brand text-white' : 'border-line text-ink hover:border-brand hover:text-brand'"
                    @click="updateQuery({ category: undefined })"
                >
                    {{ t('catalog.all') }}
                </button>
                <button
                    v-for="item in categories.items"
                    :key="item.id"
                    type="button"
                    role="tab"
                    :aria-selected="category === item.slug"
                    class="shrink-0 cursor-pointer rounded-lg border px-4 py-2 text-sm transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand"
                    :class="category === item.slug ? 'border-brand bg-brand text-white' : 'border-line text-ink hover:border-brand hover:text-brand'"
                    @click="updateQuery({ category: item.slug })"
                >
                    {{ tr(item) }}
                </button>
            </div>

            <!-- Filtr qatori -->
            <div class="mt-6 flex flex-col gap-3 border-b border-line pb-5 lg:flex-row lg:items-center lg:justify-between">
                <p class="text-sm text-muted lg:shrink-0">
                    <template v-if="q">
                        {{ t('catalog.searchResults', { q }) }} ·
                    </template>
                    {{ t('catalog.found', { n: total }) }}
                    <button
                        v-if="q || brand || gender || tier"
                        type="button"
                        class="ml-2 cursor-pointer text-brand underline-offset-4 hover:underline"
                        @click="updateQuery({ q: undefined, brand: undefined, gender: undefined, tier: undefined })"
                    >
                        {{ t('catalog.clearSearch') }}
                    </button>
                </p>

                <div class="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center lg:justify-end">
                <el-radio-group
                    class="shrink-0 flex-nowrap!"
                    :model-value="gender"
                    :aria-label="t('catalog.gender')"
                    @update:model-value="updateQuery({ gender: ($event as string) || undefined })"
                >
                    <el-radio-button value="">{{ t('catalog.all') }}</el-radio-button>
                    <el-radio-button
                        v-for="g in SHOP_GENDERS"
                        :key="g"
                        :value="g"
                    >
                        {{ perfume.gender(g) }}
                    </el-radio-button>
                </el-radio-group>
                <el-select
                    :model-value="tier || undefined"
                    clearable
                    class="w-full! sm:w-40!"
                    :placeholder="t('catalog.allTiers')"
                    :aria-label="t('catalog.tier')"
                    @update:model-value="updateQuery({ tier: $event || undefined })"
                >
                    <el-option
                        v-for="x in TIERS"
                        :key="x"
                        :value="x"
                        :label="perfume.tier(x)"
                    />
                </el-select>
                <el-select
                    :model-value="brand || undefined"
                    filterable
                    clearable
                    class="w-full! sm:w-52!"
                    :placeholder="t('catalog.allBrands')"
                    :aria-label="t('catalog.brand')"
                    @update:model-value="updateQuery({ brand: $event || undefined })"
                >
                    <el-option
                        v-for="b in brands"
                        :key="b.name"
                        :value="b.name"
                        :label="`${b.name} (${b.count})`"
                    />
                </el-select>
                <el-select
                    :model-value="sort"
                    :empty-values="[null, undefined]"
                    class="w-full! sm:w-60!"
                    :aria-label="t('catalog.sort.label')"
                    @update:model-value="updateQuery({ sort: $event })"
                >
                    <el-option
                        v-for="option in sortOptions"
                        :key="option.value"
                        :value="option.value"
                        :label="option.label"
                    />
                </el-select>
                </div>
            </div>

            <!-- Mahsulotlar -->
            <div
                v-if="loading"
                class="mt-8 grid grid-cols-2 gap-x-4 gap-y-10 sm:gap-x-6 lg:grid-cols-4"
            >
                <ProductCardSkeleton
                    v-for="n in 8"
                    :key="n"
                />
            </div>

            <el-result
                v-else-if="failed"
                icon="error"
                :title="t('common.loadError')"
            >
                <template #extra>
                    <el-button
                        type="primary"
                        @click="load"
                    >
                        {{ t('common.retry') }}
                    </el-button>
                </template>
            </el-result>

            <div
                v-else-if="products.length === 0"
                class="py-20 text-center"
            >
                <p class="text-lg font-medium text-ink">{{ t('catalog.emptyTitle') }}</p>
                <p class="mt-2 text-sm text-muted">{{ t('catalog.emptyText') }}</p>
                <el-button
                    class="mt-6! rounded-lg!"
                    @click="router.push('/catalog')"
                >
                    {{ t('catalog.showAll') }}
                </el-button>
            </div>

            <div
                v-else
                class="mt-8 grid grid-cols-2 gap-x-4 gap-y-10 sm:gap-x-6 lg:grid-cols-4"
            >
                <ProductCard
                    v-for="product in products"
                    :key="product.id"
                    :product="product"
                />
            </div>

            <div
                v-if="total > PAGE_SIZE"
                class="mt-12 flex justify-center"
            >
                <el-pagination
                    background
                    layout="prev, pager, next"
                    :pager-count="5"
                    :total="total"
                    :page-size="PAGE_SIZE"
                    :current-page="page"
                    @current-change="updateQuery({ page: $event })"
                />
            </div>
        </div>
    </div>
</template>
