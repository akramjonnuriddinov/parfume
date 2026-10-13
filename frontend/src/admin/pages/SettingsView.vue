<!-- SettingsView.vue -->
<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, type FormInstance } from 'element-plus'
import { adminApi } from '../../api/admin'
import { ApiError } from '../../api/http'
import type { Settings } from '../../api/types'
import { useSettingsStore } from '../../stores/settings'
import AdminCard from '../components/AdminCard.vue'

const { t, locale } = useI18n()
const settingsStore = useSettingsStore()

const loading = ref(true)
const saving = ref(false)
const form = reactive<Settings>({})

const passwordRef = ref<FormInstance>()
const changingPassword = ref(false)
const passwords = reactive({ current: '', next: '', confirm: '' })

const passwordRules = computed(() => {
    void locale.value
    return {
        current: [{ required: true, message: t('admin.settings.passwordRequired'), trigger: 'blur' }],
        next: [
            { required: true, message: t('admin.settings.passwordRequired'), trigger: 'blur' },
            { min: 6, max: 72, message: t('admin.settings.passwordLength'), trigger: 'blur' },
        ],
        confirm: [
            {
                validator: (_r: unknown, value: string, cb: (e?: Error) => void) =>
                    value !== passwords.next ? cb(new Error(t('admin.settings.passwordMismatch'))) : cb(),
                trigger: 'blur',
            },
        ],
    }
})

onMounted(async () => {
    try {
        Object.assign(form, await adminApi.settings())
    } catch {
        ElMessage.error(t('common.loadError'))
    } finally {
        loading.value = false
    }
})

async function save(): Promise<void> {
    saving.value = true
    try {
        const saved = await adminApi.saveSettings({ ...form })
        Object.assign(form, saved)
        settingsStore.set(saved)
        ElMessage.success(t('common.saved'))
    } catch {
        ElMessage.error(t('common.saveError'))
    } finally {
        saving.value = false
    }
}

async function changePassword(): Promise<void> {
    if (!(await passwordRef.value?.validate().catch(() => false))) return
    changingPassword.value = true
    try {
        await adminApi.changePassword(passwords.current, passwords.next)
        ElMessage.success(t('admin.settings.passwordChanged'))
        passwords.current = passwords.next = passwords.confirm = ''
        passwordRef.value?.resetFields()
    } catch (err) {
        ElMessage.error(
            err instanceof ApiError && err.status === 422 ? t('admin.settings.passwordWrong') : t('common.saveError')
        )
    } finally {
        changingPassword.value = false
    }
}
</script>

<template>
    <div
        v-loading="loading"
        class="grid gap-6 xl:grid-cols-[1fr_380px]"
    >
        <AdminCard :title="t('admin.settings.contacts')">
            <el-form
                label-position="top"
                @submit.prevent="save"
            >
                <div class="grid gap-x-5 sm:grid-cols-2">
                    <el-form-item :label="t('admin.settings.phone')">
                        <el-input
                            v-model="form.phone"
                            placeholder="+998 90 123 45 67"
                        />
                    </el-form-item>
                    <el-form-item :label="t('admin.settings.phoneSecondary')">
                        <el-input v-model="form.phone_secondary" />
                    </el-form-item>
                    <el-form-item label="Email">
                        <el-input
                            v-model="form.email"
                            type="email"
                        />
                    </el-form-item>
                    <el-form-item label="Telegram URL">
                        <el-input
                            v-model="form.telegram_url"
                            placeholder="https://t.me/..."
                        />
                    </el-form-item>
                    <el-form-item label="Instagram URL">
                        <el-input
                            v-model="form.instagram_url"
                            placeholder="https://instagram.com/..."
                        />
                    </el-form-item>
                </div>

                <h3 class="mt-2 mb-4 border-t border-line pt-5 text-sm font-medium text-ink">
                    {{ t('admin.settings.address') }}
                </h3>
                <div class="grid gap-x-5 sm:grid-cols-3">
                    <el-form-item
                        v-for="lang in (['uz', 'ru', 'en'] as const)"
                        :key="lang"
                        :label="t(`admin.lang.${lang}`)"
                    >
                        <el-input
                            v-model="form[`address_${lang}`]"
                            type="textarea"
                            :rows="2"
                        />
                    </el-form-item>
                </div>

                <h3 class="mt-2 mb-4 border-t border-line pt-5 text-sm font-medium text-ink">
                    {{ t('admin.settings.workHours') }}
                </h3>
                <div class="grid gap-x-5 sm:grid-cols-3">
                    <el-form-item
                        v-for="lang in (['uz', 'ru', 'en'] as const)"
                        :key="lang"
                        :label="t(`admin.lang.${lang}`)"
                    >
                        <el-input v-model="form[`work_hours_${lang}`]" />
                    </el-form-item>
                </div>

                <el-button
                    type="primary"
                    native-type="submit"
                    :loading="saving"
                >
                    {{ t('common.save') }}
                </el-button>
            </el-form>
        </AdminCard>

        <AdminCard :title="t('admin.settings.password')">
            <el-form
                ref="passwordRef"
                :model="passwords"
                :rules="passwordRules"
                label-position="top"
                @submit.prevent="changePassword"
            >
                <el-form-item
                    :label="t('admin.settings.currentPassword')"
                    prop="current"
                >
                    <el-input
                        v-model="passwords.current"
                        type="password"
                        show-password
                        autocomplete="current-password"
                    />
                </el-form-item>
                <el-form-item
                    :label="t('admin.settings.newPassword')"
                    prop="next"
                >
                    <el-input
                        v-model="passwords.next"
                        type="password"
                        show-password
                        autocomplete="new-password"
                    />
                </el-form-item>
                <el-form-item
                    :label="t('admin.settings.confirmPassword')"
                    prop="confirm"
                >
                    <el-input
                        v-model="passwords.confirm"
                        type="password"
                        show-password
                        autocomplete="new-password"
                    />
                </el-form-item>
                <el-button
                    native-type="submit"
                    :loading="changingPassword"
                    class="w-full!"
                >
                    {{ t('admin.settings.changePassword') }}
                </el-button>
            </el-form>
        </AdminCard>
    </div>
</template>
