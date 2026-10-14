<!-- HeroSlider.vue — crossfading photo card; each slide opens its perfume -->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const slides = [
    { image: '/images/hero/imagination.webp', brand: 'Louis Vuitton', name: 'Imagination', productId: 59 },
    { image: '/images/hero/layton.webp', brand: 'Parfums de Marly', name: 'Layton', productId: 21 },
    { image: '/images/hero/ombre-nomade.webp', brand: 'Louis Vuitton', name: 'Ombre Nomade', productId: 29 },
    { image: '/images/hero/blue-talisman.webp', brand: 'Ex Nihilo', name: 'Blue Talisman', productId: 54 },
    { image: '/images/hero/pacific-chill.webp', brand: 'Louis Vuitton', name: 'Pacific Chill', productId: 32 },
    { image: '/images/hero/aventus.webp', brand: 'Creed', name: 'Aventus', productId: 35 },
]

const active = ref(0)
const paused = ref(false)
// Autoplay is driven by the progress bar's CSS animation: when it ends we
// move on, so pausing the animation pauses the slider too.
const autoplay = ref(false)
const slide = computed(() => slides[active.value])

function goTo(index: number): void {
    active.value = (index + slides.length) % slides.length
}

let touchX: number | null = null

function onTouchStart(e: TouchEvent): void {
    touchX = e.touches[0]?.clientX ?? null
}

function onTouchEnd(e: TouchEvent): void {
    const endX = e.changedTouches[0]?.clientX
    if (touchX === null || endX === undefined) return
    const dx = endX - touchX
    touchX = null
    if (Math.abs(dx) > 40) {
        e.preventDefault() // a swipe is not a tap on the link
        goTo(active.value + (dx < 0 ? 1 : -1))
    }
}

onMounted(() => {
    autoplay.value = !window.matchMedia('(prefers-reduced-motion: reduce)').matches
})
</script>

<template>
    <div
        class="flex items-center justify-center px-6 pt-2 pb-10 sm:px-10 lg:py-10"
        role="region"
        aria-roledescription="carousel"
        :aria-label="t('home.hero.slides')"
        @mouseenter="paused = true"
        @mouseleave="paused = false"
        @focusin="paused = true"
        @focusout="paused = false"
    >
        <div class="w-full max-w-[270px] sm:max-w-[320px] lg:max-w-[310px]">
            <div class="relative">
                <!-- Orqadagi varaq -->
                <span
                    class="absolute inset-0 translate-x-4 translate-y-2 rotate-[4deg] rounded-2xl bg-white/70 shadow-sm"
                    aria-hidden="true"
                />

                <router-link
                    :to="`/products/${slide.productId}`"
                    class="relative block aspect-[4/5] overflow-hidden rounded-2xl bg-white shadow-[0_24px_60px_-24px_rgba(28,28,28,0.45)] focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand"
                    :aria-label="`${slide.brand} ${slide.name}`"
                    @touchstart.passive="onTouchStart"
                    @touchend="onTouchEnd"
                >
                    <img
                        v-for="(s, i) in slides"
                        :key="s.productId"
                        :src="s.image"
                        :alt="i === active ? `${s.brand} ${s.name}` : ''"
                        class="absolute inset-0 h-full w-full object-cover transition-[opacity,transform] duration-[900ms] ease-out"
                        :class="i === active ? 'scale-100 opacity-100' : 'scale-[1.04] opacity-0'"
                        :loading="i === 0 ? 'eager' : 'lazy'"
                        :fetchpriority="i === 0 ? 'high' : 'auto'"
                        decoding="async"
                        draggable="false"
                    >
                </router-link>
            </div>

            <!-- Nom va boshqaruv -->
            <div class="mt-6 flex items-end justify-between gap-4">
                <router-link
                    :to="`/products/${slide.productId}`"
                    class="group min-w-0"
                    aria-live="polite"
                >
                    <span class="block truncate text-[11px] font-medium tracking-[0.18em] text-gold uppercase">{{ slide.brand }}</span>
                    <span class="mt-1 flex items-center gap-1.5 text-lg font-medium text-ink">
                        <span class="truncate">{{ slide.name }}</span>
                        <svg class="size-4 shrink-0 text-muted transition-transform group-hover:translate-x-0.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <path d="M5 12h14m-6-6 6 6-6 6" />
                        </svg>
                    </span>
                </router-link>
                <div class="hidden shrink-0 gap-2 sm:flex">
                    <button
                        type="button"
                        class="flex size-9 cursor-pointer items-center justify-center rounded-full border border-line text-ink transition-colors hover:border-ink"
                        :aria-label="t('home.hero.prev')"
                        @click="goTo(active - 1)"
                    >
                        <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <path d="m15 6-6 6 6 6" />
                        </svg>
                    </button>
                    <button
                        type="button"
                        class="flex size-9 cursor-pointer items-center justify-center rounded-full border border-line text-ink transition-colors hover:border-ink"
                        :aria-label="t('home.hero.next')"
                        @click="goTo(active + 1)"
                    >
                        <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <path d="m9 6 6 6-6 6" />
                        </svg>
                    </button>
                </div>
            </div>

            <!-- Jarayon chiziqlari -->
            <div class="mt-4 flex gap-1.5">
                <button
                    v-for="(s, i) in slides"
                    :key="s.productId"
                    type="button"
                    class="relative h-5 flex-1 cursor-pointer"
                    :aria-label="t('home.hero.goTo', { n: i + 1 })"
                    :aria-current="i === active"
                    @click="goTo(i)"
                >
                    <span class="absolute inset-x-0 top-1/2 h-[2px] -translate-y-1/2 overflow-hidden rounded-full bg-ink/15">
                        <!-- Distinct keys: a finished bar must never reuse the animated element. -->
                        <span
                            v-if="i === active && autoplay"
                            :key="`run-${active}`"
                            class="hero-progress absolute inset-y-0 left-0 bg-ink"
                            :style="{ animationPlayState: paused ? 'paused' : 'running' }"
                            @animationend="goTo(active + 1)"
                        />
                        <span
                            v-else-if="i <= active"
                            key="done"
                            class="absolute inset-0 bg-ink"
                        />
                    </span>
                </button>
            </div>
        </div>
    </div>
</template>

<style scoped>
.hero-progress {
    animation: hero-progress 5s linear forwards;
}

@keyframes hero-progress {
    from {
        width: 0;
    }

    to {
        width: 100%;
    }
}
</style>
