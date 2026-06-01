<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { Activity, FolderGit2, KeyRound, Landmark, MoveRight } from '@lucide/vue'
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

import { bootstrap } from '@/app/bootstrap'
import { featureCatalog } from '@/app/features'
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
      description="Recent repositories, organizations, SSH keys, and repository modules in one simple workspace view."
    >
      <template #actions>
        <Badge>{{ authStore.currentUser?.username }}</Badge>
      </template>
    </PageHeader>

    <ViewState
      :loading="workspaceLoading"
      :empty="workspaceEmpty"
      empty-eyebrow="Workspace"
      empty-title="This workspace is empty."
      empty-description="Create a repository, add an organization, or register an SSH key to get started."
      wrapper-class="section-stack"
    >
      <template #loading>
        <Card>
          <CardSkeletonGrid :count="4" wrapper-class="grid gap-4 md:grid-cols-2 xl:grid-cols-4" item-class="h-28" />
        </Card>
        <div class="grid gap-6 xl:grid-cols-[1.4fr_0.6fr]">
          <Card class="space-y-4">
            <div class="panel-header">
              <div class="space-y-2">
                <Skeleton class="h-4 w-36" />
                <Skeleton class="h-7 w-56" />
              </div>
              <Skeleton class="h-9 w-24" />
            </div>
            <CardSkeletonGrid :count="4" wrapper-class="grid gap-4" item-class="h-24" />
          </Card>
          <div class="space-y-6">
            <Card class="space-y-4">
              <Skeleton class="h-4 w-32" />
              <CardSkeletonGrid :count="4" wrapper-class="grid gap-3" item-class="h-16" />
            </Card>
            <Card class="space-y-4">
              <Skeleton class="h-4 w-40" />
              <div class="flex flex-wrap gap-2">
                <Skeleton v-for="index in 4" :key="index" class="h-7 w-24" />
              </div>
            </Card>
          </div>
        </div>
      </template>

      <template #empty>
        <Card class="space-y-6">
          <div class="space-y-2">
            <p class="eyebrow">Workspace</p>
            <h3 class="text-xl font-semibold text-zinc-50">Start by creating your first repository.</h3>
            <p class="text-sm leading-6 text-zinc-400">
              Once you add data, this screen will show recent repositories, organizations, SSH keys, and enabled repository modules.
            </p>
          </div>
          <div class="flex flex-wrap gap-3">
            <Button :as="RouterLink" :to="{ name: 'repositories' }">
              Open repositories
              <MoveRight class="size-4" />
            </Button>
            <Button :as="RouterLink" :to="{ name: 'keys' }" variant="secondary">
              Add SSH key
            </Button>
          </div>
        </Card>
      </template>

      <Card>
        <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <div class="space-y-2 border-b border-zinc-800 pb-4 md:border-b-0 md:border-r md:pb-0 md:pr-4">
            <p class="text-sm font-medium text-zinc-400">Repositories</p>
            <p class="text-3xl font-semibold text-zinc-50">{{ repositories.length }}</p>
            <p class="text-sm text-zinc-500">Visible to this account.</p>
          </div>
          <div class="space-y-2 border-b border-zinc-800 pb-4 md:border-b-0 xl:border-r xl:pb-0 xl:px-4">
            <p class="text-sm font-medium text-zinc-400">Organizations</p>
            <p class="text-3xl font-semibold text-zinc-50">{{ organizations.length }}</p>
            <p class="text-sm text-zinc-500">Shared namespaces.</p>
          </div>
          <div class="space-y-2 border-b border-zinc-800 pb-4 md:border-b-0 md:border-r md:pb-0 md:pr-4 xl:px-4">
            <p class="text-sm font-medium text-zinc-400">SSH Keys</p>
            <p class="text-3xl font-semibold text-zinc-50">{{ keys.length }}</p>
            <p class="text-sm text-zinc-500">Registered identities.</p>
          </div>
          <div class="space-y-2 md:pl-4">
            <p class="text-sm font-medium text-zinc-400">Repository Modules</p>
            <p class="text-3xl font-semibold text-zinc-50">{{ enabledRepositoryModules.length }}</p>
            <p class="text-sm text-zinc-500">Enabled repository tabs.</p>
          </div>
        </div>
      </Card>

      <div class="grid gap-6 xl:grid-cols-[1.4fr_0.6fr]">
        <Card class="space-y-4">
          <div class="panel-header">
            <div>
              <p class="eyebrow">Recent Repositories</p>
              <h3 class="mt-1 text-lg font-semibold text-zinc-50">Resume active codebases</h3>
            </div>
            <Button :as="RouterLink" :to="{ name: 'repositories' }" variant="secondary">
              All repositories
              <MoveRight class="size-4" />
            </Button>
          </div>

          <div class="divide-y divide-zinc-800">
            <RouterLink
              v-for="repository in recentRepositories"
              :key="`${repository.owner}/${repository.name}`"
              :to="{ name: 'repository-code', params: { owner: repository.owner, repo: repository.name } }"
              class="group block py-4 first:pt-0 last:pb-0"
            >
              <div class="flex items-start justify-between gap-3 transition group-hover:text-zinc-50">
                <div class="min-w-0">
                  <p class="truncate font-mono text-sm font-medium text-zinc-100 group-hover:text-zinc-50">
                    {{ repository.owner }}/{{ repository.name }}
                  </p>
                  <p class="mt-2 text-sm leading-6 text-zinc-400">
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

        <div class="space-y-6">
          <Card class="space-y-4">
            <div>
              <p class="eyebrow">Workspace</p>
              <h3 class="mt-1 text-lg font-semibold text-zinc-50">Primary sections</h3>
            </div>
            <div class="divide-y divide-zinc-800">
              <div class="flex items-start gap-3 py-3 first:pt-0 last:pb-0">
                <FolderGit2 class="mt-0.5 size-4 text-zinc-400" />
                <div>
                  <p class="text-sm font-medium text-zinc-100">Repositories</p>
                  <p class="text-sm text-zinc-500">Browse codebases and create new repositories.</p>
                </div>
              </div>
              <div class="flex items-start gap-3 py-3 first:pt-0 last:pb-0">
                <Landmark class="mt-0.5 size-4 text-zinc-400" />
                <div>
                  <p class="text-sm font-medium text-zinc-100">Organizations</p>
                  <p class="text-sm text-zinc-500">Manage shared namespaces and members.</p>
                </div>
              </div>
              <div class="flex items-start gap-3 py-3 first:pt-0 last:pb-0">
                <KeyRound class="mt-0.5 size-4 text-zinc-400" />
                <div>
                  <p class="text-sm font-medium text-zinc-100">SSH Keys</p>
                  <p class="text-sm text-zinc-500">Register keys for clone and push access.</p>
                </div>
              </div>
              <div class="flex items-start gap-3 py-3 first:pt-0 last:pb-0">
                <Activity class="mt-0.5 size-4 text-zinc-400" />
                <div>
                  <p class="text-sm font-medium text-zinc-100">Repository Modules</p>
                  <p class="text-sm text-zinc-500">Code, access, automation, and activity.</p>
                </div>
              </div>
            </div>
          </Card>

          <Card class="space-y-4">
            <div>
              <p class="eyebrow">Enabled Modules</p>
              <h3 class="mt-1 text-lg font-semibold text-zinc-50">Repository navigation</h3>
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
