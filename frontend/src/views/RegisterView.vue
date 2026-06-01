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
const attempted = ref(false)

const usernameError = computed(() => (attempted.value && username.value.trim() === '' ? 'Username is required.' : ''))
const passwordError = computed(() =>
  attempted.value && password.value.length < 12 ? 'Password must be at least 12 characters.' : '',
)
const formValid = computed(() => username.value.trim() !== '' && password.value.length >= 12)

async function handleSubmit() {
  attempted.value = true
  errorMessage.value = ''
  if (!formValid.value) {
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
    <div class="mx-auto grid min-h-[calc(100vh-2rem)] max-w-5xl items-center gap-12 lg:grid-cols-[1fr_420px]">
      <div class="space-y-8">
        <div class="space-y-4">
          <p class="eyebrow">Create Account</p>
          <h1 class="max-w-3xl text-4xl font-semibold tracking-tight text-zinc-50 md:text-5xl">
            Create a single developer identity for the whole workspace.
          </h1>
          <p class="max-w-2xl text-base leading-7 text-zinc-400">
            Sign in to the UI, call the API, push over HTTP, and then add SSH access when you are ready.
          </p>
        </div>

        <div class="space-y-4 border-t border-zinc-800 pt-6">
          <div class="flex items-start gap-3">
            <div class="mt-1 size-2 rounded-full bg-zinc-500" />
            <div>
              <p class="text-sm font-medium text-zinc-100">Real backend contract</p>
              <p class="text-sm text-zinc-500">The browser sits on top of the same server routes and data model used everywhere else.</p>
            </div>
          </div>
          <div class="flex items-start gap-3">
            <div class="mt-1 size-2 rounded-full bg-zinc-500" />
            <div>
              <p class="text-sm font-medium text-zinc-100">Repository and org workflows</p>
              <p class="text-sm text-zinc-500">Create repositories, organizations, collaborators, and automation from the same workspace.</p>
            </div>
          </div>
          <div class="flex items-start gap-3">
            <div class="mt-1 size-2 rounded-full bg-zinc-500" />
            <div>
              <p class="text-sm font-medium text-zinc-100">Clear validation</p>
              <p class="text-sm text-zinc-500">Required fields and password rules are enforced before the request is sent.</p>
            </div>
          </div>
        </div>
      </div>

      <Card class="p-6 md:p-8">
        <div class="space-y-2">
          <p class="eyebrow">Account</p>
          <h2 class="text-2xl font-semibold text-zinc-50">Create your account</h2>
          <p class="text-sm leading-6 text-zinc-400">
            Use a password with at least 12 characters.
          </p>
        </div>

        <form class="mt-6 space-y-5" @submit.prevent="handleSubmit">
          <div>
            <label class="field-label" for="register-username">Username</label>
            <Input id="register-username" v-model="username" autocomplete="username" placeholder="yash" required />
            <p v-if="usernameError" class="mt-1 text-xs text-red-400">{{ usernameError }}</p>
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
            <p class="mt-1 text-xs text-zinc-500">Minimum 12 characters.</p>
            <p v-if="passwordError" class="mt-1 text-xs text-red-400">{{ passwordError }}</p>
          </div>

          <div v-if="errorMessage" class="rounded-md border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-300">
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
