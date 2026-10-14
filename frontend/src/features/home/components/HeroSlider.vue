<!-- HeroSlider.vue — swipeable photo slider; each slide opens its perfume -->
<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
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

const AUTOPLAY_MS = 4500

const track = ref<HTMLElement | null>(null)
const active = ref(0)
// Where the slider is heading; differs from `active` while a smooth scroll
// runs, so quick repeated clicks keep counting from the right slide.
let target: number | null = null
let timer: number | undefined
let paused = false

// Native scroll-snap does the swiping; we only follow and drive it.
function onScroll(): void {
    const el = track.value
    if (!el) return
    active.value = Math.round(el.scrollLeft / el.clientWidth)
    if (target !== null && Math.abs(el.scrollLeft - target * el.clientWidth) < 2) target = null
}

function goTo(index: number): void {
    const el = track.value
    if (!el) return
    target = (index + slides.length) % slides.length
    el.scrollTo({ left: target * el.clientWidth, behavior: 'smooth' })
}

function current(): number {
    return target ?? active.value
}

function start(): void {
    stop()
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return
    timer = window.setInterval(() => {
        if (!paused && !document.hidden) goTo(current() + 1)
    }, AUTOPLAY_MS)
}

function stop(): void {
    if (timer !== undefined) window.clearInterval(timer)
    timer = undefined
}

function pause(value: boolean): void {
    paused = value
}

// A swipe takes over from any scroll we started.
function onTouch(): void {
    target = null
    start()
}

// After a manual move, give the visitor a full interval before moving on.
function manual(index: number): void {
    goTo(index)
    start()
}

onMounted(start)
onBeforeUnmount(stop)
</script>

<template>
    <div
        class="relative h-full min-h-[340px] overflow-hidden sm:min-h-[420px]"
        role="region"
        aria-roledescription="carousel"
        :aria-label="t('home.hero.slides')"
        @mouseenter="pause(true)"
        @mouseleave="pause(false)"
        @focusin="pause(true)"
        @focusout="pause(false)"
        @touchstart.passive="onTouch"
    >
        <div
            ref="track"
            class="hero-track flex h-full snap-x snap-mandatory overflow-x-auto"
            @scroll.passive="onScroll"
        >
            <router-link
                v-for="(slide, i) in slides"
                :key="slide.productId"
                :to="`/products/${slide.productId}`"
                class="group relative block h-full w-full shrink-0 snap-center snap-always"
                role="group"
                aria-roledescription="slide"
                :aria-label="`${slide.brand} ${slide.name}`"
                :tabindex="i === active ? 0 : -1"
            >
                <img
                    :src="slide.image"
                    :alt="`${slide.brand} ${slide.name}`"
                    class="absolute inset-0 h-full w-full object-cover"
                    :loading="i === 0 ? 'eager' : 'lazy'"
                    :fetchpriority="i === 0 ? 'high' : 'auto'"
                    decoding="async"
                    draggable="false"
                >
                <span class="absolute bottom-14 left-4 flex items-center gap-3 rounded-xl bg-white/90 px-4 py-2.5 shadow-sm backdrop-blur-sm sm:left-6">
                    <span class="flex flex-col">
                        <span class="text-[11px] font-medium tracking-[0.14em] text-gold uppercase">{{ slide.brand }}</span>
                        <span class="text-sm font-medium text-ink">{{ slide.name }}</span>
                    </span>
                    <span class="flex items-center gap-1 text-xs text-brand">
                        {{ t('home.hero.view') }}
                        <svg class="size-3.5 transition-transform group-hover:translate-x-0.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <path d="M5 12h14m-6-6 6 6-6 6" />
                        </svg>
                    </span>
                </span>
            </router-link>
        </div>

        <!-- Strelkalar (kompyuterda) -->
        <button
            type="button"
            class="absolute top-1/2 left-3 hidden size-10 -translate-y-1/2 cursor-pointer items-center justify-center rounded-full bg-white/85 text-ink shadow-sm transition hover:bg-white sm:flex"
            :aria-label="t('home.hero.prev')"
            @click="manual(current() - 1)"
        >
            <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="m15 6-6 6 6 6" />
            </svg>
        </button>
        <button
            type="button"
            class="absolute top-1/2 right-3 hidden size-10 -translate-y-1/2 cursor-pointer items-center justify-center rounded-full bg-white/85 text-ink shadow-sm transition hover:bg-white sm:flex"
            :aria-label="t('home.hero.next')"
            @click="manual(current() + 1)"
        >
            <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="m9 6 6 6-6 6" />
            </svg>
        </button>

        <!-- Nuqtalar -->
        <div class="absolute bottom-4 left-1/2 flex -translate-x-1/2 gap-2 rounded-full bg-black/25 px-3 py-2 backdrop-blur-sm">
            <button
                v-for="(slide, i) in slides"
                :key="slide.productId"
                type="button"
                class="h-2 cursor-pointer rounded-full transition-all"
                :class="i === active ? 'w-6 bg-white' : 'w-2 bg-white/60 hover:bg-white/80'"
                :aria-label="t('home.hero.goTo', { n: i + 1 })"
                :aria-current="i === active"
                @click="manual(i)"
            />
        </div>
    </div>
</template>

<style scoped>
.hero-track {
    scrollbar-width: none;
}

.hero-track::-webkit-scrollbar {
    display: none;
}
</style>
