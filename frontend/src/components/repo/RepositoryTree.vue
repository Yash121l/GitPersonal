<script setup lang="ts">
import { ChevronRight, FileCode2, FolderTree } from '@lucide/vue'
import { useVirtualizer } from '@tanstack/vue-virtual'
import { computed, ref } from 'vue'

import type { RepositoryTreeEntry } from '@/lib/api'
import { cn, formatBytes } from '@/lib/utils'

const props = defineProps<{
  entries: RepositoryTreeEntry[]
  activePath?: string
}>()

const emit = defineEmits<{
  select: [entry: RepositoryTreeEntry]
}>()

const parentRef = ref<HTMLElement | null>(null)
const rowVirtualizer = useVirtualizer(
  computed(() => ({
    count: props.entries.length,
    getScrollElement: () => parentRef.value,
    estimateSize: () => 52,
    overscan: 8,
  })),
)

const items = computed(() => rowVirtualizer.value.getVirtualItems())
</script>

<template>
  <div ref="parentRef" class="h-[28rem] overflow-auto rounded-xl border border-zinc-800 bg-black/20 scrollbar-subtle">
    <div :style="{ height: `${rowVirtualizer.getTotalSize()}px`, position: 'relative' }">
      <button
        v-for="item in items"
        :key="props.entries[item.index]?.path"
        type="button"
        :class="
          cn(
            'absolute left-0 flex w-full items-center justify-between gap-3 border-b border-zinc-800 px-4 py-3 text-left text-zinc-300 transition hover:bg-zinc-900/80 hover:text-zinc-100',
            activePath === props.entries[item.index]?.path ? 'bg-zinc-900 text-zinc-50 hover:bg-zinc-900' : '',
          )
        "
        :style="{ transform: `translateY(${item.start}px)` }"
        @click="emit('select', props.entries[item.index]!)"
      >
        <div class="min-w-0">
          <div class="flex items-center gap-3">
            <component
              :is="props.entries[item.index]?.type === 'tree' ? FolderTree : FileCode2"
              class="size-4 shrink-0"
            />
            <span class="truncate text-sm font-medium">{{ props.entries[item.index]?.name }}</span>
          </div>
          <p class="ml-7 mt-1 text-xs text-inherit/70">
            {{ props.entries[item.index]?.type === 'tree' ? 'Directory' : formatBytes(props.entries[item.index]?.size_bytes) }}
          </p>
        </div>
        <ChevronRight v-if="props.entries[item.index]?.type === 'tree'" class="size-4 shrink-0 opacity-70" />
      </button>
    </div>
  </div>
</template>
