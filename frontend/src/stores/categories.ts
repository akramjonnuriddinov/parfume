import { defineStore } from 'pinia'
import { ref } from 'vue'
import { publicApi } from '../api/public'
import type { Category } from '../api/types'

export const useCategoriesStore = defineStore('categories', () => {
    const items = ref<Category[]>([])
    const loaded = ref(false)
    let pending: Promise<void> | null = null

    function load(): Promise<void> {
        if (!pending) {
            pending = publicApi
                .categories()
                .then((data) => {
                    items.value = data
                    loaded.value = true
                })
                .catch(() => {
                    pending = null
                })
        }
        return pending
    }

    return { items, loaded, load }
})
