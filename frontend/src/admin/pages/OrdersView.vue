<!-- OrdersView.vue -->
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { adminApi } from '../../api/admin'
import { ORDER_STATUSES, type Order } from '../../api/types'
import { formatPhone, orderNumber } from '../../utils/format'
import { useLocalized } from '../../utils/useLocalized'
import AdminCard from '../components/AdminCard.vue'
import OrderStatusTag from '../components/OrderStatusTag.vue'

const PAGE_SIZE = 20

const { t } = useI18n()
const { price, date } = useLocalized()
const route = useRoute()
const router = useRouter()

const orders = ref<Order[]>([])
const total = ref(0)
const loading = ref(false)
const search = ref('')

const status = computed(() => (typeof route.query.status === 'string' ? route.query.status : ''))
const q = computed(() => (typeof route.query.q === 'string' ? route.query.q : ''))
const page = computed(() => Math.max(1, Number(route.query.page) || 1))

function setQuery(patch: Record<string, string | number | undefined>): void {
    const merged: Record<string, string | number | undefined> = { status: status.value, q: q.value, page: undefined, ...patch }
    const query: Record<string, string> = {}
    for (const [k, v] of Object.entries(merged)) {
        if (v !== undefined && v !== '' && !(k === 'page' && v === 1)) query[k] = String(v)
    }
    router.push({ path: '/admin/orders', query })
}

async function load(): Promise<void> {
    loading.value = true
    try {
        const result = await adminApi.orders({ status: status.value, q: q.value, page: page.value, limit: PAGE_SIZE })
        orders.value = result.items
        total.value = result.total
    } catch {
        orders.value = []
    } finally {
        loading.value = false
    }
}

watch(
    () => route.query,
    () => {
        search.value = q.value
        load()
    },
    { immediate: true }
)
</script>

<template>
    <AdminCard :padded="false">
        <template #actions>
            <div class="flex w-full flex-col gap-3 md:flex-row md:items-center md:justify-between">
                <el-radio-group
                    :model-value="status"
                    size="default"
                    @update:model-value="setQuery({ status: $event as string })"
                >
                    <el-radio-button value="">{{ t('admin.orders.all') }}</el-radio-button>
                    <el-radio-button
                        v-for="s in ORDER_STATUSES"
                        :key="s"
                        :value="s"
                    >
                        {{ t(`admin.status.${s}`) }}
                    </el-radio-button>
                </el-radio-group>
                <form
                    class="md:w-72"
                    @submit.prevent="setQuery({ q: search.trim() })"
                >
                    <el-input
                        v-model="search"
                        size="default"
                        clearable
                        :placeholder="t('admin.orders.search')"
                        @clear="setQuery({ q: undefined })"
                    />
                </form>
            </div>
        </template>

        <el-table
            v-loading="loading"
            :data="orders"
            :empty-text="t('admin.orders.empty')"
            class="cursor-pointer"
            @row-click="(row: Order) => router.push(`/admin/orders/${row.id}`)"
        >
            <el-table-column
                :label="t('admin.orders.number')"
                width="170"
            >
                <template #default="{ row }">
                    <span class="font-medium text-ink">{{ orderNumber(row.id) }}</span>
                    <el-tag
                        v-if="row.source === 'ai_chat'"
                        size="small"
                        effect="plain"
                        class="ml-1.5"
                    >
                        {{ t('admin.orders.aiBadge') }}
                    </el-tag>
                </template>
            </el-table-column>
            <el-table-column
                :label="t('admin.orders.customer')"
                min-width="200"
            >
                <template #default="{ row }">
                    <p class="text-ink">{{ row.full_name }}</p>
                    <p class="text-xs text-muted">{{ formatPhone(row.phone) }}</p>
                </template>
            </el-table-column>
            <el-table-column
                :label="t('admin.orders.items')"
                width="110"
                align="center"
            >
                <template #default="{ row }">{{ row.items_count }}</template>
            </el-table-column>
            <el-table-column
                :label="t('admin.orders.total')"
                min-width="150"
            >
                <template #default="{ row }">
                    <span class="font-medium text-ink">{{ price(row.total) }}</span>
                </template>
            </el-table-column>
            <el-table-column
                :label="t('admin.orders.status')"
                width="150"
            >
                <template #default="{ row }">
                    <OrderStatusTag :status="row.status" />
                </template>
            </el-table-column>
            <el-table-column
                :label="t('admin.orders.date')"
                min-width="160"
            >
                <template #default="{ row }">
                    <span class="text-muted">{{ date(row.created_at) }}</span>
                </template>
            </el-table-column>
        </el-table>

        <div
            v-if="total > PAGE_SIZE"
            class="flex justify-end border-t border-line px-5 py-4"
        >
            <el-pagination
                background
                layout="total, prev, pager, next"
                :total="total"
                :page-size="PAGE_SIZE"
                :current-page="page"
                @current-change="setQuery({ page: $event })"
            />
        </div>
    </AdminCard>
</template>
