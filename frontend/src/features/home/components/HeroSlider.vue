<!-- HeroSlider.vue — a fan of perfume photo cards; the centre one opens its perfume -->
<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
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

const AUTOPLAY_MS = 5000

const active = ref(0)
// Cards start stacked and fan out once mounted (the entrance animation).
const spread = ref(false)
const slide = computed(() => slides[active.value]!)
let timer: number | undefined
const paused = ref(false)

function goTo(index: number): void {
    active.value = (index + slides.length) % slides.length
}

/** Position of a card relative to the centre one: 0, ±1, ±2… */
function offset(i: number): number {
    const n = slides.length
    return ((i - active.value + n + Math.floor(n / 2)) % n) - Math.floor(n / 2)
}

function cardStyle(i: number): Record<string, string | number> {
    const off = spread.value ? offset(i) : 0
    const abs = Math.abs(off)
    const visible = abs <= 1
    return {
        transform: `translate(-50%, -50%) translateX(${off * 62}%) rotate(${off * 8}deg) scale(${abs === 0 ? 1 : visible ? 0.8 : 0.6})`,
        opacity: visible ? 1 : 0,
        zIndex: 30 - abs * 10,
        filter: abs === 0 ? 'none' : 'brightness(0.55) saturate(0.8)',
        pointerEvents: visible ? 'auto' : 'none',
    }
}

function onCardClick(e: MouseEvent, i: number): void {
    if (offset(i) !== 0) {
        e.preventDefault()
        goTo(i)
        restart()
    }
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
        e.preventDefault() // a swipe is not a tap on the card
        goTo(active.value + (dx < 0 ? 1 : -1))
        restart()
    }
}

function restart(): void {
    if (timer !== undefined) window.clearInterval(timer)
    timer = undefined
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return
    timer = window.setInterval(() => {
        if (!paused.value && !document.hidden) goTo(active.value + 1)
    }, AUTOPLAY_MS)
}

onMounted(() => {
    window.setTimeout(() => (spread.value = true), 250)
    restart()
})
onBeforeUnmount(() => window.clearInterval(timer))
</script>

<template>
    <div
        class="relative"
        role="region"
        aria-roledescription="carousel"
        :aria-label="t('home.hero.slides')"
        @mouseenter="paused = true"
        @mouseleave="paused = false"
        @focusin="paused = true"
        @focusout="paused = false"
    >
        <div
            class="relative h-[360px] sm:h-[460px] lg:h-[500px]"
            @touchstart.passive="onTouchStart"
            @touchend="onTouchEnd"
        >
            <router-link
                v-for="(s, i) in slides"
                :key="s.productId"
                :to="`/products/${s.productId}`"
                class="hero-card absolute top-1/2 left-1/2 block aspect-[4/5] w-[210px] overflow-hidden rounded-[22px] bg-white shadow-[0_40px_80px_-30px_rgba(0,0,0,0.85)] ring-1 ring-white/10 sm:w-[270px] lg:w-[300px]"
                :style="cardStyle(i)"
                :tabindex="offset(i) === 0 ? 0 : -1"
                :aria-hidden="offset(i) !== 0"
                :aria-label="`${s.brand} ${s.name}`"
                @click="onCardClick($event, i)"
            >
                <img
                    :src="s.image"
                    alt=""
                    class="h-full w-full object-cover"
                    :loading="i < 2 || i === slides.length - 1 ? 'eager' : 'lazy'"
                    :fetchpriority="i === 0 ? 'high' : 'auto'"
                    decoding="async"
                    draggable="false"
                >
            </router-link>
        </div>

        <!-- Nom, hisoblagich va strelkalar -->
        <div class="relative z-40 mx-auto mt-2 flex max-w-[300px] items-center justify-between gap-4 sm:max-w-[360px]">
            <button
                type="button"
                class="flex size-10 shrink-0 cursor-pointer items-center justify-center rounded-full border border-white/15 text-white/80 transition-colors hover:border-gold hover:text-gold"
                :aria-label="t('home.hero.prev')"
                @click="goTo(active - 1); restart()"
            >
                <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d="m15 6-6 6 6 6" />
                </svg>
            </button>

            <router-link
                :to="`/products/${slide.productId}`"
                class="group min-w-0 text-center"
                aria-live="polite"
            >
                <Transition
                    name="hero-caption"
                    mode="out-in"
                >
                    <span
                        :key="active"
                        class="block"
                    >
                        <span class="block truncate text-[10px] font-medium tracking-[0.28em] text-gold uppercase">{{ slide.brand }}</span>
                        <span class="mt-1 block truncate font-display text-[28px] leading-tight text-white italic group-hover:text-gold">{{ slide.name }}</span>
                    </span>
                </Transition>
                <span class="mt-1 block text-[11px] tracking-[0.2em] text-white/35 tabular-nums">
                    {{ String(active + 1).padStart(2, '0') }} / {{ String(slides.length).padStart(2, '0') }}
                </span>
            </router-link>

            <button
                type="button"
                class="flex size-10 shrink-0 cursor-pointer items-center justify-center rounded-full border border-white/15 text-white/80 transition-colors hover:border-gold hover:text-gold"
                :aria-label="t('home.hero.next')"
                @click="goTo(active + 1); restart()"
            >
                <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d="m9 6 6 6-6 6" />
                </svg>
            </button>
        </div>
    </div>
</template>

<style scoped>
.hero-card {
    transition:
        transform 0.85s cubic-bezier(0.22, 0.8, 0.2, 1),
        opacity 0.6s ease,
        filter 0.6s ease;
    will-change: transform;
}

.hero-caption-enter-active,
.hero-caption-leave-active {
    transition:
        opacity 0.3s ease,
        transform 0.3s ease;
}

.hero-caption-enter-from {
    opacity: 0;
    transform: translateY(6px);
}

.hero-caption-leave-to {
    opacity: 0;
    transform: translateY(-6px);
}

@media (prefers-reduced-motion: reduce) {
    .hero-card {
        transition: none;
    }
}
</style>
