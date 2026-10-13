<!-- ProductCardSkeleton.vue -->
<template>
    <div
        class="flex flex-col overflow-hidden rounded-xl border border-slate-200"
        aria-hidden="true"
    >
        <div class="aspect-square animate-pulse bg-surface" />
        <div class="space-y-3 border-t border-slate-200 p-3">
            <div class="h-3 w-1/2 animate-pulse rounded bg-surface" />
            <div class="h-4 w-4/5 animate-pulse rounded bg-surface" />
            <div class="h-4 w-1/3 animate-pulse rounded bg-surface" />
            <div class="h-11 animate-pulse rounded-lg bg-surface" />
        </div>
    </div>
</template>
