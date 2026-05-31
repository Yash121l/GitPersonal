<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { Activity, FolderGit2, KeyRound, Landmark, MoveRight } from '@lucide/vue'
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

import { bootstrap } from '@/app/bootstrap'
import { featureCatalog } from '@/app/features'
import MetricCard from '@/components/app/MetricCard.vue'
import PageHeader from '@/components/app/PageHeader.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import { api } from '@/lib/api'
import { formatDate } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

const repositoriesQuery = useQuery({
  queryKey: ['repositories'],
  queryFn: () => api.listRepositories(),
})

const organizationsQuery = useQuery({
  queryKey: ['organizations'],
  queryFn: () => api.listOrganizations(),
})

const keysQuery = useQuery({
  queryKey: ['keys'],
  queryFn: () => api.listKeys(),
})

const repositories = computed(() => repositoriesQuery.data.value ?? [])
const organizations = computed(() => organizationsQuery.data.value ?? [])
const keys = computed(() => keysQuery.data.value ?? [])
const enabledRepositoryModules = computed(() =>
  Object.values(featureCatalog).filter(
    (feature) => feature.scope === 'repository' && bootstrap.features[feature.key],
  ),
)
const recentRepositories = computed(() =>
  [...repositories.value].sort((left, right) => right.updated_at.localeCompare(left.updated_at)).slice(0, 5),
)
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Overview"
      title="A repository workspace that scales with the product surface."
      description="Forge now has a stable workspace shell, repository-local sections, and feature-gated modules so new capabilities can be inserted without reworking the navigation model."
    >
      <template #actions>
        <Badge variant="accent">{{ authStore.currentUser?.username }}</Badge>
        <Badge>{{ bootstrap.productName }}</Badge>
      </template>
    </PageHeader>

    <div class="grid gap-4 xl:grid-cols-4">
      <MetricCard
        label="Accessible Repositories"
        :value="String(repositories.length)"
        caption="Personal, organization, and collaborator-visible repositories."
        accent="sky"
      />
      <MetricCard
        label="Organizations"
        :value="String(organizations.length)"
        caption="Shared ownership scopes available to this account."
      />
      <MetricCard
        label="SSH Keys"
        :value="String(keys.length)"
        caption="Developer identities registered for SSH transport."
      />
      <MetricCard
        label="Repository Modules"
        :value="String(enabledRepositoryModules.length)"
        caption="Feature-flagged sections available inside each repository."
        accent="emerald"
      />
    </div>

    <div class="grid gap-4 xl:grid-cols-[1.05fr_0.95fr]">
      <Card class="space-y-4">
        <div class="flex items-center justify-between gap-3">
          <div>
            <p class="eyebrow">Recent Repositories</p>
            <h3 class="mt-2 text-2xl font-semibold text-zinc-50">Jump back into active codebases.</h3>
          </div>
          <Button :as="RouterLink" :to="{ name: 'repositories' }" variant="secondary">
            All repositories
            <MoveRight class="size-4" />
          </Button>
        </div>

        <div v-if="recentRepositories.length" class="grid gap-3">
          <RouterLink
            v-for="repository in recentRepositories"
            :key="`${repository.owner}/${repository.name}`"
            :to="{ name: 'repository-code', params: { owner: repository.owner, repo: repository.name } }"
            class="rounded-xl border border-zinc-800 bg-black/30 p-5 transition hover:-translate-y-0.5 hover:border-zinc-700 hover:bg-zinc-950"
          >
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="font-mono text-base font-semibold text-zinc-50">{{ repository.owner }}/{{ repository.name }}</p>
                <p class="mt-2 text-sm text-zinc-400">
                  {{ repository.description || 'No description yet.' }}
                </p>
              </div>
              <Badge variant="accent">{{ repository.visibility }}</Badge>
            </div>
            <div class="mt-4 flex flex-wrap gap-4 text-xs text-zinc-500">
              <span>Default branch: {{ repository.default_branch }}</span>
              <span>Updated {{ formatDate(repository.updated_at) }}</span>
            </div>
          </RouterLink>
        </div>
        <div v-else class="rounded-xl border border-dashed border-zinc-800 bg-black/20 p-8 text-center">
          <p class="eyebrow">No Repositories</p>
          <h3 class="mt-2 text-lg font-semibold text-zinc-50">Nothing is visible yet.</h3>
          <p class="mt-2 text-sm text-zinc-400">
            Use the repository workspace to create your first repository or organization.
          </p>
        </div>
      </Card>

      <div class="space-y-4">
        <Card class="space-y-4">
          <div>
            <p class="eyebrow">Workspace Modules</p>
            <h3 class="mt-2 text-2xl font-semibold text-zinc-50">Stable navigation, expandable surface area.</h3>
          </div>
          <div class="grid gap-3">
            <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
              <div class="flex items-center gap-3">
                <FolderGit2 class="size-4 text-sky-300" />
                <div>
                  <p class="font-medium text-zinc-100">Repositories</p>
                  <p class="text-sm text-zinc-400">Primary source-control workflow with nested repository sections.</p>
                </div>
              </div>
            </div>
            <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
              <div class="flex items-center gap-3">
                <Landmark class="size-4 text-zinc-200" />
                <div>
                  <p class="font-medium text-zinc-100">Organizations</p>
                  <p class="text-sm text-zinc-400">Shared namespaces stay at workspace scope, not mixed into repo tabs.</p>
                </div>
              </div>
            </div>
            <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
              <div class="flex items-center gap-3">
                <KeyRound class="size-4 text-zinc-200" />
                <div>
                  <p class="font-medium text-zinc-100">SSH Keys</p>
                  <p class="text-sm text-zinc-400">Identity stays discoverable and separated from repository operations.</p>
                </div>
              </div>
            </div>
            <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
              <div class="flex items-center gap-3">
                <Activity class="size-4 text-emerald-300" />
                <div>
                  <p class="font-medium text-zinc-100">Feature Flags</p>
                  <p class="text-sm text-zinc-400">Future modules can be introduced or retired without rewiring the shell.</p>
                </div>
              </div>
            </div>
          </div>
        </Card>

        <Card class="space-y-4">
          <div class="flex items-center justify-between gap-3">
            <div>
              <p class="eyebrow">Enabled Repository Sections</p>
              <h3 class="mt-2 text-xl font-semibold text-zinc-50">Per-repository modules available today.</h3>
            </div>
            <Badge variant="accent">{{ enabledRepositoryModules.length }}</Badge>
          </div>
          <div class="flex flex-wrap gap-2">
            <Badge v-for="feature in enabledRepositoryModules" :key="feature.key">
              {{ feature.label }}
            </Badge>
          </div>
        </Card>
      </div>
    </div>
  </div>
</template>
