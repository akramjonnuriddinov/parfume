<template>
    <HeroSection
        @catalog="router.push('/catalog')"
        @decants="scrollToDecants"
    />
    <GenderSection @select="router.push({ path: '/catalog', query: { gender: $event } })" />
    <ProductsSection
        :products="products"
        :loading="loading"
        @catalog="router.push('/catalog')"
    />
    <CategoriesSection
        @select="router.push({ path: '/catalog', query: { category: $event } })"
        @catalog="router.push('/catalog')"
    />
    <DecantSection @catalog="router.push('/catalog')" />
    <BrandsSection
        :brands="brands"
        @select="router.push({ path: '/catalog', query: { brand: $event } })"
    />
    <ConsultationSection @consultation="inquiry.open()" />
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import HeroSection from '../features/home/components/HeroSection.vue'
import GenderSection from '../features/home/components/GenderSection.vue'
import CategoriesSection from '../features/home/components/CategoriesSection.vue'
import ProductsSection from '../features/home/components/ProductsSection.vue'
import DecantSection from '../features/home/components/DecantSection.vue'
import BrandsSection from '../features/home/components/BrandsSection.vue'
import ConsultationSection from '../features/home/components/ConsultationSection.vue'
import { publicApi } from '../api/public'
import type { Brand, Product } from '../api/types'
import { useInquiryStore } from '../stores/inquiry'

const router = useRouter()
const inquiry = useInquiryStore()

const products = ref<Product[]>([])
const brands = ref<Brand[]>([])
const loading = ref(true)

function scrollToDecants(): void {
    document.getElementById('raspiv')?.scrollIntoView({ behavior: 'smooth' })
}

onMounted(async () => {
    publicApi.brands().then((list) => (brands.value = list)).catch(() => undefined)
    try {
        const featured = await publicApi.products({ featured: true, limit: 8 })
        products.value = featured.items.length ? featured.items : (await publicApi.products({ limit: 8 })).items
    } catch {
        products.value = []
    } finally {
        loading.value = false
    }
})
</script>
