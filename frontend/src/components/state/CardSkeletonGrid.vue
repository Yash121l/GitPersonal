<script setup lang="ts">
import Skeleton from '@/components/ui/Skeleton.vue'

withDefaults(
  defineProps<{
    count?: number
    wrapperClass?: string
    itemClass?: string
  }>(),
  {
    count: 4,
    wrapperClass: 'grid gap-3',
    itemClass: 'h-28',
  },
)
</script>

<template>
  <div :class="wrapperClass">
    <Skeleton v-for="index in count" :key="index" :class="itemClass" />
  </div>
</template>
