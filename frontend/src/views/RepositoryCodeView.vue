<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { computed, ref, watch } from 'vue'

import EmptyState from '@/components/empty/EmptyState.vue'
import PageHeader from '@/components/app/PageHeader.vue'
import CodePreview from '@/components/repo/CodePreview.vue'
import RepositoryTree from '@/components/repo/RepositoryTree.vue'
import Badge from '@/components/ui/Badge.vue'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import Select from '@/components/ui/Select.vue'
import Skeleton from '@/components/ui/Skeleton.vue'
import { useRepositoryWorkspace } from '@/composables/useRepositoryWorkspace'
import { api } from '@/lib/api'
import { basename, formatBytes } from '@/lib/utils'

const workspace = useRepositoryWorkspace()
const { branchModel, breadcrumbs, currentBranch, selectedFile, selectedPath } = workspace
const treeFilter = ref('')

const treeQuery = useQuery({
  queryKey: ['tree', workspace.owner, workspace.repo, workspace.selectedRef, workspace.selectedPath],
  queryFn: () =>
    api.tree(workspace.owner.value, workspace.repo.value, {
      ref: workspace.selectedRef.value,
      path: workspace.selectedPath.value,
    }),
  enabled: computed(() => workspace.repositoryQuery.isSuccess.value),
})

const blobQuery = useQuery({
  queryKey: ['blob', workspace.owner, workspace.repo, workspace.selectedRef, workspace.selectedFile],
  queryFn: () =>
    api.blob(workspace.owner.value, workspace.repo.value, {
      ref: workspace.selectedRef.value,
      path: workspace.selectedFile.value,
    }),
  enabled: computed(() => Boolean(workspace.selectedFile.value)),
})

const filteredEntries = computed(() => {
  const entries = treeQuery.data.value?.entries ?? []
  const filter = treeFilter.value.trim().toLowerCase()
  if (!filter) {
    return entries
  }

  return entries.filter((entry) => entry.name.toLowerCase().includes(filter))
})

const firstVisibleEntry = computed(() => filteredEntries.value[0] ?? null)
const selectedAssetLabel = computed(() => selectedFile.value || selectedPath.value || 'root')
const treeErrorMessage = computed(() => {
  if (!treeQuery.error.value) {
    return ''
  }
  return treeQuery.error.value instanceof Error
    ? treeQuery.error.value.message
    : 'Unable to load the selected tree.'
})
const blobErrorMessage = computed(() => {
  if (!blobQuery.error.value) {
    return ''
  }
  return blobQuery.error.value instanceof Error
    ? blobQuery.error.value.message
    : 'Unable to load the selected file.'
})

watch(
  () => treeQuery.data.value?.entries,
  (entries) => {
    if (selectedFile.value || !entries?.length) {
      return
    }

    const readme = entries.find((entry) => entry.type === 'blob' && /^readme/i.test(entry.name))
    if (readme) {
      void workspace.openFile(readme.path)
    }
  },
  { immediate: true },
)
</script>

<template>
  <div class="section-stack">
    <PageHeader
      eyebrow="Code"
      title="Code"
      description="Browse the current branch, filter the active tree, and preview one file at a time."
    >
      <template #actions>
        <Badge variant="accent">{{ currentBranch }}</Badge>
        <Badge>{{ selectedAssetLabel }}</Badge>
      </template>
    </PageHeader>

    <div class="grid gap-3 2xl:grid-cols-[320px_minmax(0,1fr)_300px]">
      <Card class="space-y-4">
        <div class="flex items-end gap-3">
          <div class="min-w-0 flex-1">
            <label class="field-label">Branch</label>
            <Select v-model="branchModel">
              <option v-for="branch in workspace.branchesQuery.data.value ?? []" :key="branch.name" :value="branch.name">
                {{ branch.name }}
              </option>
            </Select>
          </div>
          <div class="min-w-0 flex-1">
            <label class="field-label">Filter current tree</label>
            <Input v-model="treeFilter" placeholder="Search files or folders" />
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-2 text-sm text-zinc-400">
          <button
            type="button"
            class="rounded-md border border-zinc-800 bg-zinc-900 px-3 py-1 hover:bg-zinc-800"
            @click="workspace.openDirectory('')"
          >
            root
          </button>
          <button
            v-for="crumb in breadcrumbs"
            :key="crumb.value"
            type="button"
            class="rounded-md border border-zinc-800 bg-zinc-900 px-3 py-1 hover:bg-zinc-800"
            @click="workspace.openDirectory(crumb.value)"
          >
            {{ crumb.label }}
          </button>
        </div>

        <div v-if="treeQuery.isLoading.value" class="space-y-3">
          <div v-for="index in 6" :key="index" class="h-12 animate-pulse rounded-xl bg-zinc-900" />
        </div>
        <EmptyState
          v-else-if="treeErrorMessage"
          eyebrow="Tree Unavailable"
          title="This branch does not have a browsable tree yet."
          :description="treeErrorMessage"
        />
        <RepositoryTree
          v-else-if="filteredEntries.length"
          :entries="filteredEntries"
          :active-path="selectedFile || selectedPath"
          @select="workspace.handleTreeSelect"
        />
        <EmptyState
          v-else
          eyebrow="No Entries"
          title="Nothing matched this tree view."
          description="Adjust the filter or switch to another branch or path."
        />
      </Card>

      <div v-if="blobErrorMessage" class="rounded-2xl border border-red-500/30 bg-red-500/10 p-8">
        <p class="eyebrow text-red-300">Preview Error</p>
        <h3 class="mt-2 text-lg font-semibold text-red-100">The selected file could not be rendered.</h3>
        <p class="mt-2 text-sm leading-6 text-red-200/80">
          {{ blobErrorMessage }}
        </p>
      </div>
      <div v-else-if="blobQuery.isLoading.value && selectedFile" class="space-y-3 rounded-2xl border border-zinc-800 bg-black/70 p-5">
        <div class="flex items-center justify-between gap-3 border-b border-zinc-800 pb-4">
          <div class="space-y-2">
            <Skeleton class="h-4 w-72" />
            <Skeleton class="h-3 w-36" />
          </div>
          <Skeleton class="h-6 w-32" />
        </div>
        <div class="grid overflow-hidden rounded-xl border border-zinc-800 md:grid-cols-[auto_1fr]">
          <div class="space-y-2 border-r border-zinc-800 bg-zinc-950/80 px-4 py-4">
            <Skeleton v-for="index in 12" :key="`line-${index}`" class="h-4 w-6" />
          </div>
          <div class="space-y-2 px-5 py-4">
            <Skeleton v-for="index in 12" :key="`content-${index}`" class="h-4 w-full" />
          </div>
        </div>
      </div>
      <CodePreview v-else :blob="blobQuery.data.value?.blob ?? null" />

      <div class="space-y-3">
        <Card class="space-y-4">
          <div class="panel-header">
            <div>
            <p class="eyebrow">Navigation Context</p>
              <h3 class="mt-1 text-lg font-semibold text-zinc-50">Current selection</h3>
            </div>
          </div>
          <div class="grid gap-3 text-sm text-zinc-400">
            <div class="rounded-lg border border-zinc-800 bg-black/30 p-3">
              <p class="eyebrow">Selected Path</p>
              <p class="mt-2 font-mono text-xs text-zinc-200">{{ selectedPath || 'root' }}</p>
            </div>
            <div class="rounded-lg border border-zinc-800 bg-black/30 p-3">
              <p class="eyebrow">Selected File</p>
              <p class="mt-2 font-mono text-xs text-zinc-200">{{ selectedFile || 'No file selected' }}</p>
            </div>
            <div class="rounded-lg border border-zinc-800 bg-black/30 p-3">
              <p class="eyebrow">Visible Entries</p>
              <p class="mt-2 text-2xl font-semibold text-zinc-50">{{ filteredEntries.length }}</p>
            </div>
          </div>
        </Card>

        <Card class="space-y-4">
          <div class="panel-header">
            <div>
            <p class="eyebrow">Current Tree</p>
              <h3 class="mt-1 text-lg font-semibold text-zinc-50">{{ selectedPath || 'root' }}</h3>
            </div>
          </div>
          <p class="text-sm text-zinc-400">
            {{ filteredEntries.length }} item{{ filteredEntries.length === 1 ? '' : 's' }} visible in this path.
          </p>
          <div v-if="firstVisibleEntry" class="rounded-lg border border-zinc-800 bg-black/30 p-3 text-sm text-zinc-400">
            Next likely file: <span class="font-semibold text-zinc-100">{{ basename(firstVisibleEntry.path) }}</span>.
          </div>
          <div
            v-if="blobQuery.data.value?.blob"
            class="rounded-lg border border-zinc-800 bg-black/30 p-3 text-sm text-zinc-400"
          >
            <p class="eyebrow">Preview</p>
            <p class="mt-2 font-mono text-xs text-zinc-200">{{ blobQuery.data.value.blob.path }}</p>
            <p class="mt-2">Rendered size: {{ formatBytes(blobQuery.data.value.blob.size_bytes) }}</p>
          </div>
        </Card>
      </div>
    </div>
  </div>
</template>
