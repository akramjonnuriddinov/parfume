import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Product } from '../api/types'

/** Global "ask the price / contact seller" dialog state. */
export const useInquiryStore = defineStore('inquiry', () => {
    const visible = ref(false)
    const product = ref<Product | null>(null)

    function open(target: Product | null = null): void {
        product.value = target
        visible.value = true
    }

    return { visible, product, open }
})
