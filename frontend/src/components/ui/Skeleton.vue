<script setup lang="ts">
import { computed } from 'vue'

import { cn } from '@/lib/utils'

defineOptions({ name: 'UiSkeleton' })

interface Props {
  class?: string
}

const props = defineProps<Props>()
const classes = computed(() => cn('animate-pulse rounded-lg bg-zinc-900', props.class))
</script>

<template>
  <div :class="classes" />
</template>
