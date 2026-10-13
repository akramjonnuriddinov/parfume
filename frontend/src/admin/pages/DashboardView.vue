<!-- DashboardView.vue -->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { adminApi } from '../../api/admin'
import { ORDER_STATUSES, type Stats } from '../../api/types'
import { formatPhone, orderNumber } from '../../utils/format'
import { useLocalized } from '../../utils/useLocalized'
import AdminCard from '../components/AdminCard.vue'
import OrderStatusTag from '../components/OrderStatusTag.vue'

const { t } = useI18n()
const { price, date } = useLocalized()
const router = useRouter()

const stats = ref<Stats | null>(null)
const loading = ref(true)
const failed = ref(false)

async function load(): Promise<void> {
    loading.value = true
    failed.value = false
    try {
        stats.value = await adminApi.stats()
    } catch {
        failed.value = true
    } finally {
        loading.value = false
    }
}

onMounted(load)

const tiles = computed(() => {
    const s = stats.value
    if (!s) return []
    return [
        { label: t('admin.dashboard.newOrders'), value: String(s.orders_by_status.new), to: '/admin/orders?status=new', accent: true },
        { label: t('admin.dashboard.ordersToday'), value: String(s.orders_today), to: '/admin/orders' },
        { label: t('admin.dashboard.revenue'), value: price(s.revenue_total), to: '/admin/orders?status=completed' },
        { label: t('admin.dashboard.unreadMessages'), value: String(s.messages_unread), to: '/admin/messages' },
    ]
})
</script>

<template>
    <div v-loading="loading && !stats">
        <el-result
            v-if="failed"
            icon="error"
            :title="t('common.loadError')"
        >
            <template #extra>
                <el-button
                    type="primary"
                    @click="load"
                >
                    {{ t('common.retry') }}
                </el-button>
            </template>
        </el-result>

        <template v-else-if="stats">
            <div class="grid grid-cols-2 gap-4 xl:grid-cols-4">
                <RouterLink
                    v-for="tile in tiles"
                    :key="tile.label"
                    :to="tile.to"
                    class="rounded-xl border bg-white p-5 transition-colors hover:border-brand"
                    :class="tile.accent ? 'border-brand/40' : 'border-line'"
                >
                    <p class="text-sm text-muted">{{ tile.label }}</p>
                    <p
                        class="mt-3 text-2xl font-semibold tracking-tight sm:text-3xl"
                        :class="tile.accent ? 'text-brand' : 'text-ink'"
                    >
                        {{ tile.value }}
                    </p>
                </RouterLink>
            </div>

            <div class="mt-6 grid gap-6 xl:grid-cols-[1fr_320px]">
                <AdminCard
                    :title="t('admin.dashboard.recentOrders')"
                    :padded="false"
                >
                    <template #actions>
                        <RouterLink
                            to="/admin/orders"
                            class="text-sm text-brand hover:underline"
                        >
                            {{ t('admin.dashboard.allOrders') }}
                        </RouterLink>
                    </template>

                    <el-table
                        :data="stats.recent_orders"
                        :empty-text="t('admin.orders.empty')"
                        class="cursor-pointer"
                        @row-click="(row: { id: number }) => router.push(`/admin/orders/${row.id}`)"
                    >
                        <el-table-column
                            :label="t('admin.orders.number')"
                            width="100"
                        >
                            <template #default="{ row }">
                                <span class="font-medium text-ink">{{ orderNumber(row.id) }}</span>
                            </template>
                        </el-table-column>
                        <el-table-column
                            :label="t('admin.orders.customer')"
                            min-width="180"
                        >
                            <template #default="{ row }">
                                <p class="text-ink">{{ row.full_name }}</p>
                                <p class="text-xs text-muted">{{ formatPhone(row.phone) }}</p>
                            </template>
                        </el-table-column>
                        <el-table-column
                            :label="t('admin.orders.total')"
                            min-width="140"
                        >
                            <template #default="{ row }">{{ price(row.total) }}</template>
                        </el-table-column>
                        <el-table-column
                            :label="t('admin.orders.status')"
                            width="140"
                        >
                            <template #default="{ row }">
                                <OrderStatusTag :status="row.status" />
                            </template>
                        </el-table-column>
                        <el-table-column
                            :label="t('admin.orders.date')"
                            min-width="150"
                        >
                            <template #default="{ row }">
                                <span class="text-muted">{{ date(row.created_at) }}</span>
                            </template>
                        </el-table-column>
                    </el-table>
                </AdminCard>

                <div class="space-y-6">
                    <AdminCard :title="t('admin.dashboard.byStatus')">
                        <ul class="space-y-3">
                            <li
                                v-for="status in ORDER_STATUSES"
                                :key="status"
                            >
                                <RouterLink
                                    :to="`/admin/orders?status=${status}`"
                                    class="flex items-center justify-between rounded-lg px-1 py-1 hover:bg-surface"
                                >
                                    <OrderStatusTag :status="status" />
                                    <span class="text-sm font-medium text-ink tabular-nums">
                                        {{ stats.orders_by_status[status] }}
                                    </span>
                                </RouterLink>
                            </li>
                        </ul>
                        <div class="mt-4 flex justify-between border-t border-line pt-4 text-sm">
                            <span class="text-muted">{{ t('admin.dashboard.totalOrders') }}</span>
                            <span class="font-semibold text-ink">{{ stats.orders_total }}</span>
                        </div>
                    </AdminCard>

                    <AdminCard :title="t('admin.dashboard.catalog')">
                        <dl class="space-y-3 text-sm">
                            <div class="flex justify-between">
                                <dt class="text-muted">{{ t('admin.dashboard.productsActive') }}</dt>
                                <dd class="font-medium text-ink">{{ stats.products_active }} / {{ stats.products_total }}</dd>
                            </div>
                            <div class="flex justify-between">
                                <dt class="text-muted">{{ t('admin.nav.categories') }}</dt>
                                <dd class="font-medium text-ink">{{ stats.categories_total }}</dd>
                            </div>
                        </dl>
                        <el-button
                            type="primary"
                            class="mt-5! w-full!"
                            @click="router.push('/admin/products/new')"
                        >
                            {{ t('admin.products.new') }}
                        </el-button>
                    </AdminCard>
                </div>
            </div>
        </template>
    </div>
</template>
