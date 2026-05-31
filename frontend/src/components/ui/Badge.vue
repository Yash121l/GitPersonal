<script setup lang="ts">
import { cva, type VariantProps } from 'class-variance-authority'
import { computed } from 'vue'

import { cn } from '@/lib/utils'

defineOptions({ name: 'UiBadge' })

const badgeVariants = cva('inline-flex items-center rounded-md border px-2 py-0.5 text-[11px] font-medium tracking-wide', {
  variants: {
    variant: {
      neutral: 'border-zinc-800 bg-zinc-900 text-zinc-300',
      accent: 'border-sky-500/30 bg-sky-500/10 text-sky-300',
      warning: 'border-amber-500/30 bg-amber-500/10 text-amber-300',
      danger: 'border-red-500/30 bg-red-500/10 text-red-300',
    },
  },
  defaultVariants: {
    variant: 'neutral',
  },
})

type BadgeVariant = VariantProps<typeof badgeVariants>['variant']

interface Props {
  variant?: BadgeVariant
  class?: string
}

const props = defineProps<Props>()
const classes = computed(() => cn(badgeVariants({ variant: props.variant }), props.class))
</script>

<template>
  <span :class="classes">
    <slot />
  </span>
</template>
