<script setup lang="ts">
import { computed, watchEffect } from 'vue'
import { RouterView, useRoute } from 'vue-router'

import { getRepositoryNavigation, isNavigationItemActive } from '@/app/navigation'
import EmptyState from '@/components/empty/EmptyState.vue'
import ViewState from '@/components/state/ViewState.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import Skeleton from '@/components/ui/Skeleton.vue'
import { provideRepositoryWorkspace } from '@/composables/useRepositoryWorkspace'
import { formatBytes } from '@/lib/utils'

const route = useRoute()
const workspace = provideRepositoryWorkspace()
const { owner, repo, currentBranch } = workspace

const repositoryNavigation = computed(() => getRepositoryNavigation())
const currentSection = computed(
  () => repositoryNavigation.value.find((item) => isNavigationItemActive(item, route)) ?? repositoryNavigation.value[0] ?? null,
)
const repositoryLoading = computed(
  () => workspace.repositoryQuery.isLoading.value && !workspace.repositoryQuery.data.value,
)

watchEffect(() => {
  const repository = workspace.repositoryQuery.data.value?.repository
  if (!repository) {
    return
  }

  const sectionTitle = currentSection.value?.label ?? 'Repository'
  document.title = `${repository.owner}/${repository.name} · ${sectionTitle} · Forge`
})
</script>

<template>
  <div v-if="workspace.repositoryQuery.isError.value" class="space-y-6">
    <EmptyState
      eyebrow="Repository"
      title="This repository could not be loaded."
      :description="
        workspace.repositoryQuery.error.value instanceof Error
          ? workspace.repositoryQuery.error.value.message
          : 'The repository may not exist or this account may not have access.'
      "
    />
  </div>

  <ViewState v-else :loading="repositoryLoading" :empty="false" wrapper-class="section-stack">
    <template #loading>
      <Card class="space-y-4">
        <div class="flex flex-wrap gap-2">
          <Skeleton class="h-6 w-20" />
          <Skeleton class="h-6 w-20" />
          <Skeleton class="h-6 w-20" />
        </div>
        <div class="grid gap-3 xl:grid-cols-[1.35fr_0.65fr]">
          <div class="space-y-3">
            <Skeleton class="h-8 w-72" />
            <Skeleton class="h-4 w-full max-w-3xl" />
            <Skeleton class="h-4 w-2/3 max-w-2xl" />
          </div>
          <div class="grid gap-3 md:grid-cols-3 xl:grid-cols-1">
            <Skeleton class="h-20" />
            <Skeleton class="h-20" />
            <Skeleton class="h-20" />
          </div>
        </div>
      </Card>
    </template>

    <Card class="space-y-4">
      <div class="flex flex-wrap items-center gap-2">
        <Badge v-if="workspace.repositoryQuery.data.value">{{ workspace.repositoryQuery.data.value.repository.owner_type }}</Badge>
        <Badge v-if="workspace.repositoryQuery.data.value" variant="accent">
          {{ workspace.repositoryQuery.data.value.repository.visibility }}
        </Badge>
        <Badge v-if="workspace.repositoryQuery.data.value" variant="warning">
          {{ workspace.repositoryQuery.data.value.repository.default_branch }}
        </Badge>
      </div>

      <div class="grid gap-3 xl:grid-cols-[1.35fr_0.65fr]">
        <div>
          <div class="flex flex-wrap items-center gap-2">
            <h2 class="font-mono text-xl font-semibold tracking-tight text-zinc-50 md:text-2xl">
              {{ owner }}/{{ repo }}
            </h2>
            <span class="rounded-md border border-zinc-800 bg-black/30 px-2 py-1 text-[10px] uppercase tracking-[0.18em] text-zinc-500">
              {{ currentSection?.label || 'Repository' }}
            </span>
          </div>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-zinc-400">
            {{
              workspace.repositoryQuery.data.value?.repository.description ||
              'No description has been added for this repository yet.'
            }}
          </p>
          <div class="meta-list mt-3">
            <span>{{ repositoryNavigation.length }} enabled modules</span>
            <span>Branch context: {{ currentBranch }}</span>
          </div>
        </div>

        <div class="grid gap-3 md:grid-cols-3 xl:grid-cols-1">
          <div class="rounded-lg border border-zinc-800 bg-black/30 p-3">
            <p class="eyebrow">Default Branch</p>
            <p class="mt-2 font-mono text-sm text-zinc-200">{{ currentBranch }}</p>
          </div>
          <div class="rounded-lg border border-zinc-800 bg-black/30 p-3">
            <p class="eyebrow">Branch Count</p>
            <p class="mt-2 text-xl font-semibold text-zinc-50">{{ workspace.branchesQuery.data.value?.length ?? 0 }}</p>
          </div>
          <div class="rounded-lg border border-zinc-800 bg-black/30 p-3">
            <p class="eyebrow">Repository Size</p>
            <p class="mt-2 text-sm font-semibold text-zinc-100">
              {{ formatBytes(workspace.repositoryQuery.data.value?.repository.size_bytes) }}
            </p>
          </div>
        </div>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-3 border-t border-zinc-800/80 pt-3">
        <p class="text-sm text-zinc-400">
          Repository modules live in the left sidebar so code, access, automation, activity, and settings stay in one consistent map.
        </p>
        <Button
          v-if="workspace.repositoryQuery.data.value"
          :as="'a'"
          :href="workspace.repositoryQuery.data.value.http_clone_url"
          target="_blank"
          rel="noreferrer"
          variant="secondary"
        >
          Open clone URL
        </Button>
      </div>
    </Card>

    <RouterView />
  </ViewState>
</template>
