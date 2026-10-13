import { createApp, watch } from 'vue'
import 'element-plus/dist/index.css'
import './assets/css/index.css'
import App from './App.vue'
import { createPinia } from 'pinia'
import { applyDocumentTitle, router } from './router'
import { i18n } from './i18n'
import { onUnauthorized } from './api/http'
import { useAuthStore } from './stores/auth'
import ElementPlus from 'element-plus'

const pinia = createPinia()
const app = createApp(App)

app.use(ElementPlus, { size: 'large', zIndex: 3000 })
app.use(pinia)
app.use(i18n)
app.use(router)

onUnauthorized(() => {
    useAuthStore().logout()
    const current = router.currentRoute.value
    if (current.meta.requiresAuth) {
        router.replace({ name: 'admin-login', query: { redirect: current.fullPath } })
    }
})

watch(i18n.global.locale, applyDocumentTitle)

app.mount('#app')
