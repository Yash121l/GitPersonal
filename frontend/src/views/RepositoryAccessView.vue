<script setup lang="ts">
import { useMutation } from '@tanstack/vue-query'
import { reactive, ref } from 'vue'

import PageHeader from '@/components/app/PageHeader.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import Select from '@/components/ui/Select.vue'
import { useRepositoryWorkspace } from '@/composables/useRepositoryWorkspace'
import { api } from '@/lib/api'

const workspace = useRepositoryWorkspace()

const collaboratorForm = reactive({
  username: '',
  role: 'read',
})
const errorMessage = ref('')
const successMessage = ref('')

const addCollaborator = useMutation({
  mutationFn: (payload: { username: string; role: string }) =>
    api.addCollaborator(workspace.owner.value, workspace.repo.value, payload),
  onSuccess: () => {
    collaboratorForm.username = ''
    collaboratorForm.role = 'read'
    errorMessage.value = ''
    successMessage.value = 'Collaborator invitation saved.'
  },
})

async function handleAddCollaborator() {
  errorMessage.value = ''
  successMessage.value = ''

  try {
    await addCollaborator.mutateAsync({
      username: collaboratorForm.username.trim(),
      role: collaboratorForm.role,
    })
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Unable to add collaborator.'
  }
}
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Access"
      title="Transport endpoints and permissions stay in one place."
      description="Repository access is separated from browsing so clone URLs, visibility, and collaborator changes stay discoverable as the product grows."
    >
      <template #actions>
        <Badge>{{ workspace.repositoryQuery.data.value?.repository.visibility }}</Badge>
      </template>
    </PageHeader>

    <div class="grid gap-4 xl:grid-cols-[1.05fr_0.95fr]">
      <Card class="space-y-4">
        <div>
          <p class="eyebrow">Clone Endpoints</p>
          <h3 class="mt-2 text-2xl font-semibold text-zinc-50">Use the same repository over HTTP and SSH.</h3>
        </div>
        <div class="grid gap-3">
          <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
            <p class="eyebrow">Smart HTTP</p>
            <p class="mt-2 break-all font-mono text-xs text-zinc-200">
              {{ workspace.repositoryQuery.data.value?.http_clone_url }}
            </p>
          </div>
          <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
            <p class="eyebrow">SSH</p>
            <p class="mt-2 break-all font-mono text-xs text-zinc-200">
              {{ workspace.repositoryQuery.data.value?.ssh_clone_url || 'SSH disabled' }}
            </p>
          </div>
          <div class="grid gap-3 md:grid-cols-3">
            <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
              <p class="eyebrow">Visibility</p>
              <p class="mt-2 text-base font-semibold text-zinc-100">
                {{ workspace.repositoryQuery.data.value?.repository.visibility }}
              </p>
            </div>
            <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
              <p class="eyebrow">Owner Type</p>
              <p class="mt-2 text-base font-semibold text-zinc-100">
                {{ workspace.repositoryQuery.data.value?.repository.owner_type }}
              </p>
            </div>
            <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
              <p class="eyebrow">Archived</p>
              <p class="mt-2 text-base font-semibold text-zinc-100">
                {{ workspace.repositoryQuery.data.value?.repository.archived ? 'Yes' : 'No' }}
              </p>
            </div>
          </div>
        </div>
      </Card>

      <div class="space-y-4">
        <Card class="space-y-4">
          <div>
            <p class="eyebrow">Collaborators</p>
            <h3 class="mt-2 text-2xl font-semibold text-zinc-50">Extend access without leaving the workspace.</h3>
          </div>
          <div class="space-y-4 rounded-xl border border-zinc-800 bg-black/30 p-4">
            <div>
              <label class="field-label">Username</label>
              <Input v-model="collaboratorForm.username" placeholder="teammate" />
            </div>
            <div>
              <label class="field-label">Role</label>
              <Select v-model="collaboratorForm.role">
                <option value="read">Read</option>
                <option value="write">Write</option>
                <option value="admin">Admin</option>
              </Select>
            </div>
            <Button :disabled="addCollaborator.isPending.value" @click="handleAddCollaborator">
              {{ addCollaborator.isPending.value ? 'Saving...' : 'Add Collaborator' }}
            </Button>
          </div>

          <div
            v-if="errorMessage"
            class="rounded-md border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300"
          >
            {{ errorMessage }}
          </div>
          <div
            v-if="successMessage"
            class="rounded-md border border-emerald-500/30 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-300"
          >
            {{ successMessage }}
          </div>
        </Card>

        <Card class="space-y-4">
          <div>
            <p class="eyebrow">Access Model</p>
            <h3 class="mt-2 text-xl font-semibold text-zinc-50">Keep the permission surface understandable.</h3>
          </div>
          <div class="grid gap-3 text-sm text-zinc-400">
            <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
              Browser access, Smart HTTP, and SSH all reuse the same repository ownership and collaborator rules.
            </div>
            <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
              Workspace-level navigation keeps organizations and keys separate from repository-local permissions.
            </div>
            <div class="rounded-xl border border-zinc-800 bg-black/30 p-4">
              Future settings can be introduced as new repository tabs without changing the surrounding shell.
            </div>
          </div>
        </Card>
      </div>
    </div>
  </div>
</template>
