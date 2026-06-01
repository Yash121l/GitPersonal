<script setup lang="ts">
import { computed } from 'vue'

import { cn } from '@/lib/utils'

defineOptions({ name: 'UiInput' })

interface Props {
  class?: string
}

const props = defineProps<Props>()
const model = defineModel<string>()
const classes = computed(() =>
  cn(
    'flex h-10 w-full rounded-md border border-zinc-800 bg-zinc-950 px-3 text-sm text-zinc-100 outline-none transition placeholder:text-zinc-500 focus:border-zinc-700 focus:ring-2 focus:ring-zinc-700/50',
    props.class,
  ),
)
</script>

<template>
  <input v-model="model" :class="classes" v-bind="$attrs" />
</template>
