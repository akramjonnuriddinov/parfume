import { defineStore } from 'pinia'
import { ref } from 'vue'
import { adminApi } from '../api/admin'

/** Sidebar badges: new orders and unread messages. */
export const useAdminCountersStore = defineStore('adminCounters', () => {
    const newOrders = ref(0)
    const unreadMessages = ref(0)

    async function refresh(): Promise<void> {
        try {
            const stats = await adminApi.stats()
            newOrders.value = stats.orders_by_status.new
            unreadMessages.value = stats.messages_unread
        } catch {
            // badges are best-effort
        }
    }

    return { newOrders, unreadMessages, refresh }
})
