<!-- CategoriesView.vue -->
<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { adminApi } from '../../api/admin'
import { ApiError } from '../../api/http'
import type { Category, CategoryInput } from '../../api/types'
import { useLocalized } from '../../utils/useLocalized'
import AdminCard from '../components/AdminCard.vue'

const { t, locale } = useI18n()
const { tr } = useLocalized()

const categories = ref<Category[]>([])
const loading = ref(false)
const dialogOpen = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)
const slugTouched = ref(false)
const formRef = ref<FormInstance>()

const emptyForm = (): CategoryInput => ({
    slug: '',
    name_uz: '',
    name_ru: '',
    name_en: '',
    description_uz: '',
    description_ru: '',
    description_en: '',
    sort_order: 0,
})
const form = reactive<CategoryInput>(emptyForm())

const rules = computed(() => {
    void locale.value
    return {
        name_uz: [{ required: true, message: t('admin.categories.nameRequired'), trigger: 'blur' }],
        slug: [
            { required: true, message: t('admin.categories.slugRequired'), trigger: 'blur' },
            { pattern: /^[a-z0-9]+(?:-[a-z0-9]+)*$/, message: t('admin.categories.slugInvalid'), trigger: 'blur' },
        ],
    }
})

function slugify(value: string): string {
    return value
        .toLowerCase()
        .replace(/[‘’ʻʼ'`]/g, '')
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-+|-+$/g, '')
        .slice(0, 80)
}

function onNameInput(): void {
    if (!slugTouched.value && editingId.value === null) {
        form.slug = slugify(form.name_uz)
    }
}

async function load(): Promise<void> {
    loading.value = true
    try {
        categories.value = await adminApi.categories()
    } finally {
        loading.value = false
    }
}

onMounted(load)

function openCreate(): void {
    editingId.value = null
    slugTouched.value = false
    Object.assign(form, emptyForm(), { sort_order: categories.value.length + 1 })
    dialogOpen.value = true
    formRef.value?.clearValidate()
}

function openEdit(category: Category): void {
    editingId.value = category.id
    slugTouched.value = true
    Object.assign(form, {
        slug: category.slug,
        name_uz: category.name_uz,
        name_ru: category.name_ru,
        name_en: category.name_en,
        description_uz: category.description_uz,
        description_ru: category.description_ru,
        description_en: category.description_en,
        sort_order: category.sort_order,
    })
    dialogOpen.value = true
    formRef.value?.clearValidate()
}

async function save(): Promise<void> {
    if (!(await formRef.value?.validate().catch(() => false))) return
    saving.value = true
    try {
        if (editingId.value === null) await adminApi.createCategory({ ...form })
        else await adminApi.updateCategory(editingId.value, { ...form })
        ElMessage.success(t('common.saved'))
        dialogOpen.value = false
        load()
    } catch (err) {
        ElMessage.error(err instanceof ApiError && err.status === 409 ? t('admin.categories.slugTaken') : t('common.saveError'))
    } finally {
        saving.value = false
    }
}

async function remove(category: Category): Promise<void> {
    try {
        await ElMessageBox.confirm(
            t('admin.categories.deleteConfirm', { name: tr(category), n: category.product_count }),
            t('common.confirm'),
            {
                type: 'warning',
                confirmButtonText: t('common.delete'),
                cancelButtonText: t('common.cancel'),
                confirmButtonClass: 'el-button--danger',
            }
        )
    } catch {
        return
    }
    try {
        await adminApi.deleteCategory(category.id)
        ElMessage.success(t('common.deleted'))
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
                <p class="text-sm text-muted">{{ t('admin.categories.hint') }}</p>
                <el-button
                    type="primary"
                    size="default"
                    :icon="Plus"
                    @click="openCreate"
                >
                    {{ t('admin.categories.new') }}
                </el-button>
            </div>
        </template>

        <el-table
            v-loading="loading"
            :data="categories"
            :empty-text="t('admin.categories.empty')"
        >
            <el-table-column
                :label="t('admin.categories.name')"
                min-width="240"
            >
                <template #default="{ row }">
                    <p class="font-medium text-ink">{{ tr(row) }}</p>
                    <p class="text-xs text-muted">{{ tr(row, 'description') }}</p>
                </template>
            </el-table-column>
            <el-table-column
                prop="slug"
                label="Slug"
                min-width="140"
            />
            <el-table-column
                :label="t('admin.nav.products')"
                width="120"
                align="center"
                prop="product_count"
            />
            <el-table-column
                :label="t('admin.products.sortOrder')"
                width="120"
                align="center"
                prop="sort_order"
            />
            <el-table-column
                width="210"
                align="right"
            >
                <template #default="{ row }">
                    <el-button
                        size="small"
                        @click="openEdit(row)"
                    >
                        {{ t('common.edit') }}
                    </el-button>
                    <el-button
                        size="small"
                        type="danger"
                        plain
                        @click="remove(row)"
                    >
                        {{ t('common.delete') }}
                    </el-button>
                </template>
            </el-table-column>
        </el-table>

        <el-dialog
            v-model="dialogOpen"
            :title="editingId === null ? t('admin.categories.new') : t('admin.categories.edit')"
            width="min(640px, calc(100vw - 32px))"
        >
            <el-form
                ref="formRef"
                :model="form"
                :rules="rules"
                label-position="top"
                @submit.prevent="save"
            >
                <div class="grid gap-x-4 sm:grid-cols-3">
                    <el-form-item
                        :label="`${t('admin.categories.name')} (UZ)`"
                        prop="name_uz"
                    >
                        <el-input
                            v-model="form.name_uz"
                            @input="onNameInput"
                        />
                    </el-form-item>
                    <el-form-item :label="`${t('admin.categories.name')} (RU)`">
                        <el-input v-model="form.name_ru" />
                    </el-form-item>
                    <el-form-item :label="`${t('admin.categories.name')} (EN)`">
                        <el-input v-model="form.name_en" />
                    </el-form-item>
                    <el-form-item :label="`${t('admin.categories.description')} (UZ)`">
                        <el-input
                            v-model="form.description_uz"
                            type="textarea"
                            :rows="2"
                        />
                    </el-form-item>
                    <el-form-item :label="`${t('admin.categories.description')} (RU)`">
                        <el-input
                            v-model="form.description_ru"
                            type="textarea"
                            :rows="2"
                        />
                    </el-form-item>
                    <el-form-item :label="`${t('admin.categories.description')} (EN)`">
                        <el-input
                            v-model="form.description_en"
                            type="textarea"
                            :rows="2"
                        />
                    </el-form-item>
                </div>
                <div class="grid gap-x-4 sm:grid-cols-2">
                    <el-form-item
                        label="Slug (URL)"
                        prop="slug"
                    >
                        <el-input
                            v-model="form.slug"
                            @input="slugTouched = true"
                        />
                    </el-form-item>
                    <el-form-item :label="t('admin.products.sortOrder')">
                        <el-input-number
                            v-model="form.sort_order"
                            :min="0"
                            :max="9999"
                            class="w-full!"
                        />
                    </el-form-item>
                </div>
                <button
                    type="submit"
                    class="hidden"
                />
            </el-form>
            <template #footer>
                <el-button @click="dialogOpen = false">{{ t('common.cancel') }}</el-button>
                <el-button
                    type="primary"
                    :loading="saving"
                    @click="save"
                >
                    {{ t('common.save') }}
                </el-button>
            </template>
        </el-dialog>
    </AdminCard>
</template>
