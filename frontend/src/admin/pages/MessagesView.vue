<!-- MessagesView.vue -->
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../api/admin'
import type { Message } from '../../api/types'
import { useAdminCountersStore } from '../../stores/adminCounters'
import { formatPhone, telHref } from '../../utils/format'
import { useLocalized } from '../../utils/useLocalized'
import AdminCard from '../components/AdminCard.vue'

const PAGE_SIZE = 20

const { t } = useI18n()
const { date } = useLocalized()
const counters = useAdminCountersStore()

const messages = ref<Message[]>([])
const total = ref(0)
const page = ref(1)
const unreadOnly = ref(false)
const loading = ref(false)

async function load(): Promise<void> {
    loading.value = true
    try {
        const result = await adminApi.messages({ unread: unreadOnly.value || undefined, page: page.value, limit: PAGE_SIZE })
        messages.value = result.items
        total.value = result.total
    } catch {
        messages.value = []
    } finally {
        loading.value = false
    }
}

onMounted(load)

async function setRead(message: Message, value: boolean): Promise<void> {
    try {
        await adminApi.markMessage(message.id, value)
        message.is_read = value
        counters.refresh()
    } catch {
        ElMessage.error(t('common.saveError'))
    }
}

async function remove(message: Message): Promise<void> {
    try {
        await ElMessageBox.confirm(t('admin.messages.deleteConfirm'), t('common.confirm'), {
            type: 'warning',
            confirmButtonText: t('common.delete'),
            cancelButtonText: t('common.cancel'),
            confirmButtonClass: 'el-button--danger',
        })
    } catch {
        return
    }
    try {
        await adminApi.deleteMessage(message.id)
        ElMessage.success(t('common.deleted'))
        counters.refresh()
        load()
    } catch {
        ElMessage.error(t('common.saveError'))
    }
}
</script>

<template>
    <AdminCard :padded="false">
        <template #actions>
            <div class="flex w-full items-center justify-between gap-3">
                <p class="text-sm text-muted">{{ t('admin.messages.hint') }}</p>
                <el-checkbox
                    v-model="unreadOnly"
                    @change="() => { page = 1; load() }"
                >
                    {{ t('admin.messages.unreadOnly') }}
                </el-checkbox>
            </div>
        </template>

        <div v-loading="loading">
            <el-empty
                v-if="!loading && messages.length === 0"
                :description="t('admin.messages.empty')"
            />
            <ul class="divide-y divide-line">
                <li
                    v-for="message in messages"
                    :key="message.id"
                    class="flex flex-col gap-3 px-5 py-4 sm:flex-row sm:items-start"
                    :class="message.is_read ? '' : 'bg-brand-soft/50'"
                >
                    <div class="min-w-0 flex-1">
                        <div class="flex flex-wrap items-center gap-x-3 gap-y-1">
                            <span
                                v-if="!message.is_read"
                                class="size-2 rounded-full bg-brand"
                                :aria-label="t('admin.messages.unread')"
                            />
                            <span class="font-medium text-ink">{{ message.full_name }}</span>
                            <a
                                :href="telHref(message.phone)"
                                class="text-sm text-brand hover:underline"
                            >
                                {{ formatPhone(message.phone) }}
                            </a>
                            <span class="text-xs text-muted">{{ date(message.created_at) }}</span>
                            <el-tag
                                v-if="message.source !== 'site'"
                                size="small"
                                :type="message.source === 'ai_quote' ? 'warning' : 'primary'"
                                effect="plain"
                            >
                                {{ message.source === 'ai_quote' ? t('admin.messages.sourceQuote') : t('admin.messages.sourceCallback') }}
                            </el-tag>
                            <RouterLink
                                v-if="message.chat_session_id"
                                :to="`/admin/chats/${message.chat_session_id}`"
                                class="text-xs text-brand hover:underline"
                            >
                                {{ t('admin.messages.openChat') }}
                            </RouterLink>
                        </div>
                        <p
                            v-if="message.product_name"
                            class="mt-2 text-sm"
                        >
                            <span class="text-muted">{{ t('admin.messages.product') }}:</span>
                            <RouterLink
                                v-if="message.product_id"
                                :to="`/admin/products/${message.product_id}`"
                                class="ml-1 font-medium text-brand hover:underline"
                            >
                                {{ message.product_name }}
                            </RouterLink>
                            <span
                                v-else
                                class="ml-1"
                            >
                                {{ message.product_name }}
                            </span>
                        </p>
                        <p
                            v-if="message.message"
                            class="mt-2 text-sm leading-6 whitespace-pre-line text-ink"
                        >
                            {{ message.message }}
                        </p>
                        <p
                            v-else
                            class="mt-2 text-sm text-muted italic"
                        >
                            {{ t('admin.messages.callback') }}
                        </p>
                    </div>
                    <div class="flex shrink-0 gap-2">
                        <el-button
                            size="small"
                            @click="setRead(message, !message.is_read)"
                        >
                            {{ message.is_read ? t('admin.messages.markUnread') : t('admin.messages.markRead') }}
                        </el-button>
                        <el-button
                            size="small"
                            type="danger"
                            plain
                            @click="remove(message)"
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
                v-model:current-page="page"
                background
                layout="total, prev, pager, next"
                :total="total"
                :page-size="PAGE_SIZE"
                @current-change="load"
            />
        </div>
    </AdminCard>
</template>
