<!-- LoginView.vue -->
<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance } from 'element-plus'
import { ApiError } from '../../api/http'
import LanguageSwitcher from '../../components/LanguageSwitcher.vue'
import { useAuthStore } from '../../stores/auth'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

const rules = computed(() => {
    void locale.value
    return {
        username: [{ required: true, message: t('admin.login.usernameRequired'), trigger: 'blur' }],
        password: [{ required: true, message: t('admin.login.passwordRequired'), trigger: 'blur' }],
    }
})

async function submit(): Promise<void> {
    if (!(await formRef.value?.validate().catch(() => false))) return
    loading.value = true
    try {
        await auth.login(form.username.trim(), form.password)
        const redirect = typeof route.query.redirect === 'string' && route.query.redirect.startsWith('/admin')
            ? route.query.redirect
            : '/admin'
        router.replace(redirect)
    } catch (err) {
        if (err instanceof ApiError && err.status === 401) ElMessage.error(t('admin.login.invalid'))
        else if (err instanceof ApiError && err.status === 429) ElMessage.error(t('form.tooMany'))
        else ElMessage.error(t('common.loadError'))
    } finally {
        loading.value = false
    }
}
</script>

<template>
    <div class="flex min-h-screen flex-col bg-surface">
        <div class="flex justify-end p-4 sm:p-6">
            <LanguageSwitcher class="bg-white" />
        </div>

        <div class="flex flex-1 items-center justify-center px-4 pb-20">
            <div class="w-full max-w-md rounded-2xl border border-line bg-white p-6 sm:p-10">
                <RouterLink
                    to="/"
                    class="inline-flex items-center gap-2.5 text-ink"
                >
                    <span
                        class="flex size-10 items-center justify-center rounded-xl bg-brand text-gold"
                        aria-hidden="true"
                    >
                        <svg class="size-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round">
                            <path d="M9.5 3h5v3.5h-5zM10.5 6.5v2M13.5 6.5v2M7 9h10a1.5 1.5 0 0 1 1.5 1.5v9A1.5 1.5 0 0 1 17 21H7a1.5 1.5 0 0 1-1.5-1.5v-9A1.5 1.5 0 0 1 7 9ZM9 15h6" />
                        </svg>
                    </span>
                    <span class="text-2xl font-semibold tracking-tight">Parfume<span class="text-gold"> Shop</span></span>
                </RouterLink>

                <h1 class="mt-8 text-2xl font-medium tracking-tight text-ink">{{ t('admin.login.title') }}</h1>
                <p class="mt-2 text-sm text-muted">{{ t('admin.login.text') }}</p>

                <el-form
                    ref="formRef"
                    :model="form"
                    :rules="rules"
                    label-position="top"
                    class="mt-8"
                    @submit.prevent="submit"
                >
                    <el-form-item
                        :label="t('admin.login.username')"
                        prop="username"
                    >
                        <el-input
                            v-model="form.username"
                            autocomplete="username"
                            autofocus
                        />
                    </el-form-item>
                    <el-form-item
                        :label="t('admin.login.password')"
                        prop="password"
                    >
                        <el-input
                            v-model="form.password"
                            type="password"
                            autocomplete="current-password"
                            show-password
                        />
                    </el-form-item>
                    <el-button
                        type="primary"
                        native-type="submit"
                        :loading="loading"
                        class="mt-2! h-12! w-full! rounded-lg! text-sm! font-medium!"
                    >
                        {{ t('admin.login.submit') }}
                    </el-button>
                </el-form>
            </div>
        </div>
    </div>
</template>
