<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import { getDefaultWorkspaceRoute } from '@/app/navigation'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const router = useRouter()

const username = ref('')
const password = ref('')
const errorMessage = ref('')
const loading = ref(false)
const passwordTooShort = computed(() => password.value.length > 0 && password.value.length < 12)

async function handleSubmit() {
  errorMessage.value = ''
  if (username.value.trim() === '') {
    errorMessage.value = 'Username is required.'
    return
  }
  if (password.value.length < 12) {
    errorMessage.value = 'Password must be at least 12 characters.'
    return
  }
  loading.value = true
  try {
    await authStore.register({
      username: username.value.trim(),
      password: password.value,
    })
    await router.push(getDefaultWorkspaceRoute())
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Unable to create account.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page-shell">
    <div class="mx-auto grid min-h-[calc(100vh-3rem)] max-w-6xl items-center gap-8 lg:grid-cols-[1.05fr_0.95fr]">
      <div class="space-y-6">
        <p class="eyebrow">Create Account</p>
        <h1 class="max-w-4xl text-5xl font-semibold leading-tight tracking-tight text-zinc-50">
          Bring a developer in once and reuse that identity everywhere.
        </h1>
        <p class="max-w-xl text-base leading-7 text-zinc-400">
          The browser sits directly on top of the real Forge backend contract: no fake repo cards, no mocked auth, and no detached front-end workflow.
        </p>

        <div class="terminal-panel overflow-hidden">
          <div class="flex items-center gap-2 border-b border-zinc-800 px-4 py-3">
            <span class="size-2 rounded-full bg-red-400/80" />
            <span class="size-2 rounded-full bg-amber-400/80" />
            <span class="size-2 rounded-full bg-emerald-400/80" />
            <span class="ml-2 text-xs text-zinc-500">forge auth bootstrap</span>
          </div>
          <div class="space-y-2 px-4 py-4 font-mono text-xs text-zinc-400">
            <p><span class="text-emerald-400">$</span> forge user create</p>
            <p class="text-zinc-500">scope.browser = enabled</p>
            <p class="text-zinc-500">scope.git_http = enabled</p>
            <p class="text-zinc-500">scope.ssh = enabled after key registration</p>
          </div>
        </div>
      </div>

      <Card class="p-8">
        <p class="eyebrow">Create Account</p>
        <h2 class="mt-2 text-3xl font-semibold text-zinc-50">Provision your first session.</h2>

        <form class="mt-8 space-y-5" @submit.prevent="handleSubmit">
          <div>
            <label class="field-label" for="register-username">Username</label>
            <Input id="register-username" v-model="username" autocomplete="username" required />
          </div>
          <div>
            <label class="field-label" for="register-password">Password</label>
            <Input
              id="register-password"
              v-model="password"
              type="password"
              autocomplete="new-password"
              minlength="12"
              required
            />
            <p class="mt-2 text-xs text-zinc-500">
              Use at least 12 characters.
            </p>
            <p v-if="passwordTooShort" class="mt-2 text-xs text-red-300">
              Password must be at least 12 characters.
            </p>
          </div>

          <div
            v-if="errorMessage"
            class="rounded-md border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300"
          >
            {{ errorMessage }}
          </div>

          <div class="flex flex-wrap gap-3">
            <Button :disabled="loading" type="submit">
              {{ loading ? 'Creating account...' : 'Create Account' }}
            </Button>
            <Button :as="RouterLink" :to="{ name: 'login' }" variant="secondary">
              Back to Sign In
            </Button>
          </div>
        </form>
      </Card>
    </div>
  </div>
</template>
