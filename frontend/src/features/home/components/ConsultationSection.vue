<!-- ConsultationSection.vue -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const emit = defineEmits<{
    consultation: []
}>()
</script>

<template>
    <section
        class="bg-white"
        aria-labelledby="consultation-title"
    >
        <div class="mx-auto max-w-[1240px] px-4 py-10 sm:px-6 lg:py-14">
            <div
                class="flex flex-col items-start justify-between gap-6 border-y border-line py-8 sm:py-10 lg:flex-row lg:items-center lg:gap-12"
            >
                <div class="max-w-xl">
                    <h2
                        id="consultation-title"
                        class="text-2xl leading-tight font-medium tracking-tight text-ink sm:text-3xl"
                    >
                        {{ t('home.consultation.title') }}
                    </h2>

                    <p class="mt-3 text-base leading-7 text-muted">
                        {{ t('home.consultation.text') }}
                    </p>
                </div>

                <el-button
                    type="primary"
                    size="large"
                    class="m-0! h-12! w-full! shrink-0! rounded-lg! border-brand! bg-brand! px-6! text-sm! font-medium! hover:border-brand-hover! hover:bg-brand-hover! sm:w-auto!"
                    @click="emit('consultation')"
                >
                    {{ t('home.consultation.cta') }}

                    <svg
                        class="ml-3 size-4"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="1.5"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        aria-hidden="true"
                    >
                        <path d="M5 12h14m-6-6 6 6-6 6" />
                    </svg>
                </el-button>
            </div>
        </div>
    </section>
</template>