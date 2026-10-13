<!-- ChatsView.vue — AI assistant conversations and unmet product demand -->
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../api/admin'
import type { ChatSessionSummary, MissingProduct } from '../../api/types'
import { orderNumber } from '../../utils/format'
import { useLocalized } from '../../utils/useLocalized'
import AdminCard from '../components/AdminCard.vue'
import OrderStatusTag from '../components/OrderStatusTag.vue'

const PAGE_SIZE = 20
const PERIODS = [0, 7, 30, 90]

const { t } = useI18n()
const { price, date } = useLocalized()
const route = useRoute()
const router = useRouter()

const tab = computed(() => (route.query.tab === 'missing' ? 'missing' : 'chats'))
const outcome = computed(() => {
    const v = route.query.outcome
    return v === 'order' || v === 'lead' ? v : ''
})
const q = computed(() => (typeof route.query.q === 'string' ? route.query.q : ''))
const page = computed(() => Math.max(1, Number(route.query.page) || 1))
const days = computed(() => Number(route.query.days) || 0)

const chats = ref<ChatSessionSummary[]>([])
const total = ref(0)
const missing = ref<MissingProduct[]>([])
const loading = ref(false)
const search = ref('')

function setQuery(patch: Record<string, string | number | undefined>): void {
    const merged: Record<string, string | number | undefined> = {
        tab: tab.value === 'missing' ? 'missing' : undefined,
        outcome: outcome.value,
        q: q.value,
        days: days.value || undefined,
        page: undefined,
        ...patch,
    }
    const query: Record<string, string> = {}
    for (const [k, v] of Object.entries(merged)) {
        if (v !== undefined && v !== '' && !(k === 'page' && v === 1)) query[k] = String(v)
    }
    router.push({ path: '/admin/chats', query })
}

async function load(): Promise<void> {
    loading.value = true
    try {
        if (tab.value === 'missing') {
            missing.value = await adminApi.missingProducts(days.value)
        } else {
            const result = await adminApi.chats({ outcome: outcome.value, q: q.value, page: page.value, limit: PAGE_SIZE })
            chats.value = result.items
            total.value = result.total
        }
    } catch {
        chats.value = []
        missing.value = []
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

async function removeChat(chat: ChatSessionSummary): Promise<void> {
    try {
        await ElMessageBox.confirm(t('admin.chats.deleteConfirm'), t('common.confirm'), {
            type: 'warning',
            confirmButtonText: t('common.delete'),
            cancelButtonText: t('common.cancel'),
            confirmButtonClass: 'el-button--danger',
        })
    } catch {
        return
    }
    try {
        await adminApi.deleteChat(chat.id)
        ElMessage.success(t('common.deleted'))
        load()
    } catch {
        ElMessage.error(t('common.saveError'))
    }
}

async function dismissMissing(item: MissingProduct): Promise<void> {
    try {
        await ElMessageBox.confirm(t('admin.chats.dismissConfirm', { name: item.name }), t('common.confirm'), {
            type: 'warning',
            confirmButtonText: t('common.delete'),
            cancelButtonText: t('common.cancel'),
        })
    } catch {
        return
    }
    try {
        await adminApi.deleteMissingProduct(item.name)
        load()
    } catch {
        ElMessage.error(t('common.saveError'))
    }
}

const maxCount = computed(() => Math.max(1, ...missing.value.map((m) => m.count)))
</script>

<template>
    <div class="space-y-4">
        <el-radio-group
            :model-value="tab"
            @update:model-value="router.push({ path: '/admin/chats', query: $event === 'missing' ? { tab: 'missing' } : {} })"
        >
            <el-radio-button value="chats">{{ t('admin.chats.tabConversations') }}</el-radio-button>
            <el-radio-button value="missing">{{ t('admin.chats.tabMissing') }}</el-radio-button>
        </el-radio-group>

        <AdminCard
            v-if="tab === 'chats'"
            :padded="false"
        >
            <template #actions>
                <div class="flex w-full flex-col gap-3 md:flex-row md:items-center md:justify-between">
                    <el-radio-group
                        :model-value="outcome"
                        @update:model-value="setQuery({ outcome: $event as string })"
                    >
                        <el-radio-button value="">{{ t('admin.chats.all') }}</el-radio-button>
                        <el-radio-button value="order">{{ t('admin.chats.withOrder') }}</el-radio-button>
                        <el-radio-button value="lead">{{ t('admin.chats.withLead') }}</el-radio-button>
                    </el-radio-group>
                    <form
                        class="md:w-72"
                        @submit.prevent="setQuery({ q: search.trim() })"
                    >
                        <el-input
                            v-model="search"
                            clearable
                            :placeholder="t('admin.chats.search')"
                            @clear="setQuery({ q: undefined })"
                        />
                    </form>
                </div>
            </template>

            <div v-loading="loading">
                <el-empty
                    v-if="!loading && chats.length === 0"
                    :description="t('admin.chats.empty')"
                />
                <ul class="divide-y divide-line">
                    <li
                        v-for="chat in chats"
                        :key="chat.id"
                        class="flex flex-col gap-3 px-5 py-4 sm:flex-row sm:items-center"
                    >
                        <RouterLink
                            :to="`/admin/chats/${chat.id}`"
                            class="min-w-0 flex-1"
                        >
                            <p class="line-clamp-2 text-sm leading-6 font-medium text-ink hover:text-brand">
                                {{ chat.first_message || '—' }}
                            </p>
                            <div class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted">
                                <span>{{ date(chat.updated_at) }}</span>
                                <span>{{ chat.locale.toUpperCase() }}</span>
                                <span>{{ t('admin.chats.messages', { n: chat.messages_count }) }}</span>
                            </div>
                        </RouterLink>
                        <div class="flex shrink-0 flex-wrap items-center gap-2">
                            <template v-if="chat.order_id">
                                <el-tag
                                    type="success"
                                    effect="plain"
                                >
                                    {{ t('admin.chats.order', { number: orderNumber(chat.order_id) }) }}
                                    <template v-if="chat.order_total !== null"> · {{ price(chat.order_total) }}</template>
                                </el-tag>
                                <OrderStatusTag
                                    v-if="chat.order_status"
                                    :status="chat.order_status"
                                />
                            </template>
                            <el-tag
                                v-if="chat.leads_count > 0"
                                type="warning"
                                effect="plain"
                            >
                                {{ t('admin.chats.leads', { n: chat.leads_count }) }}
                            </el-tag>
                            <el-button
                                size="small"
                                @click="router.push(`/admin/chats/${chat.id}`)"
                            >
                                {{ t('admin.chats.open') }}
                            </el-button>
                            <el-button
                                size="small"
                                type="danger"
                                plain
                                @click="removeChat(chat)"
                            >
                                {{ t('common.delete') }}
                            </el-button>
                        </div>
                    </li>
                </ul>
            </div>

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

        <AdminCard
            v-else
            :padded="false"
        >
            <template #actions>
                <div class="flex w-full flex-col gap-3 md:flex-row md:items-center md:justify-between">
                    <p class="max-w-xl text-sm text-muted">{{ t('admin.chats.missingHint') }}</p>
                    <el-select
                        :model-value="days"
                        class="md:w-48!"
                        :aria-label="t('admin.chats.period')"
                        @update:model-value="setQuery({ days: $event || undefined })"
                    >
                        <el-option
                            v-for="d in PERIODS"
                            :key="d"
                            :value="d"
                            :label="d ? t('admin.chats.days', { n: d }) : t('admin.chats.allTime')"
                        />
                    </el-select>
                </div>
            </template>

            <div v-loading="loading">
                <el-empty
                    v-if="!loading && missing.length === 0"
                    :description="t('admin.chats.missingEmpty')"
                />
                <ul class="divide-y divide-line">
                    <li
                        v-for="item in missing"
                        :key="item.name"
                        class="flex flex-col gap-3 px-5 py-4 sm:flex-row sm:items-start"
                    >
                        <div class="min-w-0 flex-1">
                            <div class="flex items-center gap-3">
                                <p class="text-base font-medium text-ink first-letter:uppercase">{{ item.name }}</p>
                                <span class="rounded-full bg-brand-soft px-2 py-0.5 text-xs font-semibold text-brand">
                                    {{ t('admin.chats.requests') }}: {{ item.count }}
                                </span>
                            </div>
                            <div
                                class="mt-2 h-1.5 max-w-sm overflow-hidden rounded-full bg-surface"
                                aria-hidden="true"
                            >
                                <div
                                    class="h-full rounded-full bg-brand"
                                    :style="{ width: `${(item.count / maxCount) * 100}%` }"
                                />
                            </div>
                            <p
                                v-if="item.details.length"
                                class="mt-2 text-sm text-muted"
                            >
                                {{ t('admin.chats.context') }}: {{ item.details.join(' · ') }}
                            </p>
                            <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted">
                                <span>{{ t('admin.chats.lastAsked') }}: {{ date(item.last_asked) }}</span>
                                <RouterLink
                                    v-for="(id, i) in item.session_ids.slice(0, 5)"
                                    :key="id"
                                    :to="`/admin/chats/${id}`"
                                    class="text-brand hover:underline"
                                >
                                    {{ t('admin.chats.open') }} #{{ i + 1 }}
                                </RouterLink>
                            </div>
                        </div>
                        <el-button
                            size="small"
                            plain
                            @click="dismissMissing(item)"
                        >
                            {{ t('admin.chats.dismiss') }}
                        </el-button>
                    </li>
                </ul>
            </div>
        </AdminCard>
    </div>
</template>
