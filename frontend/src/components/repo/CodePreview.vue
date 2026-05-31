<script setup lang="ts">
import { computed } from 'vue'

import type { RepositoryBlob } from '@/lib/api'
import { formatBytes } from '@/lib/utils'

const props = defineProps<{
  blob: RepositoryBlob | null
}>()

const lines = computed(() => (props.blob?.content ? props.blob.content.split('\n') : []))
</script>

<template>
  <div v-if="!blob" class="rounded-2xl border border-dashed border-zinc-800 bg-zinc-950/70 p-8 text-center">
    <p class="eyebrow">Code Preview</p>
    <h3 class="mt-2 text-lg font-semibold text-zinc-100">Select a file from the tree.</h3>
    <p class="mt-2 text-sm text-zinc-400">
      Forge only loads the current tree level and the selected blob preview, which keeps navigation responsive for larger repositories.
    </p>
  </div>

  <div v-else class="overflow-hidden rounded-2xl border border-zinc-800 bg-black/70 text-zinc-100 shadow-[0_20px_70px_rgba(0,0,0,0.45)]">
    <div class="flex flex-wrap items-center justify-between gap-3 border-b border-zinc-800 bg-zinc-950/80 px-5 py-4">
      <div>
        <p class="text-sm font-semibold">{{ blob.path }}</p>
        <p class="text-xs text-zinc-500">{{ formatBytes(blob.size_bytes) }} · {{ blob.language }}</p>
      </div>
      <p v-if="blob.truncated" class="rounded-md border border-amber-500/20 bg-amber-500/10 px-3 py-1 text-xs font-medium text-amber-300">
        Preview truncated to 256 KB
      </p>
    </div>

    <div v-if="blob.is_binary" class="p-8">
      <p class="eyebrow text-zinc-500">Binary Asset</p>
      <h3 class="mt-2 text-lg font-semibold text-white">This file is not rendered inline.</h3>
      <p class="mt-2 max-w-2xl text-sm text-zinc-400">
        Binary files are detected server-side so the browser view stays fast and avoids dumping unreadable data into the DOM.
      </p>
    </div>

    <div v-else class="grid max-h-[40rem] grid-cols-[auto_1fr] overflow-auto scrollbar-subtle text-sm">
      <div class="select-none border-r border-zinc-800 bg-zinc-950/80 px-4 py-4 text-right text-zinc-600">
        <div v-for="(_, index) in lines" :key="index" class="h-6 leading-6">
          {{ index + 1 }}
        </div>
      </div>
      <pre class="overflow-x-auto px-5 py-4"><code><div v-for="(line, index) in lines" :key="index" class="min-h-6 leading-6">{{ line || ' ' }}</div></code></pre>
    </div>
  </div>
</template>
