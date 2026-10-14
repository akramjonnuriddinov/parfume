<!-- HeroSection.vue — dark boutique-window hero with the perfume fan -->
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import HeroSlider from './HeroSlider.vue'

const { t, tm, rt } = useI18n()

const emit = defineEmits<{
    catalog: []
    decants: []
}>()

const stats = () => tm('home.hero.stats') as unknown as { value: string; label: string }[]
</script>

<template>
    <section class="relative isolate overflow-hidden bg-[#131211] text-white">
        <!-- Tilla nur va nozik chiziq -->
        <div
            class="pointer-events-none absolute inset-0 -z-10"
            aria-hidden="true"
        >
            <div class="absolute top-1/2 right-[-10%] size-[720px] -translate-y-1/2 rounded-full bg-[radial-gradient(circle,rgba(184,145,90,0.28)_0%,rgba(184,145,90,0.08)_40%,transparent_70%)]" />
            <div class="absolute -top-40 -left-40 size-[480px] rounded-full bg-[radial-gradient(circle,rgba(184,145,90,0.10)_0%,transparent_65%)]" />
            <div class="absolute inset-x-0 bottom-0 h-px bg-gradient-to-r from-transparent via-gold/40 to-transparent" />
        </div>

        <!-- Telefonda: matn → yelpig‘ich → raqamlar; kompyuterda yelpig‘ich o‘ngda -->
        <div class="mx-auto grid max-w-[1240px] items-center gap-x-10 px-4 pt-10 pb-10 sm:px-6 lg:min-h-[640px] lg:grid-cols-[1.05fr_1fr] lg:grid-rows-[1fr_auto] lg:py-16">
            <!-- Matn -->
            <div class="relative z-10 lg:col-start-1 lg:self-end">
                <p
                    class="hero-rise flex items-center gap-3 text-[11px] font-medium tracking-[0.28em] text-gold uppercase"
                    style="animation-delay: 0.05s"
                >
                    <span
                        class="h-px w-10 bg-gold"
                        aria-hidden="true"
                    />
                    {{ t('home.hero.eyebrow') }}
                </p>

                <i18n-t
                    keypath="home.hero.title"
                    tag="h1"
                    scope="global"
                    class="hero-rise mt-6 max-w-[620px] text-[44px] leading-[1.04] font-medium tracking-[-0.02em] text-white sm:text-6xl lg:text-[76px]"
                    style="animation-delay: 0.15s"
                >
                    <template #accent>
                        <span class="font-display font-medium tracking-normal text-gold italic">{{ t('home.hero.accent') }}</span>
                    </template>
                </i18n-t>

                <p
                    class="hero-rise mt-6 max-w-[440px] text-base leading-7 text-white/60 sm:text-lg sm:leading-8"
                    style="animation-delay: 0.25s"
                >
                    {{ t('home.hero.text') }}
                </p>

                <div
                    class="hero-rise mt-9 flex flex-wrap gap-3"
                    style="animation-delay: 0.35s"
                >
                    <button
                        type="button"
                        class="group inline-flex h-13 cursor-pointer items-center gap-2 rounded-full bg-gold px-7 text-sm font-medium text-[#131211] transition-colors hover:bg-[#c9a36b] focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-gold"
                        @click="emit('catalog')"
                    >
                        {{ t('home.hero.cta') }}
                        <svg class="size-4 transition-transform group-hover:translate-x-0.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <path d="M5 12h14m-6-6 6 6-6 6" />
                        </svg>
                    </button>
                    <button
                        type="button"
                        class="inline-flex h-13 cursor-pointer items-center rounded-full border border-white/20 px-7 text-sm font-medium text-white transition-colors hover:border-white/60 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-gold"
                        @click="emit('decants')"
                    >
                        {{ t('home.hero.decants') }}
                    </button>
                </div>

            </div>

            <!-- Atirlar yelpig‘ichi -->
            <HeroSlider class="lg:col-start-2 lg:row-span-2 lg:row-start-1" />

            <!-- Afzalliklar -->
            <dl
                class="hero-rise mt-4 grid max-w-[560px] lg:col-start-1 lg:mt-12 lg:self-start grid-cols-2 gap-x-6 gap-y-5 border-t border-white/10 pt-7 sm:grid-cols-4"
                style="animation-delay: 0.45s"
            >
                <div
                    v-for="s in stats()"
                    :key="rt(s.value)"
                    class="flex flex-col-reverse"
                >
                    <dt class="mt-0.5 text-xs leading-5 text-white/45">{{ rt(s.label) }}</dt>
                    <dd class="text-lg font-medium text-white">{{ rt(s.value) }}</dd>
                </div>
            </dl>
        </div>
    </section>
</template>

<style scoped>
.hero-rise {
    animation: hero-rise 0.9s cubic-bezier(0.22, 0.8, 0.2, 1) both;
}

@keyframes hero-rise {
    from {
        opacity: 0;
        transform: translateY(18px);
    }

    to {
        opacity: 1;
        transform: none;
    }
}

@media (prefers-reduced-motion: reduce) {
    .hero-rise {
        animation: none;
    }
}
</style>
