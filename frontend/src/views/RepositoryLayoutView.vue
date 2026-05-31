<script setup lang="ts">
import { computed, watchEffect } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'

import { getRepositoryNavigation, isNavigationItemActive } from '@/app/navigation'
import EmptyState from '@/components/empty/EmptyState.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import { provideRepositoryWorkspace } from '@/composables/useRepositoryWorkspace'
import { formatBytes } from '@/lib/utils'

const route = useRoute()
const workspace = provideRepositoryWorkspace()
const { owner, repo, currentBranch } = workspace

const repositoryNavigation = computed(() => getRepositoryNavigation())
const currentSection = computed(
  () => repositoryNavigation.value.find((item) => isNavigationItemActive(item, route)) ?? repositoryNavigation.value[0] ?? null,
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

  <div v-else class="space-y-6">
    <Card class="space-y-5">
      <div class="flex flex-wrap items-center gap-2">
        <Badge v-if="workspace.repositoryQuery.data.value">{{ workspace.repositoryQuery.data.value.repository.owner_type }}</Badge>
        <Badge v-if="workspace.repositoryQuery.data.value" variant="accent">
          {{ workspace.repositoryQuery.data.value.repository.visibility }}
        </Badge>
        <Badge v-if="workspace.repositoryQuery.data.value" variant="warning">
          {{ workspace.repositoryQuery.data.value.repository.default_branch }}
        </Badge>
      </div>

      <div class="grid gap-4 xl:grid-cols-[1.25fr_0.75fr]">
        <div>
          <h2 class="font-mono text-3xl font-semibold tracking-tight text-zinc-50">
            {{ owner }}/{{ repo }}
          </h2>
          <p class="mt-3 max-w-3xl text-sm leading-7 text-zinc-400">
            {{
              workspace.repositoryQuery.data.value?.repository.description ||
              'No description has been added for this repository yet.'
            }}
          </p>
        </div>

        <div class="grid gap-3 md:grid-cols-3 xl:grid-cols-1">
          <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
            <p class="eyebrow">Default Branch</p>
            <p class="mt-2 font-mono text-sm text-zinc-200">{{ currentBranch }}</p>
          </div>
          <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
            <p class="eyebrow">Branch Count</p>
            <p class="mt-2 text-2xl font-semibold text-zinc-50">{{ workspace.branchesQuery.data.value?.length ?? 0 }}</p>
          </div>
          <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
            <p class="eyebrow">Repository Size</p>
            <p class="mt-2 text-sm font-semibold text-zinc-100">
              {{ formatBytes(workspace.repositoryQuery.data.value?.repository.size_bytes) }}
            </p>
          </div>
        </div>
      </div>

      <div class="grid gap-4 xl:grid-cols-[1fr_auto] xl:items-center">
        <nav class="overflow-x-auto">
          <div class="flex min-w-max gap-2">
            <RouterLink
              v-for="item in repositoryNavigation"
              :key="item.id"
              :to="item.to({ owner: workspace.owner.value, repo: workspace.repo.value, currentQuery: route.query })"
              :class="
                [
                  'inline-flex items-center gap-2 rounded-lg border px-4 py-2 text-sm font-medium transition',
                  isNavigationItemActive(item, route)
                    ? 'border-zinc-700 bg-zinc-100 text-zinc-950'
                    : 'border-zinc-800 bg-black/20 text-zinc-400 hover:border-zinc-700 hover:bg-zinc-900 hover:text-zinc-100',
                ]
              "
            >
              <component :is="item.icon" class="size-4" />
              {{ item.label }}
            </RouterLink>
          </div>
        </nav>

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
  </div>
</template>
