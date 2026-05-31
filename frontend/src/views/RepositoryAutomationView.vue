<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { computed, reactive, ref } from 'vue'

import EmptyState from '@/components/empty/EmptyState.vue'
import PageHeader from '@/components/app/PageHeader.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import Select from '@/components/ui/Select.vue'
import { useRepositoryWorkspace } from '@/composables/useRepositoryWorkspace'
import { ApiError, api } from '@/lib/api'
import { formatDate } from '@/lib/utils'

const workspace = useRepositoryWorkspace()
const queryClient = useQueryClient()
const webhookForm = reactive({
  url: '',
  secret: '',
  event: 'repository.push',
})
const errorMessage = ref('')

const webhooksQuery = useQuery({
  queryKey: ['webhooks', workspace.owner, workspace.repo],
  queryFn: () => api.listWebhooks(workspace.owner.value, workspace.repo.value),
  retry: false,
})

const adminEnabled = computed(() => {
  const error = webhooksQuery.error.value
  return !(error instanceof ApiError && error.status === 403)
})
const inventoryError = computed(() => {
  if (!webhooksQuery.error.value || !adminEnabled.value) {
    return ''
  }
  return webhooksQuery.error.value instanceof Error
    ? webhooksQuery.error.value.message
    : 'Unable to load webhook inventory.'
})

const createWebhook = useMutation({
  mutationFn: (payload: { url: string; secret: string; events: string[] }) =>
    api.createWebhook(workspace.owner.value, workspace.repo.value, payload),
  onSuccess: async () => {
    webhookForm.url = ''
    webhookForm.secret = ''
    webhookForm.event = 'repository.push'
    errorMessage.value = ''
    await queryClient.invalidateQueries({ queryKey: ['webhooks', workspace.owner, workspace.repo] })
  },
})

const deleteWebhook = useMutation({
  mutationFn: (webhookId: number) => api.deleteWebhook(workspace.owner.value, workspace.repo.value, webhookId),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: ['webhooks', workspace.owner, workspace.repo] })
  },
})

async function handleCreateWebhook() {
  errorMessage.value = ''
  try {
    await createWebhook.mutateAsync({
      url: webhookForm.url.trim(),
      secret: webhookForm.secret.trim(),
      events: [webhookForm.event],
    })
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Unable to create webhook.'
  }
}

async function handleDeleteWebhook(webhookId: number) {
  errorMessage.value = ''
  try {
    await deleteWebhook.mutateAsync(webhookId)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Unable to delete webhook.'
  }
}
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Automation"
      title="Integration endpoints sit beside source control, not outside it."
      description="Webhooks are isolated into a repository automation module so future CI, agent, or event-driven features can grow here without inflating the code browser."
    >
      <template #actions>
        <Badge variant="accent">repository.push</Badge>
        <Badge>repository.deleted</Badge>
      </template>
    </PageHeader>

    <div class="grid gap-4 xl:grid-cols-[0.9fr_1.1fr]">
      <Card class="space-y-4">
        <div>
          <p class="eyebrow">Create Webhook</p>
          <h3 class="mt-2 text-2xl font-semibold text-zinc-50">Subscribe external systems to repository events.</h3>
        </div>

        <div v-if="adminEnabled" class="space-y-4 rounded-xl border border-zinc-800 bg-black/30 p-4">
          <div>
            <label class="field-label">Endpoint URL</label>
            <Input v-model="webhookForm.url" placeholder="https://example.com/hooks/forge" />
          </div>
          <div>
            <label class="field-label">Signing Secret</label>
            <Input v-model="webhookForm.secret" placeholder="Optional signing secret" />
          </div>
          <div>
            <label class="field-label">Event</label>
            <Select v-model="webhookForm.event">
              <option value="repository.push">repository.push</option>
              <option value="repository.deleted">repository.deleted</option>
            </Select>
          </div>
          <Button :disabled="createWebhook.isPending.value" @click="handleCreateWebhook">
            {{ createWebhook.isPending.value ? 'Creating...' : 'Create Webhook' }}
          </Button>
        </div>
        <div v-else class="rounded-xl border border-zinc-800 bg-black/30 p-4 text-sm text-zinc-400">
          This account can inspect repository automation, but webhook changes require repository admin access.
        </div>

        <div
          v-if="errorMessage"
          class="rounded-md border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300"
        >
          {{ errorMessage }}
        </div>
      </Card>

      <Card class="space-y-4">
        <div class="flex items-center justify-between gap-3">
          <div>
            <p class="eyebrow">Webhook Inventory</p>
            <h3 class="mt-2 text-2xl font-semibold text-zinc-50">Observe delivery health and destination count.</h3>
          </div>
          <Badge variant="accent">{{ webhooksQuery.data.value?.length ?? 0 }}</Badge>
        </div>

        <div v-if="webhooksQuery.isLoading.value" class="space-y-3">
          <div v-for="index in 3" :key="index" class="h-24 animate-pulse rounded-xl bg-zinc-900" />
        </div>
        <div
          v-else-if="inventoryError"
          class="rounded-md border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300"
        >
          {{ inventoryError }}
        </div>
        <div v-else-if="adminEnabled && webhooksQuery.data.value?.length" class="space-y-3">
          <div
            v-for="webhook in webhooksQuery.data.value"
            :key="webhook.id"
            class="rounded-xl border border-zinc-800 bg-black/30 p-4"
          >
            <div class="flex flex-wrap items-center gap-2">
              <Badge variant="accent">{{ webhook.events.join(', ') }}</Badge>
              <Badge>{{ webhook.url }}</Badge>
            </div>
            <p class="mt-3 text-xs text-zinc-500">
              Successes {{ webhook.success_count }} · Failures {{ webhook.failure_count }} · Last delivery
              {{ formatDate(webhook.last_delivery_at) }}
            </p>
            <p class="mt-1 text-xs text-zinc-500">
              {{ webhook.last_delivery_error || 'Last recorded delivery completed without an error message.' }}
            </p>
            <Button
              class="mt-4"
              size="sm"
              variant="ghost"
              :disabled="deleteWebhook.isPending.value"
              @click="handleDeleteWebhook(webhook.id)"
            >
              Delete
            </Button>
          </div>
        </div>
        <EmptyState
          v-else-if="adminEnabled"
          eyebrow="No Webhooks"
          title="No repository webhooks are configured."
          description="Create signed push or delete hooks for CI, deployments, or downstream automation."
        />
        <div v-else class="rounded-xl border border-zinc-800 bg-black/30 p-4 text-sm text-zinc-400">
          Repository automation is enabled, but this account cannot list or modify admin-only webhook settings.
        </div>
      </Card>
    </div>
  </div>
</template>
