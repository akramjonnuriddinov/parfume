<!-- HeroSection.vue -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import HeroSlider from './HeroSlider.vue'

const { t } = useI18n()

const emit = defineEmits<{
    catalog: []
    decants: []
}>()
</script>

<template>
    <section class="bg-white">
        <div class="mx-auto max-w-[1240px] px-4 py-6 sm:px-6 lg:py-8">
            <div class="grid overflow-hidden rounded-2xl bg-hero lg:min-h-[480px] lg:grid-cols-2">
                <!-- Matn -->
                <div class="flex flex-col items-start justify-center px-6 py-10 sm:px-10 sm:py-12 lg:p-14">
                    <p class="text-xs font-medium tracking-[0.2em] text-gold uppercase">
                        {{ t('home.hero.eyebrow') }}
                    </p>

                    <h1 class="mt-4 max-w-[500px] text-4xl leading-[1.12] font-medium tracking-tight text-ink sm:text-5xl lg:text-[56px]">
                        {{ t('home.hero.title') }}
                    </h1>

                    <p class="mt-5 max-w-[400px] text-base leading-7 text-muted">
                        {{ t('home.hero.text') }}
                    </p>

                    <div class="mt-8 flex flex-wrap gap-3">
                        <el-button
                            type="primary"
                            size="large"
                            class="m-0! h-12! rounded-lg! px-6! text-sm! font-medium!"
                            @click="emit('catalog')"
                        >
                            {{ t('home.hero.cta') }}
                        </el-button>
                        <el-button
                            size="large"
                            plain
                            class="m-0! h-12! rounded-lg! border-brand! bg-transparent! px-6! text-sm! font-medium! text-brand! hover:bg-white!"
                            @click="emit('decants')"
                        >
                            {{ t('home.hero.decants') }}
                        </el-button>
                    </div>
                </div>

                <!-- Atirlar slayderi -->
                <HeroSlider />
            </div>
        </div>
    </section>
</template>
