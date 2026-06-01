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
          <Skeleton class="h-6 w-16" />
          <Skeleton class="h-6 w-20" />
          <Skeleton class="h-6 w-16" />
        </div>
        <div class="space-y-3">
          <Skeleton class="h-8 w-72" />
          <Skeleton class="h-4 w-full max-w-3xl" />
        </div>
        <div class="grid gap-4 md:grid-cols-3">
          <Skeleton class="h-24" />
          <Skeleton class="h-24" />
          <Skeleton class="h-24" />
        </div>
      </Card>
    </template>

    <Card class="space-y-6">
      <div class="flex flex-wrap items-center gap-2">
        <Badge v-if="workspace.repositoryQuery.data.value">{{ workspace.repositoryQuery.data.value.repository.owner_type }}</Badge>
        <Badge v-if="workspace.repositoryQuery.data.value" variant="accent">
          {{ workspace.repositoryQuery.data.value.repository.visibility }}
        </Badge>
        <Badge v-if="workspace.repositoryQuery.data.value" variant="warning">
          {{ workspace.repositoryQuery.data.value.repository.default_branch }}
        </Badge>
      </div>

      <div class="space-y-3">
        <div class="flex flex-wrap items-center gap-3">
          <h2 class="font-mono text-2xl font-semibold text-zinc-50 md:text-3xl">{{ owner }}/{{ repo }}</h2>
          <Badge>{{ currentSection?.label || 'Repository' }}</Badge>
        </div>
        <p class="max-w-3xl text-sm leading-6 text-zinc-400">
          {{
            workspace.repositoryQuery.data.value?.repository.description ||
            'No description has been added for this repository yet.'
          }}
        </p>
      </div>

      <div class="grid gap-4 border-y border-zinc-800 py-4 md:grid-cols-3 md:divide-x md:divide-zinc-800 md:gap-0">
        <div class="space-y-2 md:pr-4">
          <p class="text-sm font-medium text-zinc-400">Default branch</p>
          <p class="font-mono text-sm text-zinc-100">{{ currentBranch }}</p>
        </div>
        <div class="space-y-2 md:px-4">
          <p class="text-sm font-medium text-zinc-400">Branch count</p>
          <p class="text-2xl font-semibold text-zinc-50">{{ workspace.branchesQuery.data.value?.length ?? 0 }}</p>
        </div>
        <div class="space-y-2 md:pl-4">
          <p class="text-sm font-medium text-zinc-400">Repository size</p>
          <p class="text-2xl font-semibold text-zinc-50">
            {{ formatBytes(workspace.repositoryQuery.data.value?.repository.size_bytes) }}
          </p>
        </div>
      </div>

      <div class="space-y-3 border-t border-zinc-800 pt-4">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p class="text-sm font-medium text-zinc-300">Clone URL</p>
            <p class="mt-1 overflow-x-auto font-mono text-sm text-zinc-400">
              {{ workspace.repositoryQuery.data.value?.http_clone_url || 'Unavailable' }}
            </p>
          </div>
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
      </div>
    </Card>

    <RouterView />
  </ViewState>
</template>
