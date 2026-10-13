<!-- DecantSection.vue — explains decants ("raspiv"): try before buying a full bottle -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const emit = defineEmits<{ catalog: [] }>()

const steps = ['choose', 'try', 'bottle'] as const
</script>

<template>
    <section
        id="raspiv"
        class="bg-white"
        aria-labelledby="decant-title"
    >
        <div class="mx-auto max-w-[1240px] px-4 py-10 sm:px-6 lg:py-14">
            <div class="grid gap-10 rounded-2xl bg-ink px-6 py-10 text-white sm:px-10 lg:grid-cols-[1fr_1.4fr] lg:gap-16 lg:p-14">
                <div>
                    <p class="text-xs font-medium tracking-[0.2em] text-gold uppercase">{{ t('home.decant.eyebrow') }}</p>
                    <h2
                        id="decant-title"
                        class="mt-3 text-4xl leading-tight font-medium sm:text-5xl"
                    >
                        {{ t('home.decant.title') }}
                    </h2>
                    <p class="mt-4 max-w-md text-base leading-7 text-white/70">{{ t('home.decant.text') }}</p>
                    <el-button
                        size="large"
                        class="mt-8! h-12! rounded-lg! border-gold! bg-gold! px-6! text-sm! font-medium! text-white! hover:opacity-90"
                        @click="emit('catalog')"
                    >
                        {{ t('home.decant.cta') }}
                    </el-button>
                </div>
                <ol class="grid gap-4 sm:grid-cols-3 lg:gap-6">
                    <li
                        v-for="(step, i) in steps"
                        :key="step"
                        class="rounded-xl border border-white/10 bg-white/5 p-5"
                    >
                        <span class="text-4xl font-medium text-gold [font-variant-numeric:lining-nums]">0{{ i + 1 }}</span>
                        <h3 class="mt-3 text-base font-medium">{{ t(`home.decant.steps.${step}.title`) }}</h3>
                        <p class="mt-2 text-sm leading-6 text-white/65">{{ t(`home.decant.steps.${step}.text`) }}</p>
                    </li>
                </ol>
            </div>
        </div>
    </section>
</template>
