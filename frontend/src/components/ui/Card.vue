<script setup lang="ts">
import { computed } from 'vue'

import { cn } from '@/lib/utils'

defineOptions({ name: 'UiCard' })

interface Props {
  class?: string
}

const props = defineProps<Props>()
const classes = computed(() =>
  cn(
    'rounded-2xl border border-zinc-800/90 bg-zinc-950/80 p-6 text-zinc-100 shadow-[0_20px_60px_rgba(0,0,0,0.42)] backdrop-blur-xl',
    props.class,
  ),
)
</script>

<template>
  <section :class="classes">
    <slot />
  </section>
</template>
