<!-- ProductImage.vue -->
<script setup lang="ts">
import { ref, watch } from 'vue'
import { assetUrl } from '../api/http'

const props = defineProps<{ src: string; alt?: string; eager?: boolean }>()

const failed = ref(false)
watch(
    () => props.src,
    () => (failed.value = false)
)
</script>

<template>
    <img
        v-if="src && !failed"
        :src="assetUrl(src)"
        :alt="alt ?? ''"
        class="h-full w-full object-contain"
        :loading="eager ? 'eager' : 'lazy'"
        decoding="async"
        @error="failed = true"
    />
    <span
        v-else
        class="flex h-full w-full items-center justify-center text-brand/40"
        aria-hidden="true"
    >
        <svg class="size-1/3 max-h-16 max-w-16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round">
            <rect x="3" y="3" width="18" height="18" rx="4" />
            <path d="M12 8v8M8 12h8" />
        </svg>
    </span>
</template>
