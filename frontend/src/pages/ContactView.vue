<!-- ContactView.vue -->
<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, type FormInstance } from 'element-plus'
import { ApiError } from '../api/http'
import { publicApi } from '../api/public'
import PageIntro from '../components/PageIntro.vue'
import PhoneInput from '../components/PhoneInput.vue'
import { DEFAULT_PHONE, useSettingsStore } from '../stores/settings'
import { telHref } from '../utils/format'
import { nameRules, phoneRules, toFullPhone } from '../utils/rules'
import { useLocalized } from '../utils/useLocalized'

const { t, locale } = useI18n()
const { setting } = useLocalized()
const settingsStore = useSettingsStore()

const s = computed(() => settingsStore.settings)
const phone = computed(() => s.value.phone || DEFAULT_PHONE)

const formRef = ref<FormInstance>()
const submitting = ref(false)
const sent = ref(false)
const form = reactive({ full_name: '', phone: '', message: '' })

const rules = computed(() => {
    void locale.value
    return {
        full_name: nameRules(),
        phone: phoneRules(),
        message: [{ required: true, message: t('form.messageRequired'), trigger: 'blur' }],
    }
})

async function submit(): Promise<void> {
    if (!(await formRef.value?.validate().catch(() => false))) return
    submitting.value = true
    try {
        await publicApi.createMessage({
            full_name: form.full_name.trim(),
            phone: toFullPhone(form.phone),
            message: form.message.trim(),
        })
        sent.value = true
        form.message = ''
    } catch (err) {
        ElMessage.error(err instanceof ApiError && err.status === 429 ? t('form.tooMany') : t('form.sendError'))
    } finally {
        submitting.value = false
    }
}
</script>

<template>
    <div>
        <PageIntro
            :title="t('contact.title')"
            :text="t('contact.lead')"
        />

        <section class="mx-auto grid max-w-[1240px] gap-8 px-4 py-10 sm:px-6 lg:grid-cols-[1fr_1.1fr] lg:gap-14 lg:py-14">
            <!-- Kontaktlar -->
            <div class="grid content-start gap-4 sm:grid-cols-2">
                <a
                    :href="telHref(phone)"
                    class="group rounded-2xl bg-hero p-6 transition-colors hover:bg-brand-soft"
                >
                    <p class="text-xs text-muted">{{ t('contact.phone') }}</p>
                    <p class="mt-2 text-lg font-medium text-ink group-hover:text-brand">{{ phone }}</p>
                    <p
                        v-if="s.phone_secondary"
                        class="mt-1 text-sm text-ink"
                    >
                        {{ s.phone_secondary }}
                    </p>
                </a>
                <a
                    v-if="s.telegram_url"
                    :href="s.telegram_url"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="group rounded-2xl bg-hero p-6 transition-colors hover:bg-brand-soft"
                >
                    <p class="text-xs text-muted">Telegram</p>
                    <p class="mt-2 text-lg font-medium text-ink group-hover:text-brand">{{ t('footer.telegram') }}</p>
                </a>
                <a
                    v-if="s.email"
                    :href="`mailto:${s.email}`"
                    class="group rounded-2xl bg-hero p-6 transition-colors hover:bg-brand-soft"
                >
                    <p class="text-xs text-muted">{{ t('contact.email') }}</p>
                    <p class="mt-2 text-lg font-medium break-all text-ink group-hover:text-brand">{{ s.email }}</p>
                </a>
                <a
                    v-if="s.instagram_url"
                    :href="s.instagram_url"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="group rounded-2xl bg-hero p-6 transition-colors hover:bg-brand-soft"
                >
                    <p class="text-xs text-muted">Instagram</p>
                    <p class="mt-2 text-lg font-medium text-ink group-hover:text-brand">{{ t('contact.instagram') }}</p>
                </a>
                <div
                    v-if="setting(s, 'work_hours')"
                    class="rounded-2xl bg-hero p-6"
                >
                    <p class="text-xs text-muted">{{ t('contact.hours') }}</p>
                    <p class="mt-2 text-base font-medium text-ink">{{ setting(s, 'work_hours') }}</p>
                </div>
                <div
                    v-if="setting(s, 'address')"
                    class="rounded-2xl bg-hero p-6 sm:col-span-2"
                >
                    <p class="text-xs text-muted">{{ t('contact.address') }}</p>
                    <p class="mt-2 text-base font-medium text-ink">{{ setting(s, 'address') }}</p>
                </div>
            </div>

            <!-- Forma -->
            <div class="rounded-2xl border border-line p-6 sm:p-8">
                <h2 class="text-3xl font-medium tracking-tight text-ink">{{ t('contact.formTitle') }}</h2>
                <p class="mt-2 text-sm text-muted">{{ t('contact.formText') }}</p>

                <el-alert
                    v-if="sent"
                    type="success"
                    :title="t('inquiry.sentTitle')"
                    :description="t('inquiry.sentText')"
                    show-icon
                    class="mt-6!"
                    @close="sent = false"
                />

                <el-form
                    ref="formRef"
                    :model="form"
                    :rules="rules"
                    label-position="top"
                    class="mt-6"
                    @submit.prevent="submit"
                >
                    <div class="grid gap-x-4 sm:grid-cols-2">
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
                    </div>
                    <el-form-item
                        :label="t('form.message')"
                        prop="message"
                    >
                        <el-input
                            v-model="form.message"
                            type="textarea"
                            :rows="5"
                            maxlength="2000"
                            show-word-limit
                            :placeholder="t('form.messagePlaceholder')"
                        />
                    </el-form-item>
                    <el-button
                        type="primary"
                        native-type="submit"
                        :loading="submitting"
                        class="h-12! w-full! rounded-lg! px-8! text-sm! font-medium! sm:w-auto!"
                    >
                        {{ t('form.send') }}
                    </el-button>
                </el-form>
            </div>
        </section>
    </div>
</template>
