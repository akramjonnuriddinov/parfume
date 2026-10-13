import { defineStore } from 'pinia'
import { ref } from 'vue'
import { publicApi } from '../api/public'
import type { Settings } from '../api/types'

export const DEFAULT_PHONE = '+998 90 777 40 40'

export const useSettingsStore = defineStore('settings', () => {
    const settings = ref<Settings>({})
    let loading: Promise<void> | null = null

    function load(force = false): Promise<void> {
        if (!loading || force) {
            loading = publicApi
                .settings()
                .then((data) => {
                    settings.value = data
                })
                .catch(() => {
                    loading = null
                })
        }
        return loading
    }

    function set(data: Settings): void {
        settings.value = data
    }

    return { settings, load, set }
})
