<!-- PhoneInput.vue -->
<script setup lang="ts">
import { formatLocalPhoneInput } from '../utils/format'

defineProps<{ modelValue: string; id?: string }>()

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<template>
    <el-input
        :id="id"
        :model-value="modelValue"
        type="tel"
        inputmode="tel"
        autocomplete="tel-national"
        placeholder="90 123 45 67"
        maxlength="12"
        @update:model-value="emit('update:modelValue', formatLocalPhoneInput($event))"
    >
        <template #prepend>+998</template>
    </el-input>
</template>
