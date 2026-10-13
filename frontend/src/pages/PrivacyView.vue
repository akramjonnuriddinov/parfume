<!-- PrivacyView.vue -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import PageIntro from '../components/PageIntro.vue'

const { t, tm, rt } = useI18n()

type Item = { title: string; text: string }
</script>

<template>
    <div>
        <PageIntro
            :title="t('privacy.title')"
            :text="t('privacy.lead')"
        />
        <section class="mx-auto max-w-[1240px] px-4 py-10 sm:px-6 lg:py-14">
            <div class="max-w-3xl">
                <div
                    v-for="(item, i) in (tm('privacy.sections') as Item[])"
                    :key="i"
                    class="border-t border-slate-200 py-6"
                >
                    <h2 class="text-lg font-medium text-ink">{{ rt(item.title) }}</h2>
                    <p class="mt-2 text-sm leading-7 text-slate-600">{{ rt(item.text) }}</p>
                </div>
            </div>
        </section>
    </div>
</template>
