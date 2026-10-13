<!-- CartView.vue -->
<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance } from 'element-plus'
import { ApiError } from '../api/http'
import { publicApi } from '../api/public'
import type { Locale } from '../api/types'
import PageIntro from '../components/PageIntro.vue'
import PhoneInput from '../components/PhoneInput.vue'
import ProductImage from '../components/ProductImage.vue'
import QuantityStepper from '../components/QuantityStepper.vue'
import { useCartStore } from '../stores/cart'
import { useChatStore } from '../stores/chat'
import { nameRules, phoneRules, toFullPhone } from '../utils/rules'
import { useLocalized } from '../utils/useLocalized'
import { usePerfume } from '../utils/usePerfume'

const { t, locale } = useI18n()
const { tr, price } = useLocalized()
const perfume = usePerfume()
const router = useRouter()
const cart = useCartStore()
const chat = useChatStore()

const formRef = ref<FormInstance>()
const submitting = ref(false)
const form = reactive({ full_name: '', phone: '', comment: '' })

const rules = computed(() => {
    void locale.value
    return { full_name: nameRules(), phone: phoneRules() }
})

async function syncCart(): Promise<void> {
    try {
        const removed = await cart.refresh()
        if (removed > 0) ElMessage.warning(t('cart.itemsRemoved'))
    } catch {
        // keep the local cart if the API is unreachable
    }
}

onMounted(syncCart)

async function submit(): Promise<void> {
    if (!(await formRef.value?.validate().catch(() => false))) return
    if (cart.items.length === 0) return

    submitting.value = true
    try {
        const order = await publicApi.createOrder({
            full_name: form.full_name.trim(),
            phone: toFullPhone(form.phone),
            comment: form.comment.trim(),
            locale: locale.value as Locale,
            chat_session_id: chat.sessionId || undefined,
            items: cart.items.map((item) => ({
                variant_id: item.variant_id,
                quantity: item.quantity,
            })),
        })
        cart.clear()
        router.replace({ name: 'order-success', params: { id: order.id } })
    } catch (err) {
        if (err instanceof ApiError && err.status === 422) {
            await syncCart()
            ElMessage.error(t('checkout.unavailable'))
        } else if (err instanceof ApiError && err.status === 429) {
            ElMessage.error(t('form.tooMany'))
        } else {
            ElMessage.error(t('form.sendError'))
        }
    } finally {
        submitting.value = false
    }
}
</script>

<template>
    <div>
        <PageIntro :title="t('cart.title')" />

        <div class="mx-auto max-w-[1240px] px-4 py-8 sm:px-6 lg:py-10">
            <div
                v-if="cart.items.length === 0"
                class="rounded-2xl bg-hero px-6 py-20 text-center"
            >
                <p class="text-xl font-medium text-ink">{{ t('cart.empty') }}</p>
                <p class="mt-2 text-sm text-muted">{{ t('cart.emptyHint') }}</p>
                <el-button
                    type="primary"
                    class="mt-6! h-12! rounded-lg! px-8!"
                    @click="router.push('/catalog')"
                >
                    {{ t('cart.goCatalog') }}
                </el-button>
            </div>

            <div
                v-else
                class="grid gap-8 lg:grid-cols-[1fr_400px] lg:gap-10"
            >
                <!-- Mahsulotlar ro‘yxati -->
                <section :aria-label="t('cart.title')">
                    <div class="flex items-center justify-between border-b border-line pb-4">
                        <p class="text-sm text-muted">{{ t('cart.itemsCount', { n: cart.count }) }}</p>
                        <button
                            type="button"
                            class="cursor-pointer text-sm text-muted hover:text-red-500"
                            @click="cart.clear()"
                        >
                            {{ t('cart.clear') }}
                        </button>
                    </div>

                    <ul class="divide-y divide-line">
                        <li
                            v-for="item in cart.items"
                            :key="item.variant_id"
                            class="flex gap-4 py-5"
                        >
                            <RouterLink
                                :to="`/products/${item.product_id}`"
                                class="size-20 shrink-0 overflow-hidden rounded-xl bg-surface p-2 sm:size-24"
                                :aria-label="tr(item)"
                            >
                                <ProductImage :src="item.image_url" />
                            </RouterLink>

                            <div class="flex min-w-0 flex-1 flex-col gap-3 sm:flex-row sm:items-center">
                                <div class="min-w-0 flex-1">
                                    <p class="truncate text-[11px] font-medium tracking-wider text-gold uppercase">{{ item.brand }}</p>
                                    <RouterLink
                                        :to="`/products/${item.product_id}`"
                                        class="line-clamp-2 text-sm font-medium text-ink hover:text-brand sm:text-base"
                                    >
                                        {{ tr(item) }}
                                    </RouterLink>
                                    <p class="mt-1 text-sm text-muted">{{ perfume.volume(item) }} · {{ price(item.price) }}</p>
                                </div>

                                <div class="flex items-center justify-between gap-4 sm:justify-end">
                                    <QuantityStepper
                                        size="sm"
                                        :model-value="item.quantity"
                                        @update:model-value="cart.setQuantity(item.variant_id, $event)"
                                        @remove="cart.remove(item.variant_id)"
                                    />
                                    <span class="w-32 text-right text-base font-semibold text-ink">
                                        {{ price(item.price * item.quantity) }}
                                    </span>
                                    <button
                                        type="button"
                                        :aria-label="t('cart.remove')"
                                        class="cursor-pointer rounded p-1 text-slate-400 hover:text-red-500"
                                        @click="cart.remove(item.variant_id)"
                                    >
                                        <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                            <path d="M4 7h16M10 11v6M14 11v6M5 7l1 13h12l1-13M9 7V4h6v3" />
                                        </svg>
                                    </button>
                                </div>
                            </div>
                        </li>
                    </ul>
                </section>

                <!-- Buyurtma formasi -->
                <aside class="lg:sticky lg:top-40 lg:self-start">
                    <div class="rounded-2xl border border-line bg-white p-5 sm:p-6">
                        <h2 class="text-xl font-medium text-ink">{{ t('checkout.title') }}</h2>
                        <p class="mt-1 text-sm text-muted">{{ t('checkout.text') }}</p>

                        <el-form
                            ref="formRef"
                            :model="form"
                            :rules="rules"
                            label-position="top"
                            class="mt-5"
                            @submit.prevent="submit"
                        >
                            <el-form-item
                                :label="t('form.fullName')"
                                prop="full_name"
                            >
                                <el-input
                                    v-model="form.full_name"
                                    autocomplete="name"
                                    maxlength="120"
                                    :placeholder="t('form.fullNamePlaceholder')"
                                />
                            </el-form-item>
                            <el-form-item
                                :label="t('form.phone')"
                                prop="phone"
                            >
                                <PhoneInput v-model="form.phone" />
                            </el-form-item>
                            <el-form-item
                                :label="t('checkout.comment')"
                                prop="comment"
                            >
                                <el-input
                                    v-model="form.comment"
                                    type="textarea"
                                    :rows="3"
                                    maxlength="1000"
                                    :placeholder="t('checkout.commentPlaceholder')"
                                />
                            </el-form-item>

                            <div class="space-y-2 border-t border-line pt-4 text-sm">
                                <div class="flex justify-between text-muted">
                                    <span>{{ t('cart.itemsCount', { n: cart.count }) }}</span>
                                    <span>{{ price(cart.total) }}</span>
                                </div>
                                <div class="flex justify-between text-muted">
                                    <span>{{ t('checkout.delivery') }}</span>
                                    <span>{{ t('checkout.deliveryValue') }}</span>
                                </div>
                                <div class="flex items-baseline justify-between pt-2">
                                    <span class="text-base font-medium text-ink">{{ t('cart.total') }}</span>
                                    <span class="text-2xl font-semibold text-ink">{{ price(cart.total) }}</span>
                                </div>
                            </div>

                            <el-button
                                type="primary"
                                native-type="submit"
                                :loading="submitting"
                                class="mt-5! h-12! w-full! rounded-lg! text-sm! font-medium!"
                            >
                                {{ t('checkout.submit') }}
                            </el-button>
                            <p class="mt-3 text-center text-xs leading-5 text-muted">
                                {{ t('checkout.note') }}
                            </p>
                        </el-form>
                    </div>
                </aside>
            </div>
        </div>
    </div>
</template>
