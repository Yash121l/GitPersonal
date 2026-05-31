<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { Activity, FolderGit2, KeyRound, Landmark, MoveRight } from '@lucide/vue'
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

import { bootstrap } from '@/app/bootstrap'
import { featureCatalog } from '@/app/features'
import MetricCard from '@/components/app/MetricCard.vue'
import PageHeader from '@/components/app/PageHeader.vue'
import CardSkeletonGrid from '@/components/state/CardSkeletonGrid.vue'
import ViewState from '@/components/state/ViewState.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import Skeleton from '@/components/ui/Skeleton.vue'
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
const workspaceLoading = computed(
  () =>
    repositoriesQuery.isLoading.value ||
    organizationsQuery.isLoading.value ||
    keysQuery.isLoading.value,
)
const workspaceEmpty = computed(
  () =>
    !workspaceLoading.value &&
    repositories.value.length === 0 &&
    organizations.value.length === 0 &&
    keys.value.length === 0,
)
const recentRepositories = computed(() =>
  [...repositories.value].sort((left, right) => right.updated_at.localeCompare(left.updated_at)).slice(0, 5),
)
</script>

<template>
  <div class="section-stack">
    <PageHeader
      eyebrow="Overview"
      title="Overview"
      description="Recent repositories, shared ownership, SSH access, and enabled repository modules."
    >
      <template #actions>
        <Badge variant="accent">{{ authStore.currentUser?.username }}</Badge>
        <Badge>{{ bootstrap.productName }}</Badge>
      </template>
    </PageHeader>

    <ViewState
      :loading="workspaceLoading"
      :empty="workspaceEmpty"
      empty-eyebrow="Workspace"
      empty-title="This workspace has not been populated yet."
      empty-description="Create a repository, join an organization, or register an SSH key to turn on the dashboard."
      wrapper-class="section-stack"
    >
      <template #loading>
        <CardSkeletonGrid :count="4" wrapper-class="grid gap-3 xl:grid-cols-4" item-class="h-28" />
        <div class="grid gap-3 xl:grid-cols-[1.2fr_0.8fr]">
          <Card class="space-y-3">
            <div class="panel-header">
              <div class="space-y-2">
                <Skeleton class="h-3 w-32" />
                <Skeleton class="h-7 w-52" />
              </div>
              <Skeleton class="h-9 w-24" />
            </div>
            <CardSkeletonGrid :count="4" wrapper-class="grid gap-3" item-class="h-24" />
          </Card>
          <div class="space-y-3">
            <Card class="space-y-3">
              <div class="panel-header">
                <div class="space-y-2">
                  <Skeleton class="h-3 w-32" />
                  <Skeleton class="h-7 w-44" />
                </div>
                <Skeleton class="h-6 w-20" />
              </div>
              <CardSkeletonGrid :count="4" wrapper-class="grid gap-2.5" item-class="h-18" />
            </Card>
            <Card class="space-y-3">
              <div class="panel-header">
                <div class="space-y-2">
                  <Skeleton class="h-3 w-44" />
                  <Skeleton class="h-7 w-56" />
                </div>
                <Skeleton class="h-6 w-12" />
              </div>
              <div class="flex flex-wrap gap-2">
                <Skeleton v-for="index in 4" :key="index" class="h-8 w-24" />
              </div>
            </Card>
          </div>
        </div>
      </template>

      <template #empty>
        <div class="grid gap-3 xl:grid-cols-[1.08fr_0.92fr]">
          <Card class="space-y-3">
            <div class="panel-header">
              <div>
                <p class="eyebrow">Workspace Bootstrapping</p>
                <h3 class="mt-1 text-lg font-semibold text-zinc-50">Start by creating the first namespace.</h3>
              </div>
              <Button :as="RouterLink" :to="{ name: 'repositories' }">
                Open repos
                <MoveRight class="size-4" />
              </Button>
            </div>
            <div class="grid gap-3 md:grid-cols-3">
              <MetricCard label="Repositories" value="0" caption="Visible codebases." accent="sky" />
              <MetricCard label="Organizations" value="0" caption="Shared namespaces." />
              <MetricCard label="SSH Keys" value="0" caption="Registered identities." />
            </div>
          </Card>

          <Card class="space-y-3">
            <div class="panel-header">
              <div>
                <p class="eyebrow">Enabled Repository Sections</p>
                <h3 class="mt-1 text-lg font-semibold text-zinc-50">Feature-flagged surfaces.</h3>
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
      </template>

      <div class="grid gap-3 xl:grid-cols-4">
        <MetricCard
          label="Repositories"
          :value="String(repositories.length)"
          caption="Visible to this account."
          accent="sky"
        />
        <MetricCard
          label="Organizations"
          :value="String(organizations.length)"
          caption="Shared namespaces."
        />
        <MetricCard
          label="SSH Keys"
          :value="String(keys.length)"
          caption="Registered identities."
        />
        <MetricCard
          label="Repo Modules"
          :value="String(enabledRepositoryModules.length)"
          caption="Feature-flagged tabs."
          accent="emerald"
        />
      </div>

      <div class="grid gap-3 xl:grid-cols-[1.2fr_0.8fr]">
        <Card class="space-y-3">
          <div class="panel-header">
            <div>
              <p class="eyebrow">Recent Repositories</p>
              <h3 class="mt-1 text-lg font-semibold text-zinc-50">Resume active codebases.</h3>
            </div>
            <Button :as="RouterLink" :to="{ name: 'repositories' }" variant="secondary">
              All repos
              <MoveRight class="size-4" />
            </Button>
          </div>

          <div class="grid gap-3">
            <RouterLink
              v-for="repository in recentRepositories"
              :key="`${repository.owner}/${repository.name}`"
              :to="{ name: 'repository-code', params: { owner: repository.owner, repo: repository.name } }"
              class="rounded-lg border border-zinc-800 bg-black/30 p-4 transition hover:border-zinc-700 hover:bg-zinc-950"
            >
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="font-mono text-base font-semibold text-zinc-50">{{ repository.owner }}/{{ repository.name }}</p>
                  <p class="mt-1.5 text-sm text-zinc-400">
                    {{ repository.description || 'No description yet.' }}
                  </p>
                </div>
                <Badge variant="accent">{{ repository.visibility }}</Badge>
              </div>
              <div class="meta-list mt-3">
                <span>Default branch: {{ repository.default_branch }}</span>
                <span>Updated {{ formatDate(repository.updated_at) }}</span>
              </div>
            </RouterLink>
          </div>
        </Card>

        <div class="space-y-3">
          <Card class="space-y-3">
            <div class="panel-header">
              <div>
                <p class="eyebrow">Workspace Surfaces</p>
                <h3 class="mt-1 text-lg font-semibold text-zinc-50">Primary control surfaces.</h3>
              </div>
              <Badge>{{ enabledRepositoryModules.length }} repo tabs</Badge>
            </div>
            <div class="grid gap-2.5">
              <div class="rounded-lg border border-zinc-800 bg-black/30 p-3">
                <div class="flex items-center gap-3">
                  <FolderGit2 class="size-4 text-sky-300" />
                  <div>
                    <p class="font-medium text-zinc-100">Repositories</p>
                    <p class="text-xs leading-5 text-zinc-400">Primary source-control workflow.</p>
                  </div>
                </div>
              </div>
              <div class="rounded-lg border border-zinc-800 bg-black/30 p-3">
                <div class="flex items-center gap-3">
                  <Landmark class="size-4 text-zinc-200" />
                  <div>
                    <p class="font-medium text-zinc-100">Organizations</p>
                    <p class="text-xs leading-5 text-zinc-400">Shared ownership and roles.</p>
                  </div>
                </div>
              </div>
              <div class="rounded-lg border border-zinc-800 bg-black/30 p-3">
                <div class="flex items-center gap-3">
                  <KeyRound class="size-4 text-zinc-200" />
                  <div>
                    <p class="font-medium text-zinc-100">SSH Keys</p>
                    <p class="text-xs leading-5 text-zinc-400">Identity for SSH transport.</p>
                  </div>
                </div>
              </div>
              <div class="rounded-lg border border-zinc-800 bg-black/30 p-3">
                <div class="flex items-center gap-3">
                  <Activity class="size-4 text-emerald-300" />
                  <div>
                    <p class="font-medium text-zinc-100">Feature Flags</p>
                    <p class="text-xs leading-5 text-zinc-400">Ship new surfaces safely.</p>
                  </div>
                </div>
              </div>
            </div>
          </Card>

          <Card class="space-y-3">
            <div class="panel-header">
              <div>
                <p class="eyebrow">Enabled Repository Sections</p>
                <h3 class="mt-1 text-lg font-semibold text-zinc-50">Available inside a repository.</h3>
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
    </ViewState>
  </div>
</template>
