<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { BookCopy, FolderGit2, Landmark, MoveRight, Plus } from '@lucide/vue'
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

import EmptyState from '@/components/empty/EmptyState.vue'
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

const organizations = computed(() => organizationsQuery.data.value ?? [])
const repositories = computed(() => repositoriesQuery.data.value ?? [])
</script>

<template>
  <div class="section-stack">
    <PageHeader
      eyebrow="Repositories"
      title="Repositories"
      description="Browse visible repositories and create personal or organization-owned repos."
    >
      <template #actions>
        <Badge>{{ repositories.length }} repos</Badge>
        <Badge variant="accent">{{ authStore.currentUser?.role }}</Badge>
        <Button :as="RouterLink" :to="{ name: 'new-repository' }" size="sm">
          <Plus class="size-4" />
          New
        </Button>
      </template>
    </PageHeader>

    <div class="grid gap-6 xl:grid-cols-[1.35fr_0.65fr]">
      <div class="space-y-6">
        <Card>
          <div class="grid gap-4 md:grid-cols-3">
            <div class="space-y-2 border-b border-zinc-800 pb-4 md:border-b-0 md:border-r md:pb-0 md:pr-4">
              <p class="text-sm font-medium text-zinc-400">Repositories</p>
              <p class="text-3xl font-semibold text-zinc-50">{{ repositories.length }}</p>
              <p class="text-sm text-zinc-500">Visible to this account.</p>
            </div>
            <div class="space-y-2 border-b border-zinc-800 pb-4 md:border-b-0 md:border-r md:px-4 md:pb-0">
              <p class="text-sm font-medium text-zinc-400">Personal</p>
              <p class="text-3xl font-semibold text-zinc-50">
                {{ repositories.filter((repository) => repository.owner_type === 'user').length }}
              </p>
              <p class="text-sm text-zinc-500">Repositories in user namespaces.</p>
            </div>
            <div class="space-y-2 md:pl-4">
              <p class="text-sm font-medium text-zinc-400">Organizations</p>
              <p class="text-3xl font-semibold text-zinc-50">{{ organizations.length }}</p>
              <p class="text-sm text-zinc-500">Namespaces available for shared repos.</p>
            </div>
          </div>
        </Card>

        <Card class="space-y-4">
          <div class="panel-header">
            <div>
              <p class="eyebrow">Visible Repositories</p>
              <h3 class="mt-1 text-lg font-semibold text-zinc-50">
                {{ repositories.length }} {{ repositories.length === 1 ? 'repository' : 'repositories' }}
              </h3>
            </div>
            <Badge variant="accent">{{ authStore.currentUser?.role }}</Badge>
          </div>

          <div v-if="repositoriesQuery.isLoading.value" class="space-y-3">
            <div v-for="index in 4" :key="index" class="h-24 animate-pulse rounded-lg bg-zinc-900" />
          </div>
          <div v-else-if="repositories.length" class="divide-y divide-zinc-800">
            <RouterLink
              v-for="repository in repositories"
              :key="`${repository.owner}/${repository.name}`"
              :to="{ name: 'repository-code', params: { owner: repository.owner, repo: repository.name } }"
              class="group block py-4 first:pt-0 last:pb-0"
            >
              <div class="flex flex-wrap items-center gap-2">
                <Badge>{{ repository.owner_type }}</Badge>
                <Badge variant="accent">{{ repository.visibility }}</Badge>
              </div>
              <h3 class="mt-3 font-mono text-sm font-medium text-zinc-100 group-hover:text-zinc-50">
                {{ repository.owner }}/{{ repository.name }}
              </h3>
              <p class="mt-2 text-sm leading-6 text-zinc-400">
                {{ repository.description || 'No description yet.' }}
              </p>
              <div class="meta-list mt-3">
                <span>Default branch: {{ repository.default_branch }}</span>
                <span>Updated {{ formatDate(repository.updated_at) }}</span>
              </div>
            </RouterLink>
          </div>
          <EmptyState
            v-else
            eyebrow="No Repositories"
            title="Nothing is visible to this account yet."
            description="Create a personal repository or an organization-owned repository from the dedicated creation route."
          >
            <Button :as="RouterLink" :to="{ name: 'new-repository' }">
              <Plus class="size-4" />
              New repository
            </Button>
          </EmptyState>
        </Card>
      </div>

      <Card class="space-y-5">
        <div class="space-y-2">
          <p class="eyebrow">Routes</p>
          <h3 class="text-lg font-semibold text-zinc-50">Creation workflow</h3>
          <p class="text-sm leading-6 text-zinc-400">
            Creation now lives at first-class routes, with each action leading to a dedicated workflow.
          </p>
        </div>

        <RouterLink
          :to="{ name: 'new-repository' }"
          class="group block rounded-md border border-zinc-800 bg-zinc-950 p-4 hover:border-zinc-700 hover:bg-zinc-900/60"
        >
          <div class="flex items-start gap-3">
            <div class="flex size-9 items-center justify-center rounded-md border border-zinc-800 bg-zinc-900 text-zinc-300">
              <FolderGit2 class="size-4" />
            </div>
            <div class="min-w-0">
              <p class="text-sm font-semibold text-zinc-100 group-hover:text-zinc-50">Create repository</p>
              <p class="mt-1 text-sm leading-6 text-zinc-500">Personal repos and owner-scoped defaults.</p>
              <p class="mt-3 font-mono text-xs text-zinc-500">/new</p>
            </div>
            <MoveRight class="ml-auto mt-1 size-4 shrink-0 text-zinc-500 group-hover:text-zinc-100" />
          </div>
        </RouterLink>

        <RouterLink
          :to="{ name: 'new-organization' }"
          class="group block rounded-md border border-zinc-800 bg-zinc-950 p-4 hover:border-zinc-700 hover:bg-zinc-900/60"
        >
          <div class="flex items-start gap-3">
            <div class="flex size-9 items-center justify-center rounded-md border border-zinc-800 bg-zinc-900 text-zinc-300">
              <Landmark class="size-4" />
            </div>
            <div class="min-w-0">
              <p class="text-sm font-semibold text-zinc-100 group-hover:text-zinc-50">Create organization</p>
              <p class="mt-1 text-sm leading-6 text-zinc-500">A shared namespace for teams and repos.</p>
              <p class="mt-3 font-mono text-xs text-zinc-500">/organizations/new</p>
            </div>
            <MoveRight class="ml-auto mt-1 size-4 shrink-0 text-zinc-500 group-hover:text-zinc-100" />
          </div>
        </RouterLink>

        <RouterLink
          :to="{ name: 'new-repository' }"
          class="group block rounded-md border border-zinc-800 bg-zinc-950 p-4 hover:border-zinc-700 hover:bg-zinc-900/60"
        >
          <div class="flex items-start gap-3">
            <div class="flex size-9 items-center justify-center rounded-md border border-zinc-800 bg-zinc-900 text-zinc-300">
              <BookCopy class="size-4" />
            </div>
            <div class="min-w-0">
              <p class="text-sm font-semibold text-zinc-100 group-hover:text-zinc-50">Create shared repository</p>
              <p class="mt-1 text-sm leading-6 text-zinc-500">
                Requires an organization; {{ organizations.length }} available.
              </p>
              <p class="mt-3 font-mono text-xs text-zinc-500">/new</p>
            </div>
            <MoveRight class="ml-auto mt-1 size-4 shrink-0 text-zinc-500 group-hover:text-zinc-100" />
          </div>
        </RouterLink>
      </Card>
    </div>
  </div>
</template>
