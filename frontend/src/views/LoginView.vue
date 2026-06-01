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
const attempted = ref(false)

const redirectTarget = computed<RouteLocationRaw>(() => {
  const redirect = route.query.redirect
  return typeof redirect === 'string' && redirect.startsWith('/') ? redirect : getDefaultWorkspaceRoute()
})

const usernameError = computed(() => (attempted.value && username.value.trim() === '' ? 'Username is required.' : ''))
const passwordError = computed(() => (attempted.value && password.value === '' ? 'Password is required.' : ''))
const formValid = computed(() => username.value.trim() !== '' && password.value !== '')

async function handleSubmit() {
  attempted.value = true
  errorMessage.value = ''
  if (!formValid.value) {
    return
  }

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
    <div class="mx-auto grid min-h-[calc(100vh-2rem)] max-w-5xl items-center gap-12 lg:grid-cols-[1fr_420px]">
      <div class="space-y-8">
        <div class="flex items-center gap-3">
          <div class="flex size-9 items-center justify-center rounded-md border border-zinc-800 bg-zinc-900 text-sm font-semibold text-zinc-100">
            F
          </div>
          <div>
            <p class="text-sm font-medium text-zinc-50">Forge</p>
            <p class="text-xs text-zinc-500">Self-hosted Git workspace</p>
          </div>
        </div>

        <div class="space-y-4">
          <p class="eyebrow">Sign In</p>
          <h1 class="max-w-3xl text-4xl font-semibold tracking-tight text-zinc-50 md:text-5xl">
            Self-hosted Git, simpler to operate.
          </h1>
          <p class="max-w-2xl text-base leading-7 text-zinc-400">
            Repositories, organizations, SSH keys, and repository modules in a layout built to be easy to scan and easy to navigate.
          </p>
        </div>

        <div class="space-y-4 border-t border-zinc-800 pt-6">
          <div class="flex items-start gap-3">
            <div class="mt-1 size-2 rounded-full bg-zinc-500" />
            <div>
              <p class="text-sm font-medium text-zinc-100">Repository-first navigation</p>
              <p class="text-sm text-zinc-500">Jump from the workspace to code, access, automation, and activity without fighting the layout.</p>
            </div>
          </div>
          <div class="flex items-start gap-3">
            <div class="mt-1 size-2 rounded-full bg-zinc-500" />
            <div>
              <p class="text-sm font-medium text-zinc-100">Shared identity model</p>
              <p class="text-sm text-zinc-500">The same account covers the browser, API calls, Git HTTP, and SSH keys.</p>
            </div>
          </div>
          <div class="flex items-start gap-3">
            <div class="mt-1 size-2 rounded-full bg-zinc-500" />
            <div>
              <p class="text-sm font-medium text-zinc-100">Shadcn-style baseline</p>
              <p class="text-sm text-zinc-500">Simple cards, normal radii, and restrained spacing across the whole UI.</p>
            </div>
          </div>
        </div>
      </div>

      <Card class="p-6 md:p-8">
        <div class="space-y-2">
          <p class="eyebrow">Account</p>
          <h2 class="text-2xl font-semibold text-zinc-50">Open your workspace</h2>
          <p class="text-sm leading-6 text-zinc-400">
            Use the same account that authorizes API calls, Git pushes, and SSH keys.
          </p>
        </div>

        <form class="mt-6 space-y-5" @submit.prevent="handleSubmit">
          <div>
            <label class="field-label" for="login-username">Username</label>
            <Input id="login-username" v-model="username" autocomplete="username" placeholder="yash" required />
            <p v-if="usernameError" class="mt-1 text-xs text-red-400">{{ usernameError }}</p>
          </div>
          <div>
            <label class="field-label" for="login-password">Password</label>
            <Input id="login-password" v-model="password" type="password" autocomplete="current-password" required />
            <p v-if="passwordError" class="mt-1 text-xs text-red-400">{{ passwordError }}</p>
          </div>

          <div v-if="errorMessage" class="rounded-md border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-300">
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
