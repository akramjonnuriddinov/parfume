import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'
import { publicApi } from '../api/public'
import type { Product, Variant, VariantKind } from '../api/types'

const STORAGE_KEY = 'parfumeshop-cart'

/** One line per volume: the same perfume in 10 ml and 100 ml are two lines. */
export interface CartItem {
    variant_id: number
    product_id: number
    quantity: number
    name_uz: string
    name_ru: string
    name_en: string
    brand: string
    volume_ml: number
    kind: VariantKind
    price: number
    image_url: string
}

export const MAX_QUANTITY = 99

function load(): CartItem[] {
    try {
        const raw = localStorage.getItem(STORAGE_KEY)
        const parsed = raw ? JSON.parse(raw) : []
        return Array.isArray(parsed) ? parsed.filter((i) => typeof i?.variant_id === 'number') : []
    } catch {
        return []
    }
}

export const useCartStore = defineStore('cart', () => {
    const items = ref<CartItem[]>(load())
    const drawerOpen = ref(false)

    watch(
        items,
        (value) => {
            try {
                localStorage.setItem(STORAGE_KEY, JSON.stringify(value))
            } catch {
                // storage unavailable
            }
        },
        { deep: true }
    )

    const count = computed(() => items.value.reduce((sum, item) => sum + item.quantity, 0))
    const total = computed(() => items.value.reduce((sum, item) => sum + item.price * item.quantity, 0))

    function find(variantId: number) {
        return items.value.find((item) => item.variant_id === variantId)
    }

    /** Cheapest in-stock volume, used by "add to cart" on product cards. */
    function defaultVariant(product: Product): Variant | undefined {
        const inStock = product.variants.filter((v) => v.in_stock)
        return inStock.find((v) => v.kind === 'bottle') ?? inStock[0]
    }

    function add(product: Product, variant: Variant, quantity = 1): void {
        if (!variant.in_stock) return
        const existing = find(variant.id)
        if (existing) {
            existing.quantity = Math.min(existing.quantity + quantity, MAX_QUANTITY)
            return
        }
        items.value.push({
            variant_id: variant.id,
            product_id: product.id,
            quantity: Math.min(quantity, MAX_QUANTITY),
            name_uz: product.name_uz,
            name_ru: product.name_ru,
            name_en: product.name_en,
            brand: product.brand,
            volume_ml: variant.volume_ml,
            kind: variant.kind,
            price: variant.price,
            image_url: product.image_url,
        })
    }

    function setQuantity(variantId: number, quantity: number): void {
        const item = find(variantId)
        if (item) {
            item.quantity = Math.max(1, Math.min(Math.floor(quantity) || 1, MAX_QUANTITY))
        }
    }

    function remove(variantId: number): void {
        items.value = items.value.filter((item) => item.variant_id !== variantId)
    }

    function clear(): void {
        items.value = []
    }

    /** Syncs prices/names with the server; drops volumes that are gone or out of stock. */
    async function refresh(): Promise<number> {
        if (items.value.length === 0) return 0
        const ids = [...new Set(items.value.map((item) => item.product_id))]
        const { items: products } = await publicApi.products({ ids: ids.join(','), limit: 100 })
        const variants = new Map<number, { product: Product; variant: Variant }>()
        for (const product of products) {
            for (const variant of product.variants) variants.set(variant.id, { product, variant })
        }
        const before = items.value.length
        items.value = items.value
            .filter((item) => variants.get(item.variant_id)?.variant.in_stock)
            .map((item) => {
                const { product, variant } = variants.get(item.variant_id)!
                return {
                    ...item,
                    name_uz: product.name_uz,
                    name_ru: product.name_ru,
                    name_en: product.name_en,
                    brand: product.brand,
                    volume_ml: variant.volume_ml,
                    kind: variant.kind,
                    price: variant.price,
                    image_url: product.image_url,
                }
            })
        return before - items.value.length
    }

    return { items, drawerOpen, count, total, find, defaultVariant, add, setQuantity, remove, clear, refresh }
})
