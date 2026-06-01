<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { CalendarDays, FolderGit2, Landmark } from '@lucide/vue'
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import EmptyState from '@/components/empty/EmptyState.vue'
import PageHeader from '@/components/app/PageHeader.vue'
import Badge from '@/components/ui/Badge.vue'
import Card from '@/components/ui/Card.vue'
import { api } from '@/lib/api'
import { formatDate } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const authStore = useAuthStore()

const profileUsername = computed(() => {
  const username = route.params.username
  return typeof username === 'string' ? username : authStore.currentUser?.username || ''
})

const repositoriesQuery = useQuery({
  queryKey: ['repositories'],
  queryFn: () => api.listRepositories(),
})

const organizationsQuery = useQuery({
  queryKey: ['organizations'],
  queryFn: () => api.listOrganizations(),
})

const repositories = computed(() =>
  (repositoriesQuery.data.value ?? []).filter((repository) => repository.owner === profileUsername.value),
)
const organizations = computed(() =>
  (organizationsQuery.data.value ?? []).filter(
    (organization) =>
      organization.organization_slug === profileUsername.value ||
      organization.username === profileUsername.value,
  ),
)
const publicRepositories = computed(() => repositories.value.filter((repository) => repository.visibility === 'public'))
const privateRepositories = computed(() => repositories.value.filter((repository) => repository.visibility !== 'public'))
const isCurrentUser = computed(() => profileUsername.value === authStore.currentUser?.username)
</script>

<template>
  <div class="section-stack">
    <PageHeader :eyebrow="isCurrentUser ? 'Your Profile' : 'Profile'" :title="profileUsername" description="Developer profile, repositories, and organization memberships.">
      <template #actions>
        <Badge>{{ repositories.length }} repos</Badge>
        <Badge variant="accent">{{ organizations.length }} org links</Badge>
      </template>
    </PageHeader>

    <div class="grid gap-6 xl:grid-cols-[320px_minmax(0,1fr)]">
      <Card class="space-y-5">
        <div class="flex size-20 items-center justify-center rounded-md border border-zinc-800 bg-zinc-900 text-2xl font-semibold text-zinc-100">
          {{ profileUsername.slice(0, 1).toUpperCase() }}
        </div>
        <div>
          <h2 class="text-xl font-semibold text-zinc-50">{{ profileUsername }}</h2>
          <p class="mt-1 text-sm text-zinc-500">{{ isCurrentUser ? authStore.currentUser?.role : 'developer' }}</p>
        </div>
        <div class="grid gap-3 border-y border-zinc-800 py-4">
          <div class="flex items-center justify-between gap-3 text-sm">
            <span class="text-zinc-400">Public repositories</span>
            <span class="font-semibold text-zinc-100">{{ publicRepositories.length }}</span>
          </div>
          <div class="flex items-center justify-between gap-3 text-sm">
            <span class="text-zinc-400">Private repositories</span>
            <span class="font-semibold text-zinc-100">{{ privateRepositories.length }}</span>
          </div>
          <div class="flex items-center justify-between gap-3 text-sm">
            <span class="text-zinc-400">Organizations</span>
            <span class="font-semibold text-zinc-100">{{ organizations.length }}</span>
          </div>
        </div>
        <div class="flex items-center gap-2 text-sm text-zinc-500">
          <CalendarDays class="size-4" />
          <span>{{ isCurrentUser ? `Joined ${formatDate(authStore.currentUser?.created_at || '')}` : 'University developer' }}</span>
        </div>
      </Card>

      <div class="space-y-6">
        <Card class="space-y-4">
          <div class="panel-header">
            <div>
              <p class="eyebrow">Repositories</p>
              <h3 class="mt-1 text-lg font-semibold text-zinc-50">{{ repositories.length }} visible repositories</h3>
            </div>
            <FolderGit2 class="size-5 text-zinc-500" />
          </div>
          <div v-if="repositoriesQuery.isLoading.value" class="space-y-3">
            <div v-for="index in 3" :key="index" class="h-24 animate-pulse rounded-md bg-zinc-900" />
          </div>
          <div v-else-if="repositories.length" class="divide-y divide-zinc-800">
            <RouterLink
              v-for="repository in repositories"
              :key="`${repository.owner}/${repository.name}`"
              :to="{ name: 'repository-code', params: { owner: repository.owner, repo: repository.name } }"
              class="group block py-4 first:pt-0 last:pb-0"
            >
              <div class="flex flex-wrap items-center gap-2">
                <h3 class="font-mono text-sm font-medium text-zinc-100 group-hover:text-zinc-50">
                  {{ repository.owner }}/{{ repository.name }}
                </h3>
                <Badge variant="accent">{{ repository.visibility }}</Badge>
              </div>
              <p class="mt-2 text-sm leading-6 text-zinc-400">{{ repository.description || 'No description yet.' }}</p>
              <div class="meta-list mt-3">
                <span>{{ repository.default_branch }}</span>
                <span>Updated {{ formatDate(repository.updated_at) }}</span>
              </div>
            </RouterLink>
          </div>
          <EmptyState v-else eyebrow="No Repositories" title="No visible repositories." description="Repositories appear here when this profile owns them or you have access." />
        </Card>

        <Card class="space-y-4">
          <div class="panel-header">
            <div>
              <p class="eyebrow">Organizations</p>
              <h3 class="mt-1 text-lg font-semibold text-zinc-50">Memberships</h3>
            </div>
            <Landmark class="size-5 text-zinc-500" />
          </div>
          <div v-if="organizations.length" class="grid gap-3 md:grid-cols-2">
            <RouterLink
              v-for="organization in organizations"
              :key="`${organization.organization_slug}-${organization.role}`"
              :to="{ name: 'profile', params: { username: organization.organization_slug } }"
              class="rounded-md border border-zinc-800 p-4 hover:border-zinc-700 hover:bg-zinc-900/50"
            >
              <p class="font-medium text-zinc-100">{{ organization.organization_display_name }}</p>
              <p class="mt-1 font-mono text-xs text-zinc-500">{{ organization.organization_slug }}</p>
              <Badge class="mt-3">{{ organization.role }}</Badge>
            </RouterLink>
          </div>
          <EmptyState v-else eyebrow="No Organizations" title="No organization memberships." description="Organization links appear here after membership is granted." />
        </Card>
      </div>
    </div>
  </div>
</template>
