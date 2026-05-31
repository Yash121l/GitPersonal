<script setup lang="ts">
import PageHeader from '@/components/app/PageHeader.vue'
import Badge from '@/components/ui/Badge.vue'
import Card from '@/components/ui/Card.vue'
import { useRepositoryWorkspace } from '@/composables/useRepositoryWorkspace'
import { formatBytes, formatDate } from '@/lib/utils'

const workspace = useRepositoryWorkspace()
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Activity"
      title="Health and repository metadata stay readable."
      description="This section gives the shell a dedicated place for maintenance telemetry today and future commits, jobs, and audit feeds later."
    >
      <template #actions>
        <Badge variant="accent">{{ workspace.branchesQuery.data.value?.length ?? 0 }} branches</Badge>
      </template>
    </PageHeader>

    <div class="grid gap-4 xl:grid-cols-4">
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

    <div class="grid gap-4 xl:grid-cols-[1fr_0.95fr]">
      <Card class="space-y-4">
        <div>
          <p class="eyebrow">Branches</p>
          <h3 class="mt-2 text-2xl font-semibold text-zinc-50">Current branch inventory.</h3>
        </div>
        <div class="grid gap-3 md:grid-cols-2">
          <div
            v-for="branch in workspace.branchesQuery.data.value ?? []"
            :key="branch.name"
            class="rounded-xl border border-zinc-800 bg-black/30 p-4"
          >
            <p class="font-mono text-sm font-semibold text-zinc-100">{{ branch.name }}</p>
            <p class="mt-2 text-xs text-zinc-500">
              {{ branch.name === workspace.repositoryQuery.data.value?.repository.default_branch ? 'Default branch' : 'Available branch' }}
            </p>
          </div>
        </div>
      </Card>

      <Card class="space-y-4">
        <div>
          <p class="eyebrow">Operational Notes</p>
          <h3 class="mt-2 text-2xl font-semibold text-zinc-50">Reserved space for richer telemetry.</h3>
        </div>
        <div class="grid gap-3 text-sm text-zinc-400">
          <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
            This module is where commit activity, delivery history, and audit trails can land without forcing them into the code browser.
          </div>
          <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
            Repository sections are route-based, so heavier activity views can code-split and ship only when the user needs them.
          </div>
          <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
            Feature flags can enable future activity feeds or disable incomplete ones without changing the rest of the app shell.
          </div>
        </div>
      </Card>
    </div>
  </div>
</template>
