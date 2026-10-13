<!-- DeliveryView.vue -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import PageIntro from '../components/PageIntro.vue'

const { t, tm, rt } = useI18n()

type Item = { title: string; text: string }
</script>

<template>
    <div>
        <PageIntro
            :title="t('delivery.title')"
            :text="t('delivery.lead')"
        />

        <section
            class="mx-auto max-w-[1240px] px-4 py-10 sm:px-6 lg:py-14"
            aria-labelledby="steps-title"
        >
            <h2
                id="steps-title"
                class="text-2xl font-medium tracking-tight text-ink sm:text-3xl"
            >
                {{ t('delivery.stepsTitle') }}
            </h2>
            <ol class="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                <li
                    v-for="(step, i) in (tm('delivery.steps') as Item[])"
                    :key="i"
                    class="rounded-2xl bg-hero p-6"
                >
                    <span class="flex size-10 items-center justify-center rounded-xl bg-white text-sm font-semibold text-brand">
                        {{ i + 1 }}
                    </span>
                    <h3 class="mt-5 text-base font-medium text-ink">{{ rt(step.title) }}</h3>
                    <p class="mt-2 text-sm leading-6 text-slate-600">{{ rt(step.text) }}</p>
                </li>
            </ol>
        </section>

        <section class="mx-auto grid max-w-[1240px] gap-10 px-4 py-6 sm:px-6 lg:grid-cols-2 lg:gap-16 lg:py-10">
            <div>
                <h2 class="text-3xl font-medium tracking-tight text-ink">{{ t('delivery.termsTitle') }}</h2>
                <div class="mt-6">
                    <div
                        v-for="(item, i) in (tm('delivery.terms') as Item[])"
                        :key="i"
                        class="border-t border-slate-200 py-5"
                    >
                        <h3 class="text-base font-medium text-ink">{{ rt(item.title) }}</h3>
                        <p class="mt-1.5 text-sm leading-6 text-slate-500">{{ rt(item.text) }}</p>
                    </div>
                </div>
            </div>
            <div>
                <h2 class="text-3xl font-medium tracking-tight text-ink">{{ t('delivery.paymentTitle') }}</h2>
                <div class="mt-6">
                    <div
                        v-for="(item, i) in (tm('delivery.payment') as Item[])"
                        :key="i"
                        class="border-t border-slate-200 py-5"
                    >
                        <h3 class="text-base font-medium text-ink">{{ rt(item.title) }}</h3>
                        <p class="mt-1.5 text-sm leading-6 text-slate-500">{{ rt(item.text) }}</p>
                    </div>
                </div>
            </div>
        </section>

        <section class="mx-auto max-w-[1240px] px-4 py-10 sm:px-6 lg:py-14">
            <div class="rounded-2xl border border-line p-6 sm:p-8">
                <h2 class="text-xl font-medium text-ink">{{ t('delivery.returnTitle') }}</h2>
                <p class="mt-3 max-w-3xl text-sm leading-6 text-slate-600">{{ t('delivery.returnText') }}</p>
            </div>
        </section>
    </div>
</template>
