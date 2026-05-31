<script setup lang="ts">
import { computed } from 'vue'

import Card from '@/components/ui/Card.vue'
import { cn } from '@/lib/utils'

const props = withDefaults(
  defineProps<{
    label: string
    value: string
    caption?: string
    accent?: 'default' | 'sky' | 'emerald' | 'amber'
  }>(),
  {
    caption: '',
    accent: 'default',
  },
)

const classes = computed(() =>
  cn(
    props.accent === 'sky' && 'border-sky-500/20 bg-sky-500/5',
    props.accent === 'emerald' && 'border-emerald-500/20 bg-emerald-500/5',
    props.accent === 'amber' && 'border-amber-500/20 bg-amber-500/5',
  ),
)
</script>

<template>
  <Card :class="classes">
    <p class="eyebrow">{{ props.label }}</p>
    <p class="mt-3 text-3xl font-semibold tracking-tight text-zinc-50">{{ props.value }}</p>
    <p v-if="props.caption" class="mt-2 text-sm leading-6 text-zinc-400">
      {{ props.caption }}
    </p>
  </Card>
</template>
