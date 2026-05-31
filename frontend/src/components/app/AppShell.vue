<script setup lang="ts">
import { LogOut, PanelLeftOpen, Sparkles } from '@lucide/vue'
import { computed, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { bootstrap } from '@/app/bootstrap'
import { getRepositoryNavigation, getWorkspaceNavigationGroups, isNavigationItemActive } from '@/app/navigation'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const mobileMenuOpen = ref(false)

const workspaceGroups = computed(() => getWorkspaceNavigationGroups())
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
const enabledRepositoryModules = computed(() => repositoryNavigation.value.length)
const currentScopeLabel = computed(() =>
  repositoryParams.value ? `${repositoryParams.value.owner}/${repositoryParams.value.repo}` : authStore.currentUser?.username || 'workspace',
)

async function handleLogout() {
  await authStore.logout()
  await router.push({ name: 'login' })
}
</script>

<template>
  <div class="page-shell">
    <div class="grid gap-3 xl:grid-cols-[264px_minmax(0,1fr)]">
      <aside class="hidden xl:block">
        <div class="space-y-3">
          <Card class="space-y-3">
            <div class="flex items-center gap-2.5">
              <div class="flex size-9 items-center justify-center rounded-lg border border-zinc-800 bg-zinc-900 text-sm font-semibold text-zinc-100">
                F
              </div>
              <div>
                <p class="text-base font-semibold text-zinc-50">{{ bootstrap.productName }}</p>
                <p class="text-xs text-zinc-500">Source control workspace</p>
              </div>
            </div>

            <div class="rounded-lg border border-zinc-800 bg-black/20 px-3 py-2.5">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="eyebrow">Scope</p>
                  <p class="mt-1 text-sm font-semibold text-zinc-100">{{ currentScopeLabel }}</p>
                </div>
                <span class="rounded-md border border-zinc-800 bg-black/40 px-2 py-1 text-[10px] uppercase tracking-[0.18em] text-zinc-500">
                  {{ repositoryParams ? 'repository' : 'workspace' }}
                </span>
              </div>
            </div>

            <div class="rounded-lg border border-zinc-800 bg-black/30 px-3 py-2.5">
              <p class="eyebrow">Current</p>
              <div class="mt-2">
                <p class="text-sm font-semibold text-zinc-100">{{ currentSection?.label || 'Workspace' }}</p>
                <p class="mt-1 text-xs text-zinc-500">{{ currentSection?.description || 'Stable navigation for the active surface.' }}</p>
              </div>
            </div>
          </Card>

          <Card v-for="group in workspaceGroups" :key="group.id" class="space-y-2.5">
            <p class="eyebrow">{{ group.label }}</p>
            <div class="space-y-1">
              <RouterLink
                v-for="item in group.items"
                :key="item.id"
                :to="item.to"
                :class="
                  [
                    'flex items-center gap-3 rounded-lg border px-3 py-2.5 transition',
                    isNavigationItemActive(item, route)
                      ? 'border-zinc-700 bg-zinc-100 text-zinc-950'
                      : 'border-transparent bg-black/20 text-zinc-400 hover:border-zinc-800 hover:bg-zinc-900 hover:text-zinc-100',
                  ]
                "
              >
                <component :is="item.icon" class="size-4 shrink-0" />
                <div class="min-w-0 flex-1">
                  <p class="truncate text-sm font-medium">{{ item.label }}</p>
                </div>
              </RouterLink>
            </div>
          </Card>

          <Card v-if="repositoryParams" class="space-y-2.5">
            <div class="panel-header">
              <div>
                <p class="eyebrow">Repository</p>
                <h3 class="mt-1 font-mono text-sm font-semibold text-zinc-50">
                  {{ repositoryParams.owner }}/{{ repositoryParams.repo }}
                </h3>
              </div>
              <span class="rounded-md border border-zinc-800 bg-black/20 px-2 py-1 text-[10px] uppercase tracking-[0.16em] text-zinc-500">
                {{ enabledRepositoryModules }} tabs
              </span>
            </div>
            <div class="space-y-1">
              <RouterLink
                v-for="item in repositoryNavigation"
                :key="item.id"
                :to="item.to"
                :class="
                  [
                    'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition',
                    isNavigationItemActive(item, route)
                      ? 'bg-zinc-900 text-zinc-100'
                      : 'text-zinc-400 hover:bg-zinc-900 hover:text-zinc-100',
                  ]
                "
              >
                <component :is="item.icon" class="size-4 shrink-0" />
                <span>{{ item.label }}</span>
              </RouterLink>
            </div>
          </Card>

          <Card class="space-y-3">
            <div class="panel-header">
              <div>
                <p class="eyebrow">Signed In</p>
                <p class="mt-1 text-sm font-semibold text-zinc-50">{{ authStore.currentUser?.username }}</p>
              </div>
              <span class="text-xs text-zinc-500">{{ authStore.currentUser?.role }}</span>
            </div>
            <div class="rounded-lg border border-zinc-800 bg-black/30 px-3 py-2.5">
              <div class="flex items-center gap-2 text-xs text-zinc-300">
                <Sparkles class="size-4 text-sky-300" />
                <span>{{ enabledRepositoryModules }} repository modules enabled</span>
              </div>
            </div>
            <Button class="w-full" variant="secondary" @click="handleLogout">
              <LogOut class="size-4" />
              Log out
            </Button>
          </Card>
        </div>
      </aside>

      <div class="space-y-3">
        <header class="frosted-panel flex flex-col gap-3 px-4 py-3.5 md:px-5 lg:flex-row lg:items-center lg:justify-between">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-x-2 gap-y-1 text-xs uppercase tracking-[0.18em] text-zinc-500">
              <span>{{ repositoryParams ? 'Repository' : 'Workspace' }}</span>
              <span>/</span>
              <span class="text-zinc-300">{{ currentScopeLabel }}</span>
            </div>
            <div class="mt-1.5 flex flex-wrap items-center gap-2">
              <h1 class="text-lg font-semibold tracking-tight text-zinc-50 md:text-xl">
                {{ currentSection?.label || 'Workspace' }}
              </h1>
              <span
                v-if="repositoryParams"
                class="rounded-md border border-zinc-800 bg-black/30 px-2 py-1 text-[10px] uppercase tracking-[0.18em] text-zinc-400"
              >
                {{ enabledRepositoryModules }} modules
              </span>
            </div>
            <p class="mt-1 max-w-2xl text-sm leading-6 text-zinc-400">
              {{ currentSection?.description || 'Stable navigation with repository-local modules.' }}
            </p>
          </div>

          <div class="flex items-center gap-2">
            <div class="hidden rounded-lg border border-zinc-800 bg-black/30 px-3 py-2 text-right md:block">
              <p class="text-[10px] uppercase tracking-[0.18em] text-zinc-500">Session</p>
              <p class="mt-1 text-sm font-semibold text-zinc-100">{{ authStore.currentUser?.username }}</p>
            </div>
            <Button class="xl:hidden" variant="secondary" @click="mobileMenuOpen = !mobileMenuOpen">
              <PanelLeftOpen class="size-4" />
              Menu
            </Button>
          </div>
        </header>

        <div v-if="mobileMenuOpen" class="space-y-3 xl:hidden">
          <Card v-for="group in workspaceGroups" :key="group.id" class="space-y-2">
            <p class="eyebrow">{{ group.label }}</p>
            <RouterLink
              v-for="item in group.items"
              :key="item.id"
              :to="item.to"
              class="flex items-center gap-3 rounded-lg border border-transparent bg-black/20 px-3 py-2.5 text-sm text-zinc-300 hover:border-zinc-800 hover:bg-zinc-900"
              @click="mobileMenuOpen = false"
            >
              <component :is="item.icon" class="size-4" />
              <p class="font-medium">{{ item.label }}</p>
            </RouterLink>
          </Card>

          <Card v-if="repositoryParams" class="space-y-2">
            <p class="eyebrow">Repository Map</p>
            <RouterLink
              v-for="item in repositoryNavigation"
              :key="item.id"
              :to="item.to"
              class="flex items-center gap-3 rounded-lg px-3 py-2 text-sm text-zinc-300 hover:bg-zinc-900"
              @click="mobileMenuOpen = false"
            >
              <component :is="item.icon" class="size-4" />
              <span>{{ item.label }}</span>
            </RouterLink>
          </Card>

          <Button class="w-full" variant="secondary" @click="handleLogout">
            <LogOut class="size-4" />
            Log out
          </Button>
        </div>

        <main>
          <slot />
        </main>
      </div>
    </div>
  </div>
</template>
