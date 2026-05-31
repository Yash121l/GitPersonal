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
    'rounded-xl border border-zinc-800/90 bg-zinc-950/88 p-4 text-zinc-100 shadow-[0_14px_40px_rgba(0,0,0,0.36)] backdrop-blur-xl',
    props.class,
  ),
)
</script>

<template>
  <section :class="classes">
    <slot />
  </section>
</template>
