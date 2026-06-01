<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { FolderGit2, Search, UserRound } from '@lucide/vue'
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'

import EmptyState from '@/components/empty/EmptyState.vue'
import PageHeader from '@/components/app/PageHeader.vue'
import Badge from '@/components/ui/Badge.vue'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import { api } from '@/lib/api'
import { formatDate } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const search = ref('')

const repositoriesQuery = useQuery({
  queryKey: ['repositories'],
  queryFn: () => api.listRepositories(),
})

const organizationsQuery = useQuery({
  queryKey: ['organizations'],
  queryFn: () => api.listOrganizations(),
})

const repositories = computed(() => repositoriesQuery.data.value ?? [])
const organizations = computed(() => organizationsQuery.data.value ?? [])
const normalizedSearch = computed(() => search.value.trim().toLowerCase())
const publicRepositories = computed(() =>
  repositories.value.filter((repository) => repository.visibility === 'public'),
)
const users = computed(() => {
  const names = new Set<string>()
  if (authStore.currentUser?.username) {
    names.add(authStore.currentUser.username)
  }
  for (const repository of repositories.value) {
    if (repository.owner_type === 'user') {
      names.add(repository.owner)
    }
  }
  for (const membership of organizations.value) {
    if (membership.username) {
      names.add(membership.username)
    }
  }
  return [...names].sort((left, right) => left.localeCompare(right))
})
const filteredRepositories = computed(() => {
  const query = normalizedSearch.value
  if (!query) {
    return publicRepositories.value
  }
  return publicRepositories.value.filter((repository) =>
    `${repository.owner}/${repository.name} ${repository.description}`.toLowerCase().includes(query),
  )
})
const filteredUsers = computed(() => {
  const query = normalizedSearch.value
  if (!query) {
    return users.value
  }
  return users.value.filter((username) => username.toLowerCase().includes(query))
})
</script>

<template>
  <div class="section-stack">
    <PageHeader eyebrow="Explore" title="Explore" description="Search public repositories, users, and shared namespaces.">
      <template #actions>
        <Badge>{{ filteredRepositories.length }} public repos</Badge>
        <Badge variant="accent">{{ filteredUsers.length }} users</Badge>
      </template>
    </PageHeader>

    <Card class="space-y-4">
      <label class="field-label" for="explore-search">Search</label>
      <div class="relative">
        <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-zinc-500" />
        <Input id="explore-search" v-model="search" class="pl-9" placeholder="Search users or public repositories" />
      </div>
    </Card>

    <div class="grid gap-6 xl:grid-cols-[1.25fr_0.75fr]">
      <Card class="space-y-4">
        <div class="panel-header">
          <div>
            <p class="eyebrow">Public Repositories</p>
            <h3 class="mt-1 text-lg font-semibold text-zinc-50">{{ filteredRepositories.length }} results</h3>
          </div>
          <FolderGit2 class="size-5 text-zinc-500" />
        </div>

        <div v-if="repositoriesQuery.isLoading.value" class="space-y-3">
          <div v-for="index in 4" :key="index" class="h-24 animate-pulse rounded-md bg-zinc-900" />
        </div>
        <div v-else-if="filteredRepositories.length" class="divide-y divide-zinc-800">
          <RouterLink
            v-for="repository in filteredRepositories"
            :key="`${repository.owner}/${repository.name}`"
            :to="{ name: 'repository-code', params: { owner: repository.owner, repo: repository.name } }"
            class="group block py-4 first:pt-0 last:pb-0"
          >
            <div class="flex flex-wrap items-center gap-2">
              <h3 class="font-mono text-sm font-medium text-zinc-100 group-hover:text-zinc-50">
                {{ repository.owner }}/{{ repository.name }}
              </h3>
              <Badge>{{ repository.owner_type }}</Badge>
              <Badge variant="accent">{{ repository.visibility }}</Badge>
            </div>
            <p class="mt-2 text-sm leading-6 text-zinc-400">{{ repository.description || 'No description yet.' }}</p>
            <div class="meta-list mt-3">
              <span>{{ repository.default_branch }}</span>
              <span>Updated {{ formatDate(repository.updated_at) }}</span>
            </div>
          </RouterLink>
        </div>
        <EmptyState
          v-else
          eyebrow="No Results"
          title="No public repositories matched."
          description="Try a different search term or create a public repository."
        />
      </Card>

      <Card class="space-y-4">
        <div class="panel-header">
          <div>
            <p class="eyebrow">Users</p>
            <h3 class="mt-1 text-lg font-semibold text-zinc-50">{{ filteredUsers.length }} results</h3>
          </div>
          <UserRound class="size-5 text-zinc-500" />
        </div>
        <div v-if="filteredUsers.length" class="divide-y divide-zinc-800">
          <RouterLink
            v-for="username in filteredUsers"
            :key="username"
            :to="{ name: 'profile', params: { username } }"
            class="flex items-center justify-between gap-3 py-3 text-sm first:pt-0 last:pb-0"
          >
            <span class="font-medium text-zinc-100">{{ username }}</span>
            <Badge>profile</Badge>
          </RouterLink>
        </div>
        <EmptyState v-else eyebrow="No Users" title="No users matched." description="Search by username." />
      </Card>
    </div>
  </div>
</template>
