<!-- ChatDetailView.vue — transcript of one AI assistant conversation -->
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../api/admin'
import type { ChatSessionDetail } from '../../api/types'
import { formatPhone, orderNumber, telHref } from '../../utils/format'
import { useLocalized } from '../../utils/useLocalized'
import AdminCard from '../components/AdminCard.vue'
import OrderStatusTag from '../components/OrderStatusTag.vue'

const { t } = useI18n()
const { price, date } = useLocalized()
const route = useRoute()
const router = useRouter()

const chat = ref<ChatSessionDetail | null>(null)
const loading = ref(true)
const failed = ref(false)

async function load(): Promise<void> {
    loading.value = true
    failed.value = false
    try {
        chat.value = await adminApi.chat(String(route.params.id))
    } catch {
        failed.value = true
    } finally {
        loading.value = false
    }
}

onMounted(load)

async function remove(): Promise<void> {
    if (!chat.value) return
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
        await adminApi.deleteChat(chat.value.id)
        ElMessage.success(t('common.deleted'))
        router.push('/admin/chats')
    } catch {
        ElMessage.error(t('common.saveError'))
    }
}

function leadLabel(source: string): string {
    if (source === 'ai_quote') return t('admin.messages.sourceQuote')
    if (source === 'ai_callback') return t('admin.messages.sourceCallback')
    return t('admin.messages.sourceSite')
}
</script>

<template>
    <div class="space-y-4">
        <RouterLink
            to="/admin/chats"
            class="inline-flex items-center gap-1 text-sm text-muted hover:text-brand"
        >
            <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="M15 18l-6-6 6-6" />
            </svg>
            {{ t('admin.chats.back') }}
        </RouterLink>

        <div
            v-if="loading"
            v-loading="true"
            class="h-64"
        />

        <el-result
            v-else-if="failed || !chat"
            icon="error"
            :title="t('common.loadError')"
        >
            <template #extra>
                <el-button @click="load">{{ t('common.retry') }}</el-button>
            </template>
        </el-result>

        <div
            v-else
            class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_320px]"
        >
            <AdminCard :padded="false">
                <template #actions>
                    <div class="flex w-full flex-wrap items-center gap-3">
                        <span class="text-sm text-muted">
                            {{ t('admin.chats.started') }}: {{ date(chat.created_at) }} · {{ chat.locale.toUpperCase() }} ·
                            {{ t('admin.chats.messages', { n: chat.messages_count }) }}
                        </span>
                        <el-button
                            class="ml-auto"
                            size="small"
                            type="danger"
                            plain
                            @click="remove"
                        >
                            {{ t('common.delete') }}
                        </el-button>
                    </div>
                </template>

                <ol class="space-y-4 bg-surface/60 p-5">
                    <li
                        v-for="message in chat.messages"
                        :key="message.id"
                        :class="message.role === 'user' ? 'flex flex-col items-end' : 'flex flex-col items-start'"
                    >
                        <span class="mb-1 text-[11px] text-muted">
                            {{ message.role === 'user' ? t('admin.chats.customer') : t('admin.chats.assistant') }} ·
                            {{ date(message.created_at) }}
                        </span>
                        <div
                            v-if="message.content"
                            class="max-w-[85%] rounded-2xl px-4 py-2.5 text-sm leading-6 whitespace-pre-line"
                            :class="message.role === 'user' ? 'rounded-tr-md bg-brand text-white' : 'rounded-tl-md bg-white text-ink shadow-xs'"
                        >
                            {{ message.content }}
                        </div>

                        <div
                            v-if="message.meta.products?.length"
                            class="mt-2 max-w-[85%] rounded-xl border border-line bg-white px-3 py-2 text-xs"
                        >
                            <p class="mb-1 font-medium text-muted">{{ t('admin.chats.shownProducts') }}</p>
                            <ul class="space-y-0.5">
                                <li
                                    v-for="(p, i) in message.meta.products"
                                    :key="i"
                                >
                                    <span
                                        v-if="p.group"
                                        class="text-muted"
                                    >{{ p.group }} · </span>
                                    <RouterLink
                                        :to="`/admin/products/${p.product_id}`"
                                        class="text-brand hover:underline"
                                    >
                                        {{ p.name }}
                                    </RouterLink>
                                    <template v-if="p.volume"> · {{ p.volume }}</template>
                                    × {{ p.quantity }}
                                </li>
                            </ul>
                        </div>

                        <ul
                            v-if="message.meta.notes?.length"
                            class="mt-2 max-w-[85%] space-y-1"
                        >
                            <li
                                v-for="(note, i) in message.meta.notes"
                                :key="i"
                                class="rounded-lg bg-amber-50 px-2.5 py-1 text-xs text-amber-800"
                            >
                                {{ note }}
                            </li>
                        </ul>
                    </li>
                </ol>
            </AdminCard>

            <div class="space-y-4">
                <AdminCard
                    v-if="chat.order_id"
                    :title="t('admin.chats.order', { number: orderNumber(chat.order_id) })"
                >
                    <div class="flex items-center justify-between gap-2">
                        <OrderStatusTag
                            v-if="chat.order_status"
                            :status="chat.order_status"
                        />
                        <span
                            v-if="chat.order_total !== null"
                            class="font-semibold text-ink"
                        >{{ price(chat.order_total) }}</span>
                    </div>
                    <el-button
                        class="mt-4 w-full"
                        @click="router.push(`/admin/orders/${chat.order_id}`)"
                    >
                        {{ t('admin.chats.open') }}
                    </el-button>
                </AdminCard>

                <AdminCard
                    v-if="chat.leads.length"
                    :title="t('admin.chats.relatedLeads')"
                    :padded="false"
                >
                    <ul class="divide-y divide-line">
                        <li
                            v-for="lead in chat.leads"
                            :key="lead.id"
                            class="space-y-1 px-5 py-3 text-sm"
                        >
                            <div class="flex items-center justify-between gap-2">
                                <span class="font-medium text-ink">{{ lead.full_name }}</span>
                                <el-tag
                                    size="small"
                                    effect="plain"
                                >
                                    {{ leadLabel(lead.source) }}
                                </el-tag>
                            </div>
                            <a
                                :href="telHref(lead.phone)"
                                class="text-brand hover:underline"
                            >{{ formatPhone(lead.phone) }}</a>
                            <p
                                v-if="lead.message"
                                class="text-xs leading-5 whitespace-pre-line text-muted"
                            >
                                {{ lead.message }}
                            </p>
                        </li>
                    </ul>
                </AdminCard>
            </div>
        </div>
    </div>
</template>
