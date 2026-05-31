<script setup lang="ts">
import EmptyState from '@/components/empty/EmptyState.vue'

withDefaults(
  defineProps<{
    loading: boolean
    empty: boolean
    emptyEyebrow?: string
    emptyTitle?: string
    emptyDescription?: string
    wrapperClass?: string
  }>(),
  {
    emptyEyebrow: 'No Data',
    emptyTitle: 'Nothing to show yet.',
    emptyDescription: 'This surface will populate when data becomes available.',
    wrapperClass: '',
  },
)
</script>

<template>
  <div :class="wrapperClass">
    <template v-if="loading">
      <slot name="loading" />
    </template>
    <template v-else-if="empty">
      <slot name="empty">
        <EmptyState
          :eyebrow="emptyEyebrow"
          :title="emptyTitle"
          :description="emptyDescription"
        />
      </slot>
    </template>
    <template v-else>
      <slot />
    </template>
  </div>
</template>
