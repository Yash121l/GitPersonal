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
    'rounded-lg border border-zinc-800 bg-zinc-950 p-6 text-zinc-100 shadow-sm',
    props.class,
  ),
)
</script>

<template>
  <section :class="classes">
    <slot />
  </section>
</template>
