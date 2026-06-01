<script setup lang="ts">
defineProps<{
  eyebrow?: string
  title: string
  description: string
}>()
</script>

<template>
  <div class="rounded-xl border border-dashed border-zinc-800 bg-zinc-950 p-8 text-center">
    <p v-if="eyebrow" class="text-xs font-semibold uppercase tracking-[0.18em] text-zinc-500">
      {{ eyebrow }}
    </p>
    <h3 class="mt-2 text-lg font-semibold text-zinc-100">{{ title }}</h3>
    <p class="mx-auto mt-2 max-w-2xl text-sm leading-6 text-zinc-400">{{ description }}</p>
    <div v-if="$slots.default" class="mt-4">
      <slot />
    </div>
  </div>
</template>
