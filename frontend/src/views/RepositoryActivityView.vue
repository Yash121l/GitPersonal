<script setup lang="ts">
import { computed } from 'vue'

import EmptyState from '@/components/empty/EmptyState.vue'
import PageHeader from '@/components/app/PageHeader.vue'
import Badge from '@/components/ui/Badge.vue'
import Card from '@/components/ui/Card.vue'
import Skeleton from '@/components/ui/Skeleton.vue'
import { useRepositoryWorkspace } from '@/composables/useRepositoryWorkspace'
import { formatBytes, formatDate } from '@/lib/utils'

const workspace = useRepositoryWorkspace()
const activityLoading = computed(
  () =>
    workspace.repositoryQuery.isLoading.value ||
    (workspace.branchesQuery.isLoading.value && !workspace.branchesQuery.data.value),
)
</script>

<template>
  <div class="section-stack">
    <PageHeader
      eyebrow="Activity"
      title="Activity"
      description="Repository health, branch inventory, and maintenance timestamps."
    >
      <template #actions>
        <Badge variant="accent">{{ workspace.branchesQuery.data.value?.length ?? 0 }} branches</Badge>
      </template>
    </PageHeader>

    <template v-if="activityLoading">
      <div class="grid gap-3 xl:grid-cols-4">
        <Skeleton v-for="index in 4" :key="index" class="h-28" />
      </div>
      <div class="grid gap-3 xl:grid-cols-[1fr_0.95fr]">
        <Card class="space-y-4">
          <div class="panel-header">
            <div class="space-y-2">
              <Skeleton class="h-3 w-20" />
              <Skeleton class="h-7 w-36" />
            </div>
          </div>
          <div class="grid gap-3 md:grid-cols-2">
            <Skeleton v-for="index in 4" :key="index" class="h-20" />
          </div>
        </Card>
        <Card class="space-y-4">
          <div class="panel-header">
            <div class="space-y-2">
              <Skeleton class="h-3 w-28" />
              <Skeleton class="h-7 w-44" />
            </div>
          </div>
          <div class="grid gap-3">
            <Skeleton v-for="index in 3" :key="index" class="h-16" />
          </div>
        </Card>
      </div>
    </template>

    <template v-else>
      <div class="grid gap-3 xl:grid-cols-4">
        <Card>
          <p class="eyebrow">Repository Size</p>
          <p class="mt-3 text-3xl font-semibold tracking-tight text-zinc-50">
            {{ formatBytes(workspace.repositoryQuery.data.value?.repository.size_bytes) }}
          </p>
        </Card>
        <Card>
          <p class="eyebrow">Default Branch</p>
          <p class="mt-3 text-3xl font-semibold tracking-tight text-zinc-50">
            {{ workspace.repositoryQuery.data.value?.repository.default_branch }}
          </p>
        </Card>
        <Card>
          <p class="eyebrow">Indexed</p>
          <p class="mt-3 text-lg font-semibold tracking-tight text-zinc-50">
            {{ formatDate(workspace.repositoryQuery.data.value?.repository.last_indexed_at) }}
          </p>
        </Card>
        <Card>
          <p class="eyebrow">Maintained</p>
          <p class="mt-3 text-lg font-semibold tracking-tight text-zinc-50">
            {{ formatDate(workspace.repositoryQuery.data.value?.repository.last_maintained_at) }}
          </p>
        </Card>
      </div>

      <div class="grid gap-3 xl:grid-cols-[1fr_0.95fr]">
        <Card class="space-y-4">
          <div class="panel-header">
            <div>
              <p class="eyebrow">Branches</p>
              <h3 class="mt-1 text-lg font-semibold text-zinc-50">Branch inventory</h3>
            </div>
          </div>
          <div v-if="workspace.branchesQuery.data.value?.length" class="grid gap-3 md:grid-cols-2">
            <div
              v-for="branch in workspace.branchesQuery.data.value ?? []"
              :key="branch.name"
              class="rounded-lg border border-zinc-800 bg-black/30 p-3"
            >
              <p class="font-mono text-sm font-semibold text-zinc-100">{{ branch.name }}</p>
              <p class="mt-2 text-xs text-zinc-500">
                {{ branch.name === workspace.repositoryQuery.data.value?.repository.default_branch ? 'Default branch' : 'Available branch' }}
              </p>
            </div>
          </div>
          <EmptyState
            v-else
            eyebrow="No Branches"
            title="Branch inventory is empty."
            description="Push the first branch or reindex the repository to populate this activity surface."
          />
        </Card>

        <Card class="space-y-4">
          <div class="panel-header">
            <div>
              <p class="eyebrow">Operational Notes</p>
              <h3 class="mt-1 text-lg font-semibold text-zinc-50">Future telemetry surface</h3>
            </div>
          </div>
          <div class="grid gap-3 text-sm text-zinc-400">
            <div class="rounded-lg border border-zinc-800 bg-black/30 p-3">
              Commit activity, delivery history, and audit trails can land here later.
            </div>
            <div class="rounded-lg border border-zinc-800 bg-black/30 p-3">
              Heavier activity views can stay route-based and code-split.
            </div>
            <div class="rounded-lg border border-zinc-800 bg-black/30 p-3">
              Feature flags can enable or hide future activity modules.
            </div>
          </div>
        </Card>
      </div>
    </template>
  </div>
</template>
