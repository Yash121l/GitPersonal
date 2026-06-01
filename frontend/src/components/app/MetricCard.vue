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

const accentBorder = computed(() =>
  cn(
    props.accent === 'default' && 'border-zinc-800',
    props.accent === 'sky' && 'border-sky-500/30',
    props.accent === 'emerald' && 'border-emerald-500/30',
    props.accent === 'amber' && 'border-amber-500/30',
  ),
)

const accentText = computed(() =>
  cn(
    props.accent === 'default' && 'text-zinc-50',
    props.accent === 'sky' && 'text-sky-300',
    props.accent === 'emerald' && 'text-emerald-300',
    props.accent === 'amber' && 'text-amber-300',
  ),
)
</script>

<template>
  <Card :class="cn('space-y-2', accentBorder)">
    <p class="text-sm font-medium text-zinc-400">{{ props.label }}</p>
    <p :class="cn('text-3xl font-semibold tracking-tight', accentText)">{{ props.value }}</p>
    <p v-if="props.caption" class="text-sm leading-6 text-zinc-400">
      {{ props.caption }}
    </p>
  </Card>
</template>
