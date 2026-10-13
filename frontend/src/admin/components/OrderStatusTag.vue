<!-- OrderStatusTag.vue -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { OrderStatus } from '../../api/types'

defineProps<{ status: OrderStatus }>()

const { t } = useI18n()

const types: Record<OrderStatus, 'primary' | 'warning' | 'success' | 'info'> = {
    new: 'primary',
    processing: 'warning',
    completed: 'success',
    cancelled: 'info',
}
</script>

<template>
    <el-tag
        :type="types[status]"
        effect="light"
        round
        disable-transitions
    >
        {{ t(`admin.status.${status}`) }}
    </el-tag>
</template>
