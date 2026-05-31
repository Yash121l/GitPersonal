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

async function handleLogout() {
  await authStore.logout()
  await router.push({ name: 'login' })
}
</script>

<template>
  <div class="page-shell">
    <div class="grid gap-4 xl:grid-cols-[300px_minmax(0,1fr)]">
      <aside class="hidden xl:block">
        <div class="space-y-4">
          <Card class="space-y-4">
            <div class="flex items-center gap-3">
              <div class="flex size-11 items-center justify-center rounded-lg border border-zinc-800 bg-zinc-900 text-lg font-semibold text-zinc-100">
                F
              </div>
              <div>
                <p class="text-lg font-semibold text-zinc-50">{{ bootstrap.productName }}</p>
                <p class="text-sm text-zinc-500">Developer control plane</p>
              </div>
            </div>

            <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
              <p class="eyebrow">Current Surface</p>
              <p class="mt-2 text-base font-semibold text-zinc-100">{{ currentSection?.label || 'Workspace' }}</p>
              <p class="mt-2 text-sm leading-6 text-zinc-400">
                {{ currentSection?.description || 'Stable workspace navigation with repository-local modules.' }}
              </p>
            </div>
          </Card>

          <Card v-for="group in workspaceGroups" :key="group.id" class="space-y-2">
            <p class="eyebrow">{{ group.label }}</p>
            <div class="space-y-1">
              <RouterLink
                v-for="item in group.items"
                :key="item.id"
                :to="item.to"
                :class="
                  [
                    'flex items-start gap-3 rounded-xl border px-4 py-3 transition',
                    isNavigationItemActive(item, route)
                      ? 'border-zinc-700 bg-zinc-100 text-zinc-950'
                      : 'border-transparent bg-black/20 text-zinc-400 hover:border-zinc-800 hover:bg-zinc-900 hover:text-zinc-100',
                  ]
                "
              >
                <component :is="item.icon" class="mt-0.5 size-4 shrink-0" />
                <div>
                  <p class="text-sm font-medium">{{ item.label }}</p>
                  <p class="mt-1 text-xs leading-5 text-inherit/70">{{ item.description }}</p>
                </div>
              </RouterLink>
            </div>
          </Card>

          <Card v-if="repositoryParams" class="space-y-3">
            <div>
              <p class="eyebrow">Repository Map</p>
              <h3 class="mt-2 font-mono text-lg font-semibold text-zinc-50">
                {{ repositoryParams.owner }}/{{ repositoryParams.repo }}
              </h3>
              <p class="mt-2 text-sm leading-6 text-zinc-400">
                Repository-local routes stay grouped so new features can be inserted without disturbing workspace navigation.
              </p>
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

          <Card class="space-y-4">
            <div>
              <p class="eyebrow">Signed In</p>
              <p class="mt-2 text-base font-semibold text-zinc-50">{{ authStore.currentUser?.username }}</p>
              <p class="text-sm text-zinc-500">{{ authStore.currentUser?.role }}</p>
            </div>
            <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
              <div class="flex items-center gap-2 text-sm text-zinc-300">
                <Sparkles class="size-4 text-sky-300" />
                <span>{{ enabledRepositoryModules }} repository modules enabled</span>
              </div>
              <p class="mt-2 text-xs leading-5 text-zinc-500">
                Feature flags now control which workspace and repository surfaces are available.
              </p>
            </div>
            <Button class="w-full" variant="secondary" @click="handleLogout">
              <LogOut class="size-4" />
              Log out
            </Button>
          </Card>
        </div>
      </aside>

      <div class="space-y-4">
        <header class="frosted-panel flex items-center justify-between gap-3 px-4 py-4 md:px-6">
          <div>
            <p class="eyebrow">{{ repositoryParams ? 'Repository Workspace' : 'Forge Workspace' }}</p>
            <h1 class="text-2xl font-semibold tracking-tight text-zinc-50">{{ currentSection?.label || 'Workspace' }}</h1>
            <p class="mt-2 max-w-3xl text-sm leading-6 text-zinc-400">
              {{ currentSection?.description || 'Stable navigation with feature-gated repository modules.' }}
            </p>
          </div>

          <div class="flex items-center gap-3">
            <div class="hidden rounded-xl border border-zinc-800 bg-black/30 px-4 py-3 text-right md:block">
              <p class="text-xs uppercase tracking-[0.24em] text-zinc-500">Session</p>
              <p class="mt-2 text-sm font-semibold text-zinc-100">{{ authStore.currentUser?.username }}</p>
            </div>
            <Button class="xl:hidden" variant="secondary" @click="mobileMenuOpen = !mobileMenuOpen">
              <PanelLeftOpen class="size-4" />
              Menu
            </Button>
          </div>
        </header>

        <div v-if="mobileMenuOpen" class="space-y-4 xl:hidden">
          <Card v-for="group in workspaceGroups" :key="group.id" class="space-y-2">
            <p class="eyebrow">{{ group.label }}</p>
            <RouterLink
              v-for="item in group.items"
              :key="item.id"
              :to="item.to"
              class="flex items-start gap-3 rounded-xl border border-transparent bg-black/20 px-4 py-3 text-sm text-zinc-300 hover:border-zinc-800 hover:bg-zinc-900"
              @click="mobileMenuOpen = false"
            >
              <component :is="item.icon" class="mt-0.5 size-4" />
              <div>
                <p class="font-medium">{{ item.label }}</p>
                <p class="mt-1 text-xs text-zinc-500">{{ item.description }}</p>
              </div>
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
