<!-- ProductFormView.vue -->
<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import { adminApi } from '../../api/admin'
import { publicApi } from '../../api/public'
import { CONCENTRATIONS, GENDERS, type Category, type ProductInput, type VariantInput } from '../../api/types'
import { useLocalized } from '../../utils/useLocalized'
import { usePerfume } from '../../utils/usePerfume'
import AdminCard from '../components/AdminCard.vue'
import ImageUpload from '../components/ImageUpload.vue'

const { t, locale } = useI18n()
const { tr } = useLocalized()
const perfume = usePerfume()
const route = useRoute()
const router = useRouter()

const id = route.params.id ? Number(route.params.id) : null
const isEdit = id !== null

const formRef = ref<FormInstance>()
const loading = ref(isEdit)
const saving = ref(false)
const categories = ref<Category[]>([])
const brands = ref<string[]>([])
const langTab = ref<'uz' | 'ru' | 'en'>('uz')
const variantsError = ref('')

const form = reactive<ProductInput>({
    category_id: null,
    name_uz: '',
    name_ru: '',
    name_en: '',
    description_uz: '',
    description_ru: '',
    description_en: '',
    brand: '',
    gender: 'unisex',
    concentration: 'edp',
    notes_top: '',
    notes_heart: '',
    notes_base: '',
    image_url: '',
    is_active: true,
    is_featured: false,
    sort_order: 0,
    variants: [
        { volume_ml: 10, kind: 'decant', price: 0, in_stock: true, sort_order: 0 },
        { volume_ml: 100, kind: 'bottle', price: 0, in_stock: true, sort_order: 1 },
    ],
})

const rules = computed(() => {
    void locale.value
    return {
        name_uz: [{ required: true, message: t('admin.products.nameRequired'), trigger: 'blur' }],
    }
})

onMounted(async () => {
    categories.value = await adminApi.categories().catch(() => [])
    publicApi
        .brands()
        .then((list) => (brands.value = list.map((b) => b.name)))
        .catch(() => undefined)
    if (!isEdit) return
    try {
        const p = await adminApi.product(id)
        Object.assign(form, {
            category_id: p.category_id,
            name_uz: p.name_uz,
            name_ru: p.name_ru,
            name_en: p.name_en,
            description_uz: p.description_uz,
            description_ru: p.description_ru,
            description_en: p.description_en,
            brand: p.brand,
            gender: p.gender,
            concentration: p.concentration,
            notes_top: p.notes_top,
            notes_heart: p.notes_heart,
            notes_base: p.notes_base,
            image_url: p.image_url,
            is_active: p.is_active,
            is_featured: p.is_featured,
            sort_order: p.sort_order,
            variants: p.variants.map((v) => ({ ...v })),
        })
    } catch {
        ElMessage.error(t('common.loadError'))
        router.replace('/admin/products')
    } finally {
        loading.value = false
    }
})

function queryBrands(q: string, cb: (items: { value: string }[]) => void): void {
    const needle = q.trim().toLowerCase()
    cb(brands.value.filter((b) => b.toLowerCase().includes(needle)).map((value) => ({ value })))
}

function addVariant(v: Partial<VariantInput>): void {
    form.variants.push({ volume_ml: 50, kind: 'bottle', price: 0, in_stock: true, sort_order: form.variants.length, ...v })
    variantsError.value = ''
}

function removeVariant(index: number): void {
    form.variants.splice(index, 1)
}

function validateVariants(): boolean {
    if (form.variants.length === 0) {
        variantsError.value = t('admin.products.variantsRequired')
        return false
    }
    if (form.variants.some((v) => !v.volume_ml || !v.price)) {
        variantsError.value = t('admin.products.variantsInvalid')
        return false
    }
    variantsError.value = ''
    return true
}

async function submit(): Promise<void> {
    const valid = await formRef.value?.validate().catch(() => false)
    if (!valid) {
        if (!form.name_uz.trim()) langTab.value = 'uz'
        return
    }
    if (!validateVariants()) return
    saving.value = true
    const payload: ProductInput = {
        ...form,
        category_id: form.category_id || null,
        variants: form.variants.map((v, i) => ({ ...v, sort_order: i })),
    }
    try {
        if (isEdit) {
            await adminApi.updateProduct(id, payload)
        } else {
            await adminApi.createProduct(payload)
        }
        ElMessage.success(t('common.saved'))
        router.push('/admin/products')
    } catch {
        ElMessage.error(t('common.saveError'))
    } finally {
        saving.value = false
    }
}
</script>

<template>
    <div v-loading="loading">
        <RouterLink
            to="/admin/products"
            class="mb-4 inline-flex items-center gap-1 text-sm text-muted hover:text-brand"
        >
            ← {{ t('admin.products.back') }}
        </RouterLink>

        <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            label-position="top"
            class="grid gap-6 xl:grid-cols-[1fr_360px]"
            @submit.prevent="submit"
        >
            <div class="space-y-6">
                <AdminCard :title="t('admin.products.details')">
                    <div class="grid gap-x-5 sm:grid-cols-2">
                        <el-form-item :label="t('admin.products.brand')">
                            <el-autocomplete
                                v-model="form.brand"
                                :fetch-suggestions="queryBrands"
                                :placeholder="t('admin.products.brandPlaceholder')"
                                maxlength="120"
                                class="w-full!"
                            />
                        </el-form-item>
                        <el-form-item :label="t('admin.products.category')">
                            <el-select
                                v-model="form.category_id"
                                clearable
                                :placeholder="t('admin.products.noCategory')"
                            >
                                <el-option
                                    v-for="c in categories"
                                    :key="c.id"
                                    :value="c.id"
                                    :label="tr(c)"
                                />
                            </el-select>
                        </el-form-item>
                        <el-form-item :label="t('admin.products.gender')">
                            <el-radio-group v-model="form.gender">
                                <el-radio-button
                                    v-for="g in GENDERS"
                                    :key="g"
                                    :value="g"
                                >
                                    {{ perfume.gender(g) }}
                                </el-radio-button>
                            </el-radio-group>
                        </el-form-item>
                        <el-form-item :label="t('admin.products.concentration')">
                            <el-select v-model="form.concentration">
                                <el-option
                                    v-for="c in CONCENTRATIONS"
                                    :key="c"
                                    :value="c"
                                    :label="perfume.concentration(c)"
                                />
                            </el-select>
                        </el-form-item>
                    </div>
                </AdminCard>

                <AdminCard :title="t('admin.products.content')">
                    <el-tabs v-model="langTab">
                        <el-tab-pane
                            v-for="lang in (['uz', 'ru', 'en'] as const)"
                            :key="lang"
                            :name="lang"
                            :label="t(`admin.lang.${lang}`)"
                        >
                            <el-form-item
                                :label="t('admin.products.name')"
                                :prop="`name_${lang}`"
                                :required="lang === 'uz'"
                            >
                                <el-input
                                    v-model="form[`name_${lang}`]"
                                    maxlength="200"
                                    :placeholder="lang === 'uz' ? t('admin.products.namePlaceholder') : ''"
                                />
                            </el-form-item>
                            <el-form-item :label="t('admin.products.description')">
                                <el-input
                                    v-model="form[`description_${lang}`]"
                                    type="textarea"
                                    :rows="5"
                                />
                            </el-form-item>
                        </el-tab-pane>
                    </el-tabs>
                    <p class="text-xs text-muted">{{ t('admin.products.langHint') }}</p>
                </AdminCard>

                <AdminCard :title="t('admin.products.variants')">
                    <template #actions>
                        <div class="flex flex-wrap gap-2">
                            <el-button
                                size="small"
                                :icon="Plus"
                                @click="addVariant({ volume_ml: 5, kind: 'decant' })"
                            >
                                5 ml {{ t('perfume.decant') }}
                            </el-button>
                            <el-button
                                size="small"
                                :icon="Plus"
                                @click="addVariant({ volume_ml: 10, kind: 'decant' })"
                            >
                                10 ml {{ t('perfume.decant') }}
                            </el-button>
                            <el-button
                                size="small"
                                :icon="Plus"
                                @click="addVariant({ volume_ml: 100, kind: 'bottle' })"
                            >
                                {{ t('perfume.bottle') }}
                            </el-button>
                        </div>
                    </template>
                    <p class="-mt-1 mb-4 text-xs text-muted">{{ t('admin.products.variantsHint') }}</p>
                    <div class="space-y-3">
                        <div
                            v-for="(v, i) in form.variants"
                            :key="v.id ?? `new-${i}`"
                            class="grid grid-cols-2 items-end gap-3 rounded-xl border border-line p-3 sm:grid-cols-[110px_150px_1fr_auto_auto]"
                        >
                            <label class="block">
                                <span class="mb-1 block text-xs text-muted">{{ t('admin.products.volume') }}</span>
                                <el-input-number
                                    v-model="v.volume_ml"
                                    :min="1"
                                    :max="1000"
                                    :controls="false"
                                    class="w-full!"
                                />
                            </label>
                            <label class="block">
                                <span class="mb-1 block text-xs text-muted">{{ t('admin.products.kind') }}</span>
                                <el-select v-model="v.kind">
                                    <el-option
                                        value="bottle"
                                        :label="t('perfume.bottle')"
                                    />
                                    <el-option
                                        value="decant"
                                        :label="t('perfume.decant')"
                                    />
                                </el-select>
                            </label>
                            <label class="col-span-2 block sm:col-span-1">
                                <span class="mb-1 block text-xs text-muted">{{ t('admin.products.priceSum') }}</span>
                                <el-input-number
                                    v-model="v.price"
                                    :min="0"
                                    :step="1000"
                                    :controls="false"
                                    class="w-full!"
                                />
                            </label>
                            <label class="flex h-10 items-center gap-2 text-sm text-ink">
                                <el-switch v-model="v.in_stock" />
                                {{ t('admin.products.inStock') }}
                            </label>
                            <el-button
                                :icon="Delete"
                                type="danger"
                                plain
                                :aria-label="t('common.delete')"
                                @click="removeVariant(i)"
                            />
                        </div>
                    </div>
                    <p
                        v-if="variantsError"
                        class="mt-3 text-sm text-red-600"
                    >
                        {{ variantsError }}
                    </p>
                </AdminCard>

                <AdminCard :title="t('admin.products.notes')">
                    <p class="-mt-1 mb-4 text-xs text-muted">{{ t('admin.products.notesHint') }}</p>
                    <el-form-item :label="t('product.notesTop')">
                        <el-input
                            v-model="form.notes_top"
                            maxlength="500"
                            placeholder="Bergamot, limon, pushti qalampir"
                        />
                    </el-form-item>
                    <el-form-item :label="t('product.notesHeart')">
                        <el-input
                            v-model="form.notes_heart"
                            maxlength="500"
                            placeholder="Atirgul, yasmin"
                        />
                    </el-form-item>
                    <el-form-item :label="t('product.notesBase')">
                        <el-input
                            v-model="form.notes_base"
                            maxlength="500"
                            placeholder="Vanil, mushk, sandal"
                        />
                    </el-form-item>
                </AdminCard>
            </div>

            <div class="space-y-6">
                <AdminCard :title="t('admin.products.image')">
                    <ImageUpload v-model="form.image_url" />
                </AdminCard>

                <AdminCard :title="t('admin.products.visibility')">
                    <div class="space-y-4">
                        <label class="flex items-center justify-between gap-4">
                            <span>
                                <span class="block text-sm text-ink">{{ t('admin.products.active') }}</span>
                                <span class="block text-xs text-muted">{{ t('admin.products.activeHint') }}</span>
                            </span>
                            <el-switch v-model="form.is_active" />
                        </label>
                        <label class="flex items-center justify-between gap-4">
                            <span>
                                <span class="block text-sm text-ink">{{ t('admin.products.featured') }}</span>
                                <span class="block text-xs text-muted">{{ t('admin.products.featuredHint') }}</span>
                            </span>
                            <el-switch v-model="form.is_featured" />
                        </label>
                        <el-form-item
                            :label="t('admin.products.sortOrder')"
                            class="mb-0!"
                        >
                            <el-input-number
                                v-model="form.sort_order"
                                :min="0"
                                :max="9999"
                                class="w-full!"
                            />
                        </el-form-item>
                    </div>
                </AdminCard>

                <div class="flex gap-3">
                    <el-button
                        class="flex-1"
                        @click="router.push('/admin/products')"
                    >
                        {{ t('common.cancel') }}
                    </el-button>
                    <el-button
                        type="primary"
                        native-type="submit"
                        class="flex-1"
                        :loading="saving"
                    >
                        {{ t('common.save') }}
                    </el-button>
                </div>
            </div>
        </el-form>
    </div>
</template>
