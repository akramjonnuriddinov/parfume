<!-- AboutView.vue -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import PageIntro from '../components/PageIntro.vue'
import { useInquiryStore } from '../stores/inquiry'

const { t, tm, rt } = useI18n()
const inquiry = useInquiryStore()

type Item = { title: string; text: string }
type Stat = { value: string; label: string }
</script>

<template>
    <div>
        <PageIntro
            :title="t('about.title')"
            :text="t('about.lead')"
        />

        <section class="mx-auto max-w-[1240px] px-4 py-10 sm:px-6 lg:py-14">
            <div class="grid overflow-hidden rounded-2xl bg-hero lg:grid-cols-2">
                <div class="px-6 py-10 sm:px-10 lg:p-14">
                    <h2 class="text-2xl font-medium tracking-tight text-ink sm:text-3xl">
                        {{ t('about.missionTitle') }}
                    </h2>
                    <p class="mt-4 text-base leading-7 text-slate-600">{{ t('about.missionText') }}</p>
                </div>
                <dl class="grid grid-cols-3 gap-4 border-t border-line px-6 py-10 sm:px-10 lg:border-t-0 lg:border-l lg:p-14">
                    <div
                        v-for="(stat, i) in (tm('about.stats') as Stat[])"
                        :key="i"
                        class="flex flex-col justify-center"
                    >
                        <dt class="order-2 mt-2 text-xs leading-5 text-muted sm:text-sm">{{ rt(stat.label) }}</dt>
                        <dd class="order-1 text-4xl font-medium tracking-tight text-brand sm:text-5xl">
                            {{ rt(stat.value) }}
                        </dd>
                    </div>
                </dl>
            </div>
        </section>

        <section
            class="mx-auto max-w-[1240px] px-4 py-6 sm:px-6 lg:py-10"
            aria-labelledby="values-title"
        >
            <h2
                id="values-title"
                class="text-2xl font-medium tracking-tight text-ink sm:text-3xl"
            >
                {{ t('about.valuesTitle') }}
            </h2>
            <div class="mt-8 grid gap-x-10 sm:grid-cols-2">
                <div
                    v-for="(item, i) in (tm('about.values') as Item[])"
                    :key="i"
                    class="border-t border-slate-200 py-6"
                >
                    <h3 class="text-lg font-medium text-ink">{{ rt(item.title) }}</h3>
                    <p class="mt-2 text-sm leading-6 text-slate-500">{{ rt(item.text) }}</p>
                </div>
            </div>
        </section>

        <section class="mx-auto max-w-[1240px] px-4 py-10 sm:px-6 lg:py-14">
            <div class="flex flex-col items-start justify-between gap-6 border-y border-line py-8 sm:py-10 lg:flex-row lg:items-center">
                <div class="max-w-xl">
                    <h2 class="text-2xl font-medium tracking-tight text-ink sm:text-3xl">
                        {{ t('about.b2bTitle') }}
                    </h2>
                    <p class="mt-3 text-base leading-7 text-muted">{{ t('about.b2bText') }}</p>
                </div>
                <el-button
                    type="primary"
                    class="m-0! h-12! w-full! shrink-0! rounded-lg! px-6! text-sm! font-medium! sm:w-auto!"
                    @click="inquiry.open()"
                >
                    {{ t('home.consultation.cta') }}
                </el-button>
            </div>
        </section>
    </div>
</template>
