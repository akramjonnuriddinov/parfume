import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { adminApi } from '../api/admin'
import type { Admin } from '../api/types'

const STORAGE_KEY = 'parfumeshop-admin-token'

function loadToken(): string | null {
    try {
        return localStorage.getItem(STORAGE_KEY)
    } catch {
        return null
    }
}

export const useAuthStore = defineStore('auth', () => {
    const token = ref<string | null>(loadToken())
    const admin = ref<Admin | null>(null)
    const isAuthenticated = computed(() => Boolean(token.value))

    function setToken(value: string | null): void {
        token.value = value
        try {
            if (value) localStorage.setItem(STORAGE_KEY, value)
            else localStorage.removeItem(STORAGE_KEY)
        } catch {
            // storage unavailable
        }
    }

    async function login(username: string, password: string): Promise<void> {
        const result = await adminApi.login(username, password)
        setToken(result.token)
        admin.value = result.admin
    }

    async function fetchMe(): Promise<void> {
        if (!token.value || admin.value) return
        admin.value = await adminApi.me()
    }

    function logout(): void {
        setToken(null)
        admin.value = null
    }

    return { token, admin, isAuthenticated, login, fetchMe, logout }
})
