<script setup lang="ts">
import { LogOut, Menu, Plus, Search, UserRound, X } from '@lucide/vue'
import { useQuery } from '@tanstack/vue-query'
import { computed, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { bootstrap } from '@/app/bootstrap'
import { getRepositoryNavigation, getWorkspaceNavigationGroups, isNavigationItemActive } from '@/app/navigation'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import { api } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const mobileMenuOpen = ref(false)
const search = ref('')

const repositoriesQuery = useQuery({
  queryKey: ['repositories'],
  queryFn: () => api.listRepositories(),
})

const workspaceGroups = computed(() => getWorkspaceNavigationGroups())
const repositories = computed(() => repositoriesQuery.data.value ?? [])
const searchResults = computed(() => {
  const query = search.value.trim().toLowerCase()
  if (!query) {
    return []
  }

  return repositories.value
    .filter((repository) => `${repository.owner}/${repository.name}`.toLowerCase().includes(query))
    .slice(0, 6)
})
const repositoryParams = computed(() => {
  if (typeof route.params.owner !== 'string' || typeof route.params.repo !== 'string') {
    return null
  }

  return {
    owner: route.params.owner,
    repo: route.params.repo,
  }
})
const repositoryNavigation = computed(() =>
  (() => {
    const repository = repositoryParams.value
    if (!repository) {
      return []
    }

    return getRepositoryNavigation().map((item) => ({
      ...item,
      to: item.to({
        owner: repository.owner,
        repo: repository.repo,
        currentQuery: route.query,
      }),
    }))
  })(),
)
const currentSection = computed(() => {
  const repositoryItem = repositoryNavigation.value.find((item) => isNavigationItemActive(item, route))
  if (repositoryItem) {
    return repositoryItem
  }

  const workspaceItem = workspaceGroups.value.flatMap((group) => group.items).find((item) => isNavigationItemActive(item, route))
  if (workspaceItem) {
    return workspaceItem
  }

  return null
})
const currentScopeLabel = computed(() =>
  repositoryParams.value ? `${repositoryParams.value.owner}/${repositoryParams.value.repo}` : authStore.currentUser?.username || 'workspace',
)

async function handleLogout() {
  await authStore.logout()
  await router.push({ name: 'login' })
}

function clearSearch() {
  search.value = ''
}
</script>

<template>
  <div class="min-h-screen bg-zinc-950 text-zinc-100">
    <div class="grid min-h-screen xl:grid-cols-[272px_minmax(0,1fr)]">
      <aside class="hidden min-h-screen border-r border-zinc-800 bg-zinc-950 xl:flex xl:flex-col">
        <div class="flex h-14 items-center gap-3 border-b border-zinc-800 px-4">
          <RouterLink :to="{ name: 'overview' }" class="flex min-w-0 items-center gap-3">
            <div class="flex size-8 items-center justify-center rounded-md border border-zinc-700 bg-zinc-900 text-xs font-semibold text-zinc-100">
              F
            </div>
            <div class="min-w-0">
              <p class="truncate text-sm font-semibold text-zinc-50">{{ bootstrap.productName }}</p>
              <p class="truncate text-xs text-zinc-500">{{ currentScopeLabel }}</p>
            </div>
          </RouterLink>
        </div>

        <div class="scrollbar-subtle flex-1 overflow-y-auto p-3">
          <div class="mb-4 rounded-md border border-zinc-800 bg-zinc-900/40 p-3">
            <p class="truncate text-sm font-medium text-zinc-100">{{ currentSection?.label || 'Workspace' }}</p>
            <p class="mt-1 line-clamp-2 text-xs leading-5 text-zinc-500">
              {{ currentSection?.description || currentScopeLabel }}
            </p>
          </div>

          <div v-for="group in workspaceGroups" :key="group.id" class="mb-5 space-y-2">
            <p class="px-3 text-[11px] font-medium uppercase text-zinc-500">{{ group.label }}</p>
            <nav class="space-y-1">
              <RouterLink
                v-for="item in group.items"
                :key="item.id"
                :to="item.id === 'workspace-profile' ? { name: 'profile', params: { username: authStore.currentUser?.username } } : item.to"
                :class="
                  [
                    'flex h-9 items-center gap-3 rounded-md px-3 text-sm transition',
                    isNavigationItemActive(item, route)
                      ? 'bg-zinc-100 text-zinc-950'
                      : 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-100',
                  ]
                "
              >
                <component :is="item.icon" class="size-4 shrink-0" />
                <span class="truncate">{{ item.label }}</span>
              </RouterLink>
            </nav>
          </div>

          <div v-if="repositoryParams" class="mb-5 space-y-2">
            <p class="px-3 text-[11px] font-medium uppercase text-zinc-500">Repository</p>
            <div class="mx-3 rounded-md border border-zinc-800 bg-zinc-900/60 p-3">
              <p class="truncate font-mono text-sm font-medium text-zinc-100">{{ repositoryParams.owner }}/{{ repositoryParams.repo }}</p>
            </div>
            <nav class="space-y-1">
              <RouterLink
                v-for="item in repositoryNavigation"
                :key="item.id"
                :to="item.to"
                :class="
                  [
                    'flex h-9 items-center gap-3 rounded-md px-3 text-sm transition',
                    isNavigationItemActive(item, route)
                      ? 'bg-zinc-100 text-zinc-950'
                      : 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-100',
                  ]
                "
              >
                <component :is="item.icon" class="size-4 shrink-0" />
                <span class="truncate">{{ item.label }}</span>
              </RouterLink>
            </nav>
          </div>
        </div>

        <div class="border-t border-zinc-800 p-3">
          <div class="flex items-center justify-between gap-3 rounded-md bg-zinc-900/50 p-2">
            <RouterLink :to="{ name: 'profile', params: { username: authStore.currentUser?.username } }" class="flex min-w-0 items-center gap-3">
              <div class="flex size-8 items-center justify-center rounded-md border border-zinc-800 bg-zinc-950">
                <UserRound class="size-4 text-zinc-400" />
              </div>
              <div class="min-w-0">
                <p class="truncate text-sm font-medium text-zinc-100">{{ authStore.currentUser?.username }}</p>
                <p class="truncate text-xs text-zinc-500">{{ authStore.currentUser?.role }}</p>
              </div>
            </RouterLink>
            <button class="rounded-md p-2 text-zinc-500 hover:bg-zinc-800 hover:text-zinc-100" title="Log out" type="button" @click="handleLogout">
              <LogOut class="size-4" />
            </button>
          </div>
        </div>
      </aside>

      <div class="min-w-0">
        <header class="sticky top-0 z-20 flex h-14 items-center justify-between gap-3 border-b border-zinc-800 bg-zinc-950/95 px-4 backdrop-blur md:px-6">
          <div class="flex min-w-0 items-center gap-3">
            <Button class="xl:hidden" variant="secondary" size="sm" @click="mobileMenuOpen = !mobileMenuOpen">
              <Menu v-if="!mobileMenuOpen" class="size-4" />
              <X v-else class="size-4" />
            </Button>
            <div class="relative hidden w-[min(460px,42vw)] md:block">
              <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-zinc-500" />
              <Input v-model="search" class="h-9 bg-zinc-900/80 pl-9" placeholder="Search repositories" />
              <div v-if="searchResults.length" class="absolute left-0 right-0 top-11 z-30 overflow-hidden rounded-md border border-zinc-800 bg-zinc-950 shadow-xl">
                <RouterLink
                  v-for="repository in searchResults"
                  :key="`${repository.owner}/${repository.name}`"
                  :to="{ name: 'repository-code', params: { owner: repository.owner, repo: repository.name } }"
                  class="block border-b border-zinc-800 px-3 py-2 last:border-b-0 hover:bg-zinc-900"
                  @click="clearSearch"
                >
                  <p class="truncate font-mono text-sm text-zinc-100">{{ repository.owner }}/{{ repository.name }}</p>
                  <p class="truncate text-xs text-zinc-500">{{ repository.description || repository.visibility }}</p>
                </RouterLink>
              </div>
            </div>
          </div>

          <div class="flex items-center gap-2">
            <Button :as="RouterLink" :to="{ name: 'new-repository' }" size="sm">
              <Plus class="size-4" />
              New
            </Button>
            <Button class="hidden md:inline-flex" variant="secondary" size="sm" @click="handleLogout">
              <LogOut class="size-4" />
            </Button>
          </div>
        </header>

        <div v-if="mobileMenuOpen" class="border-b border-zinc-800 bg-zinc-950 px-4 py-4 xl:hidden">
          <div class="mb-4">
            <div class="relative">
              <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-zinc-500" />
              <Input v-model="search" class="pl-9" placeholder="Search repositories" />
            </div>
          </div>
          <div class="grid gap-4 md:grid-cols-2">
            <div v-for="group in workspaceGroups" :key="group.id" class="space-y-2">
              <p class="px-1 text-[11px] font-medium uppercase text-zinc-500">{{ group.label }}</p>
              <nav class="space-y-1">
                <RouterLink
                  v-for="item in group.items"
                  :key="item.id"
                  :to="item.id === 'workspace-profile' ? { name: 'profile', params: { username: authStore.currentUser?.username } } : item.to"
                  :class="
                    [
                      'flex items-center gap-3 rounded-md px-3 py-2 text-sm transition',
                      isNavigationItemActive(item, route)
                        ? 'bg-zinc-100 text-zinc-950'
                        : 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-100',
                    ]
                  "
                  @click="mobileMenuOpen = false"
                >
                  <component :is="item.icon" class="size-4 shrink-0" />
                  <span>{{ item.label }}</span>
                </RouterLink>
              </nav>
            </div>

            <div v-if="repositoryParams" class="space-y-2">
              <p class="px-1 text-[11px] font-medium uppercase text-zinc-500">Repository</p>
              <nav class="space-y-1">
                <RouterLink
                  v-for="item in repositoryNavigation"
                  :key="item.id"
                  :to="item.to"
                  :class="
                    [
                      'flex items-center gap-3 rounded-md px-3 py-2 text-sm transition',
                      isNavigationItemActive(item, route)
                        ? 'bg-zinc-100 text-zinc-950'
                        : 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-100',
                    ]
                  "
                  @click="mobileMenuOpen = false"
                >
                  <component :is="item.icon" class="size-4 shrink-0" />
                  <span>{{ item.label }}</span>
                </RouterLink>
              </nav>
            </div>
          </div>
        </div>

        <main class="mx-auto min-h-[calc(100vh-3.5rem)] w-full max-w-[1440px] px-4 py-6 md:px-6">
          <slot />
        </main>
      </div>
    </div>
  </div>
</template>
