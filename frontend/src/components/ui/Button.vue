<script setup lang="ts">
import { cva, type VariantProps } from 'class-variance-authority'
import { computed, type Component } from 'vue'

import { cn } from '@/lib/utils'

defineOptions({ name: 'UiButton' })

const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition disabled:pointer-events-none disabled:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-zinc-700 focus-visible:ring-offset-2 focus-visible:ring-offset-zinc-950',
  {
    variants: {
      variant: {
        primary: 'bg-zinc-50 text-zinc-950 hover:bg-zinc-200',
        secondary: 'border border-zinc-800 bg-zinc-900 text-zinc-100 hover:bg-zinc-800',
        ghost: 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-100',
        danger: 'bg-red-500 text-white shadow-sm hover:bg-red-400',
      },
      size: {
        sm: 'h-8 px-3.5',
        md: 'h-10 px-4',
        lg: 'h-11 px-5 text-base',
      },
    },
    defaultVariants: {
      variant: 'primary',
      size: 'md',
    },
  },
)

type ButtonVariant = VariantProps<typeof buttonVariants>['variant']
type ButtonSize = VariantProps<typeof buttonVariants>['size']

interface Props {
  as?: string | Component
  type?: 'button' | 'submit' | 'reset'
  variant?: ButtonVariant
  size?: ButtonSize
  class?: string
}

const props = withDefaults(defineProps<Props>(), {
  as: 'button',
  type: 'button',
})

const classes = computed(() => cn(buttonVariants({ variant: props.variant, size: props.size }), props.class))
</script>

<template>
  <component :is="as" :class="classes" :type="as === 'button' ? type : undefined">
    <slot />
  </component>
</template>
