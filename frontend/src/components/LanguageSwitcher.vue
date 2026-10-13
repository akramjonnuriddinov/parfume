<!-- LanguageSwitcher.vue -->
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { LOCALES, setLocale } from '../i18n'
import type { Locale } from '../api/types'

const { locale, t } = useI18n()

const current = computed(
    () => LOCALES.find((l) => l.code === locale.value) ?? LOCALES[0]
)

function select(code: string | number | object): void {
    setLocale(code as Locale)
}
</script>

<template>
    <el-dropdown
        trigger="click"
        @command="select"
    >
        <button
            type="button"
            :aria-label="t('header.language')"
            class="flex h-11 shrink-0 cursor-pointer items-center gap-1.5 rounded-lg border border-line px-3 text-sm font-medium text-ink transition-colors hover:border-brand hover:text-brand focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand"
        >
            <svg
                class="size-4"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.7"
                stroke-linecap="round"
                stroke-linejoin="round"
                aria-hidden="true"
            >
                <circle cx="12" cy="12" r="9" />
                <path d="M3 12h18M12 3c2.5 2.7 3.8 5.7 3.8 9s-1.3 6.3-3.8 9c-2.5-2.7-3.8-5.7-3.8-9S9.5 5.7 12 3Z" />
            </svg>
            {{ current.short }}
        </button>

        <template #dropdown>
            <el-dropdown-menu>
                <el-dropdown-item
                    v-for="item in LOCALES"
                    :key="item.code"
                    :command="item.code"
                    :class="item.code === locale ? 'font-medium text-brand!' : ''"
                >
                    {{ item.label }}
                </el-dropdown-item>
            </el-dropdown-menu>
        </template>
    </el-dropdown>
</template>
