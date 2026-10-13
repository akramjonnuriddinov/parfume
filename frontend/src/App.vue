<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import uzUz from 'element-plus/es/locale/lang/uz-uz'
import ru from 'element-plus/es/locale/lang/ru'
import en from 'element-plus/es/locale/lang/en'

const { locale } = useI18n()

const elementLocales = { uz: uzUz, ru, en }
const elementLocale = computed(
    () => elementLocales[locale.value as keyof typeof elementLocales] ?? uzUz
)
</script>

<template>
    <el-config-provider :locale="elementLocale">
        <RouterView />
    </el-config-provider>
</template>
