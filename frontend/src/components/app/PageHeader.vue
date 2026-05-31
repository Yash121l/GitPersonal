<script setup lang="ts">
defineProps<{
  eyebrow?: string
  title: string
  description: string
}>()
</script>

<template>
  <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
    <div class="space-y-2">
      <p v-if="eyebrow" class="eyebrow">{{ eyebrow }}</p>
      <h2 class="text-3xl font-semibold tracking-tight text-zinc-50">{{ title }}</h2>
      <p class="max-w-3xl text-sm leading-6 text-zinc-400">{{ description }}</p>
    </div>

    <div v-if="$slots.actions" class="flex flex-wrap items-center gap-2">
      <slot name="actions" />
    </div>
  </div>
</template>
