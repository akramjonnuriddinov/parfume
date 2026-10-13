<!-- InquiryDialog.vue -->
<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, type FormInstance } from 'element-plus'
import { publicApi } from '../api/public'
import { useInquiryStore } from '../stores/inquiry'
import { useLocalized } from '../utils/useLocalized'
import { nameRules, phoneRules, toFullPhone } from '../utils/rules'
import PhoneInput from './PhoneInput.vue'

const { t, locale } = useI18n()
const { tr } = useLocalized()
const inquiry = useInquiryStore()

const formRef = ref<FormInstance>()
const submitting = ref(false)
const sent = ref(false)
const form = reactive({ full_name: '', phone: '', message: '' })

const rules = computed(() => {
    void locale.value
    return { full_name: nameRules(), phone: phoneRules() }
})

watch(
    () => inquiry.visible,
    (open) => {
        if (open) {
            sent.value = false
            form.message = ''
            formRef.value?.clearValidate()
        }
    }
)

async function submit(): Promise<void> {
    if (!(await formRef.value?.validate().catch(() => false))) return
    submitting.value = true
    try {
        await publicApi.createMessage({
            full_name: form.full_name,
            phone: toFullPhone(form.phone),
            message: form.message,
            product_id: inquiry.product?.id ?? null,
        })
        sent.value = true
    } catch {
        ElMessage.error(t('form.sendError'))
    } finally {
        submitting.value = false
    }
}
</script>

<template>
    <el-dialog
        v-model="inquiry.visible"
        :title="inquiry.product ? t('inquiry.titleProduct') : t('inquiry.title')"
        width="min(460px, calc(100vw - 32px))"
        align-center
    >
        <div
            v-if="sent"
            class="py-6 text-center"
        >
            <span class="mx-auto flex size-14 items-center justify-center rounded-full bg-brand-soft text-brand">
                <svg class="size-7" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d="m5 12 5 5L20 7" />
                </svg>
            </span>
            <p class="mt-4 text-lg font-medium text-ink">{{ t('inquiry.sentTitle') }}</p>
            <p class="mt-2 text-sm text-muted">{{ t('inquiry.sentText') }}</p>
            <el-button
                type="primary"
                class="mt-6! h-11! rounded-lg! px-8!"
                @click="inquiry.visible = false"
            >
                {{ t('common.close') }}
            </el-button>
        </div>

        <template v-else>
            <p
                v-if="inquiry.product"
                class="mb-5 rounded-lg bg-brand-soft px-4 py-3 text-sm font-medium text-brand"
            >
                {{ tr(inquiry.product) }}
            </p>
            <p
                v-else
                class="mb-5 text-sm leading-6 text-muted"
            >
                {{ t('inquiry.text') }}
            </p>

            <el-form
                ref="formRef"
                :model="form"
                :rules="rules"
                label-position="top"
                @submit.prevent="submit"
            >
                <el-form-item
                    :label="t('form.fullName')"
                    prop="full_name"
                >
                    <el-input
                        v-model="form.full_name"
                        autocomplete="name"
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
                    :label="t('form.messageOptional')"
                    prop="message"
                >
                    <el-input
                        v-model="form.message"
                        type="textarea"
                        :rows="3"
                        maxlength="2000"
                        :placeholder="t('form.messagePlaceholder')"
                    />
                </el-form-item>
                <el-button
                    type="primary"
                    native-type="submit"
                    :loading="submitting"
                    class="h-12! w-full! rounded-lg! text-sm! font-medium!"
                >
                    {{ t('form.send') }}
                </el-button>
            </el-form>
        </template>
    </el-dialog>
</template>
