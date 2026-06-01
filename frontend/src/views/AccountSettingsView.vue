<script setup lang="ts">
import { KeyRound, Settings2, ShieldCheck, UserRound } from '@lucide/vue'
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import PageHeader from '@/components/app/PageHeader.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const authStore = useAuthStore()
const section = computed(() => String(route.params.section || 'profile'))
</script>

<template>
  <div class="section-stack">
    <PageHeader eyebrow="Settings" :title="`Settings / ${section}`" description="Account identity, SSH keys, and university workspace preferences.">
      <template #actions>
        <Badge>{{ authStore.currentUser?.username }}</Badge>
      </template>
    </PageHeader>

    <div class="grid gap-6 xl:grid-cols-[260px_minmax(0,1fr)]">
      <Card class="space-y-2 p-3">
        <RouterLink
          v-for="item in [
            { name: 'profile', icon: UserRound },
            { name: 'account', icon: Settings2 },
            { name: 'security', icon: ShieldCheck },
            { name: 'keys', icon: KeyRound },
          ]"
          :key="item.name"
          :to="{ name: 'account-settings-section', params: { section: item.name } }"
          class="flex items-center gap-3 rounded-md px-3 py-2 text-sm text-zinc-400 hover:bg-zinc-900 hover:text-zinc-100"
          :class="{ 'bg-zinc-100 text-zinc-950 hover:bg-zinc-100 hover:text-zinc-950': section === item.name }"
        >
          <component :is="item.icon" class="size-4" />
          <span class="capitalize">{{ item.name }}</span>
        </RouterLink>
      </Card>

      <Card class="space-y-6">
        <div class="space-y-2">
          <p class="eyebrow">{{ section }}</p>
          <h3 class="text-lg font-semibold text-zinc-50">Account details</h3>
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <label class="field-label">Username</label>
            <Input :model-value="authStore.currentUser?.username" readonly />
          </div>
          <div>
            <label class="field-label">Role</label>
            <Input :model-value="authStore.currentUser?.role" readonly />
          </div>
        </div>
        <div class="rounded-md border border-zinc-800 bg-zinc-900/60 p-4 text-sm leading-6 text-zinc-400">
          SSH key management is available from the dedicated key screen. Profile editing and notification preferences are routed here for the full GitHub-style sitemap.
        </div>
        <Button v-if="section === 'keys'" :as="RouterLink" :to="{ name: 'keys' }" variant="secondary">
          <KeyRound class="size-4" />
          Open SSH keys
        </Button>
      </Card>
    </div>
  </div>
</template>
