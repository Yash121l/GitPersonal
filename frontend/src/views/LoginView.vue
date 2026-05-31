<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRoute, useRouter, type RouteLocationRaw } from 'vue-router'

import { getDefaultWorkspaceRoute } from '@/app/navigation'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const username = ref('')
const password = ref('')
const errorMessage = ref('')
const loading = ref(false)

const redirectTarget = computed<RouteLocationRaw>(() => {
  const redirect = route.query.redirect
  return typeof redirect === 'string' && redirect.startsWith('/') ? redirect : getDefaultWorkspaceRoute()
})

async function handleSubmit() {
  errorMessage.value = ''
  loading.value = true
  try {
    await authStore.login({
      username: username.value.trim(),
      password: password.value,
    })
    await router.push(redirectTarget.value)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Unable to sign in.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page-shell">
    <div class="mx-auto grid min-h-[calc(100vh-3rem)] max-w-6xl items-center gap-8 lg:grid-cols-[1.08fr_0.92fr]">
      <div class="space-y-6">
        <p class="eyebrow">Developer Workspace</p>
        <h1 class="max-w-4xl text-5xl font-semibold leading-tight tracking-tight text-zinc-50">
          Self-hosted Git with a control surface built for engineers.
        </h1>
        <p class="max-w-2xl text-base leading-7 text-zinc-400">
          Dark, dense, and operational by default: repository inventory, SSH identity, org access, browser code navigation, and webhook administration without leaving the same workspace.
        </p>
        <div class="grid gap-4 md:grid-cols-3">
          <Card class="border-sky-500/20 bg-sky-500/5">
            <p class="eyebrow text-sky-400">Transport</p>
            <h2 class="mt-2 text-xl font-semibold">HTTP + SSH</h2>
            <p class="mt-2 text-sm text-zinc-400">One identity model across the API, the browser, Smart HTTP, and SSH Git.</p>
          </Card>
          <Card>
            <p class="eyebrow">Repository UI</p>
            <h2 class="mt-2 text-xl font-semibold text-zinc-50">Tree + Blob</h2>
            <p class="mt-2 text-sm text-zinc-400">Branch-aware browsing with bounded previews and virtualized trees for deeper repos.</p>
          </Card>
          <Card>
            <p class="eyebrow">Operations</p>
            <h2 class="mt-2 text-xl font-semibold text-zinc-50">Automation</h2>
            <p class="mt-2 text-sm text-zinc-400">Signed webhook delivery and repo administration from the same dark control plane.</p>
          </Card>
        </div>

        <div class="terminal-panel overflow-hidden">
          <div class="flex items-center gap-2 border-b border-zinc-800 px-4 py-3">
            <span class="size-2 rounded-full bg-red-400/80" />
            <span class="size-2 rounded-full bg-amber-400/80" />
            <span class="size-2 rounded-full bg-emerald-400/80" />
            <span class="ml-2 text-xs text-zinc-500">forge status</span>
          </div>
          <div class="space-y-2 px-4 py-4 font-mono text-xs text-zinc-400">
            <p><span class="text-emerald-400">$</span> forge status --scope browser</p>
            <p class="text-zinc-500">ui.shell = ready</p>
            <p class="text-zinc-500">repo.browser = virtualized</p>
            <p class="text-zinc-500">auth.model = shared(http, ssh, api)</p>
            <p class="text-zinc-500">theme = dark developer console</p>
          </div>
        </div>
      </div>

      <Card class="p-8">
        <p class="eyebrow">Sign In</p>
        <h2 class="mt-2 text-3xl font-semibold text-zinc-50">Open your Forge workspace.</h2>
        <p class="mt-3 text-sm leading-6 text-zinc-400">
          Use the same account that authorizes API calls, Git pushes, and SSH keys.
        </p>

        <form class="mt-8 space-y-5" @submit.prevent="handleSubmit">
          <div>
            <label class="field-label" for="login-username">Username</label>
            <Input id="login-username" v-model="username" autocomplete="username" required />
          </div>
          <div>
            <label class="field-label" for="login-password">Password</label>
            <Input id="login-password" v-model="password" type="password" autocomplete="current-password" required />
          </div>

          <div
            v-if="errorMessage"
            class="rounded-md border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300"
          >
            {{ errorMessage }}
          </div>

          <div class="flex flex-wrap gap-3">
            <Button :disabled="loading" type="submit">
              {{ loading ? 'Signing in...' : 'Sign In' }}
            </Button>
            <Button :as="RouterLink" :to="{ name: 'register' }" variant="secondary">
              Create Account
            </Button>
          </div>
        </form>
      </Card>
    </div>
  </div>
</template>
