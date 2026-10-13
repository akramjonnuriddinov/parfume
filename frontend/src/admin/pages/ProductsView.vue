<!-- ProductsView.vue -->
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { adminApi } from '../../api/admin'
import type { Category, Product, Variant } from '../../api/types'
import ProductImage from '../../components/ProductImage.vue'
import { useLocalized } from '../../utils/useLocalized'
import AdminCard from '../components/AdminCard.vue'

const PAGE_SIZE = 20

const { t } = useI18n()
const { tr, price } = useLocalized()
const router = useRouter()

const products = ref<Product[]>([])
const categories = ref<Category[]>([])
const total = ref(0)
const loading = ref(false)
const filters = reactive({ q: '', category_id: '' as number | '', page: 1 })

async function load(): Promise<void> {
    loading.value = true
    try {
        const result = await adminApi.products({ ...filters, limit: PAGE_SIZE })
        products.value = result.items
        total.value = result.total
    } catch {
        products.value = []
    } finally {
        loading.value = false
    }
}

function applyFilters(): void {
    filters.page = 1
    load()
}

onMounted(async () => {
    load()
    categories.value = await adminApi.categories().catch(() => [])
})

async function toggle(product: Product, field: 'is_active' | 'is_featured', value: boolean): Promise<void> {
    const previous = product[field]
    product[field] = value
    try {
        const { id, category: _category, created_at: _c, updated_at: _u, min_price: _m, in_stock: _s, ...input } = product
        await adminApi.updateProduct(id, input)
    } catch {
        product[field] = previous
        ElMessage.error(t('common.saveError'))
    }
}

async function remove(product: Product): Promise<void> {
    try {
        await ElMessageBox.confirm(t('admin.products.deleteConfirm', { name: tr(product) }), t('common.confirm'), {
            type: 'warning',
            confirmButtonText: t('common.delete'),
            cancelButtonText: t('common.cancel'),
            confirmButtonClass: 'el-button--danger',
        })
    } catch {
        return
    }
    try {
        await adminApi.deleteProduct(product.id)
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
            <div class="flex w-full flex-col gap-3 md:flex-row md:items-center">
                <form
                    class="flex flex-1 flex-col gap-3 sm:flex-row"
                    @submit.prevent="applyFilters"
                >
                    <el-input
                        v-model="filters.q"
                        size="default"
                        clearable
                        class="sm:max-w-72"
                        :placeholder="t('admin.products.search')"
                        @clear="applyFilters"
                    />
                    <el-select
                        v-model="filters.category_id"
                        size="default"
                        clearable
                        class="sm:max-w-60"
                        :placeholder="t('admin.products.allCategories')"
                        @change="applyFilters"
                    >
                        <el-option
                            v-for="c in categories"
                            :key="c.id"
                            :value="c.id"
                            :label="tr(c)"
                        />
                    </el-select>
                </form>
                <el-button
                    type="primary"
                    size="default"
                    :icon="Plus"
                    @click="router.push('/admin/products/new')"
                >
                    {{ t('admin.products.new') }}
                </el-button>
            </div>
        </template>

        <el-table
            v-loading="loading"
            :data="products"
            :empty-text="t('admin.products.empty')"
        >
            <el-table-column
                :label="t('admin.products.product')"
                min-width="280"
            >
                <template #default="{ row }">
                    <RouterLink
                        :to="`/admin/products/${row.id}`"
                        class="flex items-center gap-3"
                    >
                        <span class="size-12 shrink-0 overflow-hidden rounded-lg bg-surface p-1">
                            <ProductImage :src="row.image_url" />
                        </span>
                        <span class="min-w-0">
                            <span class="line-clamp-2 font-medium text-ink hover:text-brand">{{ tr(row) }}</span>
                            <span class="text-xs text-muted">{{ row.brand }}</span>
                        </span>
                    </RouterLink>
                </template>
            </el-table-column>
            <el-table-column
                :label="t('admin.products.category')"
                min-width="150"
            >
                <template #default="{ row }">
                    <span class="text-muted">{{ tr(row.category) || '—' }}</span>
                </template>
            </el-table-column>
            <el-table-column
                :label="t('admin.products.price')"
                min-width="140"
            >
                <template #default="{ row }">
                    <span v-if="row.min_price !== null">{{ t('product.priceFrom', { price: price(row.min_price) }) }}</span>
                    <span
                        v-else
                        class="text-muted"
                    >
                        {{ t('product.outOfStock') }}
                    </span>
                    <span class="block text-xs text-muted">
                        {{ row.variants.map((v: Variant) => `${v.volume_ml} ml`).join(' · ') }}
                    </span>
                </template>
            </el-table-column>
            <el-table-column
                :label="t('admin.products.active')"
                width="100"
                align="center"
            >
                <template #default="{ row }">
                    <el-switch
                        :model-value="row.is_active"
                        @update:model-value="toggle(row, 'is_active', $event as boolean)"
                    />
                </template>
            </el-table-column>
            <el-table-column
                :label="t('admin.products.featured')"
                width="110"
                align="center"
            >
                <template #default="{ row }">
                    <el-switch
                        :model-value="row.is_featured"
                        @update:model-value="toggle(row, 'is_featured', $event as boolean)"
                    />
                </template>
            </el-table-column>
            <el-table-column
                width="210"
                align="right"
            >
                <template #default="{ row }">
                    <el-button
                        size="small"
                        @click="router.push(`/admin/products/${row.id}`)"
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

        <div
            v-if="total > PAGE_SIZE"
            class="flex justify-end border-t border-line px-5 py-4"
        >
            <el-pagination
                v-model:current-page="filters.page"
                background
                layout="total, prev, pager, next"
                :total="total"
                :page-size="PAGE_SIZE"
                @current-change="load"
            />
        </div>
    </AdminCard>
</template>
