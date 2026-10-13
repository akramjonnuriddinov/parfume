<!-- ImageUpload.vue -->
<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { adminApi } from '../../api/admin'
import { ApiError } from '../../api/http'
import ProductImage from '../../components/ProductImage.vue'

defineProps<{ modelValue: string }>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const { t } = useI18n()
const uploading = ref(false)
const input = ref<HTMLInputElement>()
const dragging = ref(false)

async function upload(file: File | undefined): Promise<void> {
    if (!file) return
    if (!/^image\/(jpeg|png|webp|gif)$/.test(file.type)) {
        ElMessage.error(t('admin.upload.badType'))
        return
    }
    if (file.size > 5 * 1024 * 1024) {
        ElMessage.error(t('admin.upload.tooLarge'))
        return
    }
    uploading.value = true
    try {
        const { url } = await adminApi.upload(file)
        emit('update:modelValue', url)
    } catch (err) {
        ElMessage.error(err instanceof ApiError ? err.message : t('admin.upload.failed'))
    } finally {
        uploading.value = false
        if (input.value) input.value.value = ''
    }
}

function onDrop(event: DragEvent): void {
    dragging.value = false
    upload(event.dataTransfer?.files[0])
}
</script>

<template>
    <div class="w-full">
        <div
            v-loading="uploading"
            class="relative flex aspect-square w-full max-w-72 items-center justify-center overflow-hidden rounded-xl border-2 border-dashed bg-surface p-6 transition-colors"
            :class="dragging ? 'border-brand bg-brand-soft' : 'border-line'"
            @dragover.prevent="dragging = true"
            @dragleave.prevent="dragging = false"
            @drop.prevent="onDrop"
        >
            <ProductImage
                v-if="modelValue"
                :src="modelValue"
            />
            <div
                v-else
                class="text-center text-sm text-muted"
            >
                <svg class="mx-auto size-10 text-slate-300" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <rect x="3" y="3" width="18" height="18" rx="3" />
                    <circle cx="9" cy="9" r="2" />
                    <path d="m21 15-5-5L5 21" />
                </svg>
                <p class="mt-3">{{ t('admin.upload.hint') }}</p>
                <p class="mt-1 text-xs">JPG, PNG, WEBP · 5 MB</p>
            </div>
        </div>

        <input
            ref="input"
            type="file"
            accept="image/jpeg,image/png,image/webp,image/gif"
            class="hidden"
            @change="upload(($event.target as HTMLInputElement).files?.[0])"
        />

        <div class="mt-3 flex flex-wrap gap-2">
            <el-button
                :loading="uploading"
                @click="input?.click()"
            >
                {{ modelValue ? t('admin.upload.replace') : t('admin.upload.choose') }}
            </el-button>
            <el-button
                v-if="modelValue"
                type="danger"
                plain
                @click="emit('update:modelValue', '')"
            >
                {{ t('common.delete') }}
            </el-button>
        </div>

        <el-input
            :model-value="modelValue"
            class="mt-3"
            size="default"
            :placeholder="t('admin.upload.urlPlaceholder')"
            @update:model-value="emit('update:modelValue', $event)"
        />
    </div>
</template>
