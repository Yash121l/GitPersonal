<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { KeyRound } from '@lucide/vue'
import { reactive, ref } from 'vue'

import EmptyState from '@/components/empty/EmptyState.vue'
import PageHeader from '@/components/app/PageHeader.vue'
import CardSkeletonGrid from '@/components/state/CardSkeletonGrid.vue'
import ViewState from '@/components/state/ViewState.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import Textarea from '@/components/ui/Textarea.vue'
import { api } from '@/lib/api'
import { formatDate } from '@/lib/utils'

const queryClient = useQueryClient()
const keysQuery = useQuery({
  queryKey: ['keys'],
  queryFn: () => api.listKeys(),
})

const keyForm = reactive({
  name: '',
  public_key: '',
})
const errorMessage = ref('')

const addKey = useMutation({
  mutationFn: api.createKey,
  onSuccess: async () => {
    keyForm.name = ''
    keyForm.public_key = ''
    errorMessage.value = ''
    await queryClient.invalidateQueries({ queryKey: ['keys'] })
  },
})

async function handleAddKey() {
  errorMessage.value = ''
  try {
    await addKey.mutateAsync({
      name: keyForm.name.trim(),
      public_key: keyForm.public_key.trim(),
    })
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Unable to save SSH key.'
  }
}
</script>

<template>
  <div class="section-stack">
    <PageHeader
      eyebrow="SSH Keys"
      title="SSH Keys"
      description="Register SSH public keys for clone and push access."
    />

    <div class="grid gap-3 xl:grid-cols-[0.92fr_1.08fr]">
      <Card class="space-y-4">
        <div class="panel-header">
          <div>
            <p class="eyebrow">Register Key</p>
            <h3 class="mt-1 text-lg font-semibold text-zinc-50">Add public key</h3>
          </div>
        </div>
        <div>
          <label class="field-label">Label</label>
          <Input v-model="keyForm.name" placeholder="workstation" />
        </div>
        <div>
          <label class="field-label">Public key</label>
          <Textarea
            v-model="keyForm.public_key"
            class="min-h-40 font-mono text-xs"
            placeholder="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA..."
          />
        </div>
        <Button :disabled="addKey.isPending.value" @click="handleAddKey">
          <KeyRound class="size-4" />
          {{ addKey.isPending.value ? 'Saving...' : 'Save SSH Key' }}
        </Button>

        <div
          v-if="errorMessage"
          class="rounded-md border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300"
        >
          {{ errorMessage }}
        </div>
      </Card>

      <div class="space-y-4">
        <ViewState
          :loading="keysQuery.isLoading.value"
          :empty="!keysQuery.isLoading.value && (keysQuery.data.value?.length ?? 0) === 0"
          empty-eyebrow="No Keys"
          empty-title="This account has no SSH keys yet."
          empty-description="Add a public key on the left to start cloning and pushing over SSH."
        >
          <template #loading>
            <Card class="space-y-4">
              <div class="panel-header">
                <div class="space-y-2">
                  <div class="h-3 w-32 animate-pulse rounded-lg bg-zinc-900" />
                  <div class="h-7 w-28 animate-pulse rounded-lg bg-zinc-900" />
                </div>
              </div>
              <CardSkeletonGrid :count="3" wrapper-class="grid gap-3" item-class="h-36" />
            </Card>
          </template>

          <template #empty>
            <EmptyState
              eyebrow="No Keys"
              title="This account has no SSH keys yet."
              description="Add a public key on the left to start cloning and pushing over SSH."
            />
          </template>

          <Card class="space-y-4">
            <div class="panel-header">
              <div>
                <p class="eyebrow">Registered Keys</p>
                <h3 class="mt-1 text-lg font-semibold text-zinc-50">
                  {{ keysQuery.data.value?.length ?? 0 }} key{{ (keysQuery.data.value?.length ?? 0) === 1 ? '' : 's' }}
                </h3>
              </div>
            </div>
            <div class="space-y-3">
              <div
                v-for="key in keysQuery.data.value ?? []"
                :key="key.id"
                class="rounded-lg border border-zinc-800 bg-black/30 p-4"
              >
                <div class="flex flex-wrap items-center gap-2">
                  <Badge variant="accent">{{ key.name }}</Badge>
                  <Badge>{{ key.fingerprint_sha256 }}</Badge>
                </div>
                <p class="mt-3 text-xs text-zinc-500">
                  Created {{ formatDate(key.created_at) }} · Last used {{ formatDate(key.last_used_at) }}
                </p>
                <pre class="mt-3 overflow-x-auto rounded-lg border border-zinc-800 bg-black/60 px-4 py-3 text-xs text-zinc-200">{{ key.public_key }}</pre>
              </div>
            </div>
          </Card>
        </ViewState>
      </div>
    </div>
  </div>
</template>
