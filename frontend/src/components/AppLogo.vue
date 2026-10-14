<!-- AppLogo.vue — Mariana Parfumes wordmark (transparent PNG made from assets/images/logo.jpeg) -->
<script setup lang="ts">
import logo from '../assets/images/logo.png'

withDefaults(defineProps<{ height?: string }>(), { height: 'h-10' })
</script>

<template>
    <img
        :src="logo"
        alt="Mariana Parfumes"
        width="591"
        height="180"
        class="w-auto shrink-0 select-none"
        :class="height"
        draggable="false"
    >
</template>
