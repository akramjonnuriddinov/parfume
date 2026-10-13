<!-- OrderDetailView.vue -->
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../api/admin'
import { ORDER_STATUSES, type Order, type OrderStatus } from '../../api/types'
import ProductImage from '../../components/ProductImage.vue'
import { useAdminCountersStore } from '../../stores/adminCounters'
import { formatPhone, orderNumber, telHref } from '../../utils/format'
import { useLocalized } from '../../utils/useLocalized'
import { usePerfume } from '../../utils/usePerfume'
import AdminCard from '../components/AdminCard.vue'
import OrderStatusTag from '../components/OrderStatusTag.vue'

const { t } = useI18n()
const { price, date } = useLocalized()
const perfume = usePerfume()
const route = useRoute()
const router = useRouter()
const counters = useAdminCountersStore()

const order = ref<Order | null>(null)
const loading = ref(true)
const saving = ref(false)
const status = ref<OrderStatus>('new')
const note = ref('')

const id = Number(route.params.id)

onMounted(async () => {
    try {
        order.value = await adminApi.order(id)
        status.value = order.value.status
        note.value = order.value.admin_note
    } catch {
        order.value = null
    } finally {
        loading.value = false
    }
})

async function save(): Promise<void> {
    saving.value = true
    try {
        order.value = await adminApi.updateOrder(id, status.value, note.value)
        ElMessage.success(t('common.saved'))
        counters.refresh()
    } catch {
        ElMessage.error(t('common.saveError'))
    } finally {
        saving.value = false
    }
}

async function remove(): Promise<void> {
    try {
        await ElMessageBox.confirm(t('admin.orders.deleteConfirm'), t('common.confirm'), {
            type: 'warning',
            confirmButtonText: t('common.delete'),
            cancelButtonText: t('common.cancel'),
            confirmButtonClass: 'el-button--danger',
        })
    } catch {
        return
    }
    try {
        await adminApi.deleteOrder(id)
        ElMessage.success(t('common.deleted'))
        router.replace('/admin/orders')
    } catch {
        ElMessage.error(t('common.saveError'))
    }
}
</script>

<template>
    <div v-loading="loading">
        <RouterLink
            to="/admin/orders"
            class="mb-4 inline-flex items-center gap-1 text-sm text-muted hover:text-brand"
        >
            ← {{ t('admin.orders.back') }}
        </RouterLink>

        <el-result
            v-if="!loading && !order"
            icon="warning"
            :title="t('admin.orders.notFound')"
        />

        <div
            v-else-if="order"
            class="grid gap-6 xl:grid-cols-[1fr_360px]"
        >
            <div class="space-y-6">
                <AdminCard :padded="false">
                    <template #actions>
                        <div class="flex w-full flex-wrap items-center gap-3">
                            <h2 class="text-lg font-medium text-ink">
                                {{ t('admin.orders.order') }} {{ orderNumber(order.id) }}
                            </h2>
                            <OrderStatusTag :status="order.status" />
                            <el-tag
                                v-if="order.source === 'ai_chat'"
                                effect="plain"
                            >
                                {{ t('admin.orders.aiBadge') }}
                            </el-tag>
                            <span class="ml-auto text-sm text-muted">{{ date(order.created_at) }}</span>
                        </div>
                    </template>

                    <ul class="divide-y divide-line">
                        <li
                            v-for="item in order.items"
                            :key="item.id"
                            class="flex items-center gap-4 px-5 py-4"
                        >
                            <div class="size-14 shrink-0 overflow-hidden rounded-lg bg-surface p-1.5">
                                <ProductImage :src="item.product_image" />
                            </div>
                            <div class="min-w-0 flex-1">
                                <p
                                    v-if="item.product_brand"
                                    class="text-[11px] font-medium tracking-wider text-gold uppercase"
                                >
                                    {{ item.product_brand }}
                                </p>
                                <RouterLink
                                    v-if="item.product_id"
                                    :to="`/admin/products/${item.product_id}`"
                                    class="text-sm font-medium text-ink hover:text-brand"
                                >
                                    {{ item.product_name }}
                                </RouterLink>
                                <p
                                    v-else
                                    class="text-sm font-medium text-ink"
                                >
                                    {{ item.product_name }}
                                </p>
                                <p class="mt-0.5 text-xs text-muted">
                                    <span class="font-medium text-ink">{{ perfume.volume(item) }}</span> ·
                                    {{ price(item.price) }} × {{ item.quantity }}
                                </p>
                            </div>
                            <p class="text-sm font-semibold text-ink">{{ price(item.subtotal) }}</p>
                        </li>
                    </ul>
                    <div class="flex items-center justify-between border-t border-line px-5 py-4">
                        <span class="text-sm text-muted">{{ t('cart.total') }}</span>
                        <span class="text-xl font-semibold text-ink">{{ price(order.total) }}</span>
                    </div>
                </AdminCard>

                <AdminCard
                    v-if="order.comment"
                    :title="t('admin.orders.comment')"
                >
                    <p class="text-sm leading-6 whitespace-pre-line text-ink">{{ order.comment }}</p>
                </AdminCard>
            </div>

            <div class="space-y-6">
                <AdminCard :title="t('admin.orders.customer')">
                    <dl class="space-y-4 text-sm">
                        <div>
                            <dt class="text-muted">{{ t('form.fullName') }}</dt>
                            <dd class="mt-1 font-medium text-ink">{{ order.full_name }}</dd>
                        </div>
                        <div>
                            <dt class="text-muted">{{ t('form.phone') }}</dt>
                            <dd class="mt-1">
                                <a
                                    :href="telHref(order.phone)"
                                    class="font-medium text-brand hover:underline"
                                >
                                    {{ formatPhone(order.phone) }}
                                </a>
                            </dd>
                        </div>
                        <div>
                            <dt class="text-muted">{{ t('admin.orders.language') }}</dt>
                            <dd class="mt-1 font-medium text-ink uppercase">{{ order.locale }}</dd>
                        </div>
                    </dl>
                </AdminCard>

                <AdminCard :title="t('admin.orders.manage')">
                    <el-form label-position="top">
                        <el-form-item :label="t('admin.orders.status')">
                            <el-select v-model="status">
                                <el-option
                                    v-for="s in ORDER_STATUSES"
                                    :key="s"
                                    :value="s"
                                    :label="t(`admin.status.${s}`)"
                                />
                            </el-select>
                        </el-form-item>
                        <el-form-item :label="t('admin.orders.note')">
                            <el-input
                                v-model="note"
                                type="textarea"
                                :rows="4"
                                maxlength="2000"
                                :placeholder="t('admin.orders.notePlaceholder')"
                            />
                        </el-form-item>
                        <el-button
                            type="primary"
                            class="w-full!"
                            :loading="saving"
                            @click="save"
                        >
                            {{ t('common.save') }}
                        </el-button>
                    </el-form>
                    <div class="mt-5 border-t border-line pt-4 text-xs text-muted">
                        {{ t('admin.orders.updated') }}: {{ date(order.updated_at) }}
                    </div>
                    <el-button
                        type="danger"
                        text
                        class="mt-2! -ml-3!"
                        @click="remove"
                    >
                        {{ t('admin.orders.delete') }}
                    </el-button>
                </AdminCard>
            </div>
        </div>
    </div>
</template>
