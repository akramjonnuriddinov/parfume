<script setup lang="ts">
import { onMounted } from 'vue'
import AppHeader from '../components/AppHeader.vue'
import AppFooter from '../components/AppFooter.vue'
import CartDrawer from '../components/CartDrawer.vue'
import InquiryDialog from '../components/InquiryDialog.vue'
import ChatWidget from '../components/chat/ChatWidget.vue'
import { useSettingsStore } from '../stores/settings'

const settings = useSettingsStore()
onMounted(() => settings.load())
</script>

<template>
    <div class="flex min-h-screen flex-col bg-white">
        <AppHeader />
        <main class="flex-1">
            <RouterView />
        </main>
        <AppFooter />
        <CartDrawer />
        <InquiryDialog />
        <ChatWidget />
    </div>
</template>
