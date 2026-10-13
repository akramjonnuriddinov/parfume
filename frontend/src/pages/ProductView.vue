<!-- ProductView.vue -->
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ApiError } from '../api/http'
import { publicApi } from '../api/public'
import type { Product, Variant } from '../api/types'
import PageIntro from '../components/PageIntro.vue'
import ProductCard from '../components/ProductCard.vue'
import ProductImage from '../components/ProductImage.vue'
import QuantityStepper from '../components/QuantityStepper.vue'
import { useCartStore } from '../stores/cart'
import { useInquiryStore } from '../stores/inquiry'
import { useLocalized } from '../utils/useLocalized'
import { usePerfume } from '../utils/usePerfume'

const { t, locale } = useI18n()
const { tr, price } = useLocalized()
const perfume = usePerfume()
const route = useRoute()
const router = useRouter()
const cart = useCartStore()
const inquiry = useInquiryStore()

const product = ref<Product | null>(null)
const related = ref<Product[]>([])
const loading = ref(true)
const notFound = ref(false)
const quantity = ref(1)
const selected = ref<Variant | undefined>()

const name = computed(() => tr(product.value))
const inCart = computed(() => (selected.value ? cart.find(selected.value.id) : undefined))

const pyramid = computed(() => {
    const p = product.value
    if (!p) return []
    return [
        { key: 'top', label: t('product.notesTop'), hint: t('product.notesTopHint'), notes: perfume.notes(p.notes_top) },
        { key: 'heart', label: t('product.notesHeart'), hint: t('product.notesHeartHint'), notes: perfume.notes(p.notes_heart) },
        { key: 'base', label: t('product.notesBase'), hint: t('product.notesBaseHint'), notes: perfume.notes(p.notes_base) },
    ].filter((row) => row.notes.length)
})

async function load(id: string): Promise<void> {
    loading.value = true
    notFound.value = false
    quantity.value = 1
    try {
        product.value = await publicApi.product(id)
        selected.value = cart.defaultVariant(product.value) ?? product.value.variants[0]
        const slug = product.value.category?.slug
        related.value = slug
            ? (await publicApi.products({ category: slug, limit: 5 })).items
                  .filter((p) => p.id !== product.value?.id)
                  .slice(0, 4)
            : []
    } catch (err) {
        product.value = null
        notFound.value = err instanceof ApiError && err.status === 404
    } finally {
        loading.value = false
    }
}

watch(() => route.params.id as string, (id) => id && load(id), { immediate: true })

watch([name, locale], () => {
    if (name.value && product.value) {
        document.title = `${product.value.brand} ${name.value} — Parfume Shop`
    }
})

function select(v: Variant): void {
    selected.value = v
    quantity.value = 1
}

function addToCart(): void {
    if (!product.value || !selected.value) return
    cart.add(product.value, selected.value, quantity.value)
    ElMessage.success({
        message: t('cart.added', { name: `${name.value}, ${perfume.volume(selected.value)}` }),
        grouping: true,
    })
}

function buyNow(): void {
    if (!product.value || !selected.value) return
    if (!inCart.value) cart.add(product.value, selected.value, quantity.value)
    router.push('/cart')
}
</script>

<template>
    <div>
        <div
            v-if="loading"
            class="mx-auto grid max-w-[1240px] gap-10 px-4 py-10 sm:px-6 lg:grid-cols-2"
            aria-busy="true"
        >
            <div class="aspect-square animate-pulse rounded-2xl bg-surface" />
            <div class="space-y-4">
                <div class="h-4 w-1/4 animate-pulse rounded bg-surface" />
                <div class="h-10 w-3/4 animate-pulse rounded bg-surface" />
                <div class="h-8 w-1/3 animate-pulse rounded bg-surface" />
                <div class="h-24 animate-pulse rounded bg-surface" />
            </div>
        </div>

        <el-result
            v-else-if="!product"
            :icon="notFound ? 'warning' : 'error'"
            :title="notFound ? t('product.notFound') : t('common.loadError')"
        >
            <template #extra>
                <el-button
                    type="primary"
                    @click="router.push('/catalog')"
                >
                    {{ t('cart.goCatalog') }}
                </el-button>
            </template>
        </el-result>

        <template v-else>
            <PageIntro
                :crumbs="[
                    { label: t('catalog.title'), to: '/catalog' },
                    ...(product.category
                        ? [{ label: tr(product.category), to: `/catalog?category=${product.category.slug}` }]
                        : []),
                    { label: name },
                ]"
            />

            <section class="mx-auto grid max-w-[1240px] gap-8 px-4 py-8 sm:px-6 lg:grid-cols-2 lg:gap-14 lg:py-10">
                <div class="relative flex aspect-square items-center justify-center rounded-2xl bg-hero p-10 sm:p-16 lg:sticky lg:top-40 lg:self-start">
                    <ProductImage
                        :src="product.image_url"
                        :alt="`${product.brand} ${name}`"
                        eager
                    />
                    <span class="absolute top-4 left-4 rounded-md bg-white px-2.5 py-1 text-xs font-medium text-muted">
                        {{ perfume.gender(product.gender) }}
                    </span>
                </div>

                <div class="flex flex-col">
                    <RouterLink
                        v-if="product.brand"
                        :to="{ path: '/catalog', query: { brand: product.brand } }"
                        class="self-start text-sm font-medium tracking-[0.14em] text-gold uppercase hover:underline"
                    >
                        {{ product.brand }}
                    </RouterLink>
                    <h1 class="mt-2 text-4xl leading-tight font-medium tracking-tight text-ink sm:text-5xl">
                        {{ name }}
                    </h1>
                    <p class="mt-3 text-sm text-muted">
                        {{ perfume.concentration(product.concentration) }}
                        <template v-if="product.category"> · {{ tr(product.category) }}</template>
                    </p>

                    <div class="mt-8 rounded-2xl border border-line p-5 sm:p-6">
                        <p class="text-sm font-medium text-ink">{{ t('product.chooseVolume') }}</p>
                        <div class="mt-3 grid grid-cols-2 gap-2 sm:grid-cols-3">
                            <button
                                v-for="v in product.variants"
                                :key="v.id"
                                type="button"
                                :disabled="!v.in_stock"
                                class="relative flex cursor-pointer flex-col items-start rounded-xl border-2 px-3 py-2.5 text-left transition-colors disabled:cursor-not-allowed disabled:opacity-50"
                                :class="selected?.id === v.id ? 'border-brand bg-brand-soft' : 'border-line hover:border-brand/50'"
                                :aria-pressed="selected?.id === v.id"
                                @click="select(v)"
                            >
                                <span class="text-base font-semibold text-ink">{{ v.volume_ml }} ml</span>
                                <span class="text-sm text-ink">{{ price(v.price) }}</span>
                                <span
                                    class="mt-1 text-[11px]"
                                    :class="v.kind === 'decant' ? 'font-medium text-gold' : 'text-muted'"
                                >
                                    {{ !v.in_stock ? t('product.outOfStock') : v.kind === 'decant' ? t('perfume.decant') : t('perfume.bottle') }}
                                </span>
                            </button>
                        </div>
                        <p
                            v-if="selected?.kind === 'decant'"
                            class="mt-3 rounded-lg bg-brand-soft px-3 py-2 text-xs leading-5 text-brand"
                        >
                            {{ t('product.decantHint') }}
                        </p>

                        <p class="mt-5 text-3xl font-semibold tracking-tight text-ink">
                            {{ selected ? price(selected.price * (inCart ? inCart.quantity : quantity)) : '' }}
                        </p>

                        <template v-if="selected?.in_stock">
                            <div
                                v-if="inCart"
                                class="mt-5 flex flex-col gap-3 sm:flex-row"
                            >
                                <QuantityStepper
                                    class="sm:w-40!"
                                    :model-value="inCart.quantity"
                                    @update:model-value="cart.setQuantity(selected.id, $event)"
                                    @remove="cart.remove(selected.id)"
                                />
                                <el-button
                                    type="primary"
                                    class="m-0! h-11! flex-1 rounded-lg! text-sm! font-medium!"
                                    @click="router.push('/cart')"
                                >
                                    {{ t('product.goToCart') }}
                                </el-button>
                            </div>
                            <div
                                v-else
                                class="mt-5 flex flex-col gap-3 sm:flex-row"
                            >
                                <QuantityStepper
                                    v-model="quantity"
                                    class="sm:w-40!"
                                    @remove="quantity = 1"
                                />
                                <el-button
                                    type="primary"
                                    class="m-0! h-11! flex-1 rounded-lg! text-sm! font-medium!"
                                    @click="addToCart"
                                >
                                    {{ t('product.addToCart') }}
                                </el-button>
                                <el-button
                                    plain
                                    class="m-0! h-11! flex-1 rounded-lg! border-brand! text-sm! font-medium! text-brand!"
                                    @click="buyNow"
                                >
                                    {{ t('product.buyNow') }}
                                </el-button>
                            </div>
                        </template>

                        <el-button
                            v-else
                            type="primary"
                            class="mt-5! h-11! w-full! rounded-lg! text-sm! font-medium!"
                            @click="inquiry.open(product)"
                        >
                            {{ t('product.notifyMe') }}
                        </el-button>

                        <button
                            type="button"
                            class="mt-4 cursor-pointer text-sm text-brand underline-offset-4 hover:underline"
                            @click="inquiry.open(product)"
                        >
                            {{ t('product.askQuestion') }}
                        </button>
                    </div>

                    <ul class="mt-6 grid gap-3 text-sm text-muted sm:grid-cols-2">
                        <li class="flex items-center gap-2">
                            <svg class="size-5 shrink-0 text-brand" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                <path d="M12 3 5 6v5c0 4.5 3 8.5 7 10 4-1.5 7-5.5 7-10V6z" />
                                <path d="m9 12 2 2 4-4" />
                            </svg>
                            {{ t('product.perkOriginal') }}
                        </li>
                        <li class="flex items-center gap-2">
                            <svg class="size-5 shrink-0 text-brand" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                <path d="M3 7h11v9H3zM14 10h4l3 3v3h-7" />
                                <circle cx="7" cy="18" r="1.6" />
                                <circle cx="17" cy="18" r="1.6" />
                            </svg>
                            {{ t('product.perkDelivery') }}
                        </li>
                    </ul>

                    <div
                        v-if="pyramid.length"
                        class="mt-8 border-t border-line pt-6"
                    >
                        <h2 class="text-lg font-medium text-ink">{{ t('product.notes') }}</h2>
                        <dl class="mt-4 space-y-4">
                            <div
                                v-for="row in pyramid"
                                :key="row.key"
                                class="grid gap-2 sm:grid-cols-[150px_1fr]"
                            >
                                <dt>
                                    <span class="block text-sm font-medium text-ink">{{ row.label }}</span>
                                    <span class="block text-xs text-muted">{{ row.hint }}</span>
                                </dt>
                                <dd class="flex flex-wrap gap-1.5">
                                    <span
                                        v-for="note in row.notes"
                                        :key="note"
                                        class="rounded-full border border-line bg-surface px-3 py-1 text-sm text-ink"
                                    >{{ note }}</span>
                                </dd>
                            </div>
                        </dl>
                    </div>

                    <div
                        v-if="tr(product, 'description')"
                        class="mt-8 border-t border-line pt-6"
                    >
                        <h2 class="text-lg font-medium text-ink">{{ t('product.description') }}</h2>
                        <p class="mt-3 text-base leading-7 whitespace-pre-line text-muted">
                            {{ tr(product, 'description') }}
                        </p>
                    </div>
                </div>
            </section>

            <section
                v-if="related.length"
                class="mx-auto max-w-[1240px] px-4 py-12 sm:px-6 lg:py-16"
                aria-labelledby="related-title"
            >
                <h2
                    id="related-title"
                    class="mb-8 text-3xl font-medium tracking-tight text-ink"
                >
                    {{ t('product.related') }}
                </h2>
                <div class="grid grid-cols-2 gap-x-4 gap-y-10 sm:gap-x-6 lg:grid-cols-4">
                    <ProductCard
                        v-for="item in related"
                        :key="item.id"
                        :product="item"
                    />
                </div>
            </section>
        </template>
    </div>
</template>
