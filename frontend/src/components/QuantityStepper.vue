<!-- QuantityStepper.vue -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { MAX_QUANTITY } from '../stores/cart'

const props = withDefaults(
    defineProps<{ modelValue: number; size?: 'sm' | 'md' }>(),
    { size: 'md' }
)

const emit = defineEmits<{
    'update:modelValue': [value: number]
    remove: []
}>()

const { t } = useI18n()

function decrement(): void {
    if (props.modelValue <= 1) emit('remove')
    else emit('update:modelValue', props.modelValue - 1)
}

function increment(): void {
    if (props.modelValue < MAX_QUANTITY) {
        emit('update:modelValue', props.modelValue + 1)
    }
}
</script>

<template>
    <div
        class="flex items-center justify-between rounded-lg border border-line bg-white"
        :class="size === 'sm' ? 'h-9 w-28' : 'h-11 w-full'"
    >
        <button
            type="button"
            :aria-label="t('cart.decrease')"
            class="flex h-full w-10 cursor-pointer items-center justify-center text-muted transition-colors hover:text-brand focus-visible:outline-2 focus-visible:outline-brand"
            @click="decrement"
        >
            <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
                <path d="M5 12h14" />
            </svg>
        </button>

        <span
            class="min-w-6 text-center text-sm font-medium text-ink tabular-nums"
            aria-live="polite"
        >
            {{ modelValue }}
        </span>

        <button
            type="button"
            :aria-label="t('cart.increase')"
            :disabled="modelValue >= MAX_QUANTITY"
            class="flex h-full w-10 cursor-pointer items-center justify-center text-muted transition-colors hover:text-brand focus-visible:outline-2 focus-visible:outline-brand disabled:cursor-not-allowed disabled:opacity-40"
            @click="increment"
        >
            <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
                <path d="M12 5v14M5 12h14" />
            </svg>
        </button>
    </div>
</template>
